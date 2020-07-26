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

	return PriceComponentCost{
		PriceComponent: priceComponent,
		HourlyCost:     hourlyCost,
		MonthlyCost:    hourlyCost.Mul(decimal.NewFromInt(int64(HoursInMonth))).Round(6),
	}
}

func setPriceComponentPrice(r resource.Resource, priceComponent resource.PriceComponent, queryResult gjson.Result) {
	products := queryResult.Get("data.products").Array()
	var price decimal.Decimal
	if len(products) == 0 {
		log.Warnf("No prices found for %s %s, using 0.00", r.Address(), priceComponent.Name())
		price = decimal.Zero
	} else {
		if len(products) > 1 {
			log.Warnf("Multiple prices found for %s %s, using the first price", r.Address(), priceComponent.Name())
		}
		priceStr := products[0].Get("onDemandPricing.0.priceDimensions.0.pricePerUnit.USD").String()
		price, _ = decimal.NewFromString(priceStr)
	}
	priceComponent.SetPrice(price)
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

	results := make(map[*resource.Resource]ResourceQueryResultMap, len(resources))
	for _, r := range resources {
		if !r.HasCost() {
			continue
		}
		resourceResults, err := q.RunQueries(r)
		if err != nil {
			return costBreakdowns, err
		}
		results[&r] = resourceResults

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
		costBreakdowns = append(costBreakdowns, getCostBreakdown(r, results[&r]))
	}

	sort.Slice(costBreakdowns, func(i, j int) bool {
		return costBreakdowns[i].Resource.Address() < costBreakdowns[j].Resource.Address()
	})

	return costBreakdowns, nil
}
