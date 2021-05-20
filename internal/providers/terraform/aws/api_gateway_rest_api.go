package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
	"github.com/shopspring/decimal"
)

func GetAPIGatewayRestAPIRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_api_gateway_rest_api",
		RFunc: NewAPIGatewayRestAPI,
	}
}

func NewAPIGatewayRestAPI(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	var costComponents []*schema.CostComponent
	var monthlyRequests *decimal.Decimal

	if u != nil && u.Get("monthly_requests").Exists() {
		monthlyRequests = decimalPtr(decimal.NewFromInt(u.Get("monthly_requests").Int()))

		requestLimits := []int{333000000, 667000000, 19000000000}
		apiRequestQuantities := usage.CalculateTierBuckets(*monthlyRequests, requestLimits)

		costComponents = append(costComponents, restAPICostComponent(region, "Requests (first 333M)", "0", &apiRequestQuantities[0]))

		if apiRequestQuantities[1].GreaterThan(decimal.NewFromInt(0)) {
			costComponents = append(costComponents, restAPICostComponent(region, "Requests (next 667M)", "333000000", &apiRequestQuantities[1]))
		}

		if apiRequestQuantities[2].GreaterThan(decimal.NewFromInt(0)) {
			costComponents = append(costComponents, restAPICostComponent(region, "Requests (next 19B)", "1000000000", &apiRequestQuantities[2]))
		}

		if apiRequestQuantities[3].GreaterThan(decimal.NewFromInt(0)) {
			costComponents = append(costComponents, restAPICostComponent(region, "Requests (over 20B)", "20000000000", &apiRequestQuantities[3]))
		}
	} else {
		costComponents = append(costComponents, restAPICostComponent(region, "Requests (first 333M)", "0", monthlyRequests))
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func restAPICostComponent(region string, displayName string, usageTier string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            displayName,
		Unit:            "1M requests",
		UnitMultiplier:  1000000,
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonApiGateway"),
			ProductFamily: strPtr("API Calls"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/ApiGatewayRequest/")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(usageTier),
		},
	}
}
