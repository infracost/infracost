package ibm

import (
	"fmt"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// IbmContainerVpcWorkerPool struct represents <TODO: cloud service short description>.
//
// <TODO: Add any important information about the resource and links to the
// pricing pages or documentation that might be useful to developers in the future, e.g:>
//
// Resource information: https://cloud.ibm.com/<PATH/TO/RESOURCE>/
// Pricing information: https://cloud.ibm.com/<PATH/TO/PRICING>/
type IbmContainerVpcWorkerPool struct {
	Address              string
	Region               string
	Flavor               string
	WorkerCount          int64
	ZoneCount            int64
	Entitlement          bool
	MonthlyInstanceHours *float64 `infracost_usage:"monthly_instance_hours"`
}

// IbmContainerVpcWorkerPoolUsageSchema defines a list which represents the usage schema of IbmContainerVpcWorkerPool.
var IbmContainerVpcWorkerPoolUsageSchema = []*schema.UsageItem{
	{Key: "monthly_instance_hours", DefaultValue: 0, ValueType: schema.Float64},
}

// PopulateUsage parses the u schema.UsageData into the IbmContainerVpcWorkerPool.
// It uses the `infracost_usage` struct tags to populate data into the IbmContainerVpcWorkerPool.
func (r *IbmContainerVpcWorkerPool) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid IbmContainerVpcWorkerPool struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *IbmContainerVpcWorkerPool) BuildResource() *schema.Resource {
	var attributeFilters = []*schema.AttributeFilter{
		{Key: "provider", Value: strPtr("vpc-gen2")},
		{Key: "flavor", Value: strPtr(r.Flavor)},
		{Key: "serverType", Value: strPtr("virtual")},
		{Key: "isolation", Value: strPtr("public")},
		{Key: "catalogRegion", Value: strPtr(r.Region)},
	}
	if r.Entitlement {
		attributeFilters = append(attributeFilters, &schema.AttributeFilter{
			Key: "ocpIncluded", Value: strPtr("true"),
		})
	} else {
		attributeFilters = append(attributeFilters, &schema.AttributeFilter{
			Key: "ocpIncluded", Value: strPtr(""),
		})
	}
	WorkerCount := int64(1)
	if r.WorkerCount != 0 {
		WorkerCount = r.WorkerCount
	}
	ZoneCount := int64(1)
	if r.ZoneCount != 0 {
		ZoneCount = r.ZoneCount
	}
	hourlyQuantity := WorkerCount * ZoneCount

	costComponents := []*schema.CostComponent{{
		Name:           fmt.Sprintf("VPC Container Workpool flavor: (%s) region: (%s) workers x zones: (%d) x (%d)", r.Flavor, r.Region, r.WorkerCount, r.ZoneCount),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(hourlyQuantity)),
		ProductFilter: &schema.ProductFilter{
			VendorName:       strPtr("ibm"),
			Region:           strPtr(r.Region),
			Service:          strPtr("containers-kubernetes"),
			AttributeFilters: attributeFilters,
		},
	}}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    IbmContainerVpcWorkerPoolUsageSchema,
		CostComponents: costComponents,
	}
}
