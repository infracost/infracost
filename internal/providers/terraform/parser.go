package terraform

import (
	"bytes"
	stdJson "encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/go-version"
	"gopkg.in/yaml.v3"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/hcl"
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
	providerConstraints  hcl.ProviderConstraints
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
	for cKey, cValue := range getSpecialContext(d) {
		p.ctx.ContextValues.SetValue(cKey, cValue)
	}

	if registryItem, ok := (*ResourceRegistryMap)[d.Type]; ok {
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

	resources := make([]parsedResource, len(baseResources), len(resData)+len(baseResources))
	copy(resources, baseResources)

	p.parseReferences(resData, confLoader)
	p.parseTags(resData, confLoader, providerConf)

	p.stripDataResources(resData)
	p.populateUsageData(resData, usage)

	for _, d := range resData {
		p.setRegion(confLoader, d, providerConf, vars)

		resources = append(resources, p.createParsedResource(d, d.UsageData))
	}

	return resources
}

// setRegion sets the region on the given resource data and any references it has.
func (p *Parser) setRegion(confLoader *ConfLoader, d *schema.ResourceData, providerConf gjson.Result, vars gjson.Result) {
	region := p.getRegion(confLoader, d, providerConf, vars)
	d.RawValues = schema.AddRawValue(d.RawValues, "region", region)
	d.Region = region

	for _, references := range d.ReferencesMap {
		for _, ref := range references {
			if ref.Region == "" {
				p.setRegion(confLoader, ref, providerConf, vars)
			}
		}
	}
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
	constraints := parsed.Get("infracost_provider_constraints")
	if constraints.Exists() {
		var providerConstraints hcl.ProviderConstraints
		err := json.Unmarshal([]byte(constraints.Raw), &providerConstraints)
		if err == nil {
			p.providerConstraints = providerConstraints
		}
	}

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

	urls := make([]string, 0, len(remoteUrls))
	for source := range remoteUrls {
		// Perf/memory leak: Copy gjson string slices that may be returned so we don't prevent
		// the entire underlying parsed json from being garbage collected.
		sourceCopy := strings.Clone(source)
		urls = append(urls, sourceCopy)
	}

	return urls
}

func parseProviderConfig(providerConf gjson.Result) []schema.ProviderMetadata {

	confMap := providerConf.Map()

	// Sort the metadata by configKey so any outputted JSON is deterministic
	var keys = make([]string, 0, len(confMap))
	for k := range confMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	metadatas := make([]schema.ProviderMetadata, 0, len(keys))

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

		unknownKeys := conf.Get("infracost_metadata.attributes_with_unknown_keys").Array()
		for _, unknownKey := range unknownKeys {
			vars := unknownKey.Get("missing_variables")
			if vars.IsArray() {
				vals := make([]string, 0, len(vars.Array()))
				for _, v := range vars.Array() {
					vals = append(vals, v.String())
				}
				md.AttributesWithUnknownKeys = append(md.AttributesWithUnknownKeys, schema.AttributeWithUnknownKeys{
					Attribute:        unknownKey.Get("attribute").String(),
					MissingVariables: vals,
				})
			}
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

		data := schema.NewResourceData(strings.Clone(t), strings.Clone(provider), strings.Clone(addr), nil, v)

		// Perf/memory leak: Copy gjson string slices that may be returned so we don't prevent
		// the entire underlying parsed json from being garbage collected.
		data.Metadata = gjson.ParseBytes([]byte(r.Get("infracost_metadata").Raw)).Map()
		maps.Copy(data.ProjectMetadata, p.ctx.ProjectConfig.Metadata)
		resources[strings.Clone(addr)] = data
	}

	// Recursively add any resources for child modules
	for _, m := range planVals.Get("child_modules").Array() {
		maps.Copy(resources, p.parseResourceData(isState, confLoader, providerConf, m, vars))
	}

	return resources
}

func (p *Parser) getRegion(confLoader *ConfLoader, d *schema.ResourceData, providerConf gjson.Result, vars gjson.Result) string {
	resConf := confLoader.GetResourceConfJSON(d.Address)

	// Override the region when requested
	region := overrideRegion(d, p.ctx.RunContext.Config)

	// If not overridden try getting the region from the ARN
	if region == "" {
		region = resourceRegion(d)
	}

	// Otherwise use region from the provider conf
	if region == "" {
		region = providerRegion(d, providerConf, vars, resConf)
	}

	// Perf/memory leak: Copy gjson string slices that may be returned so we don't prevent
	// the entire underlying parsed json from being garbage collected.
	return strings.Clone(region)
}

func getSpecialContext(d *schema.ResourceData) map[string]any {
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
		return map[string]any{}
	}
}

