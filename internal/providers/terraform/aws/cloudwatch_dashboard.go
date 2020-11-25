package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetCloudwatchDashboardRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_cloudwatch_dashboard",
		RFunc: NewCloudwatchDashboard,
	}
}

func NewCloudwatchDashboard(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Dashboard",
				Unit:            "months",
				UnitMultiplier:  1,
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
	}
}
