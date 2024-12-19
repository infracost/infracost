package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// CloudfrontFunction struct represents an AWS CloudFront Function. With
// CloudFront Functions, you can write lightweight functions in JavaScript
// for high-scale, latency-sensitive CDN customizations.
//
// Resource information: https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/cloudfront-functions.html
// Pricing information: https://aws.amazon.com/cloudfront/pricing/
type CloudfrontFunction struct {
	Address string
	Region  string

	MonthlyRequests *int64 `infracost_usage:"monthly_requests"`
}

// CoreType returns the name of this resource type
func (r *CloudfrontFunction) CoreType() string {
	return "CloudfrontFunction"
}

// UsageSchema defines a list which represents the usage schema of CloudfrontFunction.
func (r *CloudfrontFunction) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "MonthlyRequests", DefaultValue: 0, ValueType: schema.Int64},
	}
}

// PopulateUsage parses the u schema.UsageData into the CloudfrontFunction.
// It uses the `infracost_usage` struct tags to populate data into the CloudfrontFunction.
func (r *CloudfrontFunction) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid CloudfrontFunction struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *CloudfrontFunction) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{}

	costComponents = append(costComponents, r.monthlyRequestsCostComponent())

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *CloudfrontFunction) monthlyRequestsCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Total number of invocations",
		Unit:            "1M invocations",
		UnitMultiplier:  decimal.NewFromInt(1000000),
		MonthlyQuantity: intPtrToDecimalPtr(r.MonthlyRequests),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Service:       strPtr("AmazonCloudFront"),
			ProductFamily: strPtr("Request"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", Value: strPtr("Executions-CloudFrontFunctions")},
				{Key: "groupDescription", ValueRegex: regexPtr("CloudFront Function")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("on_demand"),
			StartUsageAmount: strPtr("2000000"),
		},
		UsageBased: true,
	}
}
