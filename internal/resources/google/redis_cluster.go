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
	NodeType     string
	NodeCount    int
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
	nodeTypeDescriptions := map[string]string{
		"REDIS_SHARED_CORE_NANO": "Shared Core Nano",
		"REDIS_STANDARD_SMALL":   "Standard Small",
		"REDIS_HIGHMEM_MEDIUM":   "Highmem Medium",
		"REDIS_HIGHMEM_XLARGE":   "Highmem XLarge",
	}

	desc, ok := nodeTypeDescriptions[strings.ToUpper(r.NodeType)]
	if !ok {
		desc = r.NodeType
	}

	name := fmt.Sprintf("Redis Cluster node (%s)", strings.ToLower(desc))
	descriptionRegex := fmt.Sprintf("Redis Cluster Node %s", desc)

	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:           name,
				Unit:           "hours",
				UnitMultiplier: decimal.NewFromInt(1),
				HourlyQuantity: decimalPtr(decimal.NewFromInt(int64(r.NodeCount))),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("gcp"),
					Region:        strPtr(r.Region),
					Service:       strPtr("Cloud Memorystore for Redis"),
					ProductFamily: strPtr("ApplicationServices"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key:"description",ValueRegex: regexPtr(descriptionRegex)},
					},
				},
			},
		},
		UsageSchema: r.UsageSchema(),
	}
}