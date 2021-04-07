package google

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

type SQLInstanceDBType int

const (
	MySQL SQLInstanceDBType = iota
	PostgreSQL
	SQLServer
)

func GetSQLInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "google_sql_database_instance",
		RFunc:               NewSQLInstance,
		ReferenceAttributes: []string{},
	}
}

func NewSQLInstance(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	name := d.Address
	tier := d.Get("settings.0").Get("tier").String()

	availabilityType := "ZONAL"
	if d.Get("settings.0").Get("availability_type").Exists() {
		availabilityType = d.Get("settings.0").Get("availability_type").String()
	}

	region := d.Get("region").String()
	dbVersion := d.Get("database_version").String()
	dbType := SQLInstanceDBVersionToDBType(dbVersion)

	diskType := "PD_SSD"
	if d.Get("settings.0").Get("disk_type").Exists() {
		diskType = d.Get("settings.0").Get("disk_type").String()
	}

	var diskSizeGB int64 = 10
	if d.Get("settings.0").Get("disk_size").Exists() {
		diskSizeGB = d.Get("settings.0").Get("disk_size").Int()
	}

	resource := &schema.Resource{
		Name:           name,
		CostComponents: []*schema.CostComponent{},
	}

	var vCPU *decimal.Decimal
	if SQLInstanceTierToResourceGroup(tier) == "" {
		cpu, _ := strconv.ParseInt(strings.Split(tier, "-")[2], 10, 32)

		vCPU = decimalPtr(decimal.NewFromInt32(int32(cpu)))
		resource.CostComponents = append(resource.CostComponents, cpuCostComponent(region, tier, availabilityType, dbType, vCPU))
	}

	var memory *decimal.Decimal
	if SQLInstanceTierToResourceGroup(tier) == "" {
		ram, _ := strconv.ParseInt(strings.Split(tier, "-")[3], 10, 32)

		memory = decimalPtr(decimal.NewFromInt32(int32(ram)).Div(decimal.NewFromInt(1024)))
		resource.CostComponents = append(resource.CostComponents, memoryCostComponent(region, tier, availabilityType, dbType, memory))
	}

	resource.CostComponents = append(resource.CostComponents, SQLInstanceStorage(region, dbType, availabilityType, diskType, diskSizeGB))
	var backupGB *decimal.Decimal
	if u != nil && u.Get("monthly_backup_gb").Exists() {
		backupGB = decimalPtr(decimal.NewFromInt(u.Get("monthly_backup_gb").Int()))
		resource.CostComponents = append(resource.CostComponents, backupCostComponent(region, backupGB))
	}

	if SQLInstanceTierToResourceGroup(tier) != "" {
		resource.CostComponents = append(resource.CostComponents, SharedSQLInstance(tier, availabilityType, dbType, region))
	} else {
		resource.CostComponents = append(resource.CostComponents, customSQLInstance(availabilityType, dbType, region, vCPU, memory))
	}

	if strings.Contains(dbVersion, "SQLSERVER") {
		resource.CostComponents = append(resource.CostComponents, SQLServerLicense(tier, dbVersion))
	}

	return resource
}

func memoryCostComponent(region string, tier string, availabilityType string, dbType SQLInstanceDBType, memory *decimal.Decimal) *schema.CostComponent {
	availabilityType = availabilityTypeDescName(availabilityType)
	dbTypeName := SQLInstanceTypeToDescriptionName(dbType)
	description := fmt.Sprintf("/%s: %s - RAM/", dbTypeName, availabilityType)

	return &schema.CostComponent{
		Name:           "Memory",
		Unit:           "GB",
		UnitMultiplier: 1,
		HourlyQuantity: memory,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(region),
			Service:       strPtr("Cloud SQL"),
			ProductFamily: strPtr("ApplicationServices"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", ValueRegex: strPtr(description)},
			},
		},
	}
}

