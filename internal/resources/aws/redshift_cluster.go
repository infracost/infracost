package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"

	"github.com/infracost/infracost/internal/usage"
	"github.com/shopspring/decimal"

	"strings"
)

type RedshiftCluster struct {
	Address                      *string
	Region                       *string
	NodeType                     *string
	NumberOfNodes                *int64
	ManagedStorageGb             *int64   `infracost_usage:"managed_storage_gb"`
	ExcessConcurrencyScalingSecs *int64   `infracost_usage:"excess_concurrency_scaling_secs"`
	SpectrumDataScannedTb        *float64 `infracost_usage:"spectrum_data_scanned_tb"`
	BackupStorageGb              *int64   `infracost_usage:"backup_storage_gb"`
}

var RedshiftClusterUsageSchema = []*schema.UsageItem{{Key: "managed_storage_gb", ValueType: schema.Int64, DefaultValue: 0}, {Key: "excess_concurrency_scaling_secs", ValueType: schema.Int64, DefaultValue: 0}, {Key: "spectrum_data_scanned_tb", ValueType: schema.Float64, DefaultValue: 0.000000}, {Key: "backup_storage_gb", ValueType: schema.Int64, DefaultValue: 0}}

func (r *RedshiftCluster) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *RedshiftCluster) BuildResource() *schema.Resource {
	region := *r.Region

	nodeType := *r.NodeType
	numberOfNodes := int64(1)
	if r.NumberOfNodes != nil {
		numberOfNodes = *r.NumberOfNodes
	}

	costComponents := []*schema.CostComponent{
		{
			Name:           fmt.Sprintf("Cluster usage (%s, %s)", "on-demand", nodeType),
			Unit:           "hours",
			UnitMultiplier: decimal.NewFromInt(1),
			HourlyQuantity: decimalPtr(decimal.NewFromInt(numberOfNodes)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonRedshift"),
				ProductFamily: strPtr("Compute Instance"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "instanceType", ValueRegex: strPtr(fmt.Sprintf("/%s/i", nodeType))},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("on_demand"),
			},
		},
	}

	if strings.HasPrefix(nodeType, "ra3") {
		var managedStorage *decimal.Decimal
		if r != nil && r.ManagedStorageGb != nil {
			managedStorage = decimalPtr(decimal.NewFromInt(*r.ManagedStorageGb))
		}
		costComponents = append(costComponents, redshiftManagedStorageCostComponent(region, nodeType, managedStorage))
	}

	if strings.HasPrefix(nodeType, "ra3") || strings.HasPrefix(nodeType, "ds2") || strings.HasPrefix(nodeType, "dc2") {
		var concurrencyScalingSeconds *decimal.Decimal
		if r != nil && r.ExcessConcurrencyScalingSecs != nil {
			concurrencyScalingSeconds = decimalPtr(decimal.NewFromInt(*r.ExcessConcurrencyScalingSecs))
		}
		costComponents = append(costComponents, redshiftConcurrencyScalingCostComponent(region, nodeType, numberOfNodes, concurrencyScalingSeconds))
	}

	var terabytesScanned *decimal.Decimal
	if r != nil && r.SpectrumDataScannedTb != nil {
		terabytesScanned = decimalPtr(decimal.NewFromFloat(*r.SpectrumDataScannedTb))
	}
	costComponents = append(costComponents, redshiftSpectrumCostComponent(region, terabytesScanned))

	if r != nil && r.BackupStorageGb != nil {
		storageSnapshotGb := decimalPtr(decimal.NewFromInt(*r.BackupStorageGb))
		storageSnapshotTiers := usage.CalculateTierBuckets(*storageSnapshotGb, []int{51200, 512000})

		if storageSnapshotTiers[0].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, redshiftStorageSnapshotCostComponent(region, "Backup storage (first 50 TB)", "0", &storageSnapshotTiers[0]))
		}

		if storageSnapshotTiers[1].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, redshiftStorageSnapshotCostComponent(region, "Backup storage (next 450 TB)", "51200", &storageSnapshotTiers[1]))
		}

		if storageSnapshotTiers[2].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, redshiftStorageSnapshotCostComponent(region, "Backup storage (over 500 TB)", "512000", &storageSnapshotTiers[2]))
		}
	} else {
		costComponents = append(costComponents, redshiftStorageSnapshotCostComponent(region, "Backup storage (first 50 TB)", "0", nil))
	}

	return &schema.Resource{
		Name:           *r.Address,
		CostComponents: costComponents, UsageSchema: RedshiftClusterUsageSchema,
	}
}

func redshiftConcurrencyScalingCostComponent(region string, nodeType string, numberOfNodes int64, concurrencySeconds *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Concurrency scaling (%s)", nodeType),
		Unit:            "node-seconds",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: concurrencySeconds,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonRedshift"),
			ProductFamily: strPtr("Redshift Concurrency Scaling"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "instanceType", ValueRegex: strPtr(fmt.Sprintf("/%s/i", nodeType))},
				{Key: "concurrencyscalingfreeusage", Value: strPtr("No")},
			},
		},
	}
}

func redshiftSpectrumCostComponent(region string, terabytesScanned *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Spectrum",
		Unit:            "TB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: terabytesScanned,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonRedshift"),
			ProductFamily: strPtr("Redshift Data Scan"),
		},
	}
}

func redshiftStorageSnapshotCostComponent(region string, displayName string, startUsageAmount string, storageSnapshot *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            displayName,
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: storageSnapshot,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonRedshift"),
			ProductFamily: strPtr("Storage Snapshot"),
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(startUsageAmount),
		},
	}
}

func redshiftManagedStorageCostComponent(region string, nodeType string, managedStorage *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Managed storage (%s)", nodeType),
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: managedStorage,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonRedshift"),
			ProductFamily: strPtr("Redshift Managed Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "instanceType", ValueRegex: strPtr(fmt.Sprintf("/%s/i", nodeType))},
			},
		},
	}
}
