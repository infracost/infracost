package cloudformation

import (
	"github.com/awslabs/goformation/v7/cloudformation"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
)

type Parser struct {
	ctx                  *config.ProjectContext
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

		// Use the CoreRFunc to generate a CoreResource if possible.  This is
		// the new/preferred way to create provider-agnostic resources that
		// support advanced features such as Infracost Cloud usage estimates
		// and actual costs.
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

func (p *Parser) parseTemplate(t *cloudformation.Template, usage schema.UsageMap) []parsedResource {
	resources := make([]parsedResource, 0, len(t.Resources))

	for name, d := range t.Resources {
		resourceData := schema.NewCFResourceData(d.AWSCloudFormationType(), "aws", name, nil, d)
		usageData := usage.Get(resourceData.Type + "." + resourceData.Address)
		resources = append(resources, p.createResource(resourceData, usageData))
	}

	return resources
}
