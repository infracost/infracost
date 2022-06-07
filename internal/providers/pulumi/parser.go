package pulumi

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/providers/pulumi/aws"
	"github.com/infracost/infracost/internal/providers/pulumi/types"
	"github.com/infracost/infracost/internal/schema"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

// These show differently in the plan JSON for Terraform 0.12 and 0.13.
var infracostProviderNames = []string{"infracost", "registry.terraform.io/infracost/infracost"}

type Parser struct {
	ctx *config.ProjectContext
}

func NewParser(ctx *config.ProjectContext) *Parser {
	return &Parser{ctx}
}

func (p *Parser) createResource(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	registryMap := GetResourceRegistryMap()

	if registryItem, ok := (*registryMap)[d.Type]; ok {
		if registryItem.NoPrice {
			return &schema.Resource{
				Name:         d.Address,
				ResourceType: d.Type,
				Tags:         d.Tags,
				IsSkipped:    true,
				NoPrice:      true,
				SkipMessage:  "Free resource.",
			}
		}

		res := registryItem.RFunc(d, u)
		if res != nil {
			res.ResourceType = d.Type
			// TODO: Figure out how to set tags.  For now, have the RFunc set them.
			// res.Tags = d.Tags
			if u != nil {
				res.EstimationSummary = u.CalcEstimationSummary()
			}
			return res
		}
	}

	return &schema.Resource{
		Name:         d.Address,
		ResourceType: d.Type,
		Tags:         d.Tags,
		IsSkipped:    true,
		SkipMessage:  "This resource is not currently supported",
	}
}

func (p *Parser) parsePreviewDigest(t types.PreviewDigest, usage map[string]*schema.UsageData, rawValues gjson.Result) ([]*schema.Resource, []*schema.Resource, error) {
	baseResources := p.loadUsageFileResources(usage)

	var resources []*schema.Resource
	var pastResources []*schema.Resource
	resources = append(resources, baseResources...)
	refResources := make(map[string]*schema.ResourceData)

	for i := range t.Steps {
		var step = t.Steps[i]
		if step.NewState.Type == "pulumi:pulumi:Stack" {
			continue
		}
		var name = step.NewState.URN.Name().String()
		var resourceType = deriveTfResourceTypes(step.NewState.Type.String())
		log.Debugf("resource type: %s", resourceType)
		var providerName = strings.Split(step.NewState.Type.String(), ":")[0]
		// this section creates a gjson raw value for infracost to search thru.
		var localInputs = step.NewState.Inputs
		localInputs["urn"] = step.URN
		localInputs["config"] = t.Config
		localInputs["dependencies"] = step.NewState.Dependencies
		localInputs["propertyDependencies"] = step.NewState.PropertyDependencies
		localInputs["region"] = parseRegion(resourceType, t.Config)
		var inputs, _ = json.Marshal(localInputs)
		var rawValues = gjson.Parse(string(inputs))
		tags := parseTags(resourceType, rawValues)
		var usageData *schema.UsageData

		if ud := usage[name]; ud != nil {
			usageData = ud
		}

		resourceData := schema.NewResourceData(resourceType, providerName, name, tags, rawValues)
		refResources[name] = resourceData
		if r := p.createResource(resourceData, usageData); r != nil {
			if step.Op == "same" {
				pastResources = append(resources, r)
			} else if step.Op == "create" {
				resources = append(resources, r)
			}
		}
	}
	p.parseReferences(refResources, rawValues)
	p.loadInfracostProviderUsageData(usage, refResources)
	return pastResources, resources, nil
}

func (p *Parser) loadUsageFileResources(u map[string]*schema.UsageData) []*schema.Resource {
	resources := make([]*schema.Resource, 0)

	for k, v := range u {
		for _, t := range GetUsageOnlyResources() {
			if strings.HasPrefix(k, fmt.Sprintf("%s.", t)) {
				d := schema.NewResourceData(t, "global", k, map[string]string{}, gjson.Result{})
				if r := p.createResource(d, v); r != nil {
					resources = append(resources, r)
				}
			}
		}
	}

	return resources
}

