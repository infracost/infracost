package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetAzureRMAppServiceCustomHostnameBindingRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_app_service_custom_hostname_binding",
		RFunc: NewAzureRMAppServiceCustomHostnameBinding,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func NewAzureRMAppServiceCustomHostnameBinding(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var sslType, sslState string
	region := "Global"
	group := d.References("resource_group_name")
	if len(group) > 0 {
		region = group[0].Get("location").String()
	}
	if d.Get("ssl_state").Type != gjson.Null {
		sslState = d.Get("ssl_state").String()
	}

	// The two approved values are IpBasedEnabled or SniEnabled
	sslState = strings.ToUpper(sslState)

	if strings.HasPrefix(sslState, "IP") {
		sslType = "IP"
	} else {
		// returning directly since SNI is currently defined as free in the Azure cost page
		return &schema.Resource{
			Name:      d.Address,
			NoPrice:   true,
			IsSkipped: true,
		}
	}

	var instanceCount int64 = 1

	costComponents := []*schema.CostComponent{
		{
			Name:            "IP SSL certificate",
			Unit:            "months",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(instanceCount)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("azure"),
				Region:        strPtr(region),
				Service:       strPtr("Azure App Service"),
				ProductFamily: strPtr("Compute"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "skuName", Value: strPtr(fmt.Sprintf("%s SSL", sslType))},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("Consumption"),
			},
		},
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