func parseAWSDefaultTags(providerConf, resConf gjson.Result) (map[string]string, []string) {
	// this only works for aws, we'll need to review when other providers support default tags
	providerKey := parseProviderKey(resConf)
	dTagsArray := providerConf.Get(fmt.Sprintf("%s.expressions.default_tags", gjsonEscape(providerKey))).Array()
	if len(dTagsArray) == 0 {
		return nil, nil
	}

	defaultTags := make(map[string]string)
	var missingAttrsCausingUnknownKeys []string
	for _, dTags := range dTagsArray {
		for k, v := range dTags.Get("tags.constant_value").Map() {
			defaultTags[k] = v.String()
		}
		for _, address := range dTags.Get("tags.missing_attributes_causing_unknown_keys").Array() {
			if address.String() == "" {
				continue
			}
			missingAttrsCausingUnknownKeys = append(missingAttrsCausingUnknownKeys, address.String())
		}
	}
	return defaultTags, missingAttrsCausingUnknownKeys
}

func parseGoogleDefaultTags(providerConf, resConf gjson.Result) (map[string]string, []string) {
	providerKey := parseProviderKey(resConf)
	defaultTags := make(map[string]string)
	for k, v := range providerConf.Get(fmt.Sprintf("%s.expressions.default_labels.constant_value", gjsonEscape(providerKey))).Map() {
		defaultTags[k] = v.String()
	}
	missingAttributes := providerConf.Get(
		fmt.Sprintf(
			"%s.expressions.default_labels.missing_attributes_causing_unknown_keys",
			gjsonEscape(providerKey),
		),
	).Array()
	missingAttrsCausingUnknownKeys := make([]string, 0, len(missingAttributes))
	for _, address := range missingAttributes {
		if address.String() == "" {
			continue
		}
		missingAttrsCausingUnknownKeys = append(missingAttrsCausingUnknownKeys, address.String())
	}

	return defaultTags, missingAttrsCausingUnknownKeys
}

func overrideRegion(d *schema.ResourceData, config *config.Config) string {
	region := ""
	providerPrefix := getProviderPrefix(d.Type)

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
		logging.Logger.Debug().Msgf("Overriding region (%s) for %s", region, d.Address)
	}

	return region
}

func resourceRegion(d *schema.ResourceData) string {
	providerPrefix := getProviderPrefix(d.Type)
	var defaultRegion string
	switch providerPrefix {
	case "aws":
		defaultRegion = aws.GetResourceRegion(d)
	case "azurerm":
		defaultRegion = azure.GetResourceRegion(d)
	case "google":
		defaultRegion = google.GetResourceRegion(d)
	default:
		logging.Logger.Debug().Msgf("Unsupported provider %s", providerPrefix)
		return ""
	}

	// let's check if the resource has a specific region lookup function. Resources
	// can define specific region lookup functions over the default provider logic,
	// as some resources require us to infer the region by traversing resource
	// references and other attributes.
	regionFunc := ResourceRegistryMap.GetRegion(d.Type)
	if regionFunc != nil {
		region := regionFunc(defaultRegion, d)
		if region != "" {
			return region
		}
	}

	return defaultRegion
}

func providerRegion(d *schema.ResourceData, providerConf gjson.Result, vars gjson.Result, resConf gjson.Result) string {
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
		providerPrefix := getProviderPrefix(d.Type)
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
				logging.Logger.Debug().Msgf("Falling back to default region (%s) for %s", region, d.Address)
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
	// Create a map of id -> resource data so we can lookup references
	idMap := make(map[string][]*schema.ResourceData)

	for _, d := range resData {

		// check for any "default" ids declared by the provider for this resource
		if f := ResourceRegistryMap.GetDefaultRefIDFunc(d.Type); f != nil {
			for _, defaultID := range f(d) {
				if _, ok := idMap[defaultID]; !ok {
					idMap[defaultID] = []*schema.ResourceData{}
				}
				if slices.Contains(idMap[defaultID], d) {
					continue
				}
				idMap[defaultID] = append(idMap[defaultID], d)
			}
		}

		// check for any "custom" ids specified by the resource and add them.
		if f := ResourceRegistryMap.GetCustomRefIDFunc(d.Type); f != nil {
			for _, customID := range f(d) {
				if _, ok := idMap[customID]; !ok {
					idMap[customID] = []*schema.ResourceData{}
				}
				if slices.Contains(idMap[customID], d) {
					continue
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
			refAttrs = ResourceRegistryMap.GetReferenceAttributes(d.Type)
		}

		for _, attr := range refAttrs {
			found := p.parseConfReferences(resData, confLoader, d, attr, ResourceRegistryMap)

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
						reverseRefAttrs := ResourceRegistryMap.GetReferenceAttributes(ref.Type)
						d.AddReference(attr, ref, reverseRefAttrs)
					}
				}
			}
		}
	}

	fixKnownModuleRefIssues(resData)
}

