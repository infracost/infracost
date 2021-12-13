package aws

import (
	"strings"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetKinesisDataAnalyticsRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_kinesisanalyticsv2_application",
		RFunc: NewKinesisDataAnalytics,
		Notes: []string{
			"Terraform doesn’t currently support Analytics Studio, but when it does they will require 2 orchestration KPUs.",
		},
	}
}

func NewKinesisDataAnalytics(ctx *config.RunContext, d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	costComponents := make([]*schema.CostComponent, 0)
	var kinesisProcessingUnits, durableApplicationBackupGb *decimal.Decimal

	if u != nil && u.Get("kinesis_processing_units").Type != gjson.Null {
		kinesisProcessingUnits = decimalPtr(decimal.NewFromInt(u.Get("kinesis_processing_units").Int()))
	}

	costComponents = append(costComponents, kinesisProcessingsCostComponent("Processing (stream)", region, kinesisProcessingUnits))

	if u != nil && u.Get("durable_application_backup_gb").Type != gjson.Null {
		durableApplicationBackupGb = decimalPtr(decimal.NewFromInt(u.Get("durable_application_backup_gb").Int()))
	}
	runtimeEnvironment := d.Get("runtime_environment").String()

	if strings.HasPrefix(strings.ToLower(runtimeEnvironment), "flink") {
		costComponents = append(costComponents, kinesisProcessingsCostComponent("Processing (orchestration)", region, decimalPtr(decimal.NewFromInt(1))))
		costComponents = append(costComponents, kinesisRunningStorageCostComponent(region, kinesisProcessingUnits))
		costComponents = append(costComponents, kinesisBackupCostComponent(region, durableApplicationBackupGb))
	}
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func kinesisProcessingsCostComponent(name, region string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           name,
		Unit:           "KPU",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonKinesisAnalytics"),
			ProductFamily: strPtr("Kinesis Analytics"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/KPU-Hour-Java/i")},
			},
		},
	}
}
func kinesisRunningStorageCostComponent(region string, quantity *decimal.Decimal) *schema.CostComponent {
	if quantity != nil {
		quantity = decimalPtr(quantity.Mul(decimal.NewFromInt(int64(50))))
	}
	return &schema.CostComponent{
		Name:            "Running storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonKinesisAnalytics"),
			ProductFamily: strPtr("Kinesis Analytics"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/RunningApplicationStorage$/i")},
			},
		},
	}
}
func kinesisBackupCostComponent(region string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Backup",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonKinesisAnalytics"),
			ProductFamily: strPtr("Kinesis Analytics"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/DurableApplicationBackups/i")},
			},
		},
	}
}
