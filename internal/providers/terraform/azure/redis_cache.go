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

	nodes := map[string]int64{
		"basic":    1,
		"standard": 2,
		"premium":  2,
	}[strings.ToLower(skuName)]

	if d.Get("replicas_per_master").Type != gjson.Null {
		nodes = 1 + d.Get("replicas_per_master").Int()
	}

	if d.Get("shard_count").Type != gjson.Null {
		nodes = 2 * d.Get("shard_count").Int()
	}

	if u != nil && u.Get("redis_nodes").Type != gjson.Null {
		nodes = u.Get("redis_nodes").Int()
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:           fmt.Sprintf("Cache usage (%s_%s), %v node", skuName, sku, nodes),
				Unit:           "hours",
				UnitMultiplier: 1,
				HourlyQuantity: decimalPtr(decimal.NewFromInt(nodes)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("azure"),
					Region:        strPtr(region),
					Service:       strPtr("Redis Cache"),
					ProductFamily: strPtr("Databases"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "productName", Value: strPtr(productName)},
						{Key: "skuName", Value: strPtr(sku)},
						{Key: "meterName", ValueRegex: strPtr(fmt.Sprintf("/^%s Cache$/", sku))},
					},
				},
				PriceFilter: &schema.PriceFilter{
					PurchaseOption: strPtr("Consumption"),
				},
			},
		},
	}

}
