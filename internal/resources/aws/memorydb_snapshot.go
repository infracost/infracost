package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// MemoryDBSnapshot represents an AWS MemoryDB snapshot
//
// Resource information: https://docs.aws.amazon.com/memorydb/latest/devguide/snapshots.html
// Pricing information: https://aws.amazon.com/memorydb/pricing/
//
// Pricing notes:
// - Snapshot storage is $0.085/GB-month
type MemoryDBSnapshot struct {
	Address         string
	Region          string
	BackupStorageGB *float64 `infracost_usage:"backup_storage_gb"`
}

func (r *MemoryDBSnapshot) CoreType() string {
	return "MemoryDBSnapshot"
}

func (r *MemoryDBSnapshot) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "backup_storage_gb", ValueType: schema.Float64, DefaultValue: 0},
	}
}

func (r *MemoryDBSnapshot) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *MemoryDBSnapshot) BuildResource() *schema.Resource {
	var monthlyBackupStorageGB *decimal.Decimal

	if r.BackupStorageGB != nil {
		monthlyBackupStorageGB = decimalPtr(decimal.NewFromFloat(*r.BackupStorageGB))
	}

	costComponent := &schema.CostComponent{
		Name:            "Backup storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: monthlyBackupStorageGB,
		UsageBased:      true,
	}

	// Set a custom price for backup storage
	costComponent.SetCustomPrice(decimalPtr(decimal.NewFromFloat(0.085)))

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: []*schema.CostComponent{costComponent},
		UsageSchema:    r.UsageSchema(),
	}
}
