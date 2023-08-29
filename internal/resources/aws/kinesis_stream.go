package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// KinesisStream struct represents <TODO: cloud service short description>.
//
// <TODO: Add any important information about the resource and links to the
// pricing pages or documentation that might be useful to developers in the future, e.g:>
//
// Resource information: https://aws.amazon.com/<PATH/TO/RESOURCE>/
// Pricing information: https://aws.amazon.com/<PATH/TO/PRICING>/
type KinesisStream struct {
	Address    string
	Region     string
	StreamMode string
	ShardCount int64

	// Usage fields
	MonthlyOnDemandDataIngestedGB      *float64 `infracost_usage:"monthly_on_demand_data_in_gb"`
	MonthlyOnDemandDataRetrievalGB     *float64 `infracost_usage:"monthly_on_demand_data_out_gb"`
	ConsumerApplicationCount           *int64   `infracost_usage:"consumer_application_count"`
	MonthlyOnDemandEFODataRetrievalGB  *float64 `infracost_usage:"monthly_on_demand_efo_data_out_gb"`
	MonthlyOnDemandExtendedRetentionGb *float64 `infracost_usage:"monthly_on_demand_extended_retention_gb"`
	MonthlyOnDemandLongTermRetentionGb *float64 `infracost_usage:"monthly_on_demand_long_term_retention_gb"`
}

// CoreType returns the name of this resource type
func (r *KinesisStream) CoreType() string {
	return "KinesisStream"
}

// UsageSchema defines a list which represents the usage schema of KinesisStream.
func (r *KinesisStream) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_on_demand_data_in_gb", DefaultValue: 0, ValueType: schema.Float64},
		{Key: "monthly_on_demand_data_out_gb", DefaultValue: 0, ValueType: schema.Float64},
		{Key: "consumer_application_count", DefaultValue: 1, ValueType: schema.Int64},
		{Key: "monthly_on_demand_efo_data_out_gb", DefaultValue: 0, ValueType: schema.Float64},
		{Key: "monthly_on_demand_extended_retention_gb", DefaultValue: 0, ValueType: schema.Float64},
		{Key: "monthly_on_demand_long_term_retention_gb", DefaultValue: 0, ValueType: schema.Float64},
	}
}

// PopulateUsage parses the u schema.UsageData into the KinesisStream.
// It uses the `infracost_usage` struct tags to populate data into the KinesisStream.
func (r *KinesisStream) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid KinesisStream struct.
// This method is called after the resource is initialized by an IaC provider.
// See providers folder for more information.

// Set some vars that come from the pricing api
var OnDemandStreamName string = "ON_DEMAND"
var ProvisionedStreamName string = "PROVISIONED"

func (r *KinesisStream) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{}
	// Depending on the stream mode, we will have different cost components
	if r.StreamMode == OnDemandStreamName {
		costComponents = append(costComponents, r.onDemandStreamCostComponent())
		costComponents = append(costComponents, r.onDemandDataIngestedCostComponent())
		costComponents = append(costComponents, r.onDemandDataRetrievalCostComponent())
		costComponents = append(costComponents, r.onDemandEfoDataRetrievalCostComponent())
		costComponents = append(costComponents, r.onDemandExtendedRetentionCostComponent())
		costComponents = append(costComponents, r.onDemandLongTermRetentionCostComponent())
	} else if r.StreamMode == ProvisionedStreamName {
		costComponents = append(costComponents, r.provisionedStreamCostComponent())
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *KinesisStream) onDemandStreamCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           OnDemandStreamName,
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonKinesis"),
			ProductFamily: strPtr("Kinesis Streams"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", Value: strPtr("OnDemand-StreamHour")},
				{Key: "operation", Value: strPtr("OnDemandStreamHr")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}

func (r *KinesisStream) onDemandDataIngestedCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Data ingested",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: floatPtrToDecimalPtr(r.MonthlyOnDemandDataIngestedGB),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonKinesis"),
			ProductFamily: strPtr("Kinesis Streams"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", Value: strPtr("OnDemand-BilledIncomingBytes")},
				{Key: "operation", Value: strPtr("OnDemandDataIngested")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}

// TODO Can we * the UnitMultiplier by ConsumerApplicationCount to get the correct price for multiple consumers?
// In the test_usage.yaml
// See how I did the provisioned stream shard count
func (r *KinesisStream) onDemandDataRetrievalCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Data retrieval",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: floatPtrToDecimalPtr(r.MonthlyOnDemandDataRetrievalGB),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonKinesis"),
			ProductFamily: strPtr("Kinesis Streams"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", Value: strPtr("OnDemand-BilledOutgoingBytes")},
				{Key: "operation", Value: strPtr("OnDemandDataRetrieval")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}

func (r *KinesisStream) onDemandEfoDataRetrievalCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Enhanced Fan Out (EFO) Data retrieval",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: floatPtrToDecimalPtr(r.MonthlyOnDemandEFODataRetrievalGB),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonKinesis"),
			ProductFamily: strPtr("Kinesis Streams"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", Value: strPtr("OnDemand-BilledOutgoingEFOBytes")},
				{Key: "operation", Value: strPtr("OnDemandEFODataRetrieval")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}

func (r *KinesisStream) onDemandExtendedRetentionCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Extended retention 24H to 7D)",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: floatPtrToDecimalPtr(r.MonthlyOnDemandExtendedRetentionGb),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonKinesis"),
			ProductFamily: strPtr("Kinesis Streams"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", Value: strPtr("OnDemand-ExtendedRetention-ByteHrs")},
				{Key: "operation", Value: strPtr("OnDemandExtendedRetentionByteHrs")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}

func (r *KinesisStream) onDemandLongTermRetentionCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Extended retention 7D+",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: floatPtrToDecimalPtr(r.MonthlyOnDemandLongTermRetentionGb),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonKinesis"),
			ProductFamily: strPtr("Kinesis Streams"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", Value: strPtr("OnDemand-LongTermRetention-ByteHrs")},
				{Key: "operation", Value: strPtr("OnDemandLongTermRetentionByteHrs")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}

func (r *KinesisStream) provisionedStreamCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           ProvisionedStreamName,
		Unit:           "hours",
		HourlyQuantity: decimalPtr(decimal.NewFromInt(r.ShardCount)),
		UnitMultiplier: decimal.NewFromInt(1),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonKinesis"),
			ProductFamily: strPtr("Kinesis Streams"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", Value: strPtr("Storage-ShardHour")},
				{Key: "operation", Value: strPtr("shardHourStorage")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}
