package base

import (
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

type testPriceComponent struct {
	name       string
	filters    []Filter
	price      decimal.Decimal
	hourlyCost decimal.Decimal
}

func (c *testPriceComponent) Name() string {
	return c.name
}

func (c *testPriceComponent) Filters() []Filter {
	return c.filters
}

func (c *testPriceComponent) Price() decimal.Decimal {
	return c.price
}

func (c *testPriceComponent) SetPrice(price decimal.Decimal) {
	c.price = price
}

func (c *testPriceComponent) HourlyCost() decimal.Decimal {
	return c.hourlyCost
}

type testResource struct {
	address         string
	subResources    []Resource
	priceComponents []PriceComponent
	references      map[string]Resource
}

func (r *testResource) Address() string {
	return r.address
}

func (r *testResource) SubResources() []Resource {
	return r.subResources
}

func (r *testResource) PriceComponents() []PriceComponent {
	return r.priceComponents
}

func (r *testResource) References() map[string]Resource {
	return r.references
}

func (r *testResource) AddReference(name string, resource Resource) {
	r.references[name] = resource
}

func (r *testResource) HasCost() bool {
	return true
}

type testQueryRunner struct {
	priceOverrides (map[Resource]map[PriceComponent]decimal.Decimal)
}

func generateResultForPrice(price decimal.Decimal) gjson.Result {
	f, _ := price.Float64()
	return gjson.Parse(fmt.Sprintf(`{"data": {"products": [{"onDemandPricing": [{ "priceDimensions": [{"unit": "Hrs","pricePerUnit": {"USD": "%f"}}]}]}]}}`, f))
}

func (r *testQueryRunner) getPrice(resource Resource, priceComponent PriceComponent) decimal.Decimal {
	if override, ok := r.priceOverrides[resource][priceComponent]; ok {
		return override
	}
	return decimal.NewFromFloat(float64(0.1))
}

func (r *testQueryRunner) RunQueries(resource Resource) (ResourceQueryResultMap, error) {
	results := make(map[*Resource]map[*PriceComponent]gjson.Result)

	for _, priceComponent := range resource.PriceComponents() {
		if _, ok := results[&resource]; !ok {
			results[&resource] = make(map[*PriceComponent]gjson.Result)
		}
		results[&resource][&priceComponent] = generateResultForPrice(r.getPrice(resource, priceComponent))
	}

	for _, subResource := range resource.SubResources() {
		for _, priceComponent := range subResource.PriceComponents() {
			if _, ok := results[&resource]; !ok {
				results[&resource] = make(map[*PriceComponent]gjson.Result)
			}
			results[&resource][&priceComponent] = generateResultForPrice(r.getPrice(resource, priceComponent))
		}
	}

	return results, nil
}