func cpuCostComponent(region string, tier string, availabilityType string, dbType SQLInstanceDBType, vCPU *decimal.Decimal) *schema.CostComponent {
	availabilityType = availabilityTypeDescName(availabilityType)
	dbTypeName := SQLInstanceTypeToDescriptionName(dbType)
	description := fmt.Sprintf("/%s: %s - vCPU/", dbTypeName, availabilityType)

	return &schema.CostComponent{
		Name:           "CPU Credits",
		Unit:           "vCPU-hours",
		UnitMultiplier: 1,
		HourlyQuantity: vCPU,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(region),
			Service:       strPtr("Cloud SQL"),
			ProductFamily: strPtr("ApplicationServices"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", ValueRegex: strPtr(description)},
			},
		},
	}
}

func SharedSQLInstance(tier, availabilityType string, dbType SQLInstanceDBType, region string) *schema.CostComponent {
	dbName := SQLInstanceTypeToDescriptionName(dbType)
	resourceGroup := SQLInstanceTierToResourceGroup(tier)
	descriptionRegex := SQLInstanceAvDBTypeToDescriptionRegex(availabilityType, dbType)

	return &schema.CostComponent{
		Name:           fmt.Sprintf("SQL instance (%s, %s)", dbName, tier),
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("gcp"),
			Region:     strPtr(region),
			Service:    strPtr("Cloud SQL"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "resourceGroup", Value: strPtr(resourceGroup)},
				{Key: "description", ValueRegex: strPtr(descriptionRegex)},
			},
		},
	}
}

func customSQLInstance(availabilityType string, dbType SQLInstanceDBType, region string, vCPU *decimal.Decimal, memory *decimal.Decimal) *schema.CostComponent {
	dbName := SQLInstanceTypeToDescriptionName(dbType)
	descriptionRegex := SQLCustomInstanceDescriptionRegex(availabilityType, dbType, vCPU, memory)

	return &schema.CostComponent{
		Name:           fmt.Sprintf("SQL instance (%s, custom)", dbName),
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("gcp"),
			Region:     strPtr(region),
			Service:    strPtr("Cloud SQL"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", ValueRegex: strPtr(descriptionRegex)},
			},
		},
	}
}

