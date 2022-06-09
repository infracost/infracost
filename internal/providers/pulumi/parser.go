package pulumi

import (
	"encoding/json"
	"fmt"
	"regexp"
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
		if resourceType == "awsx" {
			continue
		}
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

		resourceData := schema.NewPulumiResourceData(resourceType, providerName, name, tags, rawValues, string(step.URN))
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

func (p *Parser) parseReferences(resData map[string]*schema.ResourceData, conf gjson.Result) {
	registryMap := GetResourceRegistryMap()

	// Create a map of id -> resource data so we can lookup references
	idMap := make(map[string][]*schema.ResourceData)

	for _, d := range resData {

		// check for any "default" ids declared by the provider for this resource
		if f := registryMap.GetDefaultRefIDFunc(d.Type); f != nil {
			for _, defaultID := range f(d) {
				log.Debugf("defaultId %s", defaultID)
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
				//continue
			}

			// Get any values for the fields and check if they map to IDs or ARNs of any resources

			for _, refVal := range d.Get(fmt.Sprintf(`newState.inputs.%s`, attr)).Array() {
				log.Debugf("attr %s, refVal %s", fmt.Sprintf(`newState.inputs.%s`, attr), refVal)
				if refVal.String() == "" {
					continue
				}

				// Check ID map
				idRefs, ok := idMap[refVal.String()]
				log.Debugf("idRefs %s", idRefs)
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
	//This one gets the resource
	resConf := getConfJSON(conf, d.RawValues.Get("urn").String())
	// We get the matching input values mapped to dependencies ex: newState.inputs.ebsBlockDevices.#.volumeId"
	log.Debugf(fmt.Sprintf(`newState.inputs.%s`, attr))
	exps := resConf.Get(fmt.Sprintf(`newState.inputs.%s`, attr))
	log.Debugf("exps %s", exps)
	// if there are any responses iterate dependencies
	if exps.Type != gjson.Null {
		refResults := resConf.Get("newState.dependencies").Array()
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
		log.Debugf("refs %s", refs)
		found := false

		for _, ref := range refs {

			var refData *schema.ResourceData

			m := d.Address
			refAddr := fmt.Sprintf("%s%s", m, ref)
			// see if there's a resource that's an exact match on the address
			refData, ok := resData[m]
			log.Debugf("resData %s, refaddr %s, ok %s", refData, refAddr, ok)

			if ok {
				found = true
				reverseRefAttrs := registryMap.GetReferenceAttributes(refData.Type)
				d.AddReference(attr, refData, reverseRefAttrs)
			}
		}

		return found
	} else {
		return false
	}
}

func getConfJSON(conf gjson.Result, addr string) gjson.Result {
	c := conf.Get(fmt.Sprintf(`steps.#(newState.urn=="%s")`, addr))
	return c
}

func convertToUsageAttributes(j gjson.Result) map[string]gjson.Result {
	a := make(map[string]gjson.Result)

	for k, v := range j.Map() {
		a[k] = v.Get("0.value")
	}

	return a
}

func isInfracostResource(res *schema.ResourceData) bool {
	for _, p := range infracostProviderNames {
		if res.ProviderName == p {
			return true
		}
	}

	return false
}

func addressKey(addr string) string {
	r := regexp.MustCompile(`\["([^"]+)"\]`)
	m := r.FindStringSubmatch(addr)

	if len(m) > 0 {
		return m[1]
	}

	return ""
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
	providerPrefix := strings.ToLower(resourceTypeArray[0])
	var tfResourceTypes map[string]string
	switch providerPrefix {
	case "aws":
		tfResourceTypes = aws.GetAWSResourceTypes()
	case "awsx":
		return "awsx"
	default:
		log.Debugf("Unsupported provider %s", providerPrefix)
		tfResourceTypes = map[string]string{}
	}
	var midTypeArray = strings.Split(resourceTypeArray[1], "/")
	var _resourceType = strings.ToLower(resourceTypeArray[0] + "_" + midTypeArray[0] + "_" + midTypeArray[1])
	knownResourceType := tfResourceTypes[_resourceType]
	if knownResourceType == "" {
		return _resourceType
	} else {
		return knownResourceType
	}
}
