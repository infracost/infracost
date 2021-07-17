package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	"github.com/shopspring/decimal"
)

func GetAzureRMPostgreSQLFlexibleServerRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_postgresql_flexible_server",
		RFunc: NewAzureRMPostrgreSQLFlexibleServer,
	}
}

func NewAzureRMPostrgreSQLFlexibleServer(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var costComponents []*schema.CostComponent

	region := d.Get("location").String()
	sku := d.Get("sku_name").String()
	var tier, types, skuName, meterName string

	if s := strings.Split(sku, "_"); len(s) == 3 {
		tier = strings.Split(sku, "_")[0]
		types = strings.Split(sku, "_")[2]
	} else if s := strings.Split(sku, "_"); len(s) == 4 {
		tier = strings.Split(sku, "_")[0]
		types = strings.Split(sku, "_")[2]
	} else {
		log.Warnf("Unrecognised PostgreSQL Flexible Server SKU format for resource %s: %s", d.Address, sku)
		return nil
	}

	tierName := map[string]string{
		"B":  "Burstable",
		"GP": "General Purpose",
		"MO": "Memory Optimized",
	}[strings.ToUpper(tier)]

	if tierName == "" {
		log.Warnf("Unrecognised PostgreSQL tier prefix for resource %s: %s", d.Address, tierName)
		return nil
	}

	if strings.ToLower(tierName) == "burstable" {
		meterName = types
		skuName = types
	} else {
		meterName = "vCore"
		cores := types[1:]
		cores = cores[:(len(cores) - 1)]
		skuName = fmt.Sprintf("%s vCore", cores)
	}

	costComponents = append(costComponents, &schema.CostComponent{
		Name:           fmt.Sprintf("Compute (%s)", sku),
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Azure Database for PostgreSQL"),
			ProductFamily: strPtr("Databases"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", ValueRegex: strPtr(fmt.Sprintf("/Azure Database for PostgreSQL Flexible Server %s/i", tierName))},
				{Key: "skuName", ValueRegex: strPtr(fmt.Sprintf("/^%s$/i", skuName))},
				{Key: "meterName", ValueRegex: strPtr(fmt.Sprintf("/^%s$/i", meterName))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	})

	var storageGB *decimal.Decimal
	if d.Get("storage_mb").Type != gjson.Null {
		storageGB = decimalPtr(decimal.NewFromInt(d.Get("storage_mb").Int() / 1024))
	}
	costComponents = append(costComponents, &schema.CostComponent{
		Name:            "Storage",
		Unit:            "GB",
		UnitMultiplier:  1,
		MonthlyQuantity: storageGB,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Azure Database for PostgreSQL"),
			ProductFamily: strPtr("Databases"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Azure Database for PostgreSQL Flexible Server Storage")},
				{Key: "meterName", Value: strPtr("Storage Data Stored")},
			},
		},
	})

	var backupStorageGB *decimal.Decimal
	if u != nil && u.Get("additional_backup_storage_gb").Exists() {
		backupStorageGB = decimalPtr(decimal.NewFromInt(u.Get("additional_backup_storage_gb").Int()))
	}

	costComponents = append(costComponents, &schema.CostComponent{
		Name:            "Additional backup storage",
		Unit:            "GB",
		UnitMultiplier:  1,
		MonthlyQuantity: backupStorageGB,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Azure Database for PostgreSQL"),
			ProductFamily: strPtr("Databases"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Azure Database for PostgreSQL Flexible Server Backup Storage")},
				{Key: "meterName", Value: strPtr("Backup Storage LRS Data Stored")},
			},
		},
	})

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
