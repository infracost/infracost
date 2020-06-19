package base

import (
	"github.com/shopspring/decimal"
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

type priceComponentQueryResult struct {
	PriceComponent PriceComponent
	QueryResult    gjson.Result
}

type queryKey struct {
	Resource       Resource
	PriceComponent PriceComponent
}

// Batch all the queries for this resource so we can use one GraphQL call
// Use queryKeys to keep track of which query maps to which sub-resource and price component
func batchQueries(resource Resource) ([]queryKey, []GraphQLQuery) {
	queryKeys := make([]queryKey, 0)
	queries := make([]GraphQLQuery, 0)

	for _, priceComponent := range resource.PriceComponents() {
		if priceComponent.ShouldSkip() {
			continue
		}

		queryKeys = append(queryKeys, queryKey{resource, priceComponent})
		queries = append(queries, BuildQuery(priceComponent.GetFilters()))
	}

	for _, subResource := range resource.SubResources() {
		for _, priceComponent := range subResource.PriceComponents() {
			if priceComponent.ShouldSkip() {
				continue
			}

			queryKeys = append(queryKeys, queryKey{subResource, priceComponent})
			queries = append(queries, BuildQuery(priceComponent.GetFilters()))
		}
	}

	return queryKeys, queries
}

// Unpack the query results into the top-level resource results and any subresource results
func unpackQueryResults(resource Resource, queryKeys []queryKey, queryResults []gjson.Result) ([]priceComponentQueryResult, map[*Resource][]priceComponentQueryResult) {
	resourceResults := make([]priceComponentQueryResult, 0)
	subResourceResults := make(map[*Resource][]priceComponentQueryResult, 0)

	for i, queryResult := range queryResults {
		queryResource := queryKeys[i].Resource
		priceComponent := queryKeys[i].PriceComponent

		result := priceComponentQueryResult{
			priceComponent,
			queryResult,
		}

		if queryResource == resource {
			resourceResults = append(resourceResults, result)
		} else {
			if _, ok := subResourceResults[&queryResource]; !ok {
				subResourceResults[&queryResource] = make([]priceComponentQueryResult, 0)
			}

			subResourceResults[&queryResource] = append(subResourceResults[&queryResource], result)
		}
	}

	return resourceResults, subResourceResults
}

func createPriceComponentCost(priceComponent PriceComponent, queryResult gjson.Result) PriceComponentCost {
	priceStr := queryResult.Get("data.products.0.onDemandPricing.0.priceDimensions.0.pricePerUnit.USD").String()
	price, _ := decimal.NewFromString(priceStr)

	hourlyCost := priceComponent.CalculateHourlyCost(price)

	return PriceComponentCost{
		PriceComponent: priceComponent,
		HourlyCost:     hourlyCost,
		MonthlyCost:    hourlyCost.Mul(decimal.NewFromInt(int64(HoursInMonth))).Round(6),
	}
}

func GetCostBreakdown(resource Resource) (ResourceCostBreakdown, error) {
	queryKeys, queries := batchQueries(resource)

	queryResults, err := GetQueryResults(queries)
	if err != nil {
		return ResourceCostBreakdown{}, err
	}

	resourceResults, subResourceResults := unpackQueryResults(resource, queryKeys, queryResults)

	priceComponentCosts := make([]PriceComponentCost, 0, len(resourceResults))
	for _, result := range resourceResults {
		priceComponentCosts = append(priceComponentCosts, createPriceComponentCost(result.PriceComponent, result.QueryResult))
	}

	subResourceCosts := make([]ResourceCostBreakdown, 0, len(subResourceResults))
	for subResource, results := range subResourceResults {
		subResourcePriceComponentCosts := make([]PriceComponentCost, 0, len(results))
		for _, result := range results {
			subResourcePriceComponentCosts = append(subResourcePriceComponentCosts, createPriceComponentCost(result.PriceComponent, result.QueryResult))
		}
		subResourceCosts = append(subResourceCosts, ResourceCostBreakdown{
			Resource:            *subResource,
			PriceComponentCosts: subResourcePriceComponentCosts,
		})
	}

	return ResourceCostBreakdown{
		Resource:            resource,
		PriceComponentCosts: priceComponentCosts,
		SubResourceCosts:    subResourceCosts,
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
