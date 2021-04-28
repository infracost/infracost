package azure

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

func GetAzureMSSQLDatabaseRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_mssql_database",
		RFunc: NewAzureMSSQLDatabase,
		ReferenceAttributes: []string{
			"server_id",
		},
	}
}

func NewAzureMSSQLDatabase(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var costComponents []*schema.CostComponent
	server := d.References("server_id")
	region := server[0].Get("location").String()

	serviceName := "SQL Database"
	var sku string
	if d.Get("sku_name").Type != gjson.Null {
		sku = d.Get("sku_name").String()
	}

	tier, family, cores, err := mssqlSkuSplit(d.Address, sku)
	if err != nil {
		log.Warnf(string(err.Error()))
		return nil
	}

	var zoneRedundant bool
	if d.Get("zone_redundant").Type != gjson.Null {
		zoneRedundant = d.Get("zone_redundant").Bool()
	}

	productNameRegex := mssqlProductName(tier, family)
	skuName := mssqlSkuName(cores, zoneRedundant)

	if tier == "GPS" {
		var vCoreHours *decimal.Decimal
		if u != nil && u.Get("vcore_hours").Exists() {
			vCoreHours = decimalPtr(decimal.NewFromInt(u.Get("vcore_hours").Int()))
		}

		serverlessSkuName := mssqlSkuName("1", zoneRedundant)
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            fmt.Sprintf("Compute (serverless, %s)", sku),
			Unit:            "vCore-hours",
			UnitMultiplier:  1,
			MonthlyQuantity: vCoreHours,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("azure"),
				Region:        strPtr(region),
				Service:       strPtr(serviceName),
				ProductFamily: strPtr("Databases"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", ValueRegex: strPtr(productNameRegex)},
					{Key: "skuName", Value: strPtr(serverlessSkuName)},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("Consumption"),
			},
		})
	} else {
		costComponents = append(costComponents, databaseComputeInstance(region, fmt.Sprintf("Compute (provisioned, %s)", sku), serviceName, productNameRegex, skuName))
	}

	if tier == "HS" {
		var replicaCount *decimal.Decimal
		if d.Get("read_replica_count").Type != gjson.Null {
			replicaCount = decimalPtr(decimal.NewFromInt(d.Get("read_replica_count").Int()))
		}
		costComponents = append(costComponents, &schema.CostComponent{
			Name:           "Read replicas",
			Unit:           "hours",
			UnitMultiplier: 1,
			HourlyQuantity: replicaCount,
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
		})
	}

	if tier == "GP" || tier == "HS" || tier == "BC" {
		var licenseType string
		if d.Get("license_type").Type != gjson.Null {
			licenseType = d.Get("license_type").String()
			if licenseType == "LicenseIncluded" {
				costComponents = append(costComponents, sqlLicenseCostComponent(region, cores, serviceName, tier))
			}
		}
	}

	storageGb := decimalPtr(decimal.NewFromInt(5))
	if d.Get("max_size_gb").Type != gjson.Null {
		storageGb = decimalPtr(decimal.NewFromInt(d.Get("max_size_gb").Int()))
	}
	costComponents = append(costComponents, mssqlStorageComponent(storageGb, region, serviceName, tier, zoneRedundant))

	var retention *decimal.Decimal
	if tier != "HS" {
		if u != nil && u.Get("monthly_long_term_retention_storage_gb").Exists() {
			retention = decimalPtr(decimal.NewFromInt(u.Get("monthly_long_term_retention_storage_gb").Int()))
		}
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Long-term retention",
			Unit:            "GB-months",
			UnitMultiplier:  1,
			MonthlyQuantity: retention,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("azure"),
				Region:        strPtr(region),
				Service:       strPtr(serviceName),
				ProductFamily: strPtr("Databases"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", ValueRegex: strPtr("/LTR Backup Storage/")},
					{Key: "skuName", Value: strPtr("Backup RA-GRS")},
					{Key: "meterName", Value: strPtr("RA-GRS Data Stored")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("Consumption"),
			},
		})
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func mssqlSkuSplit(address, sku string) (string, string, string, error) {
	if s := strings.Split(sku, "_"); len(s) == 4 && strings.HasPrefix(sku, "GP_S_Gen") {
		return s[0] + s[1], s[2], s[3], nil
	}
	if s := strings.Split(sku, "_"); len(s) == 3 {
		if strings.HasPrefix(sku, "HS_Gen") ||
			strings.HasPrefix(sku, "HS_DC") ||
			strings.HasPrefix(sku, "GP_Gen") ||
			strings.HasPrefix(sku, "GP_Fsv2") ||
			strings.HasPrefix(sku, "GP_DC") ||
			strings.HasPrefix(sku, "BC_Gen") ||
			strings.HasPrefix(sku, "BC_M") ||
			strings.HasPrefix(sku, "BC_DC") ||
			strings.HasPrefix(sku, "BC_DC") {
			return s[0], s[1], s[2], nil
		}
	}
	return "", "", "", errors.Errorf("Unrecognised MSSQL SKU format for resource %s: %s", address, sku)
}

func mssqlProductName(tier, family string) string {
	tierName := map[string]string{
		"GPS": "General Purpose - Serverless",
		"HS":  "Hyperscale",
		"GP":  "General Purpose",
		"BC":  "Business Critical",
	}[tier]

	familyName := map[string]string{
		"Gen5":    "Compute Gen5",
		"Gen4":    "Compute Gen4",
		"DC":      "Compute DC-Series",
		"Fsv2":    "Compute FSv2 Series",
		"M":       "Compute M Series",
		"Storage": "Storage",
		"License": "SQL License",
	}[family]

	return fmt.Sprintf("/%s - %s/", tierName, familyName)
}

func mssqlSkuName(cores string, zoneRedundant bool) string {
	sku := cores + " vCore"
	if zoneRedundant {
		sku += " Zone Redundancy"
	}
	return sku
}

func sqlLicenseCostComponent(region, cores, serviceName, tier string) *schema.CostComponent {
	licenseRegion := "Global"
	if strings.Contains(region, "usgov") {
		licenseRegion = "US Gov"
	}
	if strings.Contains(region, "china") {
		licenseRegion = "China"
	}
	if strings.Contains(region, "germany") {
		licenseRegion = "Germany"
	}

	coresNum, _ := strconv.ParseInt(cores, 10, 64)

	return &schema.CostComponent{
		Name:           "SQL License",
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(coresNum)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(licenseRegion),
			Service:       strPtr(serviceName),
			ProductFamily: strPtr("Databases"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", ValueRegex: strPtr(mssqlProductName(tier, "License"))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func mssqlStorageComponent(storageGB *decimal.Decimal, region, serviceName, tier string, zoneRedundant bool) *schema.CostComponent {
	tierName := map[string]string{
		"GPS": "General Purpose",
		"HS":  "Hyperscale",
		"GP":  "General Purpose",
		"BC":  "Business Critical",
	}[tier]

	skuName := tierName
	if zoneRedundant {
		skuName += " Zone Redundancy"
	}

	productNameRegex := fmt.Sprintf("/%s - Storage/", tierName)

	return &schema.CostComponent{
		Name:            "Storage",
		Unit:            "GB-months",
		UnitMultiplier:  1,
		MonthlyQuantity: storageGB,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr(serviceName),
			ProductFamily: strPtr("Databases"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", ValueRegex: strPtr(productNameRegex)},
				{Key: "skuName", Value: strPtr(skuName)},
				{Key: "meterName", Value: strPtr("Data Stored")},
			},
		},
	}
}
