package aws

import (
	"fmt"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"strings"
)

func GetRedshiftClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_redshift_cluster",
		RFunc: NewRedshiftCluster,
	}
}

func NewRedshiftCluster(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	nodeType := d.Get("node_type").String()
	numberOfNodes := int64(1)
	if d.Get("number_of_nodes").Type != gjson.Null {
		numberOfNodes = d.Get("number_of_nodes").Int()
	}

	costComponents := []*schema.CostComponent{
		{
			Name:           fmt.Sprintf("Cluster usage (%s, %s)", "on-demand", nodeType),
			Unit:           "hours",
			UnitMultiplier: 1,
			HourlyQuantity: decimalPtr(decimal.NewFromInt(numberOfNodes)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonRedshift"),
				ProductFamily: strPtr("Compute Instance"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "instanceType", Value: strPtr(nodeType)},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("on_demand"),
			},
		},
	}

	if strings.HasPrefix(nodeType, "ra3") {
		var managedStorage *decimal.Decimal
		if u != nil && u.Get("managed_storage_gb").Type != gjson.Null {
			managedStorage = decimalPtr(decimal.NewFromInt(u.Get("managed_storage_gb").Int()))
		}
		costComponents = append(costComponents, redshiftManagedStorageCostComponent(region, nodeType, managedStorage))
	}

	if strings.HasPrefix(nodeType, "ra3") || strings.HasPrefix(nodeType, "ds2") || strings.HasPrefix(nodeType, "dc2") {
		var concurrencyScalingSeconds *decimal.Decimal
		if u != nil && u.Get("excess_concurrency_scaling_secs").Type != gjson.Null {
			concurrencyScalingSeconds = decimalPtr(decimal.NewFromInt(u.Get("excess_concurrency_scaling_secs").Int()))
		}
		costComponents = append(costComponents, redshiftConcurrencyScalingCostComponent(region, nodeType, numberOfNodes, concurrencyScalingSeconds))
	}

	var terabytesScanned *decimal.Decimal
	if u != nil && u.Get("spectrum_data_scanned_tb").Type != gjson.Null {
		terabytesScanned = decimalPtr(decimal.NewFromFloat(u.Get("spectrum_data_scanned_tb").Float()))
	}
	costComponents = append(costComponents, redshiftSpectrumCostComponent(region, terabytesScanned))

	if u != nil && u.Get("backup_storage_gb").Type != gjson.Null {
		storageSnapshotGb := decimalPtr(decimal.NewFromInt(u.Get("backup_storage_gb").Int()))
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
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func redshiftConcurrencyScalingCostComponent(region string, nodeType string, numberOfNodes int64, concurrencySeconds *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Concurrency scaling (%s)", nodeType),
		Unit:            "node-seconds",
		UnitMultiplier:  1,
		MonthlyQuantity: concurrencySeconds,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonRedshift"),
			ProductFamily: strPtr("Redshift Concurrency Scaling"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "instanceType", Value: strPtr(nodeType)},
				{Key: "concurrencyscalingfreeusage", Value: strPtr("No")},
			},
		},
	}
}

func redshiftSpectrumCostComponent(region string, terabytesScanned *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Spectrum",
		Unit:            "TB",
		UnitMultiplier:  1,
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
		Unit:            "GB-months",
		UnitMultiplier:  1,
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
		Unit:            "GB-months",
		UnitMultiplier:  1,
		MonthlyQuantity: managedStorage,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonRedshift"),
			ProductFamily: strPtr("Redshift Managed Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "instanceType", Value: strPtr(nodeType)},
			},
		},
	}
}
