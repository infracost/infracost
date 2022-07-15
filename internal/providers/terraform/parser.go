package terraform

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/providers/terraform/aws"
	"github.com/infracost/infracost/internal/providers/terraform/azure"
	"github.com/infracost/infracost/internal/providers/terraform/google"
	"github.com/infracost/infracost/internal/schema"
)

// These show differently in the plan JSON for Terraform 0.12 and 0.13.
var infracostProviderNames = []string{"infracost", "registry.terraform.io/infracost/infracost"}

type Parser struct {
	ctx                  *config.ProjectContext
	terraformVersion     string
	includePastResources bool
}

func NewParser(ctx *config.ProjectContext, includePastResources bool) *Parser {
	return &Parser{
		ctx:                  ctx,
		includePastResources: includePastResources,
	}
}

func (p *Parser) createResource(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	registryMap := GetResourceRegistryMap()

	for cKey, cValue := range getSpecialContext(d) {
		p.ctx.SetContextValue(cKey, cValue)
	}

	if registryItem, ok := (*registryMap)[d.Type]; ok {
		if registryItem.NoPrice {
			return &schema.Resource{
				Name:         d.Address,
				ResourceType: d.Type,
				Tags:         d.Tags,
				IsSkipped:    true,
				NoPrice:      true,
				SkipMessage:  "Free resource.",
				Metadata:     d.Metadata,
			}
		}

		res := registryItem.RFunc(d, u)
		if res != nil {
			res.ResourceType = d.Type
			res.Tags = d.Tags
			res.Metadata = d.Metadata

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
		Metadata:     d.Metadata,
	}
}

func (p *Parser) parseJSONResources(parsePrior bool, baseResources []*schema.Resource, usage map[string]*schema.UsageData, parsed, providerConf, conf, vars gjson.Result) []*schema.Resource {
	var resources []*schema.Resource
	resources = append(resources, baseResources...)
	var vals gjson.Result

	isState := false
	if parsePrior {
		isState = true
		vals = parsed.Get("prior_state.values.root_module")
	} else {
		vals = parsed.Get("planned_values.root_module")
		if !vals.Exists() {
			isState = true
			vals = parsed.Get("values.root_module")
		}
	}

	resData := p.parseResourceData(isState, providerConf, vals, conf, vars)

	p.parseReferences(resData, conf)
	p.loadInfracostProviderUsageData(usage, resData)
	p.stripDataResources(resData)
	p.populateUsageData(resData, usage)

	for _, d := range resData {
		if r := p.createResource(d, d.UsageData); r != nil {
			resources = append(resources, r)
		}
	}

	return resources
}

// populateUsageData finds the UsageData for each ResourceData and sets the ResourceData.UsageData field
// in case it is needed when processing a reference attribute
func (p *Parser) populateUsageData(resData map[string]*schema.ResourceData, usage map[string]*schema.UsageData) {
	for _, d := range resData {
		if ud := usage[d.Address]; ud != nil {
			d.UsageData = ud
		} else if strings.HasSuffix(d.Address, "]") {
			lastIndexOfOpenBracket := strings.LastIndex(d.Address, "[")

			if arrayUsageData := usage[fmt.Sprintf("%s[*]", d.Address[:lastIndexOfOpenBracket])]; arrayUsageData != nil {
				d.UsageData = arrayUsageData
			}
		}
	}
}

func (p *Parser) parseJSON(j []byte, usage map[string]*schema.UsageData) ([]*schema.Resource, []*schema.Resource, error) {
	baseResources := p.loadUsageFileResources(usage)

	j, _ = StripSetupTerraformWrapper(j)

	if !gjson.ValidBytes(j) {
		return baseResources, baseResources, errors.New("invalid JSON")
	}

	parsed := gjson.ParseBytes(j)

	p.terraformVersion = parsed.Get("terraform_version").String()
	providerConf := parsed.Get("configuration.provider_config")
	conf := parsed.Get("configuration.root_module")
	vars := parsed.Get("variables")

	resources := p.parseJSONResources(false, baseResources, usage, parsed, providerConf, conf, vars)
	if !p.includePastResources {
		return nil, resources, nil
	}

	pastResources := p.parseJSONResources(true, baseResources, usage, parsed, providerConf, conf, vars)
	resourceChanges := parsed.Get("resource_changes").Array()
	pastResources = stripNonTargetResources(pastResources, resources, resourceChanges)

	return pastResources, resources, nil
}

// StripSetupTerraformWrapper removes any output added from the setup-terraform
// GitHub action terraform wrapper, so we can parse the output of this as
// valid JSON. It returns the stripped out JSON and a boolean that is true
// if the wrapper output was found and removed.
func StripSetupTerraformWrapper(b []byte) ([]byte, bool) {
	headerLine := regexp.MustCompile(`(?m)^\[command\].*\n`)
	outputLine := regexp.MustCompile(`(?m)^::.*\n`)

	stripped := headerLine.ReplaceAll(b, []byte{})
	stripped = outputLine.ReplaceAll(stripped, []byte{})

	return stripped, len(stripped) != len(b)
}

func (p *Parser) loadUsageFileResources(u map[string]*schema.UsageData) []*schema.Resource {
	resources := make([]*schema.Resource, 0)

	for k, v := range u {
		for _, t := range GetUsageOnlyResources() {
			if strings.HasPrefix(k, fmt.Sprintf("%s.", t)) {
				d := schema.NewResourceData(t, "global", k, map[string]string{}, gjson.Result{})
				// set the usage data as a field on the resource data in case it is needed when
				// processing reference attributes.
				d.UsageData = v
				if r := p.createResource(d, v); r != nil {
					resources = append(resources, r)
				}
			}
		}
	}

	return resources
}

// stripNonTargetResources removes any past resources that don't exist in the
// current resources or resource_changes in the Terraform plan. When Terraform
// is run with `-target` then all resources still appear in prior_state but not
// in planned_values. This makes sure we remove any non-target resources from
// the past resources so that we only show resources matching the target.
func stripNonTargetResources(pastResources []*schema.Resource, resources []*schema.Resource, resourceChanges []gjson.Result) []*schema.Resource {
	resourceAddrMap := make(map[string]bool, len(resources))
	for _, resource := range resources {
		resourceAddrMap[resource.Name] = true
	}

	diffAddrMap := make(map[string]bool, len(resourceChanges))
	for _, change := range resourceChanges {
		diffAddrMap[change.Get("address").String()] = true
	}

	var filteredResources []*schema.Resource
	for _, resource := range pastResources {
		_, rOk := resourceAddrMap[resource.Name]
		_, dOk := diffAddrMap[resource.Name]
		if dOk || rOk {
			filteredResources = append(filteredResources, resource)
		}
	}
	return filteredResources
}

func (p *Parser) parseResourceData(isState bool, providerConf, planVals gjson.Result, conf gjson.Result, vars gjson.Result) map[string]*schema.ResourceData {
	resources := make(map[string]*schema.ResourceData)

	for _, r := range planVals.Get("resources").Array() {
		t := r.Get("type").String()
		provider := r.Get("provider_name").String()
		addr := r.Get("address").String()

		// Terraform v0.12 files have a different format for the addresses of provisioned resources
		// So we need to build the full address from the module and index
		if strings.HasPrefix(p.terraformVersion, "0.12.") && isState {
			modAddr := planVals.Get("address").String()
			if modAddr != "" && !strings.HasPrefix(addr, modAddr) {
				addr = fmt.Sprintf("%s.%s", modAddr, addr)
			}
			if r.Get("index").Type != gjson.Null {
				indexSuffix := fmt.Sprintf("[%s]", r.Get("index").Raw)
				// Check that the suffix doesn't already exist on the address. This can happen if Terraform v0.12 was
				// used to generate the state but then a different version is used to show it.
				if !strings.HasSuffix(addr, indexSuffix) {
					addr = fmt.Sprintf("%s%s", addr, indexSuffix)
				}
			}
		}

		v := r.Get("values")

		resConf := getConfJSON(conf, addr)

		// Try getting the region from the ARN
		region := resourceRegion(t, v)

		// Otherwise use region from the provider conf
		if region == "" {
			region = providerRegion(addr, providerConf, vars, t, resConf)
		}

		v = schema.AddRawValue(v, "region", region)

		tags := parseTags(t, v)

		data := schema.NewResourceData(t, provider, addr, tags, v)
		data.Metadata = r.Get("infracost_metadata").Map()
		resources[addr] = data
	}

	// Recursively add any resources for child modules
	for _, m := range planVals.Get("child_modules").Array() {
		for addr, d := range p.parseResourceData(isState, providerConf, m, conf, vars) {
			resources[addr] = d
		}
	}

	return resources
}

func getSpecialContext(d *schema.ResourceData) map[string]interface{} {
	providerPrefix := strings.Split(d.Type, "_")[0]

	switch providerPrefix {
	case "aws":
		return aws.GetSpecialContext(d)
	case "azurerm":
		return azure.GetSpecialContext(d)
	case "google":
		return google.GetSpecialContext(d)
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
	case "azurerm":
		return azure.ParseTags(resourceType, v)
	case "google":
		return google.ParseTags(resourceType, v)
	default:
		log.Debugf("Unsupported provider %s", providerPrefix)
		return map[string]string{}
	}
}

func resourceRegion(resourceType string, v gjson.Result) string {

	providerPrefix := strings.Split(resourceType, "_")[0]

	switch providerPrefix {
	case "aws":
		return aws.GetResourceRegion(resourceType, v)
	case "azurerm":
		return azure.GetResourceRegion(resourceType, v)
	case "google":
		return google.GetResourceRegion(resourceType, v)
	default:
		log.Debugf("Unsupported provider %s", providerPrefix)
		return ""
	}
}

func providerRegion(addr string, providerConf gjson.Result, vars gjson.Result, resourceType string, resConf gjson.Result) string {
	var region string

	providerKey := parseProviderKey(resConf)
	if providerKey != "" {
		region = parseRegion(providerConf, vars, providerKey)
		// Note: if the provider is passed to a module using a different alias
		// then there's no way to detect this so we just have to fallback to
		// the default provider
	}

	if region == "" {
		// Try to get the provider key from the first part of the resource
		providerPrefix := strings.Split(resourceType, "_")[0]
		region = parseRegion(providerConf, vars, providerPrefix)

		if region == "" {

			switch providerPrefix {
			case "aws":
				region = aws.DefaultProviderRegion
			case "azurerm":
				region = azure.DefaultProviderRegion
			case "google":
				region = google.DefaultProviderRegion
			}

			// Don't show this log for azurerm users since they have a different method of looking up the region.
			// A lot of Azure resources get their region from their referenced azurerm_resource_group resource
			if region != "" && providerPrefix != "azurerm" {
				log.Debugf("Falling back to default region (%s) for %s", region, addr)
			}
		}
	}

	return region
}

func parseProviderKey(resConf gjson.Result) string {
	v := resConf.Get("provider_config_key").String()
	p := strings.Split(v, ":")

	return p[len(p)-1]
}

func parseRegion(providerConf gjson.Result, vars gjson.Result, providerKey string) string {
	// Try to get constant value
	region := providerConf.Get(fmt.Sprintf("%s.expressions.region.constant_value", gjsonEscape(providerKey))).String()
	if region == "" {
		// Try to get reference
		refName := providerConf.Get(fmt.Sprintf("%s.expressions.region.references.0", gjsonEscape(providerKey))).String()
		splitRef := strings.Split(refName, ".")

		if splitRef[0] == "var" {
			// Get the region from variables
			varName := strings.Join(splitRef[1:], ".")
			varContent := vars.Get(fmt.Sprintf("%s.value", varName))

			if !varContent.IsObject() && !varContent.IsArray() {
				region = varContent.String()
			}
		}
	}

	if strings.Contains(region, "mock") {
		return ""
	}

	return region
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

	parseKnownModuleRefs(resData, conf)

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
	exps := resConf.Get("expressions").Get(attr)
	lookupStr := "references"
	if exps.IsArray() {
		lookupStr = "#.references"
	}

	refResults := exps.Get(lookupStr).Array()
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
	modNames := getModuleNames(addr)
	c := getModuleConfJSON(conf, modNames)

	if len(modNames) > 0 {
		c = c.Get("module")
	}

	return c.Get(fmt.Sprintf(`resources.#(address="%s")`, removeAddressArrayPart(addressResourcePart(addr))))
}

func getModuleConfJSON(conf gjson.Result, names []string) gjson.Result {
	if len(names) == 0 {
		return conf
	}

	// Build up the gjson search key
	p := make([]string, 0, len(names))
	for _, n := range names {
		p = append(p, fmt.Sprintf("module_calls.%s", n))
	}

	return conf.Get(strings.Join(p, ".module."))
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
		return strings.Join(p[len(p)-3:], ".")
	}

	if len(p) >= 2 {
		return strings.Join(p[len(p)-2:], ".")
	}

	return ""
}