func SQLCustomInstanceDescriptionRegex(availabilityType string, dbType SQLInstanceDBType, vCPU *decimal.Decimal, memory *decimal.Decimal) string {
	dbTypeString := SQLInstanceTypeToDescriptionName(dbType)
	availabilityTypeString := availabilityTypeDescName(availabilityType)

	descriptionRegex := fmt.Sprintf("/%s: %s - %s vCPU %s %sGB RAM/", dbTypeString, availabilityTypeString, vCPU, `\+\`, memory)
	return descriptionRegex
}

func SQLInstanceDBVersionToDBType(dbVersion string) SQLInstanceDBType {
	if strings.Contains(dbVersion, "POSTGRES") {
		return PostgreSQL
	} else if strings.Contains(dbVersion, "MYSQL") {
		return MySQL
	} else if strings.Contains(dbVersion, "SQLSERVER") {
		return SQLServer
	} else {
		return MySQL
	}
}

func SQLInstanceTierToResourceGroup(tier string) string {
	data := map[string]string{
		"db-f1-micro": "SQLGen2InstancesF1Micro",
		"db-g1-small": "SQLGen2InstancesG1Small",
	}

	return data[tier]
}

func SQLInstanceTypeToDescriptionName(dbType SQLInstanceDBType) string {
	dbTypeNames := map[SQLInstanceDBType]string{
		MySQL:      "MySQL",
		PostgreSQL: "PostgreSQL",
		SQLServer:  "SQL Server",
	}

	return dbTypeNames[dbType]
}

func availabilityTypeDescName(availabilityType string) string {
	availabilityTypeNames := map[string]string{
		"REGIONAL": "Regional",
		"ZONAL":    "Zonal",
	}

	return availabilityTypeNames[availabilityType]
}

func SQLInstanceAvDBTypeToDescriptionRegex(availabilityType string, dbType SQLInstanceDBType) string {
	dbTypeString := SQLInstanceTypeToDescriptionName(dbType)
	availabilityTypeString := availabilityTypeDescName(availabilityType)
	description := fmt.Sprintf("/%s: %s/", dbTypeString, availabilityTypeString)

	return description
}

func SQLServerLicense(tier string, dbVersion string) *schema.CostComponent {
	licenseType := SQLServerDBVersionToLicenseType(dbVersion)

	isSharedInstance := false
	sharedInstanceNames := []string{"db-f1-micro", "db-g1-small"}

	for _, tierName := range sharedInstanceNames {
		if tier == tierName {
			isSharedInstance = true
			break
		}
	}

	descriptionRegex := SQLServerTierNameToDescriptionRegex(tier, licenseType, isSharedInstance)

	cost := &schema.CostComponent{
		Name:           fmt.Sprintf("License (%s)", licenseType),
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("gcp"),
			Region:     strPtr("global"),
			Service:    strPtr("Compute Engine"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "resourceGroup", Value: strPtr("Google")},
				{Key: "description", ValueRegex: strPtr(descriptionRegex)},
			},
		},
	}

	return cost
}

func SQLServerDBVersionToLicenseType(dbVersion string) string {
	licenseType := "Standard"
	if strings.Contains(dbVersion, "ENTERPRISE") {
		licenseType = "Enterprise"
	} else if strings.Contains(dbVersion, "WEB") {
		licenseType = "Web"
	} else if strings.Contains(dbVersion, "EXPRESS") {
		licenseType = "Express"
	}
	return licenseType
}

func SQLServerTierNameToDescriptionRegex(tier, licenseType string, isSharedInstance bool) string {
	var descriptionRegex string
	if isSharedInstance {
		instanceAPINames := map[string]string{
			"db-f1-micro": "f1-micro",
			"db-g1-small": "g1-small",
		}

		descriptionRegex = fmt.Sprintf("/Licensing Fee for SQL Server 2017 %s on %s/", licenseType, instanceAPINames[tier])
	}

	return descriptionRegex
}

func SQLInstanceStorage(region string, dbType SQLInstanceDBType, availabilityType, diskType string, diskSizeGB int64) *schema.CostComponent {
	diskTypeHumanReadableNames := map[string]string{
		"PD_SSD": "SSD",
		"PD_HDD": "HDD",
	}

	diskTypeAPIResourceGroup := map[string]string{
		"PD_SSD": "SSD",
		"PD_HDD": "PDStandard",
	}

	cost := &schema.CostComponent{
		Name:            fmt.Sprintf("Storage (%s)", diskTypeHumanReadableNames[diskType]),
		Unit:            "GB-months",
		UnitMultiplier:  1,
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(diskSizeGB)),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("gcp"),
			Region:     strPtr(region),
			Service:    strPtr("Cloud SQL"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "resourceGroup", Value: strPtr(diskTypeAPIResourceGroup[diskType])},
				{Key: "description", ValueRegex: strPtr(fmt.Sprintf("/%s: %s/", SQLInstanceTypeToDescriptionName(dbType), availabilityTypeDescName(availabilityType)))},
			},
		},
	}

	return cost
}

func backupCostComponent(region string, quantity *decimal.Decimal) *schema.CostComponent {
	cost := &schema.CostComponent{
		Name:            "Backups",
		Unit:            "GB-months",
		UnitMultiplier:  1,
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("gcp"),
			Region:     strPtr(region),
			Service:    strPtr("Cloud SQL"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "resourceGroup", Value: strPtr("PDSnapshot")},
				{Key: "description", ValueRegex: strPtr("/Cloud SQL: Backups/")},
			},
		},
	}

	return cost
}
