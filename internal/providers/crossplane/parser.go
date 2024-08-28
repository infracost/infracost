package crossplane

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

var defaultProviderRegions = map[string]string{
	"aws":     "us-east-1",
	"google":  "us-central1",
	"azurerm": "us-east",
}

// Parser ...
type Parser struct {
	ctx                  *config.ProjectContext
	includePastResources bool
}

// NewParser ...
func NewParser(ctx *config.ProjectContext, includePastResources bool) *Parser {
	return &Parser{
		ctx:                  ctx,
		includePastResources: includePastResources,
	}
}

// parsedResource is used to collect a PartialResource with its corresponding ResourceData
// so the ResourceData may be used internally by the parsing job, while the PartialResource
// can be passed back up to top level functions.
type parsedResource struct {
	PartialResource *schema.PartialResource
	ResourceData    *schema.ResourceData
}

func (p *Parser) createResource(d *schema.ResourceData, u *schema.UsageData) parsedResource {
	registryMap := GetResourceRegistryMap()

	if registryItem, ok := (*registryMap)[d.Type]; ok {
		if registryItem.NoPrice {
			resource := &schema.Resource{
				Name:         d.Type + "." + d.Address,
				IsSkipped:    true,
				NoPrice:      true,
				SkipMessage:  "Free resource.",
				ResourceType: d.Type,
			}
			return parsedResource{
				PartialResource: schema.NewPartialResource(d, resource, nil, nil),
				ResourceData:    d,
			}
		}

		// Use CoreRFunc to generate a CoreResource if possible.
		if registryItem.CoreRFunc != nil {
			coreRes := registryItem.CoreRFunc(d)
			if coreRes != nil {
				return parsedResource{
					PartialResource: schema.NewPartialResource(d, nil, coreRes, nil),
					ResourceData:    d,
				}
			}
		} else {
			res := registryItem.RFunc(d, u)
			if res != nil {
				res.Name = d.Type + "." + d.Address
				if u != nil {
					res.EstimationSummary = u.CalcEstimationSummary()
				}

				return parsedResource{
					PartialResource: schema.NewPartialResource(d, res, nil, nil),
					ResourceData:    d,
				}
			}
		}
	}

	return parsedResource{
		PartialResource: schema.NewPartialResource(
			d,
			&schema.Resource{
				Name:        d.Type + "." + d.Address,
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

func (p *Parser) parseTemplates(data [][]byte, usage map[string]*schema.UsageData) ([]parsedResource, error) {
	var resources []parsedResource

	baseResources := p.loadUsageFileResources(usage)

	for _, bytes := range data {
		if !gjson.ValidBytes(bytes) {
			return nil, errors.New("invalid JSON")
		}
		resources = append(resources, p.parseTemplate(usage, gjson.ParseBytes(bytes))...)
	}

	resources = append(resources, baseResources...)

	return resources, nil
}

func (p *Parser) parseTemplate(usage map[string]*schema.UsageData, parsed gjson.Result) []parsedResource {
	var resources []parsedResource
	var parseFunc func(gjson.Result) map[string]*schema.ResourceData

	if parsed.Get("kind").String() == "Composition" {
		parseFunc = p.parseCompositeResource
	} else {
		parseFunc = p.parseSimpleResource
	}

	resData := parseFunc(parsed)

	for _, d := range resData {
		var usageData *schema.UsageData
		if ud := usage[d.Address]; ud != nil {
			usageData = ud
		} else if strings.HasSuffix(d.Address, "]") {
			lastIndexOfOpenBracket := strings.LastIndex(d.Address, "[")
			if arrayUsageData := usage[fmt.Sprintf("%s[*]", d.Address[:lastIndexOfOpenBracket])]; arrayUsageData != nil {
				usageData = arrayUsageData
			}
		}

		if r := p.createResource(d, usageData); r.PartialResource != nil {
			resources = append(resources, r)
		}
	}

	return resources
}

func (p *Parser) parseSimpleResource(parsed gjson.Result) map[string]*schema.ResourceData {
	resources := make(map[string]*schema.ResourceData)
	name, resourceType, provider, address, labels := p.getMetaData(parsed)
	spec := parsed.Get("spec")
	spec = schema.AddRawValue(spec, "name", name)
	resources[address] = schema.NewResourceData(resourceType, provider, address, &labels, spec)
	return resources
}

func (p *Parser) getMetaData(parsed gjson.Result) (string, string, string, string, map[string]string) {
	apiVersion := parsed.Get("apiVersion").String()
	kind := parsed.Get("kind").String()
	name := parsed.Get("metadata.name").String()
	provider := getProvider(apiVersion)
	labels := getLabels(parsed)
	address := fmt.Sprintf("%s.%s.%s", provider, kind, name)
	resourceType := fmt.Sprintf("%s/%s", provider, kind)
	return name, resourceType, provider, address, labels
}

func (p *Parser) parseCompositeResource(parsed gjson.Result) map[string]*schema.ResourceData {
	resources := make(map[string]*schema.ResourceData)
	if parsed.Get("spec").Get("resources").IsArray() {
		for _, r := range parsed.Get("spec.resources").Array() {
			base := r.Get("base")
			base = schema.AddRawValue(base, "name", r.Get("name").String())
			resource := p.parseSimpleResource(base)
			for key, value := range resource {
				resources[key] = value
			}
		}
	}
	return resources
}

func (p *Parser) loadUsageFileResources(u map[string]*schema.UsageData) []parsedResource {
	var resources []parsedResource
	for k, v := range u {
		for _, t := range GetUsageOnlyResources() {
			if strings.HasPrefix(k, fmt.Sprintf("%s.", t)) {
				d := schema.NewResourceData(t, "global", k, &map[string]string{}, gjson.Result{})
				resources = append(resources, p.createResource(d, v))
			}
		}
	}
	return resources
}
