package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type CloudwatchDashboard struct {
	Address string
}

func (r *CloudwatchDashboard) CoreType() string {
	return "CloudwatchDashboard"
}

func (r *CloudwatchDashboard) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *CloudwatchDashboard) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *CloudwatchDashboard) BuildResource() *schema.Resource {
	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Dashboard",
				Unit:            "months",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Service:       strPtr("AmazonCloudWatch"),
					ProductFamily: strPtr("Dashboard"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", Value: strPtr("DashboardsUsageHour")},
					},
				},
			},
		},
		UsageSchema: r.UsageSchema(),
	}
}
