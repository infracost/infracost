package aws

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetKinesisFirehoseDeliveryStreamRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_kinesis_firehose_delivery_stream",
		RFunc: NewKinesisFirehoseDeliveryStream,
	}
}

func NewKinesisFirehoseDeliveryStream(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	costComponents := make([]*schema.CostComponent, 0)
	var monthlyDataIngestedGb *decimal.Decimal
	var result []decimal.Decimal

	if u != nil && u.Get("monthly_data_ingested_gb").Type != gjson.Null {
		monthlyDataIngestedGb = decimalPtr(decimal.NewFromInt(u.Get("monthly_data_ingested_gb").Int()))
		tierLimits := []int{512_000, 1_536_000}
		result = usage.CalculateTierBuckets(*monthlyDataIngestedGb, tierLimits)

		if result[0].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, kinesisFirehoseCostComponent("first 500TB", region, "0", "512000", &result[0]))
		}
		if result[1].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, kinesisFirehoseCostComponent("next 1.5PB", region, "512000", "2048000", &result[1]))
		}
		if result[2].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, kinesisFirehoseCostComponent("next 3PB", region, "2048000", "Inf", &result[2]))
		}
	} else {
		var unknown *decimal.Decimal
		costComponents = append(costComponents, kinesisFirehoseCostComponent("first 500TB", region, "0", "512000", unknown))
	}

	if d.Get("extended_s3_configuration.0.data_format_conversion_configuration.0.enabled").Type != gjson.True {
		costComponents = append(costComponents, kinesisFirehoseConversionCostComponent(region, monthlyDataIngestedGb))
	}

	if d.Get("elasticsearch_configuration").Type != gjson.Null {
		elasticsearchConfiguration := d.Get("elasticsearch_configuration")
		if elasticsearchConfiguration.Get("0.vpc_config").Type != gjson.Null {
			costComponents = append(costComponents, kinesisFirehoseVPCCostComponent(region, monthlyDataIngestedGb))
			vpcConfigs := elasticsearchConfiguration.Get("0.vpc_config")
			zones := decimalPtr(decimal.NewFromInt(int64(len(vpcConfigs.Get("0.subnet_ids").Array()))))
			costComponents = append(costComponents, kinesisFirehoseVPCAZCostComponent(region, zones))
		}
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func kinesisFirehoseCostComponent(tier, region, startUsageAmount, endUsageAmount string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Data ingested (%s)", tier),
		Unit:            "GB",
		UnitMultiplier:  1,
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonKinesisFirehose"),
			ProductFamily: strPtr("Kinesis Firehose"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "group", Value: strPtr("Event-by-Event Processing")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(startUsageAmount),
			EndUsageAmount:   strPtr(endUsageAmount),
		},
	}
}
func kinesisFirehoseConversionCostComponent(region string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Format conversion",
		Unit:            "GB",
		UnitMultiplier:  1,
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonKinesisFirehose"),
			ProductFamily: strPtr("Kinesis Firehose"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "operation", Value: strPtr("DataFormatConversion")},
			},
		},
	}
}
func kinesisFirehoseVPCCostComponent(region string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "VPC data",
		Unit:            "GB",
		UnitMultiplier:  1,
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonKinesisFirehose"),
			ProductFamily: strPtr("Kinesis Firehose"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "operation", Value: strPtr("VpcBandwidth")},
			},
		},
	}
}
func kinesisFirehoseVPCAZCostComponent(region string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "VPC AZ deilvery",
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonKinesisFirehose"),
			ProductFamily: strPtr("Kinesis Firehose"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "operation", Value: strPtr("RunVpcInstance")},
			},
		},
	}
}
