package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/usage"
)

type APIGatewayRestAPI struct {
	Address         string
	Region          string
	MonthlyRequests *int64 `infracost_usage:"monthly_requests"`
}

func (r *APIGatewayRestAPI) CoreType() string {
	return "APIGatewayRestAPI"
}

func (r *APIGatewayRestAPI) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_requests", ValueType: schema.Int64, DefaultValue: 0},
	}
}

func (r *APIGatewayRestAPI) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *APIGatewayRestAPI) BuildResource() *schema.Resource {
	var costComponents []*schema.CostComponent
	var monthlyRequests *decimal.Decimal

	if r.MonthlyRequests != nil {
		monthlyRequests = decimalPtr(decimal.NewFromInt(*r.MonthlyRequests))

		requestLimits := []int{333000000, 667000000, 19000000000}
		apiRequestQuantities := usage.CalculateTierBuckets(*monthlyRequests, requestLimits)

		costComponents = append(costComponents, r.requestsCostComponent("Requests (first 333M)", "0", &apiRequestQuantities[0]))

		if apiRequestQuantities[1].GreaterThan(decimal.NewFromInt(0)) {
			costComponents = append(costComponents, r.requestsCostComponent("Requests (next 667M)", "333000000", &apiRequestQuantities[1]))
		}

		if apiRequestQuantities[2].GreaterThan(decimal.NewFromInt(0)) {
			costComponents = append(costComponents, r.requestsCostComponent("Requests (next 19B)", "1000000000", &apiRequestQuantities[2]))
		}

		if apiRequestQuantities[3].GreaterThan(decimal.NewFromInt(0)) {
			costComponents = append(costComponents, r.requestsCostComponent("Requests (over 20B)", "20000000000", &apiRequestQuantities[3]))
		}
	} else {
		costComponents = append(costComponents, r.requestsCostComponent("Requests (first 333M)", "0", monthlyRequests))
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *APIGatewayRestAPI) requestsCostComponent(displayName string, usageTier string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            displayName,
		Unit:            "1M requests",
		UnitMultiplier:  decimal.NewFromInt(1000000),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonApiGateway"),
			ProductFamily: strPtr("API Calls"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/ApiGatewayRequest/")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(usageTier),
		},
		UsageBased: true,
	}
}
