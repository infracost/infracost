package ibm

import (
	"fmt"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// IsInstance struct represents an IBM virtual server instance.
//
// Pricing information: https://cloud.ibm.com/kubernetes/catalog/about
type IsInstance struct {
	Address              string
	Region               string
	Profile              string // should be values from CLI 'ibmcloud is instance-profiles'
	TruncatedProfile     string
	Zone                 string
	TruncatedZone        string // should be the same as region, but with the last number removed (eg: us-south-1 -> us-south)
	IsDedicated          bool   // will be true if a dedicated_host or dedicated_host_group is specified
	MonthlyInstanceHours *int64 `infracost_usage:"monthly_instance_hours"`
}

var IsInstanceUsageSchema = []*schema.UsageItem{
	{Key: "monthly_instance_hours", DefaultValue: 0, ValueType: schema.Float64},
}

// PopulateUsage parses the u schema.UsageData into the IsInstance.
// It uses the `infracost_usage` struct tags to populate data into the IsInstance.
func (r *IsInstance) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid IsInstance struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *IsInstance) BuildResource() *schema.Resource {

	costComponents := []*schema.CostComponent{
		r.instanceRunTimeCostComponent(),
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    IsInstanceUsageSchema,
		CostComponents: costComponents,
	}
}

// charge for the dedicated host instead of the instance, if instance
// is found to be configured to run on a dedicated host
func instanceGetIsolation(isDedicated bool) string {
	if isDedicated {
		return "private"
	} else {
		return "public"
	}
}

func (r *IsInstance) instanceRunTimeCostComponent() *schema.CostComponent {
	var quantity *decimal.Decimal
	if r.MonthlyInstanceHours != nil {
		quantity = decimalPtr(decimal.NewFromInt(*r.MonthlyInstanceHours))
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("Instance usage (%s, %s, %s)", r.Profile, instanceGetIsolation(r.IsDedicated), r.Zone),
		Unit:            "hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		HourlyQuantity:  decimalPtr(decimal.NewFromInt(1)),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Region),
			Service:    strPtr("containers-kubernetes"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "flavor", ValueRegex: regexPtr(fmt.Sprintf("%s$", r.TruncatedProfile))},
				{Key: "isolation", ValueRegex: regexPtr(fmt.Sprintf("^%s$", instanceGetIsolation(r.IsDedicated)))},
			},
		},
	}
}
