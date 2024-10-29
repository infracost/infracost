package azure

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMMariaDBServerRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_mariadb_server",
		RFunc: NewAzureRMMariaDBServer,
	}
}

func NewAzureRMMariaDBServer(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Region

	var costComponents []*schema.CostComponent
	serviceName := "Azure Database for MariaDB"

	sku := d.Get("sku_name").String()
	var tier, family, cores string
	if s := strings.Split(sku, "_"); len(s) == 3 {
		tier = strings.Split(sku, "_")[0]
		family = strings.Split(sku, "_")[1]
		cores = strings.Split(sku, "_")[2]
	} else {
		logging.Logger.Warn().Msgf("Unrecognised MariaDB SKU format for resource %s: %s", d.Address, sku)
		return nil
	}

	tierName := map[string]string{
		"B":  "Basic",
		"GP": "General Purpose",
		"MO": "Memory Optimized",
	}[tier]

	if tierName == "" {
		logging.Logger.Warn().Msgf("Unrecognised MariaDB tier prefix for resource %s: %s", d.Address, tierName)
		return nil
	}

	productNameRegex := fmt.Sprintf("/%s - Compute %s/", tierName, family)
	skuName := fmt.Sprintf("%s vCore", cores)

	costComponents = append(costComponents, databaseComputeInstance(region, fmt.Sprintf("Compute (%s)", sku), serviceName, productNameRegex, skuName))

	storageGB := d.Get("storage_mb").Int() / 1024

	// MO and GP storage cost are the same, and we don't have cost component for MO Storage now
	if strings.ToLower(tier) == "mo" {
		tierName = "General Purpose"
	}
	productNameRegex = fmt.Sprintf("/%s - Storage/", tierName)

	costComponents = append(costComponents, databaseStorageComponent(region, serviceName, productNameRegex, storageGB))

	var backupStorageGB *decimal.Decimal

	if u != nil && u.Get("additional_backup_storage_gb").Exists() {
		backupStorageGB = decimalPtr(decimal.NewFromInt(u.Get("additional_backup_storage_gb").Int()))
	}

	skuName = "Backup LRS"
	if d.Get("geo_redundant_backup_enabled").Exists() && d.Get("geo_redundant_backup_enabled").Bool() {
		skuName = "Backup GRS"
	}

	costComponents = append(costComponents, databaseBackupStorageComponent(region, serviceName, skuName, backupStorageGB))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func databaseComputeInstance(region, name, serviceName, productNameRegex, skuName string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           name,
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr(serviceName),
			ProductFamily: strPtr("Databases"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", ValueRegex: strPtr(productNameRegex)},
				{Key: "skuName", Value: strPtr(skuName)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func databaseStorageComponent(region, serviceName, productNameRegex string, storageGB int64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(storageGB)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr(serviceName),
			ProductFamily: strPtr("Databases"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", ValueRegex: strPtr(productNameRegex)},
			},
		},
	}
}

func databaseBackupStorageComponent(region, serviceName, skuName string, backupStorageGB *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Additional backup storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: backupStorageGB,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr(serviceName),
			ProductFamily: strPtr("Databases"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", ValueRegex: strPtr("/Single Server - Backup Storage/")},
				{Key: "skuName", Value: strPtr(skuName)},
			},
		},
	}
}
