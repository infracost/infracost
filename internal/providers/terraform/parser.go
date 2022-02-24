package terraform

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/tidwall/gjson"
)

// These show differently in the plan JSON for Terraform 0.12 and 0.13.
var infracostProviderNames = []string{"infracost", "registry.terraform.io/infracost/infracost"}
var defaultProviderRegions = map[string]string{
	"aws":     "us-east-1",
	"google":  "us-central1",
	"azurerm": "eastus",
}

// ARN attribute mapping for resources that don't have a standard 'arn' attribute
var arnAttributeMap = map[string]string{
	"aws_cloudwatch_dashboard":     "dashboard_arn",
	"aws_db_snapshot":              "db_snapshot_arn",
	"aws_db_cluster_snapshot":      "db_cluster_snapshot_arn",
	"aws_ecs_service":              "id",
	"aws_neptune_cluster_snapshot": "db_cluster_snapshot_arn",
	"aws_docdb_cluster_snapshot":   "db_cluster_snapshot_arn",
	"aws_dms_certificate":          "certificate_arn",
	"aws_dms_endpoint":             "endpoint_arn",
	"aws_dms_replication_instance": "replication_instance_arn",
	"aws_dms_replication_task":     "replication_task_arn",
}

type Parser struct {
	ctx              *config.ProjectContext
	terraformVersion string
}

func NewParser(ctx *config.ProjectContext) *Parser {
	return &Parser{ctx, ""}
}

func (p *Parser) createResource(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	registryMap := GetResourceRegistryMap()

	if isAwsChina(d) {
		p.ctx.SetContextValue("isAWSChina", true)
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
			}
		}

		res := registryItem.RFunc(d, u)
		if res != nil {
			res.ResourceType = d.Type
			res.Tags = d.Tags
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

	pastResources := p.parseJSONResources(true, baseResources, usage, parsed, providerConf, conf, vars)
	resources := p.parseJSONResources(false, baseResources, usage, parsed, providerConf, conf, vars)

	resourceChanges := parsed.Get("resource_changes").Array()
	pastResources = stripNonTargetResources(pastResources, resources, resourceChanges)

	return pastResources, resources, nil
}

// StripTerraformWrapper removes any output added from the setup-terraform
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

		resources[addr] = schema.NewResourceData(t, provider, addr, tags, v)
	}

	// Recursively add any resources for child modules
	for _, m := range planVals.Get("child_modules").Array() {
		for addr, d := range p.parseResourceData(isState, providerConf, m, conf, vars) {
			resources[addr] = d
		}
	}

	return resources
}

func parseTags(resourceType string, v gjson.Result) map[string]string {
	tags := make(map[string]string)

	a := "tags"
	if strings.HasPrefix(resourceType, "google_") {
		a = "labels"
	}

	for k, v := range v.Get(a).Map() {
		tags[k] = v.String()
	}

	return tags
}

func resourceRegion(resourceType string, v gjson.Result) string {
	providerPrefix := strings.Split(resourceType, "_")[0]
	if providerPrefix != "aws" {
		return ""
	}

	// If a region key exists in the values use that
	if v.Get("region").String() != "" {
		return v.Get("region").String()
	}

	// Otherwise try and parse the ARN from the values
	arnAttr, ok := arnAttributeMap[resourceType]
	if !ok {
		arnAttr = "arn"
	}

	if !v.Get(arnAttr).Exists() {
		return ""
	}

	arn := v.Get(arnAttr).String()
	p := strings.Split(arn, ":")
	if len(p) < 4 {
		log.Debugf("Unexpected ARN format for %s", arn)
		return ""
	}

	return p[3]
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
			region = defaultProviderRegions[providerPrefix]

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

	// Create a map of id -> resource data and arn -> resource data so we can lookup references
	idMap := make(map[string][]*schema.ResourceData)
	arnMap := make(map[string][]*schema.ResourceData)

	for _, d := range resData {
		id := d.Get("id").String()
		if _, ok := idMap[id]; !ok {
			idMap[id] = []*schema.ResourceData{}
		}

		idMap[id] = append(idMap[id], d)

		// check for any "custom" ids specified by the resource and add them.
		if f := registryMap.GetCustomRefIDFunc(d.Type); f != nil {
			for _, customID := range f(d) {
				if _, ok := idMap[customID]; !ok {
					idMap[customID] = []*schema.ResourceData{}
				}
				idMap[customID] = append(idMap[customID], d)
			}
		}

		arnAttr, ok := arnAttributeMap[d.Type]
		if !ok {
			arnAttr = "arn"
		}

		arn := d.Get(arnAttr).String()
		if _, ok := arnMap[arn]; !ok {
			arnMap[arn] = []*schema.ResourceData{}
		}

		arnMap[arn] = append(arnMap[arn], d)
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

				// Check arn map
				arnRefs, ok := arnMap[refVal.String()]
				if ok {
					for _, ref := range arnRefs {
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
	refResults := resConf.Get("expressions").Get(attr).Get("references").Array()
	refs := make([]string, 0, len(refResults))

	for _, refR := range refResults {
		refs = append(refs, refR.String())
	}

	found := false

	for _, ref := range refs {
		if ref == "count.index" || strings.HasPrefix(ref, "var.") {
			continue
		}

		var refData *schema.ResourceData

		m := addressModulePart(d.Address)
		refAddr := fmt.Sprintf("%s%s", m, ref)

		// see if there's a resource that's an exact match on the address
		refData, ok := resData[refAddr]

		// if there's a count ref value then try with the array index of the count ref
		if !ok && containsString(refs, "count.index") {
			a := fmt.Sprintf("%s[%d]", refAddr, addressCountIndex(d.Address))
			refData, ok = resData[a]

			if ok {
				log.Debugf("reference specifies a count: using resource %s for %s.%s", a, d.Address, attr)
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

	return strings.Join(p[len(p)-2:], ".")
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

func isAwsChina(d *schema.ResourceData) bool {
	return strings.HasPrefix(d.Type, "aws_") && strings.HasPrefix(d.Get("region").String(), "cn-")
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
