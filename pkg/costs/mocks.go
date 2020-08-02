package costs

import (
	"fmt"
	"infracost/pkg/resource"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

type testQueryRunner struct {
	priceOverrides (map[resource.Resource]map[resource.PriceComponent]decimal.Decimal)
}

func generateResultForPrice(price decimal.Decimal) gjson.Result {
	f, _ := price.Float64()
	return gjson.Parse(fmt.Sprintf(`{"data": {"products": [{"prices": [{"USD": "%f"}]}]}}`, f))
}

func (q *testQueryRunner) getPrice(resource resource.Resource, priceComponent resource.PriceComponent) decimal.Decimal {
	if override, ok := q.priceOverrides[resource][priceComponent]; ok {
		return override
	}
	return decimal.NewFromFloat(float64(0.1))
}

func (q *testQueryRunner) RunQueries(r resource.Resource) (ResourceQueryResultMap, error) {
	results := make(map[resource.Resource]map[resource.PriceComponent]gjson.Result)

	for _, priceComponent := range r.PriceComponents() {
		if _, ok := results[r]; !ok {
			results[r] = make(map[resource.PriceComponent]gjson.Result)
		}
		results[r][priceComponent] = generateResultForPrice(q.getPrice(r, priceComponent))
	}

	for _, subResource := range resource.FlattenSubResources(r) {
		for _, priceComponent := range subResource.PriceComponents() {
			if _, ok := results[subResource]; !ok {
				results[subResource] = make(map[resource.PriceComponent]gjson.Result)
			}
			results[subResource][priceComponent] = generateResultForPrice(q.getPrice(subResource, priceComponent))
		}
	}

	return results, nil
}
