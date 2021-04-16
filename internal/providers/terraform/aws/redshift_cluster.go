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
			Name:           fmt.Sprintf("Cluster Usage (%s, %s)", "on-demand", nodeType),
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

	if u != nil && u.Get("excess_concurrency_scaling_sec").Type != gjson.Null {
		concurrencyScalingSeconds := u.Get("excess_concurrency_scaling_sec").Int()
		if concurrencyScalingSeconds > 0 {
			costComponents = append(costComponents, redshiftConcurrencyScalingCostComponent(region, nodeType, numberOfNodes, concurrencyScalingSeconds))
		}
	}

	if u != nil && u.Get("spectrum_data_scanned_tb").Type != gjson.Null {
		terabytesScanned := u.Get("spectrum_data_scanned_tb").Float()
		if terabytesScanned > 0 {
			costComponents = append(costComponents, redshiftSpectrumCostComponent(region, terabytesScanned))
		}
	}

	if u != nil && u.Get("backup_storage_gb").Type != gjson.Null {
		storageSnapshotGb := decimalPtr(decimal.NewFromInt(u.Get("backup_storage_gb").Int()))
		storageSnapshotTiers := usage.CalculateTierBuckets(*storageSnapshotGb, []int{51200, 512000})

		if storageSnapshotTiers[0].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, redshiftStorageSnapshotCostComponent(region, "Backup Storage (first 50 TB)", "0", &storageSnapshotTiers[0]))
		}

		if storageSnapshotTiers[1].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, redshiftStorageSnapshotCostComponent(region, "Backup Storage (next 450 TB)", "51200", &storageSnapshotTiers[1]))
		}

		if storageSnapshotTiers[2].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, redshiftStorageSnapshotCostComponent(region, "Backup Storage (over 500 TB)", "512000", &storageSnapshotTiers[2]))
		}
	}

	if strings.HasPrefix(nodeType, "ra3") {
		var managedStorage *decimal.Decimal
		if u != nil && u.Get("managed_storage_gb").Type != gjson.Null {
			managedStorage = decimalPtr(decimal.NewFromInt(u.Get("managed_storage_gb").Int()))
		}
		costComponents = append(costComponents, redshiftManagedStorageCostComponent(region, nodeType, managedStorage))
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func redshiftConcurrencyScalingCostComponent(region string, nodeType string, numberOfNodes int64, concurrencySeconds int64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Concurrency Scaling (%s)", nodeType),
		Unit:            "Node-seconds", // maybe this should just be 'seconds' but the descrpiption is ""$0.00007 per Redshift Concurrency Scaling DC2.L Node-second"
		UnitMultiplier:  1,
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(numberOfNodes * concurrencySeconds)),
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

func redshiftSpectrumCostComponent(region string, terabytesScanned float64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Spectrum",
		Unit:            "terabytes",
		UnitMultiplier:  1,
		MonthlyQuantity: decimalPtr(decimal.NewFromFloat(terabytesScanned)),
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
		Name:            fmt.Sprintf("Managed Storage (%s)", nodeType),
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
