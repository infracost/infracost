package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"

	log "github.com/sirupsen/logrus"

	"strings"

	"github.com/shopspring/decimal"
)

type LightsailInstance struct {
	Address  string
	Region   string
	BundleID string
}

var LightsailInstanceUsageSchema = []*schema.UsageItem{}

func (r *LightsailInstance) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *LightsailInstance) BuildResource() *schema.Resource {
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

	if strings.Contains(strings.ToLower(r.BundleID), "_win_") {
		operatingSystem = "Windows"
		operatingSystemLabel = "Windows"
	}

	bundlePrefix := strings.Split(strings.ToLower(r.BundleID), "_")[0]

	specs, ok := bundlePrefixMappings[bundlePrefix]
	if !ok {
		log.Warnf("Skipping resource %s. Unrecognized bundle_id %s", r.Address, r.BundleID)
		return nil
	}

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
						{Key: "operatingSystem", Value: strPtr(operatingSystem)},
						{Key: "vcpu", Value: strPtr(specs.vcpu)},
						{Key: "memory", Value: strPtr(specs.memory)},
					},
				},
				PriceFilter: &schema.PriceFilter{
					EndUsageAmount: strPtr("Inf"),
				},
			},
		},
		UsageSchema: LightsailInstanceUsageSchema,
	}
}
