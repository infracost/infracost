package aws

import (
	"fmt"
	"github.com/infracost/infracost/internal/schema"
	"strings"

	"github.com/shopspring/decimal"
)

func GetSSMParameterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_ssm_parameter",
		RFunc: NewSSMParameter,
	}
}

func NewSSMParameter(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, parameterStorageCostComponent(d))
	costComponents = append(costComponents, apiThroughputCostComponent(d, u))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func parameterStorageCostComponent(d *schema.ResourceData) *schema.CostComponent {
	region := d.Get("region").String()

	if d.Get("tier").String() == "Standard" {
		return &schema.CostComponent{
			Name:           "Parameter storage - standard",
			Unit:           "Hours",
			UnitMultiplier: 1,
			HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		}
	}

	return &schema.CostComponent{
		Name:           "Parameter storage - advanced",
		Unit:           "Hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AWSSystemsManager"),
			ProductFamily: strPtr("AWS Systems Manager"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/Advanced-Param-Tier1/")},
			},
		},
	}
}

func apiThroughputCostComponent(d *schema.ResourceData, u *schema.UsageData) *schema.CostComponent {
	region := d.Get("region").String()

	var parameterType string
	var tierType *string

	monthlyRequests := decimal.Zero

	if u != nil && u.Get("api_throughput_tier").Exists() {
		parameterType = u.Get("api_throughput_tier").String()
	} else {
		parameterType = d.Get("tier").String()
	}

	if u != nil && u.Get("monthly_requests").Exists() {
		monthlyRequests = decimal.NewFromInt(u.Get("monthly_requests").Int())
	}

	switch parameterType {
	case "Standard":
		return &schema.CostComponent{
			Name:            "API interactions - standard",
			Unit:            "Requests",
			UnitMultiplier:  1,
			MonthlyQuantity: decimalPtr(monthlyRequests),
		}
	case "Advanced":
		tierType = strPtr("/Param-Processed-Tier1/")
	case "Higher":
		tierType = strPtr("/Param-Processed-Tier2/")
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("API interactions - %s", strings.ToLower(parameterType)),
		Unit:            "Requests",
		UnitMultiplier:  10000,
		MonthlyQuantity: decimalPtr(monthlyRequests),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AWSSystemsManager"),
			ProductFamily: strPtr("API Request"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: tierType},
			},
		},
	}
}
