package aws

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetKenesisFirehoseDeliveryStreamRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_kinesis_firehose_delivery_stream",
		RFunc: NewKenesisFirehoseDeliveryStream,
	}
}

func NewKenesisFirehoseDeliveryStream(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	costComponents := make([]*schema.CostComponent, 0)
	var monthlyDataIngestedGb *decimal.Decimal
	var requestQuantities []decimal.Decimal
	halfAMillion := []int{512000}
	oneAndHalfPB := decimal.NewFromInt(1536000)

	if u != nil && u.Get("monthly_data_ingested_gb").Type != gjson.Null {
		monthlyDataIngestedGb = decimalPtr(decimal.NewFromInt(u.Get("monthly_data_ingested_gb").Int()))
		requestQuantities = usage.CalculateTierBuckets(*monthlyDataIngestedGb, halfAMillion)
		firstGB := requestQuantities[0]
		nextOneAndHalfGB := requestQuantities[1]

		if nextOneAndHalfGB.Equal(decimal.Zero) {

			costComponents = append(costComponents, kenesisFirehosCostComponent("first 500TB", region, "0", "512000", &firstGB))
		}
		if nextOneAndHalfGB.GreaterThan(oneAndHalfPB) {
			nextThreeGB := nextOneAndHalfGB.Sub(oneAndHalfPB)
			costComponents = append(costComponents, kenesisFirehosCostComponent("first 500TB", region, "0", "512000", decimalPtr(decimal.NewFromInt(int64(halfAMillion[0])))))
			costComponents = append(costComponents, kenesisFirehosCostComponent("next 1.5PB", region, "512000", "2048000", &oneAndHalfPB))
			costComponents = append(costComponents, kenesisFirehosCostComponent("next 3PB", region, "2048000", "Inf", &nextThreeGB))
		} else {
			costComponents = append(costComponents, kenesisFirehosCostComponent("first 500TB", region, "0", "512000", decimalPtr(decimal.NewFromInt(int64(halfAMillion[0])))))
			costComponents = append(costComponents, kenesisFirehosCostComponent("next 1.5PB", region, "512000", "2048000", &nextOneAndHalfGB))
		}
	} else {
		var unknown *decimal.Decimal

		costComponents = append(costComponents, kenesisFirehosCostComponent("first 500TB", region, "0", "512000", unknown))
		costComponents = append(costComponents, kenesisFirehosCostComponent("next 1.5PB", region, "512000", "2048000", unknown))
		costComponents = append(costComponents, kenesisFirehosCostComponent("next 3PB", region, "2048000", "Inf", unknown))
	}

	if d.Get("extended_s3_configuration.0.data_format_conversion_configuration.0.enabled").Type != gjson.Null {
		enabled := d.Get("extended_s3_configuration.0.data_format_conversion_configuration.0.enabled")
		if enabled.Type == gjson.True {
			costComponents = append(costComponents, kenesisFirehosConversationCostComponent(region, monthlyDataIngestedGb))
		}
	}

	if d.Get("elasticsearch_configuration").Type != gjson.Null {
		elasticsearchConfiguration := d.Get("elasticsearch_configuration")
		if elasticsearchConfiguration.Get("0.vpc_config").Type != gjson.Null {
			costComponents = append(costComponents, kenesisFirehosVPCCostComponent("VPC data", region, "VpcBandwidth", monthlyDataIngestedGb))
			vpcConfigs := elasticsearchConfiguration.Get("0.vpc_config")
			zones := decimalPtr(decimal.NewFromInt(int64(len(vpcConfigs.Get("0.subnet_ids").Array()))))
			costComponents = append(costComponents, kenesisFirehosVPCCostComponent("VPC AZ deilvery", region, "RunVpcInstance", zones))
		}
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func kenesisFirehosCostComponent(tier, region, startUsageAmount, endUsageAmount string, quantity *decimal.Decimal) *schema.CostComponent {
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
func kenesisFirehosConversationCostComponent(region string, quantity *decimal.Decimal) *schema.CostComponent {
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
func kenesisFirehosVPCCostComponent(name, region, operation string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "GB",
		UnitMultiplier:  1,
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonKinesisFirehose"),
			ProductFamily: strPtr("Kinesis Firehose"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "operation", ValueRegex: strPtr(fmt.Sprintf("/%s/", operation))},
			},
		},
	}
}
