package aws

import (
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// CloudHSMv2HSM struct represents an HSM module in a CloudHSM cluster.
//
// The HSM module is charged a hourly rate. Terraform allows you to specify the instance
// type of the HSM cluster, but at the moment AWS only supports one instance type, so
// each module has a set price depending on region.
//
// The cluster is counted as a free resource, but each cluster can have up to 32 HSM modules.
//
// Resource information: https://aws.amazon.com/cloudhsm/
// Pricing information: https://aws.amazon.com/cloudhsm/pricing/
type CloudHSMv2HSM struct {
	Address string
	Region  string

	MonthlyHours *float64 `infracost_usage:"monthly_hrs"`
}

// CoreType returns the name of this resource type
func (r *CloudHSMv2HSM) CoreType() string {
	return "CloudHSMv2HSM"
}

// UsageSchema defines a list which represents the usage schema of CloudHSMv2HSM.
func (r *CloudHSMv2HSM) UsageSchema() []*schema.UsageItem {
	hours, _ := schema.HourToMonthUnitMultiplier.Float64()

	return []*schema.UsageItem{
		{Key: "monthly_hrs", DefaultValue: hours, ValueType: schema.Float64},
	}
}

// PopulateUsage parses the u schema.UsageData into the CloudHSMv2HSM.
// It uses the `infracost_usage` struct tags to populate data into the CloudHSMv2HSM.
func (r *CloudHSMv2HSM) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid CloudHSMv2HSM struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *CloudHSMv2HSM) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		r.hsmCostComponent(),
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *CloudHSMv2HSM) hsmCostComponent() *schema.CostComponent {
	quantity := schema.HourToMonthUnitMultiplier
	if r.MonthlyHours != nil {
		quantity = decimal.NewFromFloat(*r.MonthlyHours)
	}

	return &schema.CostComponent{
		Name:            "HSM usage",
		Unit:            "hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(quantity),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("CloudHSM"),
			ProductFamily: strPtr("Dedicated-Host"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "instanceFamily", Value: strPtr("CloudHSM-v2")},
				{Key: "usagetype", ValueRegex: regexPtr("CloudHSMv2Usage$")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
		UsageBased: true,
	}
}