func isAwsChina(d *schema.ResourceData) bool {
	return strings.HasPrefix(d.Type, "aws_") && strings.HasPrefix(d.Get("region").String(), "cn-")
}

func getSpecialContext(d *schema.ResourceData) map[string]interface{} {
	providerPrefix := strings.Split(d.Type, "_")[0]

	switch providerPrefix {
	case "aws":
		return aws.GetSpecialContext(d)
	default:
		log.Debugf("Unsupported provider %s", providerPrefix)
		return map[string]interface{}{}
	}
}

func parseTags(resourceType string, v gjson.Result) map[string]string {

	providerPrefix := strings.Split(resourceType, "_")[0]

	switch providerPrefix {
	case "aws":
		return aws.ParseTags(resourceType, v)
	default:
		log.Debugf("Unsupported provider %s", providerPrefix)
		return map[string]string{}
	}
}

func parseRegion(resourceType string, v map[string]string) string {
	var region string
	providerPrefix := strings.Split(resourceType, "_")[0]

	switch providerPrefix {
	case "aws":
		region = v["aws:region"]
		if region == "" {
			region = aws.DefaultProviderRegion
		}
		return region
	default:
		log.Debugf("Unsupported provider %s", providerPrefix)
		return ""
	}
}

func (p *Parser) loadInfracostProviderUsageData(u map[string]*schema.UsageData, resData map[string]*schema.ResourceData) {
	log.Debugf("Loading usage data from Infracost provider resources")

	for _, d := range resData {
		if isInfracostResource(d) {
			p.ctx.SetContextValue("terraformInfracostProviderEnabled", true)

			for _, ref := range d.References("resources") {
				if _, ok := u[ref.Address]; !ok {
					u[ref.Address] = schema.NewUsageData(ref.Address, convertToUsageAttributes(d.RawValues))
				} else {
					log.Debugf("Skipping loading usage for resource %s since it has already been defined", ref.Address)
				}
			}
		}
	}
}

func (p *Parser) stripDataResources(resData map[string]*schema.ResourceData) {
	for addr, d := range resData {
		if strings.HasPrefix(addressResourcePart(d.Address), "data.") {
			delete(resData, addr)
		}
	}
}

func (p *Parser) parseReferences(resData map[string]*schema.ResourceData, conf gjson.Result) {
	registryMap := GetResourceRegistryMap()

	// Create a map of id -> resource data so we can lookup references
	idMap := make(map[string][]*schema.ResourceData)

	for _, d := range resData {

		// check for any "default" ids declared by the provider for this resource
		if f := registryMap.GetDefaultRefIDFunc(d.Type); f != nil {
			for _, defaultID := range f(d) {
				if _, ok := idMap[defaultID]; !ok {
					idMap[defaultID] = []*schema.ResourceData{}
				}
				idMap[defaultID] = append(idMap[defaultID], d)
			}
		}

		// check for any "custom" ids specified by the resource and add them.
		if f := registryMap.GetCustomRefIDFunc(d.Type); f != nil {
			for _, customID := range f(d) {
				if _, ok := idMap[customID]; !ok {
					idMap[customID] = []*schema.ResourceData{}
				}
				idMap[customID] = append(idMap[customID], d)
			}
		}

	}

	//parseKnownModuleRefs(resData, conf)

	for _, d := range resData {
		var refAttrs []string

		if isInfracostResource(d) {
			refAttrs = []string{"resources"}
		} else {
			refAttrs = registryMap.GetReferenceAttributes(d.Type)
		}

		for _, attr := range refAttrs {
			found := p.parseConfReferences(resData, conf, d, attr, registryMap)

			if found {
				continue
			}

			// Get any values for the fields and check if they map to IDs or ARNs of any resources
			for _, refVal := range d.Get(attr).Array() {
				if refVal.String() == "" {
					continue
				}

				// Check ID map
				idRefs, ok := idMap[refVal.String()]
				if ok {

					for _, ref := range idRefs {
						reverseRefAttrs := registryMap.GetReferenceAttributes(ref.Type)
						d.AddReference(attr, ref, reverseRefAttrs)
					}
				}
			}
		}
	}
}

