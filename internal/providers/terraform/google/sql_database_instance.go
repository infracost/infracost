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
		Name:  "google_sql_database_instance",
		RFunc: NewSQLInstance,
		Notes: []string{
			"Cloud SQL network costs are not yet supported.",
		},
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
	dbType := sqlInstanceDBVersionToDBType(dbVersion)

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
	if sqlInstanceTierToResourceGroup(tier) == "" {
		cpu, _ := strconv.ParseInt(strings.Split(tier, "-")[2], 10, 32)

		vCPU = decimalPtr(decimal.NewFromInt32(int32(cpu)))
		resource.CostComponents = append(resource.CostComponents, cpuCostComponent(region, tier, availabilityType, dbType, vCPU))
	}

	var memory *decimal.Decimal
	if sqlInstanceTierToResourceGroup(tier) == "" {
		ram, _ := strconv.ParseInt(strings.Split(tier, "-")[3], 10, 32)

		memory = decimalPtr(decimal.NewFromInt32(int32(ram)).Div(decimal.NewFromInt(1024)))
		resource.CostComponents = append(resource.CostComponents, memoryCostComponent(region, tier, availabilityType, dbType, memory))
	}

	resource.CostComponents = append(resource.CostComponents, sqlInstanceStorage(region, dbType, availabilityType, diskType, diskSizeGB))
	var backupGB *decimal.Decimal
	if u != nil && u.Get("backup_storage_gb").Exists() {
		backupGB = decimalPtr(decimal.NewFromInt(u.Get("backup_storage_gb").Int()))
		resource.CostComponents = append(resource.CostComponents, backupCostComponent(region, backupGB))
	}

	if d.Get("settings.0").Get("ip_configuration.0").Get("ipv4_enabled").Exists() {
		resource.CostComponents = append(resource.CostComponents, ipv4CostComponent())
	}

	if sqlInstanceTierToResourceGroup(tier) != "" {
		resource.CostComponents = append(resource.CostComponents, sharedSQLInstance(tier, availabilityType, dbType, region))
	} else {
		resource.CostComponents = append(resource.CostComponents, customSQLInstance(tier, availabilityType, dbType, region, vCPU, memory))
	}

	if strings.Contains(dbVersion, "SQLSERVER") {
		resource.CostComponents = append(resource.CostComponents, sqlServerLicense(tier, dbVersion, vCPU))
	}

	return resource
}

