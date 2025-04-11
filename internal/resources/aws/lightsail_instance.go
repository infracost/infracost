package aws

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type LightsailInstance struct {
	Address  string
	Region   string
	BundleID string
}

func (r *LightsailInstance) CoreType() string {
	return "LightsailInstance"
}

func (r *LightsailInstance) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *LightsailInstance) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *LightsailInstance) BuildResource() *schema.Resource {
	bundlePrefixMappings := map[string]string{
		"nano":    "0.5GB",
		"micro":   "1GB",
		"small":   "2GB",
		"medium":  "4GB",
		"large":   "8GB",
		"xlarge":  "16GB",
		"2xlarge": "32GB",
		"4xlarge": "64GB",
	}

	operatingSystemSuffix := ""
	operatingSystemLabel := "Linux/UNIX"

	if strings.Contains(strings.ToLower(r.BundleID), "_win_") {
		operatingSystemSuffix = "_win"
		operatingSystemLabel = "Windows"
	}

	bundlePrefix := strings.Split(strings.ToLower(r.BundleID), "_")[0]

	memory, ok := bundlePrefixMappings[bundlePrefix]
	if !ok {
		// this will end up showing a 'product not found' warning
		memory = bundlePrefix
	}

	usagetypeRegex := fmt.Sprintf("-BundleUsage:%s%s$", memory, operatingSystemSuffix)

	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:           fmt.Sprintf("Virtual server (%s)", operatingSystemLabel),
				Unit:           "hours",
				UnitMultiplier: decimal.NewFromInt(1),
				HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(r.Region),
					Service:       strPtr("AmazonLightsail"),
					ProductFamily: strPtr("Lightsail Instance"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: regexPtr(usagetypeRegex)},
					},
				},
			},
		},
		UsageSchema: r.UsageSchema(),
	}
}
