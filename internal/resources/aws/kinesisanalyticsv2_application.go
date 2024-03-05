package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"strings"

	"github.com/shopspring/decimal"
)

type KinesisAnalyticsV2Application struct {
	Address                    string
	Region                     string
	RuntimeEnvironment         string
	KinesisProcessingUnits     *int64   `infracost_usage:"kinesis_processing_units"`
	DurableApplicationBackupGB *float64 `infracost_usage:"durable_application_backup_gb"`
}

func (r *KinesisAnalyticsV2Application) CoreType() string {
	return "KinesisAnalyticsV2Application"
}

func (r *KinesisAnalyticsV2Application) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "kinesis_processing_units", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "durable_application_backup_gb", ValueType: schema.Float64, DefaultValue: 0},
	}
}

func (r *KinesisAnalyticsV2Application) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *KinesisAnalyticsV2Application) BuildResource() *schema.Resource {
	costComponents := make([]*schema.CostComponent, 0)

	var kinesisProcessingUnits *decimal.Decimal
	if r.KinesisProcessingUnits != nil {
		kinesisProcessingUnits = decimalPtr(decimal.NewFromInt(*r.KinesisProcessingUnits))
	}

	var durableApplicationBackupGB *decimal.Decimal
	if r.DurableApplicationBackupGB != nil {
		durableApplicationBackupGB = decimalPtr(decimal.NewFromFloat(*r.DurableApplicationBackupGB))
	}

	v1App := &KinesisAnalyticsApplication{
		Region:                 r.Region,
		KinesisProcessingUnits: r.KinesisProcessingUnits,
	}

	costComponents = append(costComponents, v1App.processingStreamCostComponent(kinesisProcessingUnits))

	if strings.HasPrefix(strings.ToLower(r.RuntimeEnvironment), "flink") {
		costComponents = append(costComponents, r.processingOrchestrationCostComponent())
		costComponents = append(costComponents, r.runningStorageCostComponent(kinesisProcessingUnits))
		costComponents = append(costComponents, r.backupCostComponent(durableApplicationBackupGB))
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *KinesisAnalyticsV2Application) processingOrchestrationCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "Processing (orchestration)",
		Unit:           "KPU",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonKinesisAnalytics"),
			ProductFamily: strPtr("Kinesis Analytics"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/KPU-Hour-Java/i")},
			},
		},
	}
}

func (r *KinesisAnalyticsV2Application) runningStorageCostComponent(kinesisProcessingUnits *decimal.Decimal) *schema.CostComponent {
	var quantity *decimal.Decimal
	if kinesisProcessingUnits != nil {
		quantity = decimalPtr(kinesisProcessingUnits.Mul(decimal.NewFromInt(50)))
	}

	return &schema.CostComponent{
		Name:            "Running storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonKinesisAnalytics"),
			ProductFamily: strPtr("Kinesis Analytics"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/RunningApplicationStorage$/i")},
			},
		},
		UsageBased: true,
	}
}

func (r *KinesisAnalyticsV2Application) backupCostComponent(durableApplicationBackupGB *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Backup",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: durableApplicationBackupGB,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonKinesisAnalytics"),
			ProductFamily: strPtr("Kinesis Analytics"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/DurableApplicationBackups/i")},
			},
		},
		UsageBased: true,
	}
}
