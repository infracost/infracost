package aws

import (
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// KinesisStream struct represents Kinesis Data Streams a fully managed, serverless streaming data service
//
// Resource information: https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kinesis_stream
// Pricing information: https://aws.amazon.com/kinesis/data-streams/pricing/
type KinesisStream struct {
	Address    string
	Region     string
	StreamMode string
	ShardCount int64

	// Usage fields
	// On demand
	MonthlyOnDemandDataIngestedGB      *float64 `infracost_usage:"monthly_on_demand_data_in_gb"`
	MonthlyOnDemandDataRetrievalGB     *float64 `infracost_usage:"monthly_on_demand_data_out_gb"`
	MonthlyOnDemandEFODataRetrievalGB  *float64 `infracost_usage:"monthly_on_demand_efo_data_out_gb"`
	MonthlyOnDemandExtendedRetentionGb *float64 `infracost_usage:"monthly_on_demand_extended_retention_gb"`
	MonthlyOnDemandLongTermRetentionGb *float64 `infracost_usage:"monthly_on_demand_long_term_retention_gb"`
	// Provisioned
	MonthlyProvisionedPutUnits            *float64 `infracost_usage:"monthly_provisioned_put_units"`
	MonthlyProvisionedExtendedRetentionGb *float64 `infracost_usage:"monthly_provisioned_extended_retention_gb"`
	MonthlyProvisionedLongTermRetentionGb *float64 `infracost_usage:"monthly_provisioned_long_term_retention_gb"`
	MonthlyProvisionedLongTermRetrievalGb *float64 `infracost_usage:"monthly_provisioned_long_term_retrieval_gb"`
	MonthlyProvisionedEFODataRetrievalGB  *float64 `infracost_usage:"monthly_provisioned_efo_data_out_gb"`
	MonthlyProvisionedEFOConsumerHours    *float64 `infracost_usage:"monthly_provisioned_efo_consumer_hours"`
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
		{Key: "monthly_on_demand_efo_data_out_gb", DefaultValue: 0, ValueType: schema.Float64},
		{Key: "monthly_on_demand_extended_retention_gb", DefaultValue: 0, ValueType: schema.Float64},
		{Key: "monthly_on_demand_long_term_retention_gb", DefaultValue: 0, ValueType: schema.Float64},
		{Key: "monthly_provisioned_put_units", DefaultValue: 0, ValueType: schema.Float64},
		{Key: "monthly_provisioned_extended_retention_gb", DefaultValue: 0, ValueType: schema.Float64},
		{Key: "monthly_provisioned_long_term_retention_gb", DefaultValue: 0, ValueType: schema.Float64},
		{Key: "monthly_provisioned_long_term_retrieval_gb", DefaultValue: 0, ValueType: schema.Float64},
		{Key: "monthly_provisioned_efo_data_out_gb", DefaultValue: 0, ValueType: schema.Float64},
		{Key: "monthly_provisioned_efo_consumer_hours", DefaultValue: 0, ValueType: schema.Float64},
	}
}

// PopulateUsage parses the u schema.UsageData into the KinesisStream.
// It uses the `infracost_usage` struct tags to populate data into the KinesisStream.
func (r *KinesisStream) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// Set some vars that come from the pricing api
var (
	onDemandStreamName    string = "ON_DEMAND"
	provisionedStreamName string = "PROVISIONED"
)

// BuildResource builds a schema.Resource from a valid KinesisStream struct.
// This method is called after the resource is initialized by an IaC provider.
// See providers folder for more information.
func (r *KinesisStream) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{}
	// Depending on the stream mode, we will have different cost components
	if r.StreamMode == onDemandStreamName {
		costComponents = append(costComponents, r.onDemandStreamCostComponent())
		costComponents = append(costComponents, r.onDemandDataIngestedCostComponent())
		costComponents = append(costComponents, r.onDemandDataRetrievalCostComponent())
		costComponents = append(costComponents, r.onDemandEfoDataRetrievalCostComponent())
		costComponents = append(costComponents, r.onDemandExtendedRetentionCostComponent())
		costComponents = append(costComponents, r.onDemandLongTermRetentionCostComponent())
	}
	if r.StreamMode == provisionedStreamName {
		costComponents = append(costComponents, r.provisionedStreamCostComponent())
		costComponents = append(costComponents, r.provisionedStreamPutUnitsCostComponent())
		costComponents = append(costComponents, r.provisionedExtendedRetentionCostComponent())
		costComponents = append(costComponents, r.provisionedLongTermRetentionCostComponent())
		costComponents = append(costComponents, r.provisionedLongTermRetrievalCostComponent())
		costComponents = append(costComponents, r.provisionedEfoDataRetrievalCostComponent())
		costComponents = append(costComponents, r.provisionedEfoConsumersCostComponent())
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *KinesisStream) onDemandStreamCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           onDemandStreamName,
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonKinesis"),
			ProductFamily: strPtr("Kinesis Streams"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: regexPtr("(^|-)OnDemand-StreamHour")},
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
				{Key: "usagetype", ValueRegex: regexPtr("(^|-)OnDemand-BilledIncomingBytes")},
				{Key: "operation", Value: strPtr("OnDemandDataIngested")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
		UsageBased: true,
	}
}

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
				{Key: "usagetype", ValueRegex: regexPtr("(^|-)OnDemand-BilledOutgoingBytes")},
				{Key: "operation", Value: strPtr("OnDemandDataRetrieval")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
		UsageBased: true,
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
				{Key: "usagetype", ValueRegex: regexPtr("(^|-)OnDemand-BilledOutgoingEFOBytes")},
				{Key: "operation", Value: strPtr("OnDemandEFODataRetrieval")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
		UsageBased: true,
	}
}

