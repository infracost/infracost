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
}

// CoreType returns the name of this resource type
func (r *KinesisStream) CoreType() string {
	return "KinesisStream"
}

// UsageSchema defines a list which represents the usage schema of KinesisStream.
func (r *KinesisStream) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

// PopulateUsage parses the u schema.UsageData into the KinesisStream.
// It uses the `infracost_usage` struct tags to populate data into the KinesisStream.
func (r *KinesisStream) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid KinesisStream struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.

// Set some vars that come from the pricing api
var OnDemandStreamName string = "ON_DEMAND"
var ProvisionedStreamName string = "PROVISIONED"

func (r *KinesisStream) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{}

	// Depending on the stream mode, we will have different cost components
	if r.StreamMode == OnDemandStreamName {
		costComponents = append(costComponents, r.onDemandStreamCostComponent())
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

// TODO - this is not working yet
func (r *KinesisStream) provisionedStreamCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           ProvisionedStreamName,
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonKinesis"),
			ProductFamily: strPtr("Kinesis Streams"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", Value: strPtr("Extended-ShardHour")},
				{Key: "operation", Value: strPtr("shardHourStorage")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}
