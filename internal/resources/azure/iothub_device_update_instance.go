package azure

import (
	"strings"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

type IotHubDeviceUpdateInstance struct {
	Address     string
	Region      string
	Sku         string
	DeviceCount *int64 `infracost_usage:"device_count"`
}

func (r *IotHubDeviceUpdateInstance) CoreType() string {
	return "IotHubDeviceUpdateInstance"
}

func (r *IotHubDeviceUpdateInstance) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "device_count", ValueType: schema.Int64, DefaultValue: 0},
	}
}

func (r *IotHubDeviceUpdateInstance) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *IotHubDeviceUpdateInstance) BuildResource() *schema.Resource {
	if strings.ToLower(r.Sku) != "standard" {
		return &schema.Resource{
			Name:      r.Address,
			NoPrice:   true,
			IsSkipped: true,
		}
	}

	var monthlyQuantity *decimal.Decimal
	if r.DeviceCount != nil {
		monthlyQuantity = decimalPtr(decimal.NewFromInt(*r.DeviceCount))
	}

	costComponents := []*schema.CostComponent{
		{
			Name:            "Device update devices",
			Unit:            "device",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: monthlyQuantity,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("azure"),
				Region:        strPtr(r.Region),
				Service:       strPtr("Device Update"),
				ProductFamily: strPtr("Internet of Things"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "meterName", ValueRegex: strPtr("/^Device update devices$/i")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("Consumption"),
			},
		},
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}
