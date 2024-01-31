package terraform

import (
	"bytes"
	stdJson "encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/logging"
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

// parsedResource is used to collect a PartialResource with its corresponding ResourceData so the
// ResourceData may be used internally by the parsing job, while the PartialResource can be passed
// back up to top level functions.  This allows the ResourceData to be garbage collected once the parsing
// job is complete.
type parsedResource struct {
	PartialResource *schema.PartialResource
	ResourceData    *schema.ResourceData
}

func (p *Parser) createParsedResource(d *schema.ResourceData, u *schema.UsageData) parsedResource {
	registryMap := GetResourceRegistryMap()

	for cKey, cValue := range getSpecialContext(d) {
		p.ctx.ContextValues.SetValue(cKey, cValue)
	}

	if registryItem, ok := (*registryMap)[d.Type]; ok {
		if registryItem.NoPrice {
			resource := &schema.Resource{
				Name:        d.Address,
				IsSkipped:   true,
				NoPrice:     true,
				SkipMessage: "Free resource.",
				Metadata:    d.Metadata,
			}
			return parsedResource{
				PartialResource: schema.NewPartialResource(d, resource, nil, registryItem.CloudResourceIDFunc(d)),
				ResourceData:    d,
			}

		}

		// Use the CoreRFunc to generate a CoreResource if possible.  This is
		// the new/preferred way to create provider-agnostic resources that
		// support advanced features such as Infracost Cloud usage estimates
		// and actual costs.
		if registryItem.CoreRFunc != nil {
			coreRes := registryItem.CoreRFunc(d)
			if coreRes != nil {
				return parsedResource{
					PartialResource: schema.NewPartialResource(d, nil, coreRes, registryItem.CloudResourceIDFunc(d)),
					ResourceData:    d,
				}
			}
		} else {
			res := registryItem.RFunc(d, u)
			if res != nil {
				if u != nil {
					res.EstimationSummary = u.CalcEstimationSummary()
				}

				return parsedResource{
					PartialResource: schema.NewPartialResource(d, res, nil, registryItem.CloudResourceIDFunc(d)),
					ResourceData:    d,
				}
			}
		}
	}

	return parsedResource{
		PartialResource: schema.NewPartialResource(
			d,
			&schema.Resource{
				Name:        d.Address,
				IsSkipped:   true,
				SkipMessage: "This resource is not currently supported",
				Metadata:    d.Metadata,
			},
			nil,
			[]string{},
		),
		ResourceData: d,
	}
}

func (p *Parser) parseJSONResources(parsePrior bool, baseResources []parsedResource, usage schema.UsageMap, confLoader *ConfLoader, parsed, providerConf, vars gjson.Result) []parsedResource {
	var resources []parsedResource
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

	resData := p.parseResourceData(isState, confLoader, providerConf, vals, vars)

	p.parseReferences(resData, confLoader)
	p.parseTags(resData, confLoader, providerConf)

	p.stripDataResources(resData)
	p.populateUsageData(resData, usage)

	for _, d := range resData {
		resources = append(resources, p.createParsedResource(d, d.UsageData))
	}

	return resources
}

// populateUsageData finds the UsageData for each ResourceData and sets the ResourceData.UsageData field
// in case it is needed when processing a reference attribute
func (p *Parser) populateUsageData(resData map[string]*schema.ResourceData, usage schema.UsageMap) {
	for _, d := range resData {
		d.UsageData = usage.Get(d.Address)
	}
}

type ParsedPlanConfiguration struct {
	PastResources        []*schema.PartialResource
	PastResourceDatas    []*schema.ResourceData
	CurrentResources     []*schema.PartialResource
	CurrentResourceDatas []*schema.ResourceData
	ProviderMetadata     []schema.ProviderMetadata
	RemoteModuleCalls    []string
}

func newParsedPlanConfiguration(pastResources, currentResources []parsedResource, metadatas []schema.ProviderMetadata, remoteModuleCalls []string) *ParsedPlanConfiguration {
	ppc := ParsedPlanConfiguration{
		PastResources:        make([]*schema.PartialResource, 0, len(pastResources)),
		PastResourceDatas:    make([]*schema.ResourceData, 0, len(pastResources)),
		CurrentResources:     make([]*schema.PartialResource, 0, len(currentResources)),
		CurrentResourceDatas: make([]*schema.ResourceData, 0, len(currentResources)),
		ProviderMetadata:     metadatas,
		RemoteModuleCalls:    remoteModuleCalls,
	}

	for _, parsed := range pastResources {
		ppc.PastResources = append(ppc.PastResources, parsed.PartialResource)
		ppc.PastResourceDatas = append(ppc.PastResourceDatas, parsed.ResourceData)
	}

	for _, parsed := range currentResources {
		ppc.CurrentResources = append(ppc.CurrentResources, parsed.PartialResource)
		ppc.CurrentResourceDatas = append(ppc.CurrentResourceDatas, parsed.ResourceData)
	}

	return &ppc
}

func (p *Parser) parseJSON(j []byte, usage schema.UsageMap) (*ParsedPlanConfiguration, error) {
	baseResources := p.loadUsageFileResources(usage)

	if !gjson.ValidBytes(j) {
		return newParsedPlanConfiguration(
			baseResources,
			baseResources,
			nil,
			nil,
		), errors.New("invalid JSON")
	}

	parsed := gjson.ParseBytes(j)

	p.terraformVersion = parsed.Get("terraform_version").String()
	providerConf := parsed.Get("configuration.provider_config")
	conf := parsed.Get("configuration.root_module")
	vars := parsed.Get("variables")

	providerMetadata := parseProviderConfig(providerConf)
	confLoader := NewConfLoader(conf)
	calledRemoteModules := collectModulesSourceUrls(conf.Get("module_calls.*").Array())
	resources := p.parseJSONResources(false, baseResources, usage, confLoader, parsed, providerConf, vars)
	if !p.includePastResources || !parsed.Get("prior_state").Exists() {
		return newParsedPlanConfiguration(
			nil,
			resources,
			providerMetadata,
			calledRemoteModules,
		), nil
	}

	// Check if the prior state is the same as the planned state
	// and if so we can just return pointers to the same resources
	if gjsonEqual(parsed.Get("prior_state.values.root_module"), parsed.Get("planned_values.root_module")) {
		return newParsedPlanConfiguration(
			resources,
			resources,
			providerMetadata,
			calledRemoteModules,
		), nil
	}

	pastResources := p.parseJSONResources(true, baseResources, usage, confLoader, parsed, providerConf, vars)
	resourceChanges := parsed.Get("resource_changes").Array()
	pastResources = stripNonTargetResources(pastResources, resources, resourceChanges)

	return newParsedPlanConfiguration(
		pastResources,
		resources,
		providerMetadata,
		calledRemoteModules,
	), nil
}

func collectModulesSourceUrls(moduleCalls []gjson.Result) []string {
	remoteUrls := map[string]bool{}

	for _, call := range moduleCalls {
		source := call.Get("sourceUrl").String()
		if ok := remoteUrls[source]; !ok && source != "" {
			remoteUrls[source] = true
		}

		if call.Get("module.module_calls").Exists() {
			for _, source := range collectModulesSourceUrls(call.Get("module.module_calls.*").Array()) {
				if ok := remoteUrls[source]; !ok {
					remoteUrls[source] = true
				}
			}
		}

	}

	if len(remoteUrls) == 0 {
		return nil
	}

	var urls []string
	for source := range remoteUrls {
		// Perf/memory leak: Copy gjson string slices that may be returned so we don't prevent
		// the entire underlying parsed json from being garbage collected.
		sourceCopy := strings.Clone(source)
		urls = append(urls, sourceCopy)
	}

	return urls
}

func parseProviderConfig(providerConf gjson.Result) []schema.ProviderMetadata {
	var metadatas []schema.ProviderMetadata

	confMap := providerConf.Map()

	// Sort the metadata by configKey so any outputted JSON is deterministic
	var keys = make([]string, 0, len(confMap))
	for k := range confMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		conf := confMap[k]
		md := schema.ProviderMetadata{
			// Perf/memory leak: Copy gjson string slices that may be returned so we don't prevent
			// the entire underlying parsed json from being garbage collected.
			Name:      strings.Clone(conf.Get("name").String()),
			Filename:  strings.Clone(conf.Get("infracost_metadata.filename").String()),
			StartLine: conf.Get("infracost_metadata.start_line").Int(),
			EndLine:   conf.Get("infracost_metadata.end_line").Int(),
		}

		for _, defaultTags := range conf.Get("expressions.default_tags").Array() {
			if md.DefaultTags == nil {
				md.DefaultTags = make(map[string]string)
			}

			for key, value := range defaultTags.Get("tags.constant_value").Map() {
				// Perf/memory leak: Copy gjson string slices that may be returned so we don't prevent
				// the entire underlying parsed json from being garbage collected.
				md.DefaultTags[strings.Clone(key)] = strings.Clone(value.String())
			}
		}

		metadatas = append(metadatas, md)
	}

	return metadatas
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

func (p *Parser) loadUsageFileResources(u schema.UsageMap) []parsedResource {
	resources := make([]parsedResource, 0)

	for k, v := range u.Data() {
		for _, t := range GetUsageOnlyResources() {
			if strings.HasPrefix(k, fmt.Sprintf("%s.", t)) {
				d := schema.NewResourceData(t, "global", k, nil, gjson.Result{})
				// set the usage data as a field on the resource data in case it is needed when
				// processing reference attributes.
				d.UsageData = v
				resources = append(resources, p.createParsedResource(d, v))
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
func stripNonTargetResources(pastResources []parsedResource, resources []parsedResource, resourceChanges []gjson.Result) []parsedResource {
	resourceAddrMap := make(map[string]bool, len(resources))
	for _, resource := range resources {
		resourceAddrMap[resource.PartialResource.Address] = true
	}

	diffAddrMap := make(map[string]bool, len(resourceChanges))
	for _, change := range resourceChanges {
		diffAddrMap[change.Get("address").String()] = true
	}

	var filteredResources []parsedResource
	for _, resource := range pastResources {
		_, rOk := resourceAddrMap[resource.PartialResource.Address]
		_, dOk := diffAddrMap[resource.PartialResource.Address]
		if dOk || rOk {
			filteredResources = append(filteredResources, resource)
		}
	}
	return filteredResources
}

func (p *Parser) parseResourceData(isState bool, confLoader *ConfLoader, providerConf, planVals, vars gjson.Result) map[string]*schema.ResourceData {
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

		resConf := confLoader.GetResourceConfJSON(addr)

		// Override the region when requested
		region := overrideRegion(addr, t, p.ctx.RunContext.Config)

		// If not overridden try getting the region from the ARN
		if region == "" {
			region = resourceRegion(t, v)
		}

		// Otherwise use region from the provider conf
		if region == "" {
			region = providerRegion(addr, providerConf, vars, t, resConf)
		}

		// Perf/memory leak: Copy gjson string slices that may be returned so we don't prevent
		// the entire underlying parsed json from being garbage collected.
		v = schema.AddRawValue(v, "region", strings.Clone(region))
		data := schema.NewResourceData(strings.Clone(t), strings.Clone(provider), strings.Clone(addr), nil, v)

		// Perf/memory leak: Copy gjson string slices that may be returned so we don't prevent
		// the entire underlying parsed json from being garbage collected.
		data.Metadata = gjson.ParseBytes([]byte(r.Get("infracost_metadata").Raw)).Map()
		resources[strings.Clone(addr)] = data
	}

	// Recursively add any resources for child modules
	for _, m := range planVals.Get("child_modules").Array() {
		for addr, d := range p.parseResourceData(isState, confLoader, providerConf, m, vars) {
			resources[addr] = d
		}
	}

	return resources
}

func getSpecialContext(d *schema.ResourceData) map[string]interface{} {
	providerPrefix := getProviderPrefix(d.Type)

	switch providerPrefix {
	case "aws":
		return aws.GetSpecialContext(d)
	case "azurerm":
		return azure.GetSpecialContext(d)
	case "google":
		return google.GetSpecialContext(d)
	default:
		logging.Logger.Debug().Msgf("Unsupported provider %s", providerPrefix)
		return map[string]interface{}{}
	}
}

func parseDefaultTags(providerConf, resConf gjson.Result) *map[string]string {
	// this only works for aws, we'll need to review when other providers support default tags
	providerKey := parseProviderKey(resConf)
	dTagsArray := providerConf.Get(fmt.Sprintf("%s.expressions.default_tags", gjsonEscape(providerKey))).Array()
	if len(dTagsArray) == 0 {
		return nil
	}

	defaultTags := make(map[string]string)
	for _, dTags := range dTagsArray {
		for k, v := range dTags.Get("tags.constant_value").Map() {
			defaultTags[k] = v.String()
		}
	}

	return &defaultTags
}

func overrideRegion(addr string, resourceType string, config *config.Config) string {
	region := ""
	providerPrefix := getProviderPrefix(resourceType)

	switch providerPrefix {
	case "aws":
		region = config.AWSOverrideRegion
	case "azurerm":
		region = config.AzureOverrideRegion
	case "google":
		region = config.GoogleOverrideRegion
	default:
		logging.Logger.Debug().Msgf("Unsupported provider %s", providerPrefix)
	}

	if region != "" {
		logging.Logger.Debug().Msgf("Overriding region (%s) for %s", region, addr)
	}

	return region
}

func resourceRegion(resourceType string, v gjson.Result) string {
	providerPrefix := getProviderPrefix(resourceType)

	switch providerPrefix {
	case "aws":
		return aws.GetResourceRegion(resourceType, v)
	case "azurerm":
		return azure.GetResourceRegion(resourceType, v)
	case "google":
		return google.GetResourceRegion(resourceType, v)
	default:
		logging.Logger.Debug().Msgf("Unsupported provider %s", providerPrefix)
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
		providerPrefix := getProviderPrefix(resourceType)
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
				logging.Logger.Debug().Msgf("Falling back to default region (%s) for %s", region, addr)
			}
		}
	}

	return region
}

func getProviderPrefix(resourceType string) string {
	providerPrefix := strings.Split(resourceType, "_")
	if len(providerPrefix) == 0 {
		return ""
	}

	return providerPrefix[0]
}

func parseProviderKey(resConf gjson.Result) string {
	return resConf.Get("provider_config_key").String()
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

func (p *Parser) stripDataResources(resData map[string]*schema.ResourceData) {
	for addr, d := range resData {
		if strings.HasPrefix(addressResourcePart(d.Address), "data.") {
			delete(resData, addr)
		}
	}
}

func (p *Parser) parseReferences(resData map[string]*schema.ResourceData, confLoader *ConfLoader) {
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

	parseKnownModuleRefs(resData, confLoader)

	for _, d := range resData {
		var refAttrs []string

		if isInfracostResource(d) {
			refAttrs = []string{"resources"}
		} else {
			refAttrs = registryMap.GetReferenceAttributes(d.Type)
		}

		for _, attr := range refAttrs {
			found := p.parseConfReferences(resData, confLoader, d, attr, registryMap)

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

func (p *Parser) parseConfReferences(resData map[string]*schema.ResourceData, confLoader *ConfLoader, d *schema.ResourceData, attr string, registryMap *ResourceRegistryMap) bool {
	// Check if there's a reference in the conf
	resConf := confLoader.GetResourceConfJSON(d.Address)
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
					logging.Logger.Debug().Msgf("reference specifies a count: using resource %s for %s.%s", a, d.Address, attr)
				}
			} else if containsString(refs, "each.key") {
				a := fmt.Sprintf("%s[\"%s\"]", refAddr, addressKey(d.Address))
				refData, ok = resData[a]

				if ok {
					logging.Logger.Debug().Msgf("reference specifies a key: using resource %s for %s.%s", a, d.Address, attr)
				}
			}
		}

		// if still not found, see if there's a matching resource with an [0] array part
		if !ok {
			a := fmt.Sprintf("%s[0]", refAddr)
			refData, ok = resData[a]

			if ok {
				logging.Logger.Debug().Msgf("reference does not specify a count: using resource %s for for %s.%s", a, d.Address, attr)
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

func (p *Parser) parseTags(data map[string]*schema.ResourceData, confLoader *ConfLoader, providerConf gjson.Result) {
	for _, resourceData := range data {
		providerPrefix := getProviderPrefix(resourceData.Type)
		var tags *map[string]string
		switch providerPrefix {
		case "aws":
			resConf := confLoader.GetResourceConfJSON(resourceData.Address)
			defaultTags := parseDefaultTags(providerConf, resConf)
			tags = aws.ParseTags(defaultTags, resourceData)
		case "azurerm":
			tags = azure.ParseTags(resourceData)
		case "google":
			tags = google.ParseTags(resourceData)
		default:
			logging.Logger.Debug().Msgf("Unsupported provider %s", providerPrefix)
		}

		resourceData.Tags = tags
	}
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
func parseKnownModuleRefs(resData map[string]*schema.ResourceData, confLoader *ConfLoader) {
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
		modNames := getModuleNames(d.Address)
		modSource := confLoader.GetModuleConfJSON(modNames).Get("source").String()

		for _, knownRef := range knownRefs {
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

func gjsonEqual(a, b gjson.Result) bool {
	var err error

	var aOut bytes.Buffer
	err = stdJson.Compact(&aOut, []byte(a.Raw))
	if err != nil {
		logging.Logger.Debug().Msgf("error indenting JSON: %s", err)
		return false
	}

	var bOut bytes.Buffer
	err = stdJson.Compact(&bOut, []byte(b.Raw))
	if err != nil {
		logging.Logger.Debug().Msgf("error indenting JSON: %s", err)
		return false
	}

	return aOut.String() == bOut.String()
}

type ConfLoader struct {
	conf        gjson.Result
	moduleCache map[string]gjson.Result
	// Seems like we can't cache module resources because the providerConfigKey can be different.
	// resourceCache map[string]gjson.Result
}

func NewConfLoader(conf gjson.Result) *ConfLoader {
	return &ConfLoader{
		conf:        conf,
		moduleCache: make(map[string]gjson.Result),
		// resourceCache: make(map[string]gjson.Result),
	}
}

func (l *ConfLoader) GetModuleConfJSON(names []string) gjson.Result {
	if len(names) == 0 {
		return l.conf
	}

	// Build up the gjson search key
	p := make([]string, 0, len(names))
	for _, n := range names {
		p = append(p, fmt.Sprintf("module_calls.%s", n))
	}

	key := strings.Join(p, ".module.")
	if c, ok := l.moduleCache[key]; ok {
		return c
	}

	c := l.conf.Get(key)
	l.moduleCache[key] = c

	return c
}

func (l *ConfLoader) GetResourceConfJSON(addr string) gjson.Result {
	modNames := getModuleNames(addr)
	moduleConf := l.GetModuleConfJSON(modNames)

	if len(modNames) > 0 {
		moduleConf = moduleConf.Get("module")
	}

	key := fmt.Sprintf(`resources.#(address="%s")`, removeAddressArrayPart(addressResourcePart(addr)))
	// if c, ok := l.resourceCache[key]; ok {
	//	 return c
	// }

	c := moduleConf.Get(key)
	// l.resourceCache[key] = c

	return c
}