func memoryCostComponent(region string, tier string, availabilityType string, dbType SQLInstanceDBType, memory *decimal.Decimal) *schema.CostComponent {
	availabilityType = availabilityTypeDescName(availabilityType)
	dbTypeName := sqlInstanceTypeToDescriptionName(dbType)
	description := fmt.Sprintf("/%s: %s - RAM/", dbTypeName, availabilityType)

	return &schema.CostComponent{
		Name:           fmt.Sprintf("Memory (%s)", strings.ToLower(availabilityType)),
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
	dbTypeName := sqlInstanceTypeToDescriptionName(dbType)
	description := fmt.Sprintf("/%s: %s - vCPU/", dbTypeName, availabilityType)

	return &schema.CostComponent{
		Name:           fmt.Sprintf("vCPUs (%s)", strings.ToLower(availabilityType)),
		Unit:           "hours",
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

func sharedSQLInstance(tier, availabilityType string, dbType SQLInstanceDBType, region string) *schema.CostComponent {
	resourceGroup := sqlInstanceTierToResourceGroup(tier)
	descriptionRegex := sqlInstanceAvDBTypeToDescriptionRegex(availabilityType, dbType)

	return &schema.CostComponent{
		Name:           fmt.Sprintf("SQL instance (%s, %s)", tier, strings.ToLower(availabilityType)),
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

func customSQLInstance(tier, availabilityType string, dbType SQLInstanceDBType, region string, vCPU *decimal.Decimal, memory *decimal.Decimal) *schema.CostComponent {
	descriptionRegex := SQLCustomInstanceDescriptionRegex(availabilityType, dbType, vCPU, memory)

	return &schema.CostComponent{
		Name:           fmt.Sprintf("SQL instance (%s, %s)", tier, strings.ToLower(availabilityType)),
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
	dbTypeString := sqlInstanceTypeToDescriptionName(dbType)
	availabilityTypeString := availabilityTypeDescName(availabilityType)

	descriptionRegex := fmt.Sprintf("/%s: %s - %s vCPU %s %sGB RAM/", dbTypeString, availabilityTypeString, vCPU, `\+\`, memory)
	return descriptionRegex
}

func sqlInstanceDBVersionToDBType(dbVersion string) SQLInstanceDBType {
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

func sqlInstanceTierToResourceGroup(tier string) string {
	data := map[string]string{
		"db-f1-micro": "SQLGen2InstancesF1Micro",
		"db-g1-small": "SQLGen2InstancesG1Small",
	}

	return data[tier]
}

func sqlInstanceTypeToDescriptionName(dbType SQLInstanceDBType) string {
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

func sqlInstanceAvDBTypeToDescriptionRegex(availabilityType string, dbType SQLInstanceDBType) string {
	dbTypeString := sqlInstanceTypeToDescriptionName(dbType)
	availabilityTypeString := availabilityTypeDescName(availabilityType)
	description := fmt.Sprintf("/%s: %s/", dbTypeString, availabilityTypeString)

	return description
}

func sqlServerLicense(tier string, dbVersion string, vCPUs *decimal.Decimal) *schema.CostComponent {
	licenseType := sqlServerDBVersionToLicenseType(dbVersion)

	isSharedInstance := false
	sharedInstanceNames := []string{"db-f1-micro", "db-g1-small"}

	for _, tierName := range sharedInstanceNames {
		if tier == tierName {
			isSharedInstance = true
			break
		}
	}

	descriptionRegex := sqlServerTierNameToDescriptionRegex(tier, licenseType, isSharedInstance, vCPUs)

	return sqlServerCostComponent(vCPUs, descriptionRegex, licenseType)
}

func sqlServerCostComponent(vCPUs *decimal.Decimal, descriptionRegex, licenseType string) *schema.CostComponent {
	multiplier := decimal.NewFromInt(1)
	if vCPUs.LessThan(decimal.NewFromInt(4)) {
		multiplier = decimal.NewFromInt(4)
	}

	return &schema.CostComponent{
		Name:           fmt.Sprintf("License (%s)", licenseType),
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(multiplier),
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
}

func sqlServerDBVersionToLicenseType(dbVersion string) string {
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

func sqlServerTierNameToDescriptionRegex(tier, licenseType string, isSharedInstance bool, vCPUs *decimal.Decimal) string {
	var descriptionRegex string
	if isSharedInstance {
		instanceAPINames := map[string]string{
			"db-f1-micro": "f1-micro",
			"db-g1-small": "g1-small",
		}

		descriptionRegex = fmt.Sprintf("/Licensing Fee for SQL Server 2017 %s on %s/", licenseType, instanceAPINames[tier])
	} else {
		descriptionRegex = fmt.Sprintf("/Licensing Fee for SQL Server 2017 Standard on VM with %s/", vCPUs)
	}

	return descriptionRegex
}

func sqlInstanceStorage(region string, dbType SQLInstanceDBType, availabilityType, diskType string, diskSizeGB int64) *schema.CostComponent {
	diskTypeHumanReadableNames := map[string]string{
		"PD_SSD": "SSD",
		"PD_HDD": "HDD",
	}

	diskTypeAPIResourceGroup := map[string]string{
		"PD_SSD": "SSD",
		"PD_HDD": "PDStandard",
	}

	cost := &schema.CostComponent{
		Name:            fmt.Sprintf("Storage (%s, %s)", diskTypeHumanReadableNames[diskType], strings.ToLower(availabilityType)),
		Unit:            "GB-months",
		UnitMultiplier:  1,
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(diskSizeGB)),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("gcp"),
			Region:     strPtr(region),
			Service:    strPtr("Cloud SQL"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "resourceGroup", Value: strPtr(diskTypeAPIResourceGroup[diskType])},
				{Key: "description", ValueRegex: strPtr(fmt.Sprintf("/%s: %s/", sqlInstanceTypeToDescriptionName(dbType), availabilityTypeDescName(availabilityType)))},
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

func ipv4CostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "IP address (if unused)",
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("gcp"),
			Region:     strPtr("global"),
			Service:    strPtr("Cloud SQL"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "resourceGroup", Value: strPtr("IpAddress")},
				{Key: "description", ValueRegex: strPtr("/IP address idling - hour/")},
			},
		},
	}
}
