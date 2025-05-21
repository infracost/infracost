package google

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

var redisNodeTypeNames = map[string]string{
	"REDIS_SHARED_CORE_NANO": "shared core nano",
	"REDIS_STANDARD_SMALL":   "standard small",
	"REDIS_HIGHMEM_MEDIUM":   "highmem medium",
	"REDIS_HIGHMEM_XLARGE":   "highmem xlarge",
}

var redisNodeTypeDescSuffixes = map[string]string{
	"REDIS_SHARED_CORE_NANO": "Shared Core Nano",
	"REDIS_STANDARD_SMALL":   "Standard Small",
	"REDIS_HIGHMEM_MEDIUM":   "Default",
	"REDIS_HIGHMEM_XLARGE":   "Highmem XLarge",
}

type RedisCluster struct {
	Address          string
	Region           string
	NodeType         string
	NodeCount        int
	AOFProvisionedGB int64
	AOFEnabled       bool
	BackupsEnabled   bool
	BackupStorageGB  *float64 `infracost_usage:"backup_storage_gb"`
}

func (r *RedisCluster) CoreType() string {
	return "RedisCluster"
}

func (r *RedisCluster) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "backup_storage_gb", ValueType: schema.Float64, DefaultValue: 0},
	}
}

func (r *RedisCluster) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *RedisCluster) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{}

	costComponents = append(costComponents, r.nodeTypeCostComponent())

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

func (r *RedisCluster) nodeTypeCostComponent() *schema.CostComponent {
	nameSuffix, ok := redisNodeTypeNames[strings.ToUpper(r.NodeType)]
	if !ok {
		nameSuffix = r.NodeType
	}
	name := fmt.Sprintf("Cluster node (%s)", strings.ToLower(nameSuffix))

	descSuffix, ok := redisNodeTypeDescSuffixes[strings.ToUpper(r.NodeType)]
	if !ok {
		descSuffix = r.NodeType
	}
	descriptionRegex := fmt.Sprintf("Redis Cluster Node %s", descSuffix)

	return &schema.CostComponent{
		Name:           name,
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(int64(r.NodeCount))),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Cloud Memorystore for Redis"),
			ProductFamily: strPtr("ApplicationServices"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", ValueRegex: regexPtr(descriptionRegex)},
			},
		},
	}
}

func (r *RedisCluster) aofCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "AOF persistence",
		Unit:           "GB",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(r.AOFProvisionedGB)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Cloud Memorystore for Redis"),
			ProductFamily: strPtr("ApplicationServices"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", ValueRegex: regexPtr("Memorystore for Redis Cluster: AOF Storage")},
			},
		},
	}
}

func (r *RedisCluster) backupCostComponent() *schema.CostComponent {
	var backupGB *decimal.Decimal
	if r.BackupStorageGB != nil {
		backupGB = decimalPtr(decimal.NewFromFloat(*r.BackupStorageGB).Mul(schema.HourToMonthUnitMultiplier))
	}

	return &schema.CostComponent{
		Name:            "Backups",
		Unit:            "GB",
		UnitMultiplier:  schema.HourToMonthUnitMultiplier,
		MonthlyQuantity: backupGB,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Cloud Memorystore for Redis"),
			ProductFamily: strPtr("ApplicationServices"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", ValueRegex: regexPtr("Memorystore for Redis Cluster: Backups")},
			},
		},
	}
}
