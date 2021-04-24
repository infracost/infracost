package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

func GetAzureMariaDBServerRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_mariadb_server",
		RFunc: NewAzureMariaDBServer,
	}
}

func NewAzureMariaDBServer(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var costComponents []*schema.CostComponent
	serviceName := "Azure Database for MariaDB"

	region := d.Get("location").String()
	sku := d.Get("sku_name").String()
	var tier, family, cores string
	if s := strings.Split(sku, "_"); len(s) == 3 {
		tier = strings.Split(sku, "_")[0]
		family = strings.Split(sku, "_")[1]
		cores = strings.Split(sku, "_")[2]
	} else {
		log.Warnf("Unrecognised MariaDB SKU format for resource %s: %s", d.Address, sku)
		return nil
	}

	tierName := map[string]string{
		"B":  "Basic",
		"GP": "General Purpose",
		"MO": "Memory Optimized",
	}[tier]

	if tierName == "" {
		log.Warnf("Unrecognised MariaDB tier prefix for resource %s: %s", d.Address, tierName)
		return nil
	}

	productNameRegex := fmt.Sprintf("/%s - Compute %s/", tierName, family)
	skuName := fmt.Sprintf("%s vCore", cores)

	costComponents = append(costComponents, databaseComputeInstance(region, serviceName, sku, productNameRegex, skuName))

	storageGB := d.Get("storage_mb").Int() / 1024

	// MO and GP storage cost are the same, and we don't have cost component for MO Storage now
	if tier == "MO" {
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

func databaseComputeInstance(region, serviceName, sku, productNameRegex, skuName string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           fmt.Sprintf("Compute (%s)", sku),
		Unit:           "hours",
		UnitMultiplier: 1,
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
	}
}

func databaseStorageComponent(region, serviceName, productNameRegex string, storageGB int64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Storage",
		Unit:            "GB-months",
		UnitMultiplier:  1,
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
		Unit:            "GB-months",
		UnitMultiplier:  1,
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