func (p *Parser) parseConfReferences(resData map[string]*schema.ResourceData, conf gjson.Result, d *schema.ResourceData, attr string, registryMap *ResourceRegistryMap) bool {
	// Check if there's a reference in the conf
	resConf := getConfJSON(conf, d.Address)
	log.Debugf("resConf %s %s", d.Address, attr)
	exps := resConf.Get(attr)
	lookupStr := "dependencies"
	if exps.IsArray() {
		lookupStr = "#.dependencies"
	}

	refResults := exps.Get(lookupStr).Array()
	log.Debugf("Dependencies %s", refResults)
	refs := make([]string, 0, len(refResults))

	for _, refR := range refResults {
		if refR.Type == gjson.JSON {
			arr := refR.Array()
			for _, r := range arr {
				refs = append(refs, r.String())
			}
			continue
		}

		refs = append(refs, refR.String())
	}

	found := false

	for _, ref := range refs {
		if ref == "count.index" || ref == "each.key" || strings.HasPrefix(ref, "var.") {
			continue
		}

		var refData *schema.ResourceData

		m := addressModulePart(d.Address)
		refAddr := fmt.Sprintf("%s%s", m, ref)

		// see if there's a resource that's an exact match on the address
		refData, ok := resData[refAddr]

		// if there's a count ref value then try with the array index of the count ref
		if !ok {
			if containsString(refs, "count.index") {
				a := fmt.Sprintf("%s[%d]", refAddr, addressCountIndex(d.Address))
				refData, ok = resData[a]

				if ok {
					log.Debugf("reference specifies a count: using resource %s for %s.%s", a, d.Address, attr)
				}
			} else if containsString(refs, "each.key") {
				a := fmt.Sprintf("%s[\"%s\"]", refAddr, addressKey(d.Address))
				refData, ok = resData[a]

				if ok {
					log.Debugf("reference specifies a key: using resource %s for %s.%s", a, d.Address, attr)
				}
			}
		}

		// if still not found, see if there's a matching resource with an [0] array part
		if !ok {
			a := fmt.Sprintf("%s[0]", refAddr)
			refData, ok = resData[a]

			if ok {
				log.Debugf("reference does not specify a count: using resource %s for for %s.%s", a, d.Address, attr)
			}
		}

		if ok {
			found = true
			reverseRefAttrs := registryMap.GetReferenceAttributes(refData.Type)
			d.AddReference(attr, refData, reverseRefAttrs)
		}
	}

	return found
}

func convertToUsageAttributes(j gjson.Result) map[string]gjson.Result {
	a := make(map[string]gjson.Result)

	for k, v := range j.Map() {
		a[k] = v.Get("0.value")
	}

	return a
}

func getConfJSON(conf gjson.Result, addr string) gjson.Result {
	return conf
}

func getModuleConfJSON(conf gjson.Result, names []string) gjson.Result {
	return conf
}

func isInfracostResource(res *schema.ResourceData) bool {
	for _, p := range infracostProviderNames {
		if res.ProviderName == p {
			return true
		}
	}

	return false
}

// addressResourcePart parses a resource addr and returns resource suffix (without the module prefix).
// For example: `module.name1.module.name2.resource` will return `name2.resource`.
func addressResourcePart(addr string) string {
	p := splitAddress(addr)

	if len(p) >= 3 && p[len(p)-3] == "data" {
		return strings.Join(p[len(p)-3:], ":")
	}

	return strings.Join(p[len(p)-2:], ":")
}

