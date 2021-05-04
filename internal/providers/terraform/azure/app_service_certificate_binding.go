package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetAzureRMAppServiceCertificateBindingRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_app_service_certificate_binding",
		RFunc: NewAzureRMAppServiceCertificateBinding,
	}
}

func NewAzureRMAppServiceCertificateBinding(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {

	var sslType string
	sslState := d.Get("ssl_state").String()

	// The two approved values are IpBasedEnabled or SniEnabled
	sslState = strings.ToUpper(sslState)[0:2]

	if sslState == "IP" {
		sslType = "IP"
	} else {
		sslType = "SNI"

		// returning directly since SNI is currently defined as free in the Azure cost page
		return &schema.Resource{
			NoPrice:   true,
			IsSkipped: true,
		}
	}

	var instanceCount int64 = 1

	costComponents := []*schema.CostComponent{
		{
			Name:            "IP SSL certificate",
			Unit:            "months",
			UnitMultiplier:  1,
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(instanceCount)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("azure"),
				Region:        strPtr("Global"),
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
