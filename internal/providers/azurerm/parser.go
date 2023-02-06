package azurerm

import (
	"encoding/json"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/providers/azurerm/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

type usageMap map[string]*schema.UsageData

type Parser struct {
	ctx *config.ProjectContext
}

func NewParser(ctx *config.ProjectContext) *Parser {
	return &Parser{ctx}
}

// Same as providers/terraform/parser.go:createPartialResource
func (p *Parser) createPartialResource(d *schema.ResourceData, u *schema.UsageData) *schema.PartialResource {
	registryMap := resources.GetRegistryMap()

	if registryItem, ok := (*registryMap)[d.Type]; ok {
		if registryItem.NoPrice {
			return &schema.PartialResource{
				ResourceData: d,
				Resource: &schema.Resource{
					Name:         d.Address,
					ResourceType: d.Type,
					Tags:         d.Tags,
					IsSkipped:    true,
					NoPrice:      true,
					SkipMessage:  "Free resource.",
				},
				CloudResourceIDs: registryItem.CloudResourceIDFunc(d),
			}
		}

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

func (p *Parser) parse(parseBefore bool, j []byte, usage usageMap) ([]*schema.PartialResource, []*schema.PartialResource, error) {
	var resources []*schema.PartialResource
	var pastResources []*schema.PartialResource
	// TODO: baseresources??
	// baseResources := p.loadUsageFileResources(usage)

	// TODO: StripSetupAzureRMWrapper??
	var whatif WhatIf
	err := json.Unmarshal(j, &whatif)
	if err != nil {
		return pastResources, resources, errors.New("Failed to unmarshal whatif operation result")
	}

	if whatif.Status != "Succeeded" {
		return pastResources, resources, errors.New("WhatIf operation was not successful")
	}

	resources, err = p.parseResources(false, &whatif, usage)
	if err != nil {
		return pastResources, resources, err
	}

	if parseBefore {
		pastResources, err = p.parseResources(false, &whatif, usage)
		if err != nil {
			return pastResources, resources, err
		}
	}

	return pastResources, resources, nil
}

// TODO: need baseresources like TF provider?
func (p *Parser) parseResources(parseBefore bool, whatif *WhatIf, usage usageMap) ([]*schema.PartialResource, error) {
	var resources = make([]*schema.PartialResource, 0)

	// Parse both 'before' state and 'after' state when it's present
	for _, change := range whatif.Properties.Changes {
		var rd *schema.ResourceData
		var err error
		rd, err = p.parseResourceData(&change, parseBefore)

		if err != nil {
			return nil, err
		}
		res := p.createPartialResource(rd, rd.UsageData)
		resources = append(resources, res)
	}

	return resources, nil
}

// TODO: This is not exhaustive yet, probably need to do something with 'Delta' and 'WhatIfPropertyChange'
func (p *Parser) parseResourceData(change *ResourceSnapshot, isBefore bool) (*schema.ResourceData, error) {
	var parsed gjson.Result
	var resData *schema.ResourceData

	if isBefore {
		parsed = change.Before()
	} else {
		parsed = change.After()
	}

	armType := parsed.Get("type")
	resId := parsed.Get("id")
	if !armType.Exists() || !resId.Exists() {
		return nil, errors.New("Could not get resource type and id from WhatIf")
	}

	tfType := resources.GetTFResourceFromAzureRMType(armType.Str)
	resData = schema.NewAzureRMResourceData(tfType, resId.Str, parsed)

	return resData, nil
}
