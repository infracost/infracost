package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"strings"

	"github.com/shopspring/decimal"
)

type Kinesisanalyticsv2Application struct {
	Address                    *string
	Region                     *string
	RuntimeEnvironment         *string
	KinesisProcessingUnits     *int64   `infracost_usage:"kinesis_processing_units"`
	DurableApplicationBackupGb *float64 `infracost_usage:"durable_application_backup_gb"`
}

var Kinesisanalyticsv2ApplicationUsageSchema = []*schema.UsageItem{{Key: "kinesis_processing_units", ValueType: schema.Int64, DefaultValue: 0}, {Key: "durable_application_backup_gb", ValueType: schema.Float64, DefaultValue: 0}}

func (r *Kinesisanalyticsv2Application) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *Kinesisanalyticsv2Application) BuildResource() *schema.Resource {
	region := *r.Region
	costComponents := make([]*schema.CostComponent, 0)
	var kinesisProcessingUnits, durableApplicationBackupGb *decimal.Decimal

	if r.KinesisProcessingUnits != nil {
		kinesisProcessingUnits = decimalPtr(decimal.NewFromInt(*r.KinesisProcessingUnits))
	}

	costComponents = append(costComponents, kinesisProcessingsCostComponent("Processing (stream)", region, kinesisProcessingUnits))

	if r.DurableApplicationBackupGb != nil {
		durableApplicationBackupGb = decimalPtr(decimal.NewFromFloat(*r.DurableApplicationBackupGb))
	}
	runtimeEnvironment := *r.RuntimeEnvironment

	if strings.HasPrefix(strings.ToLower(runtimeEnvironment), "flink") {
		costComponents = append(costComponents, kinesisProcessingsCostComponent("Processing (orchestration)", region, decimalPtr(decimal.NewFromInt(1))))
		costComponents = append(costComponents, kinesisRunningStorageCostComponent(region, kinesisProcessingUnits))
		costComponents = append(costComponents, kinesisBackupCostComponent(region, durableApplicationBackupGb))
	}
	return &schema.Resource{
		Name:           *r.Address,
		CostComponents: costComponents, UsageSchema: Kinesisanalyticsv2ApplicationUsageSchema,
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
