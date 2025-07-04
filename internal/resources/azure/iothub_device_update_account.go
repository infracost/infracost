package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

type IotHubDeviceUpdateAccount struct {
	Address              string
	Region               string
	HasStandardInstance  *bool `infracost_usage:"has_standard_instance"`
}

func (r *IotHubDeviceUpdateAccount) CoreType() string {
	return "IotHubDeviceUpdateAccount"
}

func (r *IotHubDeviceUpdateAccount) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "has_standard_instance", ValueType: schema.Bool, DefaultValue: false},
	}
}

func (r *IotHubDeviceUpdateAccount) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *IotHubDeviceUpdateAccount) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{}

	if r.HasStandardInstance != nil && *r.HasStandardInstance {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:           "Device update tenants",
			Unit:           "tenant",
			UnitMultiplier: decimal.NewFromInt(1),
			HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("azure"),
				Region:        strPtr(r.Region),
				Service:       strPtr("Device Update"),
				ProductFamily: strPtr("Internet of Things"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "meterName", ValueRegex: strPtr("/^Device update tenants$/i")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("Consumption"),
			},
		})
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}
