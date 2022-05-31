package pulumi

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/providers/pulumi/types"
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

	for i := range t.Steps {
		var step = t.Steps[i]
		if step.NewState.Type == "pulumi:pulumi:Stack" {
			continue
		}
		var name = step.NewState.URN.URNName()
		var resourceType = step.NewState.Type.String()
		var providerName = strings.Split(step.NewState.Type.String(), ":")[0]
		var localInputs = step.NewState.Inputs
		localInputs["config"] = t.Config
		localInputs["dependencies"] = step.NewState.Dependencies
		localInputs["propertyDependencies"] = step.NewState.PropertyDependencies
		var inputs, _ = json.Marshal(localInputs)

		tags := map[string]string{} // TODO: Where do I get tags? Day 2 Issue
		var usageData *schema.UsageData

		if ud := usage[name]; ud != nil {
			usageData = ud
		}

		resourceData := schema.NewResourceData(resourceType, providerName, name, tags, gjson.Parse(string(inputs)))
		if r := p.createResource(resourceData, usageData); r != nil {
			if step.Op == "same" {
				pastResources = append(resources, r)
			} else if step.Op == "create" {
				resources = append(resources, r)
			}
		}
	}

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
