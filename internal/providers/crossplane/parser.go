package crossplane

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

var defaultProviderRegions = map[string]string{
	"aws":     "us-east-1",
	"google":  "us-central1",
	"azurerm": "eastus",
}

// Parser ...
type Parser struct {
	ctx *config.ProjectContext
}

// NewParser ...
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
			res.Tags = d.Tags
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

func (p *Parser) parseJSON(data [][]byte, usage map[string]*schema.UsageData) ([]*schema.Resource, []*schema.Resource, error) {
	var resources []*schema.Resource
	baseResources := p.loadUsageFileResources(usage)
	// Process each Crossplane template
	for _, bytes := range data {
		if !gjson.ValidBytes(bytes) {
			return baseResources, baseResources, errors.New("invalid JSON")
		}
		resources = append(resources, p.parseJSONResources(false, usage, gjson.ParseBytes(bytes))...)
	}
	resources = append(resources, baseResources...)
	return nil, resources, nil
}

func (p *Parser) parseJSONResources(parsePrior bool, usage map[string]*schema.UsageData, parsed gjson.Result) []*schema.Resource {
	resData := map[string]*schema.ResourceData{}
	var resources []*schema.Resource

	if parsed.Get("kind").String() == "Composition" {
		resData = p.parseCompositeResource(parsed)
	} else {
		resData = p.parseSimpleResourse(parsed)
	}

	// p.parseReferences(resData, conf)
	// p.stripDataResources(resData)

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
		if r := p.createResource(d, usageData); r != nil {
			resources = append(resources, r)
		}
	}

	return resources
}

func (p *Parser) parseSimpleResourse(parsed gjson.Result) map[string]*schema.ResourceData {
	resources := make(map[string]*schema.ResourceData)
	name, resourceType, provider, address, labels := p.getMetaData(parsed)
	spec := parsed.Get("spec")
	spec = schema.AddRawValue(spec, "name", name)
	resources[address] = schema.NewResourceData(resourceType, provider, address, labels, spec)
	return resources
}

func (p *Parser) getMetaData(parsed gjson.Result) (string, string, string, string, map[string]string) {
	apiVersion := parsed.Get("apiVersion").String()
	kind := parsed.Get("kind").String()
	name := parsed.Get("metadata.name").String()
	provider := getProvider(apiVersion)
	labels := getLabels(parsed)
	address := apiVersion + "/" + kind
	resourceType := provider + "/" + kind
	return name, resourceType, provider, address, labels
}

func (p *Parser) parseCompositeResource(parsed gjson.Result) map[string]*schema.ResourceData {
	resources := make(map[string]*schema.ResourceData)
	if parsed.Get("spec").Get("resources").IsArray() {
		for _, r := range parsed.Get("spec").Get("resources").Array() {
			base := r.Get("base")
			kind := base.Get("kind").String()
			switch kind {
			case "Provider", "ProviderConfig", "CompositeResourceDefinition", "ResourceGroup", "ProviderConfigUsage", "Account":
				log.Infof("Skipping: %s", kind)
			default:
				//TODO: Process r.Get("patches") then update base before calling parseSimpleResourse
				resource := p.parseSimpleResourse(base)
				for key, value := range resource {
					resources[key] = value
				}
			}
		}
	}
	return resources
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
