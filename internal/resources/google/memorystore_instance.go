package google

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

type memorystoreValkeyNodeType struct {
	name        string
	description string
}

var memorystoreValkeyNodeTypes = map[string]memorystoreValkeyNodeType{
	"SHARED_CORE_NANO": {name: "shared core nano", description: "Shared Core Nano"},
	"CUSTOM_PICO":      {name: "custom pico", description: "Custom Pico"},
	"CUSTOM_MICRO":     {name: "custom micro", description: "Custom Micro"},
	"CUSTOM_MINI":      {name: "custom mini", description: "Custom Mini"},
	"STANDARD_SMALL":   {name: "standard small", description: "Standard Small"},
	"HIGHMEM_MEDIUM":   {name: "highmem medium", description: "Highmem Medium"},
	"HIGHCPU_MEDIUM":   {name: "highcpu medium", description: "Highcpu Medium"},
	"STANDARD_LARGE":   {name: "standard large", description: "Standard Large"},
	"HIGHMEM_XLARGE":   {name: "highmem xlarge", description: "Highmem XLarge"},
	"HIGHMEM_2XLARGE":  {name: "highmem 2xlarge", description: "Highmem 2xlarge"},
}

type MemorystoreInstance struct {
	Address          string
	Region           string
	NodeType         string
	NodeCount        *int64
	AOFProvisionedGB *float64
	AOFEnabled       bool
	BackupsEnabled   bool
	BackupStorageGB  *float64 `infracost_usage:"backup_storage_gb"`
}

func (r *MemorystoreInstance) CoreType() string {
	return "MemorystoreInstance"
}

func (r *MemorystoreInstance) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "backup_storage_gb", ValueType: schema.Float64, DefaultValue: 0},
	}
}

func (r *MemorystoreInstance) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *MemorystoreInstance) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{r.nodeCostComponent()}

	if r.AOFEnabled {
		costComponents = append(costComponents, r.aofCostComponent())
	}

	if r.BackupsEnabled {
		costComponents = append(costComponents, r.backupCostComponent())
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *MemorystoreInstance) nodeCostComponent() *schema.CostComponent {
	nodeType := memorystoreValkeyNodeTypes[strings.ToUpper(r.NodeType)]
	descriptionRegex := fmt.Sprintf("^%s Node", nodeType.description)

	return &schema.CostComponent{
		Name:           fmt.Sprintf("Node (%s)", nodeType.name),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: intPtrToDecimalPtr(r.NodeCount),
		ProductFilter:  r.productFilter(descriptionRegex),
	}
}

func (r *MemorystoreInstance) aofCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "AOF persistence",
		Unit:           "GB",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: floatPtrToDecimalPtr(r.AOFProvisionedGB),
		ProductFilter:  r.productFilter("^Memorystore for Valkey: AOF Storage"),
	}
}

func (r *MemorystoreInstance) backupCostComponent() *schema.CostComponent {
	var backupGB *decimal.Decimal
	if r.BackupStorageGB != nil {
		backupGB = decimalPtr(decimal.NewFromFloat(*r.BackupStorageGB).Mul(schema.HourToMonthUnitMultiplier))
	}

	return &schema.CostComponent{
		Name:            "Backups",
		Unit:            "GB",
		UnitMultiplier:  schema.HourToMonthUnitMultiplier,
		MonthlyQuantity: backupGB,
		ProductFilter:   r.productFilter("^Memorystore for Valkey: Backups"),
	}
}

func (r *MemorystoreInstance) productFilter(descriptionRegex string) *schema.ProductFilter {
	return &schema.ProductFilter{
		VendorName:    strPtr("gcp"),
		Region:        strPtr(r.Region),
		Service:       strPtr("Cloud Memorystore"),
		ProductFamily: strPtr("ApplicationServices"),
		AttributeFilters: []*schema.AttributeFilter{
			{Key: "resourceGroup", Value: strPtr("Valkey")},
			{Key: "description", ValueRegex: regexPtr(descriptionRegex)},
		},
	}
}
