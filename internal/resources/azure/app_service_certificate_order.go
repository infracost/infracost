package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type AppServiceCertificateOrder struct {
	Address     string
	ProductType string
}

func (r *AppServiceCertificateOrder) CoreType() string {
	return "AppServiceCertificateOrder"
}

func (r *AppServiceCertificateOrder) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *AppServiceCertificateOrder) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *AppServiceCertificateOrder) BuildResource() *schema.Resource {
	region := "Global"

	if strings.HasPrefix(region, "usgov") {
		region = "US Gov"
	}

	productType := "Standard"
	if r.ProductType != "" {
		productType = r.ProductType
	}
	productType = strings.ToLower(productType)

	costComponents := []*schema.CostComponent{
		{
			Name:           fmt.Sprintf("SSL certificate (%s)", productType),
			Unit:           "years",
			UnitMultiplier: decimal.NewFromInt(1),

			MonthlyQuantity: decimalPtr(decimal.NewFromInt(1).Div(decimal.NewFromInt(12))),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("azure"),
				Region:        strPtr(region),
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
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}
