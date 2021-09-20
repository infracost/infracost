package cloudformation

import (
	"fmt"
	"strings"

	"github.com/awslabs/goformation/v4/cloudformation"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

type Parser struct {
	ctx *config.ProjectContext
}

func NewParser(ctx *config.ProjectContext) *Parser {
	return &Parser{ctx}
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

func (p *Parser) parseTemplate(t *cloudformation.Template, usage map[string]*schema.UsageData) ([]*schema.Resource, []*schema.Resource, error) {
	baseResources := p.loadUsageFileResources(usage)

	var resources []*schema.Resource
	resources = append(resources, baseResources...)

	for name, d := range t.Resources {
		tags := map[string]string{} // TODO: Where do I get tags?
		var usageData *schema.UsageData

		if ud := usage[name]; ud != nil {
			usageData = ud
		} else if strings.HasSuffix(name, "]") {
			lastIndexOfOpenBracket := strings.LastIndex(name, "[")

			if arrayUsageData := usage[fmt.Sprintf("%s[*]", name[:lastIndexOfOpenBracket])]; arrayUsageData != nil {
				usageData = arrayUsageData
			}
		}
		resourceData := schema.NewCFResourceData(d.AWSCloudFormationType(), "aws", name, tags, d)

		if r := p.createResource(resourceData, usageData); r != nil {
			resources = append(resources, r)
		}
	}

	return resources, resources, nil
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
