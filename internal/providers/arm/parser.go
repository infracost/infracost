package arm

import (
	"strings"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/providers/arm/azure"
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

type Parser struct {
	ctx                  *config.ProjectContext
	includePastResources bool
}

func NewParser(ctx *config.ProjectContext, includePastResources bool) *Parser {
	return &Parser{ctx: ctx, includePastResources: includePastResources}
}

type parsedResource struct {
	PartialResource *schema.PartialResource
	ResourceData    *schema.ResourceData
}

func (p *Parser) ParseJSON(data gjson.Result, usage schema.UsageMap) ([]*parsedResource, error) {
	parsedResources := []*parsedResource{}

	resourceData, _ := p.parseResourceData(&data)

	p.populateUsageData(resourceData, usage)

	for _, d := range resourceData {

		parsedResource := p.createParsedResource(d, d.UsageData)
		parsedResources = append(parsedResources, &parsedResource)

	}
	return parsedResources, nil
}

func (p *Parser) createParsedResource(d *schema.ResourceData, u *schema.UsageData) parsedResource {
	for cKey, cValue := range azure.GetSpecialContext(d) {
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

func (p *Parser) parseResourceData(data *gjson.Result) (map[string]*schema.ResourceData, error) {
	resources := make(map[string]*schema.ResourceData)
	for _, res := range data.Array() {
		t := res.Get("type").String()
		if t == "Microsoft.Compute/virtualMachines" {
			// There is no official naming for Linux or Windows virtual machines, so we need to modify the resource name to obtain the resource from the registry
			GetOSResourceType(&t, &res)
		}
		// New address will consist of the official Microsoft resource type, with an extra '/' at the end to separate the resource type from the name
		newAddress := strings.Clone(t) + "/" + res.Get("name").Str
		resData := schema.NewResourceData(strings.Clone(t), "azurerm", newAddress, nil, res)
		resData.Region = res.Get("location").String()
		resources[strings.Clone(resData.Address)] = resData
	}
	return resources, nil
}

/*
**
There doesn't seem to be an official Microsoft resource name that indicates whether the VM is Linux or Windows
So, we have to check the OS type from the properties and modified the register and mappping accordingly
In a previous version of the Infracost code, the azure_virtual_machine function used to check the os type and create a linux/windows virtual machine accordingly
**
*/
func GetOSResourceType(resourceType *string, data *gjson.Result) {
	name := "Microsoft.Compute/virtualMachines"
	os := "Linux"
	if data.Get("storage_image_reference.0.offer").Type != gjson.Null {
		if strings.ToLower((data.Get("storage_image_reference.0.offer")).String()) == "windowsserver" {
			os = "Windows"
		}
	}
	if strings.ToLower((data.Get("storage_image_reference.0.offer")).String()) == "windows" {
		os = "Windows"
	}
	*resourceType = name + "/" + os

}

func (p *Parser) populateUsageData(resData map[string]*schema.ResourceData, usage schema.UsageMap) {
	for _, d := range resData {
		d.UsageData = usage.Get(d.Address)
	}
}
