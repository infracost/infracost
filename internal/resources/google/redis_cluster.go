package google

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

type RedisCluster struct {
	Address      string
	Region       string
	Tier         string
	MemorySizeGB float64
}

func (r *RedisCluster) CoreType() string {
	return "RedisCluster"
}

func (r *RedisCluster) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *RedisCluster) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *RedisCluster) BuildResource() *schema.Resource {
	serviceTier := "Standard"

	tierMapping := map[string]string{
		"STANDARD_HA": "Standard",
		"BASIC":       "Basic",
	}

	if mapped, ok := tierMapping[strings.ToUpper(r.Tier)]; ok {
		serviceTier = mapped
	}

	name := fmt.Sprintf("Redis cluster (%s)", strings.ToLower(serviceTier))
	description := fmt.Sprintf("/Redis Cluster Capacity %s/", serviceTier)

	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:           name,
				Unit:           "GB",
				UnitMultiplier: schema.HourToMonthUnitMultiplier,
				HourlyQuantity: decimalPtr(decimal.NewFromFloat(r.MemorySizeGB)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("gcp"),
					Region:        strPtr(r.Region),
					Service:       strPtr("Cloud Memorystore for Redis Cluster"),
					ProductFamily: strPtr("ApplicationServices"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "description", ValueRegex: strPtr(description)},
					},
				},
			},
		},
		UsageSchema: r.UsageSchema(),
	}
}