// addressModulePart parses a resource addr and returns module prefix.
// For example: `module.name1.module.name2.resource` will return `module.name1.module.name2.`.
func addressModulePart(addr string) string {
	ap := splitAddress(addr)

	var mp []string

	if len(ap) >= 3 && ap[len(ap)-3] == "data" {
		mp = ap[:len(ap)-3]
	} else {
		mp = ap[:len(ap)-2]
	}

	if len(mp) == 0 {
		return ""
	}

	return fmt.Sprintf("%s.", strings.Join(mp, "."))
}

func addressCountIndex(addr string) int {
	r := regexp.MustCompile(`\[(\d+)\]`)
	m := r.FindStringSubmatch(addr)

	if len(m) > 0 {
		i, _ := strconv.Atoi(m[1]) // TODO: unhandled error

		return i
	}

	return -1
}

func addressKey(addr string) string {
	r := regexp.MustCompile(`\["([^"]+)"\]`)
	m := r.FindStringSubmatch(addr)

	if len(m) > 0 {
		return m[1]
	}

	return ""
}

func removeAddressArrayPart(addr string) string {
	r := regexp.MustCompile(`([^\[]+)`)
	m := r.FindStringSubmatch(addressResourcePart(addr))

	return m[1]
}

// splitAddress splits the address by `.`, but ignores any `.`s quoted in the array part of the address
func splitAddress(addr string) []string {
	quoted := false
	return strings.FieldsFunc(addr, func(r rune) bool {
		if r == '"' {
			quoted = !quoted
		}
		return !quoted && r == '.'
	})
}

func containsString(a []string, s string) bool {
	for _, i := range a {
		if i == s {
			return true
		}
	}

	return false
}

func gjsonEscape(s string) string {
	s = strings.ReplaceAll(s, ".", `\.`)
	s = strings.ReplaceAll(s, "*", `\*`)
	s = strings.ReplaceAll(s, "?", `\?`)

	return s
}

// Parses known modules to create references for specific resources in that module
// This is useful if the module uses a `dynamic` block which means the references aren't defined in the plan JSON
// See https://github.com/hashicorp/terraform/issues/28346 for more info
func parseKnownModuleRefs(resData map[string]*schema.ResourceData, conf gjson.Result) {
	knownRefs := []struct {
		SourceAddrSuffix string
		DestAddrSuffix   string
		Attribute        string
		ModuleSource     string
	}{
		{
			SourceAddrSuffix: "aws_autoscaling_group.workers_launch_template",
			DestAddrSuffix:   "aws_launch_template.workers_launch_template",
			Attribute:        "launch_template",
			ModuleSource:     "terraform-aws-modules/eks/aws",
		},
		{
			SourceAddrSuffix: "aws_autoscaling_group.this",
			DestAddrSuffix:   "aws_launch_template.this",
			Attribute:        "launch_template",
			ModuleSource:     "terraform-aws-modules/autoscaling/aws",
		},
		{
			SourceAddrSuffix: "aws_autoscaling_group.this",
			DestAddrSuffix:   "aws_launch_configuration.this",
			Attribute:        "launch_configuration",
			ModuleSource:     "terraform-aws-modules/autoscaling/aws",
		},
	}
	log.Debugf("known refs %s", knownRefs)
}

// This function takes the string type from pulumi and converts it to the terraform type
// a good bit of the infracost internals are dependent on the tf naming, it was easier to convert.
func deriveTfResourceTypes(resourceType string) string {
	var resourceTypeArray = strings.Split(resourceType, ":")
	var midTypeArray = strings.Split(resourceTypeArray[1], "/")
	var _resourceType = resourceTypeArray[0] + "_" + midTypeArray[0] + "_" + midTypeArray[1]
	tfResourceTypes := map[string]string{
		"aws_ec2_eip": "aws_eip",
	}
	knownResourceType := tfResourceTypes[_resourceType]
	if knownResourceType == "" {
		return _resourceType
	} else {
		return knownResourceType
	}
}
