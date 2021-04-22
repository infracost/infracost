package azure

import (
	"fmt"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"strings"
)

func GetAzureRMAppServiceCertificateOrderRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_app_service_certificate_order",
		RFunc: NewAzureRMAppServiceCertificateOrder,
	}
}

func NewAzureRMAppServiceCertificateOrder(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	location := "Global"
	if strings.HasPrefix(location, "usgov") {
		location = "US Gov"
	}

	productType := "Standard"
	if d.Get("product_type").Type != gjson.Null {
		productType = d.Get("product_type").String()
	}

	costComponents := []*schema.CostComponent{
		{
			Name:           fmt.Sprintf("SSL certificate (%s)", productType),
			Unit:           "years",
			UnitMultiplier: 1,
			// Convert yearly price to monthly
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(1).Div(decimal.NewFromInt(12))),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("azure"),
				Region:        strPtr(location),
				Service:       strPtr("Azure App Service"),
				ProductFamily: strPtr("Compute"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "skuName", ValueRegex: strPtr(fmt.Sprintf("/%s SSL - 1 Year/i", productType))},
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
