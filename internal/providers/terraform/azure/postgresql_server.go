package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

func GetAzureRMPostgreSQLServerRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_postgresql_server",
		RFunc: NewAzureRMPostrgreSQLServer,
	}
}

func NewAzureRMPostrgreSQLServer(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Region

	var costComponents []*schema.CostComponent
	serviceName := "Azure Database for PostgreSQL"

	sku := d.Get("sku_name").String()
	var tier, family, cores string
	if s := strings.Split(sku, "_"); len(s) == 3 {
		tier = strings.Split(sku, "_")[0]
		family = strings.Split(sku, "_")[1]
		cores = strings.Split(sku, "_")[2]
	} else {
		logging.Logger.Warn().Msgf("Unrecognised PostgreSQL SKU format for resource %s: %s", d.Address, sku)
		return nil
	}

	tierName := map[string]string{
		"B":  "Basic",
		"GP": "General Purpose",
		"MO": "Memory Optimized",
	}[tier]

	if tierName == "" {
		logging.Logger.Warn().Msgf("Unrecognised PostgreSQL tier prefix for resource %s: %s", d.Address, tierName)
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
