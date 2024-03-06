package aws

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/usage"
)

type RedshiftCluster struct {
	Address                      string
	Region                       string
	NodeType                     string
	Nodes                        *int64
	ManagedStorageGB             *float64 `infracost_usage:"managed_storage_gb"`
	ExcessConcurrencyScalingSecs *int64   `infracost_usage:"excess_concurrency_scaling_secs"`
	SpectrumDataScannedTB        *float64 `infracost_usage:"spectrum_data_scanned_tb"`
	BackupStorageGB              *float64 `infracost_usage:"backup_storage_gb"`
}

func (r *RedshiftCluster) CoreType() string {
	return "RedshiftCluster"
}

func (r *RedshiftCluster) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "managed_storage_gb", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "excess_concurrency_scaling_secs", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "spectrum_data_scanned_tb", ValueType: schema.Float64, DefaultValue: 0.0},
		{Key: "backup_storage_gb", ValueType: schema.Float64, DefaultValue: 0},
	}
}

func (r *RedshiftCluster) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *RedshiftCluster) BuildResource() *schema.Resource {
	numberOfNodes := int64(1)
	if r.Nodes != nil {
		numberOfNodes = *r.Nodes
	}

	costComponents := []*schema.CostComponent{
		{
			Name:           fmt.Sprintf("Cluster usage (%s, %s)", "on-demand", r.NodeType),
			Unit:           "hours",
			UnitMultiplier: decimal.NewFromInt(1),
			HourlyQuantity: decimalPtr(decimal.NewFromInt(numberOfNodes)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(r.Region),
				Service:       strPtr("AmazonRedshift"),
				ProductFamily: strPtr("Compute Instance"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "instanceType", ValueRegex: strPtr(fmt.Sprintf("/%s/i", r.NodeType))},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("on_demand"),
			},
		},
	}

	if strings.HasPrefix(r.NodeType, "ra3") {
		var managedStorage *decimal.Decimal
		if r.ManagedStorageGB != nil {
			managedStorage = decimalPtr(decimal.NewFromFloat(*r.ManagedStorageGB))
		}

		costComponents = append(costComponents, r.managedStorageCostComponent(managedStorage))
	}

	if strings.HasPrefix(r.NodeType, "ra3") || strings.HasPrefix(r.NodeType, "ds2") || strings.HasPrefix(r.NodeType, "dc2") {
		var concurrencyScalingSeconds *decimal.Decimal
		if r.ExcessConcurrencyScalingSecs != nil {
			concurrencyScalingSeconds = decimalPtr(decimal.NewFromInt(*r.ExcessConcurrencyScalingSecs))
		}

		costComponents = append(costComponents, r.concurrencyScalingCostComponent(numberOfNodes, concurrencyScalingSeconds))
	}

	var tbScanned *decimal.Decimal
	if r.SpectrumDataScannedTB != nil {
		tbScanned = decimalPtr(decimal.NewFromFloat(*r.SpectrumDataScannedTB))
	}

	costComponents = append(costComponents, r.spectrumCostComponent(tbScanned))

	if r.BackupStorageGB != nil {
		storageSnapshotGB := decimal.NewFromFloat(*r.BackupStorageGB)
		storageSnapshotTiers := usage.CalculateTierBuckets(storageSnapshotGB, []int{51200, 512000})

		if storageSnapshotTiers[0].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, r.storageSnapshotCostComponent("Backup storage (first 50 TB)", "0", &storageSnapshotTiers[0]))
		}

		if storageSnapshotTiers[1].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, r.storageSnapshotCostComponent("Backup storage (next 450 TB)", "51200", &storageSnapshotTiers[1]))
		}

		if storageSnapshotTiers[2].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, r.storageSnapshotCostComponent("Backup storage (over 500 TB)", "512000", &storageSnapshotTiers[2]))
		}
	} else {
		costComponents = append(costComponents, r.storageSnapshotCostComponent("Backup storage (first 50 TB)", "0", nil))
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *RedshiftCluster) concurrencyScalingCostComponent(numberOfNodes int64, concurrencySeconds *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Concurrency scaling (%s)", r.NodeType),
		Unit:            "node-seconds",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: concurrencySeconds,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonRedshift"),
			ProductFamily: strPtr("Redshift Concurrency Scaling"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "instanceType", ValueRegex: strPtr(fmt.Sprintf("/%s/i", r.NodeType))},
				{Key: "concurrencyscalingfreeusage", Value: strPtr("No")},
			},
		},
		UsageBased: true,
	}
}

func (r *RedshiftCluster) spectrumCostComponent(tbScanned *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Spectrum",
		Unit:            "TB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: tbScanned,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonRedshift"),
			ProductFamily: strPtr("Redshift Data Scan"),
		},
		UsageBased: true,
	}
}

func (r *RedshiftCluster) storageSnapshotCostComponent(displayName string, startUsageAmount string, storageSnapshot *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            displayName,
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: storageSnapshot,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonRedshift"),
			ProductFamily: strPtr("Storage Snapshot"),
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(startUsageAmount),
		},
		UsageBased: true,
	}
}

func (r *RedshiftCluster) managedStorageCostComponent(managedStorage *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Managed storage (%s)", r.NodeType),
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: managedStorage,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonRedshift"),
			ProductFamily: strPtr("Redshift Managed Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "instanceType", ValueRegex: strPtr(fmt.Sprintf("/%s/i", r.NodeType))},
			},
		},
		UsageBased: true,
	}
}
