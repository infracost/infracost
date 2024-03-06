package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type KinesisAnalyticsApplication struct {
	Address                string
	Region                 string
	KinesisProcessingUnits *int64 `infracost_usage:"kinesis_processing_units"`
}

func (r *KinesisAnalyticsApplication) CoreType() string {
	return "KinesisAnalyticsApplication"
}

func (r *KinesisAnalyticsApplication) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "kinesis_processing_units", ValueType: schema.Int64, DefaultValue: 0},
	}
}

func (r *KinesisAnalyticsApplication) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *KinesisAnalyticsApplication) BuildResource() *schema.Resource {
	var kinesisProcessingUnits *decimal.Decimal
	if r.KinesisProcessingUnits != nil {
		kinesisProcessingUnits = decimalPtr(decimal.NewFromInt(*r.KinesisProcessingUnits))
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: []*schema.CostComponent{r.processingStreamCostComponent(kinesisProcessingUnits)},
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *KinesisAnalyticsApplication) processingStreamCostComponent(kinesisProcessingUnits *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "Processing (stream)",
		Unit:           "KPU",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: kinesisProcessingUnits,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonKinesisAnalytics"),
			ProductFamily: strPtr("Kinesis Analytics"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/KPU-Hour-Java/i")},
			},
		},
		UsageBased: true,
	}
}
