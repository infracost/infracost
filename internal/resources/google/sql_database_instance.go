package google

import (
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type SQLDatabaseInstance struct {
	Address              string
	DiskSize             int64
	UseIPV4              bool
	ReplicaConfiguration string
	Tier                 string
	Edition              string
	AvailabilityType     string
	Region               string
	DatabaseVersion      string
	DiskType             string
	BackupStorageGB      *float64 `infracost_usage:"backup_storage_gb"`
}

func (r *SQLDatabaseInstance) CoreType() string {
	return "SQLDatabaseInstance"
}

func (r *SQLDatabaseInstance) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "backup_storage_gb", ValueType: schema.Float64, DefaultValue: 0},
	}
}

var lightweightRAM = decimal.NewFromInt(3840)
var standardRAMRatio = decimal.NewFromInt(3840)
var highmemRAMRatio = decimal.NewFromInt(6656)

func (r *SQLDatabaseInstance) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *SQLDatabaseInstance) BuildResource() *schema.Resource {
	var resource *schema.Resource

	if strings.EqualFold(r.Edition, "enterprise_plus") {
		logging.Logger.Warn().Msgf("edition %s of %s is not yet supported", r.Edition, r.Address)
		return nil
	}

	replica := false
	if r.ReplicaConfiguration != "" {
		replica = true
	}

	resource = r.costComponents(false)
	if replica {
		resource.SubResources = append(resource.SubResources, r.costComponents(true))
	}

	return resource
}

type SQLInstanceDBType int

const (
	MySQL SQLInstanceDBType = iota
	PostgreSQL
	SQLServer
)

func (r *SQLDatabaseInstance) costComponents(replica bool) *schema.Resource {
	var costComponents []*schema.CostComponent

	if r.AvailabilityType == "" || replica {
		r.AvailabilityType = "ZONAL"
	}

	if r.DiskType == "" {
		r.DiskType = "PD_SSD"
	}

	if r.DiskSize == 0 {
		r.DiskSize = 10
	}

	if r.IsShared() {
		costComponents = append(costComponents, r.sharedInstanceCostComponent())
	} else if r.IsLegacy() && r.dbType() == MySQL {
		costComponents = append(costComponents, r.legacyMySQLInstanceCostComponent())
	} else {
		costComponents = append(costComponents, r.instanceCostComponents()...)
	}

	costComponents = append(costComponents, r.storageCostComponent())

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

	name := r.Address
	if replica {
		name = "Replica"
	}

	return &schema.Resource{
		Name:           name,
		CostComponents: costComponents,
	}
}

func (r *SQLDatabaseInstance) sharedInstanceCostComponent() *schema.CostComponent {
	var resourceGroup string
	if strings.EqualFold(r.Tier, "db-f1-micro") {
		resourceGroup = "SQLGen2InstancesF1Micro"
	} else if strings.EqualFold(r.Tier, "db-g1-small") {
		resourceGroup = "SQLGen2InstancesG1Small"
	} else {
		logging.Logger.Warn().Msgf("tier %s of %s is not supported", r.Tier, r.Address)
		return nil
	}

	descriptionRegex := fmt.Sprintf("Cloud SQL for %s: %s(?! - Extended).*", r.dbTypeDescName(), r.availabilityTypeDescName())

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
				{Key: "description", ValueRegex: regexPtr(descriptionRegex)},
			},
		},
	}
}

func (r *SQLDatabaseInstance) legacyMySQLInstanceCostComponent() *schema.CostComponent {
	var resourceGroup string
	if r.IsStandard() {
		resourceGroup = "SQLGen2InstancesN1Standard"
	} else if r.IsHighMem() {
		resourceGroup = "SQLGen2InstancesN1Highmem"
	}

	vCPUs, err := r.vCPUs()
	if err != nil {
		logging.Logger.Warn().Msgf("vCPU of tier %s of %s is not parsable", r.Tier, r.Address)
		return nil
	}

	descriptionRegex := fmt.Sprintf("Cloud SQL for %s: %s - %s vCPU", r.dbTypeDescName(), r.availabilityTypeDescName(), vCPUs)

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
				{Key: "description", ValueRegex: regexPtr(descriptionRegex)},
			},
		},
	}
}

