package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"

	log "github.com/sirupsen/logrus"
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

	resource = r.sqlDatabaseInstanceCostComponents(false, r.Address)
	if replica {
		resource.SubResources = append(resource.SubResources, r.sqlDatabaseInstanceCostComponents(true, "Replica"))
	}

	return resource
}

type SQLInstanceDBType int

const (
	MySQL SQLInstanceDBType = iota
	PostgreSQL
	SQLServer
)

func (r *SQLDatabaseInstance) dbType() SQLInstanceDBType {
	return r.sqlInstanceDBVersionToDBType()
}

func (r *SQLDatabaseInstance) sqlDatabaseInstanceCostComponents(replica bool, name string) *schema.Resource {
	var costComponents []*schema.CostComponent
	tier := r.Tier

	if r.AvailabilityType == "" || replica {
		r.AvailabilityType = "ZONAL"
	}

	if r.DiskType == "" {
		r.DiskType = "PD_SSD"
	}

	if r.DiskSize == 0 {
		r.DiskSize = 10
	}

	if r.sqlInstanceTierToResourceGroup() != "" && r.dbType() != SQLServer {
		costComponents = append(costComponents, r.sharedSQLInstance())
	} else if r.sqlInstanceTierToResourceGroup() == "" && strings.Contains(tier, "db-custom-") {
		splittedTier := strings.Split(tier, "-")
		cpu, err := strconv.ParseInt(splittedTier[2], 10, 32)
		if err != nil {
			log.Warnf("cpu of tier %s of %s is not parsable", tier, r.Address)
			return nil
		}
		vCPU := decimalPtr(decimal.NewFromInt32(int32(cpu)))

		costComponents = append(costComponents, r.cpuCostComponent(vCPU))

		if len(splittedTier) < 4 {
			log.Warnf("tier %s of %s has no ram data", tier, r.Address)
			return nil
		}
		ram, err := strconv.ParseInt(splittedTier[3], 10, 32)
		if err != nil {
			log.Warnf("ram of tier %s of %s is not parsable", tier, r.Address)
			return nil
		}
		memory := decimalPtr(decimal.NewFromInt32(int32(ram)).Div(decimal.NewFromInt(1024)))

		costComponents = append(costComponents, r.memoryCostComponent(memory))
	} else if strings.Contains(tier, "db-n1-") && r.dbType() == MySQL {
		costComponents = append(costComponents, r.sharedSQLInstance())
	}

	costComponents = append(costComponents, r.sqlInstanceStorage())

	if !replica {
		var backupGB *decimal.Decimal
		if r.BackupStorageGB != nil {
			backupGB = decimalPtr(decimal.NewFromFloat(*r.BackupStorageGB))
		}
		costComponents = append(costComponents, r.backupCostComponent(backupGB))

		if r.UseIPV4 {
			costComponents = append(costComponents, r.ipv4CostComponent())
		}
	}

	return &schema.Resource{
		Name:           name,
		CostComponents: costComponents,
	}
}

func (r *SQLDatabaseInstance) memoryCostComponent(memory *decimal.Decimal) *schema.CostComponent {
	availabilityType := r.availabilityTypeDescName()
	dbTypeName := r.sqlInstanceTypeToDescriptionName()
	description := fmt.Sprintf("/%s: %s - RAM/", dbTypeName, availabilityType)

	return &schema.CostComponent{
		Name:           fmt.Sprintf("Memory (%s)", strings.ToLower(availabilityType)),
		Unit:           "GB",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: memory,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Cloud SQL"),
			ProductFamily: strPtr("ApplicationServices"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", ValueRegex: strPtr(description)},
			},
		},
	}
}

func (r *SQLDatabaseInstance) cpuCostComponent(vCPU *decimal.Decimal) *schema.CostComponent {
	availabilityType := r.availabilityTypeDescName()
	dbTypeName := r.sqlInstanceTypeToDescriptionName()
	description := fmt.Sprintf("/%s: %s - vCPU/", dbTypeName, availabilityType)

	return &schema.CostComponent{
		Name:           fmt.Sprintf("vCPUs (%s)", strings.ToLower(availabilityType)),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: vCPU,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Cloud SQL"),
			ProductFamily: strPtr("ApplicationServices"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", ValueRegex: strPtr(description)},
			},
		},
	}
}

