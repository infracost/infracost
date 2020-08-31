package prices

import (
	"fmt"
	"infracost/pkg/config"
	"infracost/pkg/schema"

	"github.com/prometheus/common/log"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)


func PopulatePrices(resources []*schema.Resource) error {
	q := NewGraphQLQueryRunner(fmt.Sprintf("%s/graphql", config.Config.ApiUrl))

	for _, resource := range resources {
		err := GetPrices(resource, q)
		if err != nil {
			return err
		}
	}

	return nil
}


func GetPrices(resource *schema.Resource, q QueryRunner) error {
	queryResult, err := q.RunQueries(resource)
	if err != nil {
		return err
	}

	for _, queryResult := range queryResult {
		setCostComponentPrice(queryResult.Resource, queryResult.CostComponent, queryResult.Result)
	}

	return nil
}

func setCostComponentPrice(resource *schema.Resource, costComponent *schema.CostComponent, result gjson.Result) {
	var priceVal decimal.Decimal

	products := result.Get("data.products").Array()
	if len(products) == 0 {
		log.Warnf("No products found for %s %s, using 0.00", resource.Name, costComponent.Name)
		priceVal = decimal.Zero
		costComponent.SetPrice(priceVal)
		return
	}
	if len(products) > 1 {
		log.Warnf("Multiple products found for %s %s, using the first product", resource.Name, costComponent.Name)
	}

	prices := products[0].Get("prices").Array()
	if len(prices) == 0 {
		log.Warnf("No prices found for %s %s, using 0.00", resource.Name, costComponent.Name)
		priceVal = decimal.Zero
		costComponent.SetPrice(priceVal)
		return
	}
	if len(prices) > 1 {
		log.Warnf("Multiple prices found for %s %s, using the first price", resource.Name, costComponent.Name)
	}

	priceVal, _ = decimal.NewFromString(prices[0].Get("USD").String())
	costComponent.SetPrice(priceVal)
	costComponent.SetPriceHash(prices[0].Get("priceHash").String())
}
