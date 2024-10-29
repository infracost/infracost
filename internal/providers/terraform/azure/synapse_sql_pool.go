package azure

import (
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMSynapseSQLPoolRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_synapse_sql_pool",
		RFunc: NewAzureRMSynapseSQLPool,
		ReferenceAttributes: []string{
			"synapse_workspace_id",
		},
		Notes: []string{"the total costs consist of several resources that should be viewed as a whole"},
		GetRegion: func(defaultRegion string, d *schema.ResourceData) string {
			return lookupRegion(d, []string{"synapse_workspace_id"})
		},
	}
}

func NewAzureRMSynapseSQLPool(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Region

	costComponents := make([]*schema.CostComponent, 0)

	sku := ""
	if d.Get("sku_name").Type != gjson.Null {
		sku = d.Get("sku_name").String()
	}
	costComponents = append(costComponents, synapseDedicatedSQLPoolCostComponent(region, "DWU blocks", sku))

	var storage *decimal.Decimal
	if u != nil && u.Get("storage_tb").Type != gjson.Null {
		storage = decimalPtr(decimal.NewFromInt(u.Get("storage_tb").Int()))
	}
	costComponents = append(costComponents, synapseDedicatedSQLPoolStorageCostComponent(region, "Storage", storage))

	disasterRecoveryEnabled := true
	if u != nil && u.Get("disaster_recovery_enabled").Type != gjson.Null {
		disasterRecoveryEnabled = u.Get("disaster_recovery_enabled").Bool()
	}
	if disasterRecoveryEnabled {
		costComponents = append(costComponents, synapseDedicatedSQLPoolDisasterRecoveryStorageCostComponent(region, "Geo-redundant disaster recovery", storage))
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func synapseDedicatedSQLPoolCostComponent(region, name, sku string) *schema.CostComponent {

	return &schema.CostComponent{
		Name:            fmt.Sprintf("%s (%s)", name, sku),
		Unit:            "hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(730)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Azure Synapse Analytics"),
			ProductFamily: strPtr("Analytics"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productId", Value: strPtr("DZH318Z0BZ1B")},
				{Key: "skuName", Value: strPtr(sku)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func synapseDedicatedSQLPoolStorageCostComponent(region, name string, quantity *decimal.Decimal) *schema.CostComponent {

	return &schema.CostComponent{
		Name:            name,
		Unit:            "TB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Azure Synapse Analytics"),
			ProductFamily: strPtr("Analytics"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productId", Value: strPtr("DZH318Z0BXTR")},
				{Key: "skuName", Value: strPtr("Standard LRS")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func synapseDedicatedSQLPoolDisasterRecoveryStorageCostComponent(region, name string, quantity *decimal.Decimal) *schema.CostComponent {

	if quantity != nil {
		quantity = decimalPtr(quantity.Mul(decimal.NewFromInt(1000)))
	}

	return &schema.CostComponent{
		Name:            name,
		Unit:            "TB",
		UnitMultiplier:  decimal.NewFromInt(1000),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Azure Synapse Analytics"),
			ProductFamily: strPtr("Analytics"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productId", Value: strPtr("DZH318Z0BXTP")},
				{Key: "skuName", Value: strPtr("Standard RA-GRS")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