func (p *Parser) parseConfReferences(resData map[string]*schema.ResourceData, confLoader *ConfLoader, d *schema.ResourceData, attr string, registryMap *RegistryItemMap) bool {
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

type YorConfig struct {
	Name  string `yaml:"name"`
	Value struct {
		Default string `yaml:"default"`
	} `yaml:"value"`
	TagGroups []struct {
		Name string `yaml:"name"`
		Tags []struct {
			Name  string `yaml:"name"`
			Value struct {
				Default string `yaml:"default"`
			} `yaml:"value"`
		} `yaml:"tags"`
	} `yaml:"tag_groups"`
}

func (p *Parser) parseYorTagsFromConfigFile(path string, tags map[string]string) {

	data, err := os.ReadFile(path)
	if err != nil {
		logging.Logger.Debug().Msgf("failed to read yor config: %s", err)
		return
	}

	var conf YorConfig
	if err := yaml.Unmarshal(data, &conf); err != nil {
		logging.Logger.Debug().Msgf("failed to unmarshal yor config: %s", err)
		return
	}

	// single tag style config
	if conf.Name != "" && conf.Value.Default != "" {
		tags[conf.Name] = conf.Value.Default
	}

	// tag group style config
	for _, tg := range conf.TagGroups {
		for _, t := range tg.Tags {
			if t.Name == "" || t.Value.Default == "" {
				continue
			}
			tags[t.Name] = t.Value.Default
		}
	}
}

func (p *Parser) parseYorTagsFromJSON(rawJSON string, tags map[string]string) {
	var simpleTags map[string]string
	if err := stdJson.Unmarshal([]byte(rawJSON), &simpleTags); err == nil {
		for k, v := range simpleTags {
			if k == "" || v == "" {
				continue
			}
			tags[k] = v
		}
	} else {
		logging.Logger.Debug().Msgf("failed to unmarshal yor simple tags json: %s", err)
	}
}

func (p *Parser) parseTags(data map[string]*schema.ResourceData, confLoader *ConfLoader, providerConf gjson.Result) {
	awsTagParsingConfig := aws.TagParsingConfig{PropagateDefaultsToVolumeTags: hcl.ConstraintsAllowVersionOrAbove(p.providerConstraints.AWS, hcl.AWSVersionConstraintVolumeTags)}

	externalTags := make(map[string]string)
	if path := p.ctx.ProjectConfig.YorConfigPath; path != "" {
		if root := p.ctx.RunContext.Config.RootPath; root != "" && !filepath.IsAbs(path) {
			path = filepath.Join(root, path)
		}
		p.parseYorTagsFromConfigFile(path, externalTags)
	}
	if yorSimpleTags := os.Getenv("YOR_SIMPLE_TAGS"); yorSimpleTags != "" {
		p.parseYorTagsFromJSON(yorSimpleTags, externalTags)
	}

	for _, resourceData := range data {

		var missingVarsCausingUnknownTagKeys []string
		var missingVarsCausingUnknownDefaultTagKeys []string
		providerPrefix := getProviderPrefix(resourceData.Type)
		defaultTagSupport := p.areDefaultTagsSupported(providerPrefix)
		var tags map[string]string
		var defaultTags map[string]string
		resConf := confLoader.GetResourceConfJSON(resourceData.Address)
		switch providerPrefix {
		case "aws":
			if defaultTagSupport {
				defaultTags, missingVarsCausingUnknownDefaultTagKeys = parseAWSDefaultTags(providerConf, resConf)
			}
			tags, missingVarsCausingUnknownTagKeys = aws.ParseTags(externalTags, defaultTags, resourceData, awsTagParsingConfig)
		case "azurerm":
			tags, missingVarsCausingUnknownTagKeys = azure.ParseTags(externalTags, resourceData)
		case "google":
			if defaultTagSupport {
				defaultTags, missingVarsCausingUnknownDefaultTagKeys = parseGoogleDefaultTags(providerConf, resConf)
			}
			tags, missingVarsCausingUnknownTagKeys = google.ParseTags(resourceData, externalTags, defaultTags)
		default:
			logging.Logger.Debug().Msgf("Unsupported provider %s", providerPrefix)
		}

		if conf := providerConf.Get(gjsonEscape(parseProviderKey(resConf))); conf.Exists() {
			if metadata := conf.Get("infracost_metadata"); metadata.Exists() {
				providerLink := metadata.Get("filename").String()
				if providerLine := metadata.Get("start_line").Int(); providerLine > 0 {
					providerLink = fmt.Sprintf("%s:%d", providerLink, providerLine)
				}
				resourceData.ProviderLink = providerLink
			}
		}

		if tags != nil {
			resourceData.Tags = &tags
		}
		if defaultTags != nil {
			resourceData.DefaultTags = &defaultTags
		}
		resourceData.ProviderSupportsDefaultTags = defaultTagSupport
		resourceData.TagPropagation = p.getTagPropagationInfo(resourceData)
		resourceData.MissingVarsCausingUnknownTagKeys = missingVarsCausingUnknownTagKeys
		resourceData.MissingVarsCausingUnknownDefaultTagKeys = missingVarsCausingUnknownDefaultTagKeys
	}
}

var (
	versionAWSProviderForDefaultTagSupport = version.Must(version.NewVersion("3.38.0"))
	versionGoogleProviderForDefaultTags    = version.Must(version.NewVersion("5.0.0"))
)

func (p *Parser) areDefaultTagsSupported(providerPrefix string) bool {
	switch providerPrefix {
	case "aws":
		// default tags were added in aws provider v3.38.0 - if the constraints allow a version before this,
		// we can't rely on default tag support
		return hcl.ConstraintsAllowVersionOrAbove(p.providerConstraints.AWS, versionAWSProviderForDefaultTagSupport)
	case "google":
		// default tags (labels) were added in google provider v5.0.0 - if the constraints allow a version before this,
		// we can't rely on default tag support
		return hcl.ConstraintsAllowVersionOrAbove(p.providerConstraints.Google, versionGoogleProviderForDefaultTags)
	default:
		return false
	}
}

func (p *Parser) getTagPropagationInfo(resource *schema.ResourceData) *schema.TagPropagation {
	if expected, ok := aws.ExpectedPropagations[resource.Type]; ok {
		propagateTags := resource.GetStringOrDefault(expected.Attribute, "")
		propagation := &schema.TagPropagation{
			From:      &propagateTags,
			To:        expected.To,
			Attribute: expected.Attribute,
		}
		hasRequired := true
		for _, required := range expected.Requires {
			hasRequired = hasRequired && resource.Get(required).Exists()
		}
		propagation.HasRequiredAttributes = hasRequired
		if expected.RefMap != nil {
			if attr, ok := expected.RefMap[propagateTags]; ok {
				if ref, ok := resource.ReferencesMap[attr]; ok && len(ref) == 1 {
					propagation.Tags = ref[0].Tags
				}
			}
		}
		return propagation
	}
	return nil
}

func isInfracostResource(res *schema.ResourceData) bool {
	return slices.Contains(infracostProviderNames, res.ProviderName)
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
	r := regexp.MustCompile(`\[(\d+)\]$`)
	m := r.FindStringSubmatch(addr)

	if len(m) > 0 {
		i, _ := strconv.Atoi(m[1]) // TODO: unhandled error

		return i
	}

	return -1
}

func addressKey(addr string) string {
	r := regexp.MustCompile(`\["([^"]+)"\]$`)
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
	return slices.Contains(a, s)
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

// fixKnownModuleRefIssues deals with edge cases where the module returns an id that doesn't work for the resource
// this is exemplified by the s3-bucket module which coalesces a number of id's starting with "aws_s3_bucket_policy."
// this means we're referencing the wrong resource in the plan JSON. This function fixes that by replacing the reference.
// the intention is to be laser focused in the application of this where we know the specific conditions it will occur
func fixKnownModuleRefIssues(resData map[string]*schema.ResourceData) {
	knownRefs := []struct {
		SourceResourceType  string
		AttributeName       string
		TargetResource      string
		ReplacementResource string
	}{
		{
			SourceResourceType:  "aws_s3_bucket_lifecycle_configuration",
			AttributeName:       "bucket",
			TargetResource:      "aws_s3_bucket_policy",
			ReplacementResource: "aws_s3_bucket",
		},
	}

	for _, d := range resData {
		for _, knownRef := range knownRefs {
			if d.Type == knownRef.SourceResourceType {
				for _, ref := range d.ReferencesMap[knownRef.AttributeName] {
					if ref.Type == knownRef.TargetResource {
						targetAddress := strings.Replace(ref.Address, knownRef.TargetResource, knownRef.ReplacementResource, 1)

						for _, target := range resData {
							// find possible targets and ensure that its from the same module as existing reference
							if target.Type == knownRef.ReplacementResource && target.Address == targetAddress {
								// replace the reference
								d.ReplaceReference(knownRef.AttributeName, ref, target)

							}
						}
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
