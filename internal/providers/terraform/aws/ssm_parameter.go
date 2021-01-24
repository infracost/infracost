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

	costComponents = append(costComponents, parameterStorageCostComponent(d, u))
	costComponents = append(costComponents, apiThroughputCostComponent(d, u))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func parameterStorageCostComponent(d *schema.ResourceData, u *schema.UsageData) *schema.CostComponent {
	region := d.Get("region").String()

	var parameterStorageHours decimal.Decimal

	if d.Get("tier").String() == "Standard" {
		return nil
	}

	if u != nil && u.Get("parameter_storage_hours").Exists() {
		parameterStorageHours = decimal.NewFromInt(u.Get("parameter_storage_hours").Int())
	} else {
		parameterStorageHours = decimal.NewFromInt(730)
	}

	return &schema.CostComponent{
		Name:           "Parameter storage (advanced)",
		Unit:           "Parameter-hours",
		UnitMultiplier: 1,
		MonthlyQuantity: decimalPtr(parameterStorageHours),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AWSSystemsManager"),
			ProductFamily: strPtr("AWS Systems Manager"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/PS-Advanced-Param-Tier1/")},
			},
		},
	}
}

func apiThroughputCostComponent(d *schema.ResourceData, u *schema.UsageData) *schema.CostComponent {
	region := d.Get("region").String()

	var parameterType string
	var tierType *string

	monthlyApiInteractions := decimal.Zero

	if u != nil && u.Get("api_throughput_tier").Exists() {
		parameterType = u.Get("api_throughput_tier").String()
	} else {
		parameterType = d.Get("tier").String()
	}

	if u != nil && u.Get("monthly_api_interactions").Exists() {
		monthlyApiInteractions = decimal.NewFromInt(u.Get("monthly_api_interactions").Int())
	}

	switch parameterType {
	case "Advanced", "Higher":
		tierType = strPtr("/PS-Param-Processed-Tier2/")
	default:
		return nil
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("API interactions (%s)", strings.ToLower(parameterType)),
		Unit:            "Interactions",
		UnitMultiplier:  10000,
		MonthlyQuantity: decimalPtr(monthlyApiInteractions),
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
