package idem

import (
	"fmt"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/providers/idem/aws"
	"strings"

	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
)

// TODO: What is this?
// These show differently in the plan JSON for Terraform 0.12 and 0.13.
var infracostProviderNames = []string{"infracost", "registry.terraform.io/infracost/infracost"}

type Parser struct {
	ctx                  *config.ProjectContext
	includePastResources bool
}

func NewParser(ctx *config.ProjectContext, includePastResources bool) *Parser {
	return &Parser{ctx: ctx,
		includePastResources: includePastResources}
}

func (p *Parser) createResource(d *schema.ResourceData, u *schema.UsageData) *schema.PartialResource {
	registryMap := GetResourceRegistryMap()

	if isAwsChina(d) {
		p.ctx.SetContextValue("isAWSChina", true)
	}

	if registryItem, ok := (*registryMap)[d.Type]; ok {
		if registryItem.NoPrice {
			return &schema.PartialResource{
				ResourceData: d,
				Resource: &schema.Resource{
					Name:        d.Address,
					IsSkipped:   true,
					NoPrice:     true,
					SkipMessage: "Free resource.",
				},
				CloudResourceIDs: registryItem.CloudResourceIDFunc(d),
			}
		}

		// TODO: What's the difference?
		// Use the CoreRFunc to generate a CoreResource if possible.  This is
		// the new/preferred way to create provider-agnostic resources that
		// support advanced features such as Infracost Cloud usage estimates
		// and actual costs.
		if registryItem.CoreRFunc != nil {
			coreRes := registryItem.CoreRFunc(d)
			if coreRes != nil {
				return &schema.PartialResource{ResourceData: d, CoreResource: coreRes, CloudResourceIDs: registryItem.CloudResourceIDFunc(d)}
			}
		} else {
			res := registryItem.RFunc(d, u)
			if res != nil {
				if u != nil {
					res.EstimationSummary = u.CalcEstimationSummary()
				}

				return &schema.PartialResource{ResourceData: d, Resource: res, CloudResourceIDs: registryItem.CloudResourceIDFunc(d)}
			}
		}
	}

	return &schema.PartialResource{
		ResourceData: d,
		Resource: &schema.Resource{
			Name:        d.Address,
			IsSkipped:   true,
			SkipMessage: "This resource is not currently supported",
		},
	}
}

func (p *Parser) parseJSONResources(parsePrior bool, baseResources []*schema.PartialResource, usage schema.UsageMap, parsed gjson.Result) []*schema.PartialResource {
	var resources []*schema.PartialResource
	resources = append(resources, baseResources...)

	resData := p.parseResourceData(parsed, parsePrior)

	p.parseReferences(resData)
	p.populateUsageData(resData, usage)

	for _, d := range resData {
		if r := p.createResource(d, d.UsageData); r != nil {
			resources = append(resources, r)
		}
	}
	return resources
}

func (p *Parser) parseTemplate(parsed gjson.Result, usage schema.UsageMap) ([]*schema.PartialResource, []*schema.PartialResource, error) {
	baseResources := p.loadUsageFileResources(usage)
	resources := p.parseJSONResources(false, baseResources, usage, parsed)
	if !p.includePastResources {
		return nil, resources, nil
	}

	pastResources := p.parseJSONResources(true, baseResources, usage, parsed)

	return pastResources, resources, nil
}

func (p *Parser) parseResourceData(parsed gjson.Result, parsePrior bool) map[string]*schema.ResourceData {
	resources := make(map[string]*schema.ResourceData)
	parsed.ForEach(func(key, r gjson.Result) bool {
		addr := r.Get("name").String()
		t := r.Get("ref").String()
		provider := r.Get("acct_details.provider").String()

		// Some Idem utility states do not have type
		if t == "" {
			return true
		}
		var val gjson.Result

		if parsePrior {
			val = r.Get("old_state")
		} else {
			val = r.Get("new_state")
		}

		if val.Type == gjson.Null {
			return true
		}

		val = schema.AddRawValue(val, "region", r.Get("acct_details.region_name").String())

		tags := parseTags(t, val)

		resources[addr] = schema.NewResourceData(t, provider, addr, tags, val)
		return true
	})

	return resources
}

func (p *Parser) parseReferences(resData map[string]*schema.ResourceData) {
	registryMap := GetResourceRegistryMap()

	// Create a map of id -> resource data so we can lookup references
	idMap := make(map[string][]*schema.ResourceData)

	for _, d := range resData {

		// check for any "default" ids declared by the provider for this resource
		if f := registryMap.GetDefaultRefIDFunc(d.Type); f != nil {
			for _, defaultID := range f(d) {
				var reId = d.Type + ":" + defaultID
				if _, ok := idMap[reId]; !ok {
					idMap[reId] = []*schema.ResourceData{}
				}
				idMap[reId] = append(idMap[reId], d)
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
			var splitRef = strings.Split(attr, ":")

			// Get any values for the fields and check if they map to IDs or ARNs of any resources
			for _, refVal := range d.Get(splitRef[1]).Array() {
				if refVal.String() == "" {
					continue
				}

				// Check ID map
				var id = splitRef[0] + ":" + refVal.String()
				idRefs, ok := idMap[id]
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

// populateUsageData finds the UsageData for each ResourceData and sets the ResourceData.UsageData field
// in case it is needed when processing a reference attribute
func (p *Parser) populateUsageData(resData map[string]*schema.ResourceData, usage schema.UsageMap) {
	for _, d := range resData {
		d.UsageData = usage.Get(d.Address)
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

func (p *Parser) loadUsageFileResources(u schema.UsageMap) []*schema.PartialResource {
	resources := make([]*schema.PartialResource, 0)

	for k, v := range u.Data() {
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

func isAwsChina(d *schema.ResourceData) bool {
	return strings.HasPrefix(d.Type, "states.aws") && strings.HasPrefix(d.Get("region").String(), "cn-")
}

func getProviderPrefix(resourceType string) string {
	// Idem type format is states.{provider type}.{resource path}.present
	providerPrefix := strings.Split(resourceType, ".")
	if len(providerPrefix) < 2 {
		return ""
	}

	return providerPrefix[1]
}

func parseTags(resourceType string, v gjson.Result) map[string]string {
	providerPrefix := getProviderPrefix(resourceType)

	switch providerPrefix {
	case "aws":
		return aws.ParseTags(resourceType, v)
	default:
		logging.Logger.Debugf("Unsupported provider %s", providerPrefix)
		return map[string]string{}
	}
}
