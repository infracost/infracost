package aws

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"
	log "github.com/sirupsen/logrus"

	"strings"

	"github.com/shopspring/decimal"
)

func GetLightsailInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_lightsail_instance",
		RFunc: NewLightsailInstance,
	}
}

func NewLightsailInstance(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	region := d.Get("region").String()

	type bundleSpecs struct {
		vcpu   string
		memory string
	}

	bundlePrefixMappings := map[string]bundleSpecs{
		"nano":    {"1", "0.5GB"},
		"micro":   {"1", "1GB"},
		"small":   {"1", "2GB"},
		"medium":  {"2", "4GB"},
		"large":   {"2", "8GB"},
		"xlarge":  {"4", "16GB"},
		"2xlarge": {"8", "32GB"},
	}

	operatingSystem := "Linux"
	operatingSystemLabel := "Linux/UNIX"

	if strings.Contains(d.Get("bundle_id").String(), "_win_") {
		operatingSystem = "Windows"
		operatingSystemLabel = "Windows"
	}

	bundlePrefix := strings.Split(d.Get("bundle_id").String(), "_")[0]

	specs, ok := bundlePrefixMappings[bundlePrefix]
	if !ok {
		log.Warnf("Skipping resource %s. Unrecognised bundle_id %s", d.Address, d.Get("bundle_id").String())
		return nil
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:           fmt.Sprintf("Virtual server (%s)", operatingSystemLabel),
				Unit:           "hours",
				HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonLightsail"),
					ProductFamily: strPtr("Lightsail Instance"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "operatingSystem", Value: strPtr(operatingSystem)},
						{Key: "vcpu", Value: strPtr(specs.vcpu)},
						{Key: "memory", Value: strPtr(specs.memory)},
					},
				},
			},
		},
	}
}
