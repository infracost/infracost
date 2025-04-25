package google

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

type RedisCluster struct {
	Address      	string
	Region       	string
	NodeType     	string
	NodeCount    	int
	ProvisionedGB 	int64
	AOFEnabled 		bool
	BackupsEnabled 	bool
	BackupStorageGB *float64 `infracost_usage:"backup_storage_gb"`
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
		Name:				r.Address,
		CostComponents:		costComponents,
		UsageSchema: 		r.UsageSchema(),
	}
	}

func (r *RedisCluster) nodeTypeCostComponent() *schema.CostComponent {
	nodeTypeDescriptions := map[string]string{
		"REDIS_SHARED_CORE_NANO": "Shared Core Nano",
		"REDIS_STANDARD_SMALL":   "Standard Small",
		"REDIS_HIGHMEM_MEDIUM":   "Highmem Medium",
		"REDIS_HIGHMEM_XLARGE":   "Highmem XLarge",
	}

	desc, ok := nodeTypeDescriptions[strings.ToUpper(r.NodeType)]
	if !ok {
		desc = r.NodeType
	}

	name := fmt.Sprintf("Redis Cluster Node (%s)", strings.ToLower(desc))
	descriptionRegex := fmt.Sprintf("Redis Cluster Node %s", desc)

	return &schema.CostComponent{
		Name:				name,
		Unit:				"hours",
		UnitMultiplier: 	decimal.NewFromInt(1),
		HourlyQuantity: 	decimalPtr(decimal.NewFromInt(int64(r.NodeCount))),
		ProductFilter: &schema.ProductFilter{
			VendorName:		strPtr("gcp"),
			Region:			strPtr(r.Region),
			Service:		strPtr("Cloud Memorystore Redis"),
			ProductFamily:	strPtr("ApplicationServices"),
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
		HourlyQuantity: decimalPtr(decimal.NewFromInt(r.ProvisionedGB)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Cloud Memorystore for Redis"),
			ProductFamily: strPtr("Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", ValueRegex: regexPtr("AOF Persistence")},
			},
		},
	}
}

func (r *RedisCluster) backupCostComponent() *schema.CostComponent {
	var backupGB *decimal.Decimal
	if r.BackupStorageGB != nil {
		backupGB = decimalPtr(decimal.NewFromFloat(*r.BackupStorageGB))
	}

	return &schema.CostComponent{
		Name:           "Backups",
		Unit:           "GB",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: backupGB,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Cloud Memorystore for Redis"),
			ProductFamily: strPtr("Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", ValueRegex: regexPtr("Backup Storage")},
			},
		},
	}
}