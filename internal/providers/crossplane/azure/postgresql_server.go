package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	log "github.com/sirupsen/logrus"

	"github.com/shopspring/decimal"
)

// GetAzureRMPostgreSQLServerRegistryItem ...
func GetAzureRMPostgreSQLServerRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "database.azure.crossplane.io/PostgreSQLServer",
		RFunc: NewAzureRMPostrgreSQLServer,
	}
}

// NewAzureRMPostrgreSQLServer ...
// Reference: https://doc.crds.dev/github.com/crossplane/provider-azure/database.azure.crossplane.io/PostgreSQLServer/v1beta1@v0.16.1
func NewAzureRMPostrgreSQLServer(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{})

	var costComponents []*schema.CostComponent
	serviceName := "Azure Database for PostgreSQL"
	forProvider := d.Get("forProvider")
	skuObj := forProvider.Get("sku")
	tier := skuObj.Get("tier").String()
	family := skuObj.Get("family").String()
	capacity := skuObj.Get("capacity").String()

	tierName := map[string]string{
		"Basic":           "Basic",
		"GeneralPurpose":  "General Purpose",
		"MemoryOptimized": "Memory Optimized",
	}[tier]

	if tierName == "" {
		log.Warnf("Unrecognised PostgreSQL tier for resource %s: %v", d.Address, skuObj)
		return nil
	}

	productNameRegex := fmt.Sprintf("/%s - Compute %s/", tierName, family)
	skuName := fmt.Sprintf("%s vCore", capacity)

	costComponents = append(costComponents, databaseComputeInstance(region, fmt.Sprintf("Compute (%s)", tier+"_"+family+"_"+capacity), serviceName, productNameRegex, skuName))

	storageProfile := forProvider.Get("storageProfile")

	storageGB := storageProfile.Get("storageMB").Int() / 1024

	// MemoryOptimized and GeneralPurpose storage cost are the same, and we don't have cost component for MemoryOptimized Storage now
	if strings.ToLower(tierName) == "memoryoptimized" {
		tier = "General Purpose"
	}
	productNameRegex = fmt.Sprintf("/%s - Storage/", tierName)

	costComponents = append(costComponents, databaseStorageComponent(region, serviceName, productNameRegex, storageGB))

	//This option is not available
	var backupStorageGB *decimal.Decimal
	if u != nil && u.Get("additional_backup_storage_gb").Exists() {
		backupStorageGB = decimalPtr(decimal.NewFromInt(u.Get("additional_backup_storage_gb").Int()))
	}

	skuName = "Backup LRS"
	if storageProfile.Get("geoRedundantBackup").Exists() && storageProfile.Get("geoRedundantBackup").String() == "Enabled" {
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
