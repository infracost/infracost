package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type ConfigConfigurationRecorder struct {
	Address                  string
	Region                   string
	MonthlyConfigItems       *int64 `infracost_usage:"monthly_config_items"`
	MonthlyCustomConfigItems *int64 `infracost_usage:"monthly_custom_config_items"`
}

func (r *ConfigConfigurationRecorder) CoreType() string {
	return "ConfigConfigurationRecorder"
}

func (r *ConfigConfigurationRecorder) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_config_items", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_custom_config_items", ValueType: schema.Int64, DefaultValue: 0},
	}
}

func (r *ConfigConfigurationRecorder) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ConfigConfigurationRecorder) BuildResource() *schema.Resource {
	var monthlyConfigItems *decimal.Decimal
	if r.MonthlyConfigItems != nil {
		monthlyConfigItems = decimalPtr(decimal.NewFromInt(*r.MonthlyConfigItems))
	}

	var monthlyCustomConfigItems *decimal.Decimal
	if r.MonthlyCustomConfigItems != nil {
		monthlyCustomConfigItems = decimalPtr(decimal.NewFromInt(*r.MonthlyCustomConfigItems))
	}

	costComponents := []*schema.CostComponent{}

	costComponents = append(costComponents, &schema.CostComponent{
		Name:            "Config items",
		Unit:            "records",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: monthlyConfigItems,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AWSConfig"),
			ProductFamily: strPtr("Management Tools - AWS Config"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: regexPtr("ConfigurationItemRecorded$")},
			},
		},
		UsageBased: true,
	})

	costComponents = append(costComponents, &schema.CostComponent{
		Name:            "Custom config items",
		Unit:            "records",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: monthlyCustomConfigItems,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AWSConfig"),
			ProductFamily: strPtr("Management Tools - AWS Config"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: regexPtr("CustomConfigItemRecorded$")},
			},
		},
		UsageBased: true,
	})

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}