func (r *SQLDatabaseInstance) sharedSQLInstance() *schema.CostComponent {
	resourceGroup := r.sqlInstanceTierToResourceGroup()
	descriptionRegex := "/" + r.sqlInstanceAvDBTypeToDescription()

	var vCPU string
	if strings.Contains(r.Tier, "db-n1-standard") || strings.Contains(r.Tier, "db-n1-highmem") {
		vCPU = (strings.Split(r.Tier, "-")[3])
		descriptionRegex += " - " + vCPU + "/"
	} else {
		descriptionRegex += "/"
	}

	return &schema.CostComponent{
		Name:           fmt.Sprintf("SQL instance (%s, %s)", r.Tier, strings.ToLower(r.AvailabilityType)),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("gcp"),
			Region:     strPtr(r.Region),
			Service:    strPtr("Cloud SQL"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "resourceGroup", Value: strPtr(resourceGroup)},
				{Key: "description", ValueRegex: strPtr(descriptionRegex)},
			},
		},
	}
}

func (r *SQLDatabaseInstance) sqlInstanceDBVersionToDBType() SQLInstanceDBType {
	if strings.Contains(r.DatabaseVersion, "POSTGRES") {
		return PostgreSQL
	} else if strings.Contains(r.DatabaseVersion, "MYSQL") {
		return MySQL
	} else if strings.Contains(r.DatabaseVersion, "SQLSERVER") {
		return SQLServer
	} else {
		return MySQL
	}
}

func (r *SQLDatabaseInstance) sqlInstanceTierToResourceGroup() string {
	data := map[string]string{
		"db-f1-micro": "SQLGen2InstancesF1Micro",
		"db-g1-small": "SQLGen2InstancesG1Small",
	}

	if data[r.Tier] != "" {
		return data[r.Tier]
	} else if strings.Contains(r.Tier, "db-n1-standard") {
		return "SQLGen2InstancesN1Standard"
	} else if strings.Contains(r.Tier, "db-n1-highmem") {
		return "SQLGen2InstancesN1Highmem"
	}

	return ""
}

func (r *SQLDatabaseInstance) sqlInstanceTypeToDescriptionName() string {
	dbTypeNames := map[SQLInstanceDBType]string{
		MySQL:      "MySQL",
		PostgreSQL: "PostgreSQL",
		SQLServer:  "SQL Server",
	}

	return dbTypeNames[r.dbType()]
}

func (r *SQLDatabaseInstance) availabilityTypeDescName() string {
	availabilityTypeNames := map[string]string{
		"REGIONAL": "Regional",
		"ZONAL":    "Zonal",
	}

	return availabilityTypeNames[r.AvailabilityType]
}

func (r *SQLDatabaseInstance) sqlInstanceAvDBTypeToDescription() string {
	dbTypeString := r.sqlInstanceTypeToDescriptionName()
	availabilityTypeString := r.availabilityTypeDescName()

	description := fmt.Sprintf("%s: %s", dbTypeString, availabilityTypeString)

	return description
}

func (r *SQLDatabaseInstance) sqlInstanceStorage() *schema.CostComponent {
	diskType := r.DiskType
	diskTypeHumanReadableNames := map[string]string{
		"PD_SSD": "SSD",
		"PD_HDD": "HDD",
	}

	diskTypeAPIResourceGroup := map[string]string{
		"PD_SSD": "SSD",
		"PD_HDD": "PDStandard",
	}

	if r.dbType() == SQLServer {
		diskType = "PD_SSD"
	}

	cost := &schema.CostComponent{
		Name:            fmt.Sprintf("Storage (%s, %s)", diskTypeHumanReadableNames[diskType], strings.ToLower(r.AvailabilityType)),
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(r.DiskSize)),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("gcp"),
			Region:     strPtr(r.Region),
			Service:    strPtr("Cloud SQL"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "resourceGroup", Value: strPtr(diskTypeAPIResourceGroup[diskType])},
				{Key: "description", ValueRegex: strPtr(fmt.Sprintf("/%s: %s/", r.sqlInstanceTypeToDescriptionName(), r.availabilityTypeDescName()))},
			},
		},
	}

	return cost
}

func (r *SQLDatabaseInstance) backupCostComponent(quantity *decimal.Decimal) *schema.CostComponent {
	cost := &schema.CostComponent{
		Name:            "Backups",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("gcp"),
			Region:     strPtr(r.Region),
			Service:    strPtr("Cloud SQL"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "resourceGroup", Value: strPtr("PDSnapshot")},
				{Key: "description", ValueRegex: strPtr("/Cloud SQL: Backups/")},
			},
		},
	}

	return cost
}

func (r *SQLDatabaseInstance) ipv4CostComponent() *schema.CostComponent {
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
