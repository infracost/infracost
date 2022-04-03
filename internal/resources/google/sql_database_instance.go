package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

type SQLDatabaseInstance struct {
	Address              string
	DiskSize             int64
	UseIPV4              bool
	ReplicaConfiguration string
	Tier                 string
	AvailabilityType     string
	Region               string
	DatabaseVersion      string
	DiskType             string
	BackupStorageGB      *float64 `infracost_usage:"backup_storage_gb"`
}

var SQLDatabaseInstanceUsageSchema = []*schema.UsageItem{{Key: "backup_storage_gb", ValueType: schema.Float64, DefaultValue: 0}}

func (r *SQLDatabaseInstance) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *SQLDatabaseInstance) BuildResource() *schema.Resource {
	var resource *schema.Resource

	replica := false
	if r.ReplicaConfiguration != "" {
		replica = true
	}

	resource = sqlDatabaseInstanceCostComponents(r, false, r.Address)
	if replica {
		resource.SubResources = append(resource.SubResources, sqlDatabaseInstanceCostComponents(r, true, "Replica"))
	}

	return resource
}

type SQLInstanceDBType int

const (
	MySQL SQLInstanceDBType = iota
	PostgreSQL
	SQLServer
)

func sqlDatabaseInstanceCostComponents(r *SQLDatabaseInstance, replica bool, name string) *schema.Resource {
	var costComponents []*schema.CostComponent
	tier := r.Tier

	availabilityType := "ZONAL"
	if r.AvailabilityType != "" && !replica {
		availabilityType = r.AvailabilityType
	}

	region := r.Region
	dbVersion := r.DatabaseVersion
	dbType := sqlInstanceDBVersionToDBType(dbVersion)

	diskType := "PD_SSD"
	if r.DiskType != "" {
		diskType = r.DiskType
	}

	var diskSizeGB int64 = 10
	if r.DiskSize != 0 {
		diskSizeGB = r.DiskSize
	}

	if sqlInstanceTierToResourceGroup(tier) != "" && dbType != SQLServer {
		costComponents = append(costComponents, sharedSQLInstance(tier, availabilityType, dbType, region))
	} else if sqlInstanceTierToResourceGroup(tier) == "" && strings.Contains(tier, "db-custom-") {
		cpu, _ := strconv.ParseInt(strings.Split(tier, "-")[2], 10, 32)
		vCPU := decimalPtr(decimal.NewFromInt32(int32(cpu)))

		costComponents = append(costComponents, cpuCostComponent(region, tier, availabilityType, dbType, vCPU))

		ram, _ := strconv.ParseInt(strings.Split(tier, "-")[3], 10, 32)
		memory := decimalPtr(decimal.NewFromInt32(int32(ram)).Div(decimal.NewFromInt(1024)))

		costComponents = append(costComponents, memoryCostComponent(region, tier, availabilityType, dbType, memory))
	} else if strings.Contains(tier, "db-n1-") && dbType == MySQL {
		costComponents = append(costComponents, sharedSQLInstance(tier, availabilityType, dbType, region))
	}

	costComponents = append(costComponents, sqlInstanceStorage(region, dbType, availabilityType, diskType, diskSizeGB))

	if !replica {
		var backupGB *decimal.Decimal
		if r.BackupStorageGB != nil {
			backupGB = decimalPtr(decimal.NewFromFloat(*r.BackupStorageGB))
		}
		costComponents = append(costComponents, backupCostComponent(region, backupGB))

		if r.UseIPV4 {
			costComponents = append(costComponents, ipv4CostComponent())
		}
	}

	return &schema.Resource{
		Name:           name,
		CostComponents: costComponents,
	}
}

func memoryCostComponent(region string, tier string, availabilityType string, dbType SQLInstanceDBType, memory *decimal.Decimal) *schema.CostComponent {
	availabilityType = availabilityTypeDescName(availabilityType)
	dbTypeName := sqlInstanceTypeToDescriptionName(dbType)
	description := fmt.Sprintf("/%s: %s - RAM/", dbTypeName, availabilityType)

	return &schema.CostComponent{
		Name:           fmt.Sprintf("Memory (%s)", strings.ToLower(availabilityType)),
		Unit:           "GB",
		UnitMultiplier: decimal.NewFromInt(1),
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
		UnitMultiplier: decimal.NewFromInt(1),
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
	descriptionRegex := "/" + sqlInstanceAvDBTypeToDescription(availabilityType, dbType)

	var vCPU string
	if strings.Contains(tier, "db-n1-standard") || strings.Contains(tier, "db-n1-highmem") {
		vCPU = (strings.Split(tier, "-")[3])
		descriptionRegex += " - " + vCPU + "/"
	} else {
		descriptionRegex += "/"
	}

	return &schema.CostComponent{
		Name:           fmt.Sprintf("SQL instance (%s, %s)", tier, strings.ToLower(availabilityType)),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
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

	if data[tier] != "" {
		return data[tier]
	} else if strings.Contains(tier, "db-n1-standard") {
		return "SQLGen2InstancesN1Standard"
	} else if strings.Contains(tier, "db-n1-highmem") {
		return "SQLGen2InstancesN1Highmem"
	}

	return ""
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

func sqlInstanceAvDBTypeToDescription(availabilityType string, dbType SQLInstanceDBType) string {
	dbTypeString := sqlInstanceTypeToDescriptionName(dbType)
	availabilityTypeString := availabilityTypeDescName(availabilityType)

	description := fmt.Sprintf("%s: %s", dbTypeString, availabilityTypeString)

	return description
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

	if dbType == SQLServer {
		diskType = "PD_SSD"
	}

	cost := &schema.CostComponent{
		Name:            fmt.Sprintf("Storage (%s, %s)", diskTypeHumanReadableNames[diskType], strings.ToLower(availabilityType)),
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
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
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
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
		UnitMultiplier: decimal.NewFromInt(1),
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
