package azurerm

import (
	"encoding/json"
	"fmt"

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
					SkipMessage:  "Free resource",
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

func (p *Parser) parse(j []byte, usage usageMap) ([]*schema.PartialResource, []*schema.PartialResource, error) {
	var resources []*schema.PartialResource
	var pastResources []*schema.PartialResource
	var whatif WhatIf

	err := json.Unmarshal(j, &whatif)
	if err != nil {
		return pastResources, resources, errors.New("Failed to unmarshal whatif operation result")
	}

	if whatif.Status != "Succeeded" {
		return pastResources, resources, errors.New("WhatIf operation was not successful")
	}

	for _, change := range whatif.Changes {
		before, after, err := p.parseChange(&change, usage)
		if err != nil {
			return nil, nil, err
		}

		if after != nil {
			resources = append(resources, after)
		}

		if before != nil {
			pastResources = append(pastResources, before)
		}
	}

	return pastResources, resources, nil
}

// TODO: need baseresources like TF provider?
func (p *Parser) parseChange(change *ResourceSnapshot, usage usageMap) (*schema.PartialResource, *schema.PartialResource, error) {
	var after *schema.PartialResource
	var before *schema.PartialResource

	beforeData := change.Before()
	afterData := change.After()

	if afterData.Get("id").Exists() {
		afterRd, err := p.parseResourceData(afterData)
		if err != nil {
			return nil, nil, err
		}
		after = p.createPartialResource(afterRd, afterRd.UsageData)
	}

	// Parse before snapshot only when it's present
	if beforeData.Get("id").Exists() {
		beforeRd, err := p.parseResourceData(beforeData)
		if err != nil {
			return nil, nil, err
		}
		before = p.createPartialResource(beforeRd, beforeRd.UsageData)
	}

	return before, after, nil
}

// TODO: This is not exhaustive yet, probably need to do something with 'Delta' and 'WhatIfPropertyChange'
func (p *Parser) parseResourceData(data gjson.Result) (*schema.ResourceData, error) {
	var resData *schema.ResourceData

	armType := data.Get("type")
	resId := data.Get("id")
	if !armType.Exists() || !resId.Exists() {
		return nil, errors.New(fmt.Sprintf("Failed to parse resource data"))
	}

	tfType := resources.GetTFResourceFromAzureRMType(armType.Str, data)
	if len(tfType) == 0 {
		return nil, errors.New(fmt.Sprintf("Could not convert AzureRM type '%s' to TF type", armType.Str))
	}

	resData = schema.NewAzureRMResourceData(tfType, resId.Str, data)

	return resData, nil
}
