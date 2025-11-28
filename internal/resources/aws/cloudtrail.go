package aws

import (
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

var (
	cloudTrailServiceName = strPtr("AWSCloudTrail")

	cloudTrailManagementEvent = "Management events (additional copies)"
	cloudTrailDataEvent       = "Data events"
	cloudTrailInsightEvent    = "Insight events"

	cloudTrailBillingMultiplier = decimal.NewFromInt(100000)
)

// Cloudtrail struct represents a cloudtrail instance to monitor activity across a set of resources.
// AWS Cloudtrail monitors and records account activity across infrastructure, keeping an audit log of activity.
// This is mostly used for security purposes.
//
// Resource information: https://aws.amazon.com/cloudtrail/
// Pricing information: https://aws.amazon.com/cloudtrail/pricing/
type Cloudtrail struct {
	Address                 string
	Region                  string
	IncludeManagementEvents bool
	IncludeInsightEvents    bool

	MonthlyAdditionalManagementEvents *float64 `infracost_usage:"monthly_additional_management_events"`
	MonthlyDataEvents                 *float64 `infracost_usage:"monthly_data_events"`
	MonthlyInsightEvents              *float64 `infracost_usage:"monthly_insight_events"`
}

func (r *Cloudtrail) CoreType() string {
	return "Cloudtrail"
}

func (r *Cloudtrail) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_additional_management_events", DefaultValue: 0, ValueType: schema.Float64},
		{Key: "monthly_data_events", DefaultValue: 0, ValueType: schema.Float64},
		{Key: "monthly_insight_events", DefaultValue: 0, ValueType: schema.Float64},
	}
}

// PopulateUsage parses the u schema.UsageData into the Cloudtrail.
// It uses the `infracost_usage` struct tags to populate data into the Cloudtrail.
func (r *Cloudtrail) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid Cloudtrail struct.
// It returns Cloudtrail as a schema.Resource with 3 main cost components. All cost components are defined as "events".
// All cost components are charged per 100k events delivered/analyzed.
//
//  1. Additional Management events delivered to S3, charged at $2.00 per 100k management events delivered.
//     Management events are normally priced as free, however if a user specifies an additional replication of events
//     this is charged. We only show this cost therefore if Cloudtrail.IncludeManagementEvents is set. This is set at
//     a per IAC basis.
//  2. Data events delivered to S3, charged at $0.10 per 100k events delivered.
//  3. CloudTrail Insights, charged at $0.35 per 100k events analyzed. This again is configured optionally on a Cloudtrail
//     instance. Hence, we only include the cost component if Cloudtrail.IncludeInsightEvents. This is set at
//     a per IAC basis.
//
// This method is called after the resource is initialised by an IaC provider. See providers folder for more information.
func (r *Cloudtrail) BuildResource() *schema.Resource {
	var costComponents []*schema.CostComponent

	if r.IncludeManagementEvents || r.MonthlyAdditionalManagementEvents != nil {
		costComponents = append(costComponents, r.managementEventCostComponent())
	}

	costComponents = append(costComponents, r.dataEventsCostComponent())

	if r.IncludeInsightEvents || r.MonthlyInsightEvents != nil {
		costComponents = append(costComponents, r.insightEventsCostComponent())
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *Cloudtrail) managementEventCostComponent() *schema.CostComponent {
	var quantity *decimal.Decimal
	if r.MonthlyAdditionalManagementEvents != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*r.MonthlyAdditionalManagementEvents))
	}

	return r.eventCostComponent(cloudTrailManagementEvent, quantity)
}

func (r *Cloudtrail) dataEventsCostComponent() *schema.CostComponent {
	var quantity *decimal.Decimal
	if r.MonthlyDataEvents != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*r.MonthlyDataEvents))
	}

	return r.eventCostComponent(cloudTrailDataEvent, quantity)
}

func (r *Cloudtrail) insightEventsCostComponent() *schema.CostComponent {
	var quantity *decimal.Decimal
	if r.MonthlyInsightEvents != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*r.MonthlyInsightEvents))
	}

	return r.eventCostComponent(cloudTrailInsightEvent, quantity)
}

func (r *Cloudtrail) eventCostComponent(name string, quantity *decimal.Decimal) *schema.CostComponent {
	productFamily := "Management Tools - AWS CloudTrail Paid Events Recorded"
	if name == cloudTrailDataEvent {
		productFamily = "Management Tools - AWS CloudTrail Data Events Recorded"
	}

	var attrFilters []*schema.AttributeFilter
	if name == cloudTrailInsightEvent {
		productFamily = "Management Tools - AWS CloudTrail Insights Events"
		attrFilters = []*schema.AttributeFilter{
			{Key: "usagetype", ValueRegex: regexPtr(".*-InsightsEvents$")},
		}
	}

	return &schema.CostComponent{
		Name:            name,
		Unit:            "100k events",
		UnitMultiplier:  cloudTrailBillingMultiplier,
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:       vendorName,
			Region:           strPtr(r.Region),
			Service:          cloudTrailServiceName,
			ProductFamily:    strPtr(productFamily),
			AttributeFilters: attrFilters,
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
		UsageBased: true,
	}
}