func (r *SQLDatabaseInstance) instanceCostComponents() []*schema.CostComponent {
	cpuDescRegex := fmt.Sprintf("%s: %s - vCPU", r.dbTypeDescName(), r.availabilityTypeDescName())
	memDescRegex := fmt.Sprintf("%s: %s - RAM", r.dbTypeDescName(), r.availabilityTypeDescName())

	vCPUs, err := r.vCPUs()
	if err != nil {
		logging.Logger.Warn().Msgf("vCPU of tier %s of %s is not parsable: %s", r.Tier, r.Address, err)
		return nil
	}

	mem, err := r.memory()
	if err != nil {
		logging.Logger.Warn().Msgf("memory of tier %s of %s is not parsable: %s", r.Tier, r.Address, err)
		return nil
	}

	return []*schema.CostComponent{
		{
			Name:           fmt.Sprintf("vCPUs (%s, %s)", vCPUs, strings.ToLower(r.AvailabilityType)),
			Unit:           "hours",
			UnitMultiplier: decimal.NewFromInt(1),
			HourlyQuantity: decimalPtr(vCPUs),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("gcp"),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cloud SQL"),
				ProductFamily: strPtr("ApplicationServices"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "description", ValueRegex: regexPtr(cpuDescRegex)},
				},
			},
		},
		{
			Name:           fmt.Sprintf("Memory (%s GB, %s)", mem, strings.ToLower(r.AvailabilityType)),
			Unit:           "GB",
			UnitMultiplier: schema.HourToMonthUnitMultiplier,
			HourlyQuantity: decimalPtr(mem),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("gcp"),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cloud SQL"),
				ProductFamily: strPtr("ApplicationServices"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "description", ValueRegex: regexPtr(memDescRegex)},
				},
			},
		},
	}
}

func (r *SQLDatabaseInstance) dbType() SQLInstanceDBType {
	if strings.HasPrefix(strings.ToLower(r.DatabaseVersion), "postgres") {
		return PostgreSQL
	} else if strings.HasPrefix(strings.ToLower(r.DatabaseVersion), "mysql") {
		return MySQL
	} else if strings.HasPrefix(strings.ToLower(r.DatabaseVersion), "sqlserver") {
		return SQLServer
	} else {
		return MySQL
	}
}

func (r *SQLDatabaseInstance) dbTypeDescName() string {
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

func (r *SQLDatabaseInstance) diskTypeDescName() string {
	diskTypeDescs := map[string]string{
		"PD_SSD": "Standard storage",
		"PD_HDD": "Low cost storage",
	}

	return diskTypeDescs[r.DiskType]
}

func (r *SQLDatabaseInstance) vCPUs() (decimal.Decimal, error) {
	p := strings.Split(r.Tier, "-")

	if len(p) < 3 {
		return decimal.Decimal{}, fmt.Errorf("tier %s has no vCPU data", r.Tier)
	}

	if r.IsCustom() {
		return decimal.NewFromString(p[2])
	}

	return decimal.NewFromString(p[len(p)-1])
}

func (r *SQLDatabaseInstance) memory() (decimal.Decimal, error) {
	if r.IsCustom() {
		p := strings.Split(r.Tier, "-")

		if len(p) < 4 {
			return decimal.Decimal{}, fmt.Errorf("tier %s has no RAM data", r.Tier)
		}

		v, err := decimal.NewFromString(p[len(p)-1])
		if err != nil {
			return decimal.Decimal{}, err
		}

		return v.Div(decimal.NewFromInt(1024)), nil
	} else if r.IsStandard() || r.IsHighMem() {
		vCPUs, err := r.vCPUs()
		if err != nil {
			return decimal.Decimal{}, err
		}

		if r.IsStandard() {
			return vCPUs.Mul(standardRAMRatio).Div(decimal.NewFromInt(1024)), nil
		} else if r.IsHighMem() {
			return vCPUs.Mul(highmemRAMRatio).Div(decimal.NewFromInt(1024)), nil
		}
	} else if r.isLightweight() {
		return lightweightRAM.Div(decimal.NewFromInt(1024)), nil
	}

	return decimal.Decimal{}, fmt.Errorf("tier %s has no RAM data", r.Tier)
}

func (r *SQLDatabaseInstance) storageCostComponent() *schema.CostComponent {
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

	return &schema.CostComponent{
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
				{Key: "description", ValueRegex: regexPtr(fmt.Sprintf("%s: %s - %s", r.dbTypeDescName(), r.availabilityTypeDescName(), r.diskTypeDescName()))},
			},
		},
	}
}

func (r *SQLDatabaseInstance) backupCostComponent(quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
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
		UsageBased: true,
	}
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

func (r *SQLDatabaseInstance) IsCustom() bool {
	return strings.HasPrefix(strings.ToLower(r.Tier), "db-custom-")
}

func (r *SQLDatabaseInstance) IsShared() bool {
	return strings.EqualFold(r.Tier, "db-f1-micro") || strings.EqualFold(r.Tier, "db-g1-small")
}

func (r *SQLDatabaseInstance) IsLegacy() bool {
	return strings.HasPrefix(strings.ToLower(r.Tier), "db-n1-")
}

func (r *SQLDatabaseInstance) IsStandard() bool {
	return strings.HasPrefix(strings.ToLower(r.Tier), "db-n1-standard") || strings.HasPrefix(strings.ToLower(r.Tier), "db-standard")
}

func (r *SQLDatabaseInstance) IsHighMem() bool {
	return strings.HasPrefix(strings.ToLower(r.Tier), "db-n1-highmem") || strings.HasPrefix(strings.ToLower(r.Tier), "db-highmem")
}

func (r *SQLDatabaseInstance) isLightweight() bool {
	return strings.HasPrefix(strings.ToLower(r.Tier), "db-lightweight")
}
