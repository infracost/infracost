package costs

import (
	"infracost/pkg/resource"
	"sort"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

var HoursInMonth = 730

type PriceComponentCost struct {
	PriceComponent resource.PriceComponent
	PriceHash      string
	HourlyCost     decimal.Decimal
	MonthlyCost    decimal.Decimal
}

type ResourceCostBreakdown struct {
	Resource            resource.Resource
	PriceComponentCosts []PriceComponentCost
	SubResourceCosts    []ResourceCostBreakdown
}

type PriceComponentQueryMap struct {
	Resource       resource.Resource
	PriceComponent resource.PriceComponent
	Query          GraphQLQuery
}

type queryKey struct {
	Resource       resource.Resource
	PriceComponent resource.PriceComponent
}

func createPriceComponentCost(priceComponent resource.PriceComponent, queryResult gjson.Result) PriceComponentCost {
	hourlyCost := priceComponent.HourlyCost()

	priceHash := queryResult.Get("data.products.0.prices.0.priceHash").String()

	return PriceComponentCost{
		PriceComponent: priceComponent,
		PriceHash:      priceHash,
		HourlyCost:     hourlyCost,
		MonthlyCost:    hourlyCost.Mul(decimal.NewFromInt(int64(HoursInMonth))).Round(6),
	}
}

func setPriceComponentPrice(r resource.Resource, priceComponent resource.PriceComponent, queryResult gjson.Result) {
	var priceVal decimal.Decimal

	products := queryResult.Get("data.products").Array()
	if len(products) == 0 {
		log.Warnf("No products found for %s %s, using 0.00", r.Address(), priceComponent.Name())
		priceVal = decimal.Zero
		priceComponent.SetPrice(priceVal)
		return
	}
	if len(products) > 1 {
		log.Warnf("Multiple products found for %s %s, using the first product", r.Address(), priceComponent.Name())
	}

	prices := products[0].Get("prices").Array()
	if len(prices) == 0 {
		log.Warnf("No prices found for %s %s, using 0.00", r.Address(), priceComponent.Name())
		priceVal = decimal.Zero
		priceComponent.SetPrice(priceVal)
		return
	}
	if len(prices) > 1 {
		log.Warnf("Multiple prices found for %s %s, using the first price", r.Address(), priceComponent.Name())
	}

	priceVal, _ = decimal.NewFromString(prices[0].Get("USD").String())
	priceComponent.SetPrice(priceVal)
}

func getCostBreakdown(r resource.Resource, results ResourceQueryResultMap) ResourceCostBreakdown {
	priceComponentCosts := make([]PriceComponentCost, 0, len(r.PriceComponents()))
	for _, priceComponent := range r.PriceComponents() {
		result := results[r][priceComponent]
		priceComponentCosts = append(priceComponentCosts, createPriceComponentCost(priceComponent, result))
	}

	subResourceCosts := make([]ResourceCostBreakdown, 0, len(r.SubResources()))
	for _, subResource := range r.SubResources() {
		subResourceCosts = append(subResourceCosts, getCostBreakdown(subResource, results))
	}

	return ResourceCostBreakdown{
		Resource:            r,
		PriceComponentCosts: priceComponentCosts,
		SubResourceCosts:    subResourceCosts,
	}
}

func GenerateCostBreakdowns(q QueryRunner, resources []resource.Resource) ([]ResourceCostBreakdown, error) {
	costBreakdowns := make([]ResourceCostBreakdown, 0, len(resources))

	results := make(map[resource.Resource]ResourceQueryResultMap, len(resources))
	for _, r := range resources {
		if !r.HasCost() {
			continue
		}
		resourceResults, err := q.RunQueries(r)
		if err != nil {
			return costBreakdowns, err
		}
		results[r] = resourceResults

		for _, priceComponentResults := range resourceResults {
			for priceComponent, result := range priceComponentResults {
				setPriceComponentPrice(r, priceComponent, result)
			}
		}
	}

	for _, r := range resources {
		if !r.HasCost() {
			continue
		}
		costBreakdowns = append(costBreakdowns, getCostBreakdown(r, results[r]))
	}

	sort.Slice(costBreakdowns, func(i, j int) bool {
		return costBreakdowns[i].Resource.Address() < costBreakdowns[j].Resource.Address()
	})

	return costBreakdowns, nil
}