// addressModulePart parses a resource addr and returns module prefix.
// For example: `module.name1.module.name2.resource` will return `module.name1.module.name2.`.
func addressModulePart(addr string) string {
	ap := splitAddress(addr)

	var mp []string

	if len(ap) >= 3 && ap[len(ap)-3] == "data" {
		mp = ap[:len(ap)-3]
	} else if len(ap) >= 2 {
		mp = ap[:len(ap)-2]
	}

	if len(mp) == 0 {
		return ""
	}

	return fmt.Sprintf("%s.", strings.Join(mp, "."))
}

func getModuleNames(addr string) []string {
	r := regexp.MustCompile(`module\.([^\.\[]*)`)
	matches := r.FindAllStringSubmatch(addressModulePart(addr), -1)

	if matches == nil {
		return []string{}
	}

	n := make([]string, 0, len(matches))
	for _, m := range matches {
		n = append(n, m[1])
	}

	return n
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

	if len(m) == 0 {
		return ""
	}

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

	for _, d := range resData {
		for _, knownRef := range knownRefs {
			modNames := getModuleNames(d.Address)
			modSource := getModuleConfJSON(conf, modNames).Get("source").String()
			matches := strings.HasSuffix(removeAddressArrayPart(d.Address), knownRef.SourceAddrSuffix) && modSource == knownRef.ModuleSource

			if matches {
				countIndex := addressCountIndex(d.Address)

				for _, destD := range resData {
					suffix := fmt.Sprintf("%s[%d]", knownRef.DestAddrSuffix, countIndex)
					if cmp.Equal(getModuleNames(destD.Address), modNames) && strings.HasSuffix(destD.Address, suffix) {
						d.AddReference(knownRef.Attribute, destD, []string{})
					}
				}
			}
		}
	}
}
