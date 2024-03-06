package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/usage"
)

type KinesisFirehoseDeliveryStream struct {
	Address                     string
	Region                      string
	DataFormatConversionEnabled bool
	VPCDeliveryEnabled          bool
	VPCDeliveryAZs              int64
	MonthlyDataIngestedGB       *float64 `infracost_usage:"monthly_data_ingested_gb"`
}

func (r *KinesisFirehoseDeliveryStream) CoreType() string {
	return "KinesisFirehoseDeliveryStream"
}

func (r *KinesisFirehoseDeliveryStream) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_data_ingested_gb", ValueType: schema.Float64, DefaultValue: 0},
	}
}

func (r *KinesisFirehoseDeliveryStream) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *KinesisFirehoseDeliveryStream) BuildResource() *schema.Resource {
	costComponents := make([]*schema.CostComponent, 0)

	if r.MonthlyDataIngestedGB != nil {
		tierLimits := []int{512_000, 1_536_000}

		result := usage.CalculateTierBuckets(decimal.NewFromFloat(*r.MonthlyDataIngestedGB), tierLimits)

		if result[0].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, r.dataIngestedCostComponent("first 500TB", "0", "512000", &result[0]))
		}
		if result[1].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, r.dataIngestedCostComponent("next 1.5PB", "512000", "2048000", &result[1]))
		}
		if result[2].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, r.dataIngestedCostComponent("next 3PB", "2048000", "Inf", &result[2]))
		}
	} else {
		costComponents = append(costComponents, r.dataIngestedCostComponent("first 500TB", "0", "512000", nil))
	}

	if r.DataFormatConversionEnabled {
		costComponents = append(costComponents, r.formatConversionCostComponent())
	}

	if r.VPCDeliveryEnabled {
		costComponents = append(costComponents, r.vpcDataCostComponent())
		costComponents = append(costComponents, r.vpcDeliveryCostComponent())
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *KinesisFirehoseDeliveryStream) dataIngestedCostComponent(tier, startUsageAmount, endUsageAmount string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Data ingested (%s)", tier),
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonKinesisFirehose"),
			ProductFamily: strPtr("Kinesis Firehose"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "group", Value: strPtr("Event-by-Event Processing")},
				{Key: "sourcetype", Value: strPtr("")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(startUsageAmount),
			EndUsageAmount:   strPtr(endUsageAmount),
		},
		UsageBased: true,
	}
}

func (r *KinesisFirehoseDeliveryStream) formatConversionCostComponent() *schema.CostComponent {
	var monthlyDataIngestedGB *decimal.Decimal
	if r.MonthlyDataIngestedGB != nil {
		monthlyDataIngestedGB = decimalPtr(decimal.NewFromFloat(*r.MonthlyDataIngestedGB))
	}

	return &schema.CostComponent{
		Name:            "Format conversion",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: monthlyDataIngestedGB,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonKinesisFirehose"),
			ProductFamily: strPtr("Kinesis Firehose"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "operation", Value: strPtr("DataFormatConversion")},
			},
		},
	}
}

func (r *KinesisFirehoseDeliveryStream) vpcDataCostComponent() *schema.CostComponent {
	var monthlyDataIngestedGB *decimal.Decimal
	if r.MonthlyDataIngestedGB != nil {
		monthlyDataIngestedGB = decimalPtr(decimal.NewFromFloat(*r.MonthlyDataIngestedGB))
	}

	return &schema.CostComponent{
		Name:            "VPC data",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: monthlyDataIngestedGB,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonKinesisFirehose"),
			ProductFamily: strPtr("Kinesis Firehose"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "operation", Value: strPtr("VpcBandwidth")},
			},
		},
	}
}

func (r *KinesisFirehoseDeliveryStream) vpcDeliveryCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "VPC AZ delivery",
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(r.VPCDeliveryAZs)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonKinesisFirehose"),
			ProductFamily: strPtr("Kinesis Firehose"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "operation", Value: strPtr("RunVpcInstance")},
			},
		},
	}
}
