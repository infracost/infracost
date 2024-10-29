package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type AppServiceCertificateBinding struct {
	Address  string
	Region   string
	SSLState string
}

func (r *AppServiceCertificateBinding) CoreType() string {
	return "AppServiceCertificateBinding"
}

func (r *AppServiceCertificateBinding) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *AppServiceCertificateBinding) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *AppServiceCertificateBinding) BuildResource() *schema.Resource {
	region := r.Region

	var sslType string

	sslState := strings.ToUpper(r.SSLState)

	if strings.HasPrefix(sslState, "IP") {
		sslType = "IP"
	} else {

		return &schema.Resource{
			Name:        r.Address,
			NoPrice:     true,
			IsSkipped:   true,
			UsageSchema: r.UsageSchema(),
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
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}