func (r *KinesisStream) onDemandExtendedRetentionCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Extended retention (24H to 7D)",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: floatPtrToDecimalPtr(r.MonthlyOnDemandExtendedRetentionGb),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonKinesis"),
			ProductFamily: strPtr("Kinesis Streams"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: regexPtr("(^|-)OnDemand-ExtendedRetention-ByteHrs")},
				{Key: "operation", Value: strPtr("OnDemandExtendedRetentionByteHrs")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
		UsageBased: true,
	}
}

func (r *KinesisStream) onDemandLongTermRetentionCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Long term retention (7D+)",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: floatPtrToDecimalPtr(r.MonthlyOnDemandLongTermRetentionGb),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonKinesis"),
			ProductFamily: strPtr("Kinesis Streams"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: regexPtr("(^|-)OnDemand-LongTermRetention-ByteHrs")},
				{Key: "operation", Value: strPtr("OnDemandLongTermRetentionByteHrs")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
		UsageBased: true,
	}
}

func (r *KinesisStream) provisionedStreamCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           provisionedStreamName,
		Unit:           "hours",
		HourlyQuantity: decimalPtr(decimal.NewFromInt(r.ShardCount)),
		UnitMultiplier: decimal.NewFromInt(1),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonKinesis"),
			ProductFamily: strPtr("Kinesis Streams"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: regexPtr("(^|-)Storage-ShardHour")},
				{Key: "operation", Value: strPtr("shardHourStorage")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}

func (r *KinesisStream) provisionedStreamPutUnitsCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Put request unit",
		Unit:            "units",
		MonthlyQuantity: floatPtrToDecimalPtr(r.MonthlyProvisionedPutUnits),
		UnitMultiplier:  decimal.NewFromInt(1),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonKinesis"),
			ProductFamily: strPtr("Kinesis Streams"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: regexPtr("(^|-)PutRequestPayloadUnits")},
				{Key: "operation", Value: strPtr("PutRequest")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
		UsageBased: true,
	}
}

func (r *KinesisStream) provisionedExtendedRetentionCostComponent() *schema.CostComponent {
	return &schema.CostComponent{

		Name:            "Extended retention (24H to 7D)",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: floatPtrToDecimalPtr(r.MonthlyProvisionedExtendedRetentionGb),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonKinesis"),
			ProductFamily: strPtr("Kinesis Streams"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: regexPtr("(^|-)Extended-ShardHour")},
				{Key: "operation", Value: strPtr("shardHourStorage")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
		UsageBased: true,
	}
}

func (r *KinesisStream) provisionedLongTermRetentionCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Long term retention (7D+)",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: floatPtrToDecimalPtr(r.MonthlyProvisionedLongTermRetentionGb),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonKinesis"),
			ProductFamily: strPtr("Kinesis Streams"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: regexPtr("(^|-)LongTermRetention-ByteHrs")},
				{Key: "operation", Value: strPtr("LongTermRetentionByteHrs")},
			},
		},
		UsageBased: true,
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}

func (r *KinesisStream) provisionedLongTermRetrievalCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Extended retention retrieval (7D+)",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: floatPtrToDecimalPtr(r.MonthlyProvisionedLongTermRetrievalGb),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonKinesis"),
			ProductFamily: strPtr("Kinesis Streams"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: regexPtr("(^|-)LongTermRetention-ReadBytes")},
				{Key: "operation", Value: strPtr("LongTermRetentionDataRetrieval")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
		UsageBased: true,
	}
}

func (r *KinesisStream) provisionedEfoDataRetrievalCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Enhanced Fan Out (EFO) Data retrieval",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: floatPtrToDecimalPtr(r.MonthlyProvisionedEFODataRetrievalGB),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonKinesis"),
			ProductFamily: strPtr("Kinesis Streams"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: regexPtr("(^|-)ReadBytes")},
				{Key: "operation", Value: strPtr("EnhancedFanoutDataRetrieval")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
		UsageBased: true,
	}
}

func (r *KinesisStream) provisionedEfoConsumersCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Enhanced Fan Out (EFO)",
		Unit:            "consumer-shard hour",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: floatPtrToDecimalPtr(r.MonthlyProvisionedEFOConsumerHours),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonKinesis"),
			ProductFamily: strPtr("Kinesis Streams"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: regexPtr("(^|-)EnhancedFanoutHour")},
				{Key: "operation", Value: strPtr("ConsumerHour")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
		UsageBased: true,
	}
}
