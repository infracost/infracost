package azure

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

var managedRedisFamilyNames = map[string]string{
	"Balanced":         "Balanced",
	"ComputeOptimized": "Compute Optimized",
	"FlashOptimized":   "Flash Optimized",
	"MemoryOptimized":  "Memory Optimized",
}

// ManagedRedis represents an Azure Managed Redis instance.
//
// Resource information: https://learn.microsoft.com/azure/redis/overview
// Pricing information: https://azure.microsoft.com/en-us/pricing/details/managed-redis/
type ManagedRedis struct {
	Address       string
	Region        string
	SKU           string
	InstanceCount int64
}

func (r *ManagedRedis) CoreType() string {
	return "ManagedRedis"
}

func (r *ManagedRedis) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *ManagedRedis) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ManagedRedis) BuildResource() *schema.Resource {
	productName, shortSKU, ok := r.parseSKU()
	if !ok {
		return &schema.Resource{
			Name:        r.Address,
			IsSkipped:   true,
			SkipMessage: fmt.Sprintf("Unrecognized Azure Managed Redis SKU %q", r.SKU),
		}
	}

	instanceCount := r.InstanceCount
	if instanceCount == 0 {
		instanceCount = 2
	}

	return &schema.Resource{
		Name:        r.Address,
		UsageSchema: r.UsageSchema(),
		CostComponents: []*schema.CostComponent{
			{
				Name:           fmt.Sprintf("Cache usage (%s)", r.SKU),
				Unit:           "instances",
				UnitMultiplier: schema.HourToMonthUnitMultiplier,
				HourlyQuantity: decimalPtr(decimal.NewFromInt(instanceCount)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr(vendorName),
					Region:        strPtr(r.Region),
					Service:       strPtr("Redis Cache"),
					ProductFamily: strPtr("Databases"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "productName", Value: strPtr(productName)},
						{Key: "skuName", Value: strPtr(shortSKU)},
						{Key: "meterName", Value: strPtr(fmt.Sprintf("%s Cache Instance", shortSKU))},
					},
				},
				PriceFilter: priceFilterConsumption,
			},
		},
	}
}

func (r *ManagedRedis) parseSKU() (string, string, bool) {
	rawFamily, rawSKU, ok := strings.Cut(strings.TrimSpace(r.SKU), "_")
	if !ok {
		return "", "", false
	}

	familyName, ok := managedRedisFamilyNames[rawFamily]
	if !ok || strings.TrimSpace(rawSKU) == "" {
		return "", "", false
	}

	return fmt.Sprintf("Azure Managed Redis - %s", familyName), rawSKU, true
}
