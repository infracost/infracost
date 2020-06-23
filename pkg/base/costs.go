package base

import (
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
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
	SubResourceCosts    []ResourceCostBreakdown
}

type PriceComponentQueryMap struct {
	Resource       Resource
	PriceComponent PriceComponent
	Query          GraphQLQuery
}

type queryKey struct {
	Resource       Resource
	PriceComponent PriceComponent
}

func createPriceComponentCost(priceComponent PriceComponent, queryResult gjson.Result) PriceComponentCost {
	hourlyCost := priceComponent.HourlyCost()

	return PriceComponentCost{
		PriceComponent: priceComponent,
		HourlyCost:     hourlyCost,
		MonthlyCost:    hourlyCost.Mul(decimal.NewFromInt(int64(HoursInMonth))).Round(6),
	}
}

func setPriceComponentPrice(priceComponent PriceComponent, queryResult gjson.Result) {
	products := queryResult.Get("data.products").Array()
	var price decimal.Decimal
	if len(products) == 0 {
		log.Warnf("No prices found for %s %s, using 0.00", priceComponent.Resource().Address(), priceComponent.Name())
		price = decimal.Zero
	} else {
		if len(products) > 1 {
			log.Warnf("Multiple prices found for %s %s, using the first price", priceComponent.Resource().Address(), priceComponent.Name())
		}
		priceStr := products[0].Get("onDemandPricing.0.priceDimensions.0.pricePerUnit.USD").String()
		price, _ = decimal.NewFromString(priceStr)
	}
	priceComponent.SetPrice(price)
}

func getCostBreakdown(resource Resource, results ResourceQueryResultMap) ResourceCostBreakdown {
	priceComponentCosts := make([]PriceComponentCost, 0, len(resource.PriceComponents()))
	for _, priceComponent := range resource.PriceComponents() {
		result := results[&resource][&priceComponent]
		priceComponentCosts = append(priceComponentCosts, createPriceComponentCost(priceComponent, result))
	}

	subResourceCosts := make([]ResourceCostBreakdown, 0, len(resource.SubResources()))
	for _, subResource := range resource.SubResources() {
		subResourcePriceComponentCosts := make([]PriceComponentCost, 0, len(subResource.PriceComponents()))
		for _, priceComponent := range subResource.PriceComponents() {
			result := results[&subResource][&priceComponent]
			subResourcePriceComponentCosts = append(subResourcePriceComponentCosts, createPriceComponentCost(priceComponent, result))
		}
		subResourceCosts = append(subResourceCosts, ResourceCostBreakdown{
			Resource:            subResource,
			PriceComponentCosts: subResourcePriceComponentCosts,
		})
	}

	return ResourceCostBreakdown{
		Resource:            resource,
		PriceComponentCosts: priceComponentCosts,
		SubResourceCosts:    subResourceCosts,
	}
}

func GenerateCostBreakdowns(q QueryRunner, resources []Resource) ([]ResourceCostBreakdown, error) {
	costBreakdowns := make([]ResourceCostBreakdown, 0, len(resources))

	results := make(map[*Resource]ResourceQueryResultMap, len(resources))
	for _, resource := range resources {
		resourceResults, err := q.RunQueries(resource)
		if err != nil {
			return costBreakdowns, err
		}
		results[&resource] = resourceResults

		for _, priceComponentResults := range resourceResults {
			for priceComponent, result := range priceComponentResults {
				setPriceComponentPrice(*priceComponent, result)
			}
		}
	}

	for _, resource := range resources {
		if !resource.HasCost() {
			continue
		}
		costBreakdowns = append(costBreakdowns, getCostBreakdown(resource, results[&resource]))
	}

	return costBreakdowns, nil
}
