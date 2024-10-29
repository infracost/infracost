package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type RedisInstance struct {
	Address      string
	Region       string
	Tier         string
	MemorySizeGB float64
}

func (r *RedisInstance) CoreType() string {
	return "RedisInstance"
}

func (r *RedisInstance) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *RedisInstance) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *RedisInstance) BuildResource() *schema.Resource {
	serviceTier := "Basic"

	var tierMapping = map[string]string{
		"BASIC":       "Basic",
		"STANDARD_HA": "Standard",
	}

	if r.Tier != "" {
		serviceTier = tierMapping[r.Tier]
	}

	var memorySize = r.MemorySizeGB
	var capacityTier string

	if memorySize >= 1 && memorySize <= 4 {
		capacityTier = "M1"
	} else if memorySize >= 5 && memorySize <= 10 {
		capacityTier = "M2"
	} else if memorySize >= 11 && memorySize <= 35 {
		capacityTier = "M3"
	} else if memorySize >= 36 && memorySize <= 100 {
		capacityTier = "M4"
	} else {
		capacityTier = "M5"
	}

	description := fmt.Sprintf("/Redis Capacity %s %s/", serviceTier, capacityTier)
	name := fmt.Sprintf("Redis instance (%s, %s)", strings.ToLower(serviceTier), capacityTier)

	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:           name,
				Unit:           "GB",
				UnitMultiplier: schema.HourToMonthUnitMultiplier,
				HourlyQuantity: decimalPtr(decimal.NewFromFloat(memorySize)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("gcp"),
					Region:        strPtr(r.Region),
					Service:       strPtr("Cloud Memorystore for Redis"),
					ProductFamily: strPtr("ApplicationServices"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "description", ValueRegex: strPtr(description)},
					},
				},
			},
		}, UsageSchema: r.UsageSchema(),
	}
}
