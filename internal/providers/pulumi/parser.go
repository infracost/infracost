package pulumi

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/providers/pulumi/aws"
	"github.com/infracost/infracost/internal/schema"
	"github.com/pulumi/pulumi/sdk/v3/go/common/display"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

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

func (p *Parser) parsePreviewDigest(t display.PreviewDigest, usage schema.UsageMap, rawValues gjson.Result) ([]*schema.Resource, []*schema.Resource, error) {
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
		// log.Debugf("resource type: %s", resourceType)
		if resourceType == "awsx" || resourceType == "unknown" {
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

		if ud := usage.Get(name); ud != nil {
			usageData = ud
		}

		resourceData := schema.NewPulumiResourceData(resourceType, providerName, name, tags, rawValues, string(step.URN))
		refResources[name] = resourceData
		// You have to load this in the loop so it will find the resources.
		p.parseReferences(refResources, rawValues)
		p.loadInfracostProviderUsageData(usage, refResources)
		if r := p.createResource(resourceData, usageData); r != nil {
			if step.Op == "same" {
				pastResources = append(pastResources, r)
			} else if step.Op == "create" {
				resources = append(resources, r)
			}
		}
	}
	return pastResources, resources, nil
}

func (p *Parser) loadUsageFileResources(u schema.UsageMap) []*schema.Resource {
	resources := make([]*schema.Resource, 0)

	for k, v := range u.Data() {
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

func (p *Parser) loadInfracostProviderUsageData(u schema.UsageMap, resData map[string]*schema.ResourceData) {
	log.Debugf("Loading usage data from Infracost provider resources")

	for _, d := range resData {
		if isInfracostResource(d) {
			p.ctx.SetContextValue("terraformInfracostProviderEnabled", true)

			for _, ref := range d.References("resources") {
				address := ref.Address
				resource := u.Get(address)
				if resource == nil {
					u.Data()[address] = schema.NewUsageData(ref.Address, convertToUsageAttributes(d.RawValues))
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
			// Get any values for the fields and check if they map to IDs or ARNs of any resources
			// Replaced refExists with _ as we only need to know the attribute exists
			for i, refExists := range d.RawValues.Get(attr).Array() {
				// log.Debugf("i %s, attr %s, refVal %s", i, fmt.Sprintf(`newState.inputs.%s`, attr), refExists)
				log.Debugf("Searching for %s", refExists)
				attrFirst := strings.Split(attr, ".")[0]
				searchString := fmt.Sprintf(`propertyDependencies.%s`, attrFirst)
				// log.Debugf("searchString %s", searchString)
				refVal := d.RawValues.Get(searchString).Array()[i]
				// Check ID map
				idRefs, ok := idMap[refVal.String()]
				// log.Debugf("idRefs %s, ok %s", idRefs, ok)
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
	case "pulumi":
		return strings.ToLower(strings.Join(resourceTypeArray, "_"))
	case "kubernetes":
		log.Debugf("supported kubernetes provider %s", providerPrefix)
		return strings.ToLower(strings.ReplaceAll(strings.Join(resourceTypeArray, "_"), "/", "_"))
	default:
		log.Debugf("Unsupported provider %s", providerPrefix)
		return "unknown"
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
