package google

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetRedisInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_redis_instance",
		RFunc: NewRedisInstance,
	}
}

func NewRedisInstance(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	serviceTier := "Basic"

	var tierMapping = map[string]string{
		"BASIC":       "Basic",
		"STANDARD_HA": "Standard",
	}

	if d.Get("tier").Exists() {
		serviceTier = tierMapping[d.Get("tier").String()]
	}

	var memorySize = d.Get("memory_size_gb").Int()
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
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:           name,
				Unit:           "GB-hours",
				UnitMultiplier: 1,
				HourlyQuantity: decimalPtr(decimal.NewFromInt(memorySize)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("gcp"),
					Region:        strPtr(region),
					Service:       strPtr("Cloud Memorystore for Redis"),
					ProductFamily: strPtr("ApplicationServices"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "description", ValueRegex: strPtr(description)},
					},
				},
			},
		},
	}
}
