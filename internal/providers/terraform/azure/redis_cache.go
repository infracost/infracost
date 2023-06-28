package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetAzureRMRedisCacheRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_redis_cache",
		RFunc: NewAzureRMRedisCache,
	}
}

func NewAzureRMRedisCache(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{})

	skuName := d.Get("sku_name").String()
	family := d.Get("family").String()
	capacity := d.Get("capacity").String()

	sku := family + capacity
	productName := fmt.Sprintf("Azure Redis Cache %s", skuName)

	nodesPerShard := map[string]int64{
		"basic":    1,
		"standard": 2,
		"premium":  2,
	}[strings.ToLower(skuName)]

	shards := int64(1)

	if strings.EqualFold(skuName, "premium") {
		if d.Get("shard_count").Type != gjson.Null {
			shards = d.Get("shard_count").Int()
		}

		if d.Get("replicas_per_primary").Type != gjson.Null {
			nodesPerShard = 1 + d.Get("replicas_per_primary").Int()
		} else if d.Get("replicas_per_master").Type != gjson.Null {
			nodesPerShard = 1 + d.Get("replicas_per_master").Int()
		}
	}

	nodes := shards * nodesPerShard

	// Standard and Premium caches are billed per 2 nodes
	qty := decimal.NewFromInt(nodes)
	mul := schema.HourToMonthUnitMultiplier
	if strings.EqualFold(skuName, "standard") || strings.EqualFold(skuName, "premium") {
		qty = qty.Div(decimal.NewFromInt(2))
		mul = mul.Div(decimal.NewFromInt(2))
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:           fmt.Sprintf("Cache usage (%s_%s%s)", skuName, family, capacity),
				Unit:           "nodes",
				UnitMultiplier: mul,
				HourlyQuantity: decimalPtr(qty),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("azure"),
					Region:        strPtr(region),
					Service:       strPtr("Redis Cache"),
					ProductFamily: strPtr("Databases"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "productName", ValueRegex: strPtr(fmt.Sprintf("/^%s$/i", productName))},
						{Key: "skuName", ValueRegex: strPtr(fmt.Sprintf("/^%s$/i", sku))},
						{Key: "meterName", ValueRegex: strPtr(fmt.Sprintf("/^%s Cache$/i", sku))},
					},
				},
				PriceFilter: &schema.PriceFilter{
					PurchaseOption: strPtr("Consumption"),
				},
			},
		},
	}

}
