package base

import (
	"github.com/shopspring/decimal"
)

var HoursInMonth = 730

type PriceComponentCost struct {
	PriceComponent PriceComponent
	HourlyCost     decimal.Decimal
	MonthlyCost    decimal.Decimal
}

type ResourceCostBreakdown struct {
	Resource            Resource
	PriceComponentCosts []PriceComponentCost
}

func GetCostBreakdown(resource Resource) (ResourceCostBreakdown, error) {
	queries := make([]GraphQLQuery, 0, len(resource.PriceComponents()))
	queriedPriceComponents := make([]PriceComponent, 0, len(queries))

	for _, priceComponent := range resource.PriceComponents() {
		if priceComponent.ShouldSkip() {
			continue
		}

		queries = append(queries, BuildQuery(priceComponent.GetFilters()))
		queriedPriceComponents = append(queriedPriceComponents, priceComponent)
	}

	queryResults, err := GetQueryResults(queries)
	if err != nil {
		return ResourceCostBreakdown{}, err
	}

	priceComponentCosts := make([]PriceComponentCost, 0)
	for i, queryResult := range queryResults {
		priceComponent := queriedPriceComponents[i]

		priceStr := queryResult.Get("data.products.0.onDemandPricing.0.priceDimensions.0.pricePerUnit.USD").String()
		price, _ := decimal.NewFromString(priceStr)

		hourlyCost := priceComponent.CalculateHourlyCost(price)

		priceComponentCosts = append(priceComponentCosts, PriceComponentCost{
			PriceComponent: priceComponent,
			HourlyCost:     hourlyCost,
			MonthlyCost:    hourlyCost.Mul(decimal.NewFromInt(int64(HoursInMonth))),
		})
	}

	return ResourceCostBreakdown{
		Resource:            resource,
		PriceComponentCosts: priceComponentCosts,
	}, nil
}

func GetCostBreakdowns(resources []Resource) ([]ResourceCostBreakdown, error) {
	costBreakdowns := make([]ResourceCostBreakdown, 0, len(resources))

	for _, resource := range resources {
		costBreakdown, err := GetCostBreakdown(resource)
		if err != nil {
			return costBreakdowns, err
		}
		costBreakdowns = append(costBreakdowns, costBreakdown)
	}

	return costBreakdowns, nil
}
