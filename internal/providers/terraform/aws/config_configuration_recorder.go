package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetConfigurationRecorderItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_config_configuration_recorder",
		RFunc: NewConfigurationRecorder,
	}
}

func NewConfigurationRecorder(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	var monthlyConfigItems *decimal.Decimal
	if u != nil && u.Get("monthly_config_items").Exists() {
		monthlyConfigItems = decimalPtr(decimal.NewFromInt(u.Get("monthly_config_items").Int()))
	}

	var monthlyCustomConfigItems *decimal.Decimal
	if u != nil && u.Get("monthly_custom_config_items").Exists() {
		monthlyCustomConfigItems = decimalPtr(decimal.NewFromInt(u.Get("monthly_custom_config_items").Int()))
	}

	costComponents := []*schema.CostComponent{}

	costComponents = append(costComponents, &schema.CostComponent{
		Name:            "Config items",
		Unit:            "records",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: monthlyConfigItems,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AWSConfig"),
			ProductFamily: strPtr("Management Tools - AWS Config"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/ConfigurationItemRecorded/")},
			},
		},
	})

	costComponents = append(costComponents, &schema.CostComponent{
		Name:            "Custom config items",
		Unit:            "records",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: monthlyCustomConfigItems,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AWSConfig"),
			ProductFamily: strPtr("Management Tools - AWS Config"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/CustomConfigItemRecorded/")},
			},
		},
	})

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
