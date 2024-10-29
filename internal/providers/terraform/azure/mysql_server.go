package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

func GetAzureRMMySQLServerRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_mysql_server",
		RFunc: NewAzureRMMySQLServer,
	}
}

func NewAzureRMMySQLServer(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Region

	serviceName := "Azure Database for MySQL"
	var costComponents []*schema.CostComponent

	sku := d.Get("sku_name").String()
	var tier, family, cores string
	if s := strings.Split(sku, "_"); len(s) == 3 {
		tier = strings.Split(sku, "_")[0]
		family = strings.Split(sku, "_")[1]
		cores = strings.Split(sku, "_")[2]
	} else {
		logging.Logger.Warn().Msgf("Unrecognised MySQL SKU format for resource %s: %s", d.Address, sku)
		return nil
	}

	tierName := map[string]string{
		"B":  "Basic",
		"GP": "General Purpose",
		"MO": "Memory Optimized",
	}[tier]

	if tierName == "" {
		logging.Logger.Warn().Msgf("Unrecognised MySQL tier prefix for resource %s: %s", d.Address, tierName)
		return nil
	}

	productNameRegex := fmt.Sprintf("/%s - Compute %s/", tierName, family)
	skuName := fmt.Sprintf("%s vCore", cores)

	name := fmt.Sprintf("Compute (%s)", sku)
	logging.Logger.Debug().Msgf("'Multiple products found' are safe to ignore for '%s' due to limitations in the Azure API.", name)

	costComponents = append(costComponents, databaseComputeInstance(region, name, serviceName, productNameRegex, skuName))

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
