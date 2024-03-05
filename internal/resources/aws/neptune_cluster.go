package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"

	"github.com/shopspring/decimal"
)

type NeptuneCluster struct {
	Address               string
	Region                string
	BackupRetentionPeriod int64
	StorageGB             *float64 `infracost_usage:"storage_gb"`
	MonthlyIORequests     *int64   `infracost_usage:"monthly_io_requests"`
	BackupStorageGB       *float64 `infracost_usage:"backup_storage_gb"`
}

func (r *NeptuneCluster) CoreType() string {
	return "NeptuneCluster"
}

func (r *NeptuneCluster) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "storage_gb", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "monthly_io_requests", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "backup_storage_gb", ValueType: schema.Float64, DefaultValue: 0},
	}
}

func (r *NeptuneCluster) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *NeptuneCluster) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		r.storageCostComponent(),
		r.ioRequestsCostComponent(),
	}

	if r.BackupRetentionPeriod > 1 {
		costComponents = append(costComponents, r.backupStorageCostComponent())
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *NeptuneCluster) storageCostComponent() *schema.CostComponent {
	var storageGB *decimal.Decimal
	if r.StorageGB != nil {
		storageGB = decimalPtr(decimal.NewFromFloat(*r.StorageGB))
	}

	return &schema.CostComponent{
		Name:            "Storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: storageGB,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(r.Region),
			Service:    strPtr("AmazonNeptune"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: regexPtr("^([A-Z]{3}\\d-|Global-|EU-)?StorageUsage$")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
		UsageBased: true,
	}
}

func (r *NeptuneCluster) ioRequestsCostComponent() *schema.CostComponent {
	var monthlyIORequests *decimal.Decimal
	if r.MonthlyIORequests != nil {
		monthlyIORequests = decimalPtr(decimal.NewFromInt(*r.MonthlyIORequests))
	}

	return &schema.CostComponent{
		Name:            "I/O requests",
		Unit:            "1M request",
		UnitMultiplier:  decimal.NewFromInt(int64(1000000)),
		MonthlyQuantity: monthlyIORequests,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(r.Region),
			Service:    strPtr("AmazonNeptune"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s$/i", "StorageIOUsage"))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
		UsageBased: true,
	}
}

func (r *NeptuneCluster) backupStorageCostComponent() *schema.CostComponent {
	var backupStorageGB *decimal.Decimal
	if r.BackupStorageGB != nil {
		backupStorageGB = decimalPtr(decimal.NewFromFloat(*r.BackupStorageGB))
	}

	return &schema.CostComponent{
		Name:            "Backup storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: backupStorageGB,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(r.Region),
			Service:    strPtr("AmazonNeptune"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/BackupUsage$/i")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
		UsageBased: true,
	}
}
