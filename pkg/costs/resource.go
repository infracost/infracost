package costs

import (
	"infracost/pkg/schema"

	"github.com/prometheus/common/log"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

var hourToMonthMultiplier = decimal.NewFromInt(730)

type Resource struct {
	schemaResource *schema.Resource
	SubResources   []*Resource
	CostComponents []*CostComponent
}

func NewResource(schemaResource *schema.Resource) *Resource {
	costComponents := make([]*CostComponent, 0, len(schemaResource.CostComponents))
	for _, schemaCostComponent := range schemaResource.CostComponents {
		costComponents = append(costComponents, NewCostComponent(schemaCostComponent))
	}

	subResources := make([]*Resource, 0, len(schemaResource.SubResources))
	for _, schemaSubResource := range schemaResource.SubResources {
		subResources = append(subResources, NewResource(schemaSubResource))
	}

	return &Resource{
		schemaResource: schemaResource,
		SubResources:   subResources,
		CostComponents: costComponents,
	}
}

func (r *Resource) Name() string {
	return r.schemaResource.Name
}

func (r *Resource) CalculateCosts(q QueryRunner) error {
	queryResult, err := q.RunQueries(r)
	if err != nil {
		return err
	}

	for _, queryResult := range queryResult {
		setCostComponentPrice(queryResult.Resource, queryResult.CostComponent, queryResult.Result)
	}

	return nil
}

func (r *Resource) HourlyCost() decimal.Decimal {
	hourlyCost := decimal.Zero

	for _, costComponent := range r.CostComponents {
		hourlyCost = hourlyCost.Add(costComponent.HourlyCost())
	}

	for _, subResource := range r.SubResources {
		hourlyCost = hourlyCost.Add(subResource.HourlyCost())
	}

	return hourlyCost
}

func (r *Resource) MonthlyCost() decimal.Decimal {
	return r.HourlyCost().Mul(hourToMonthMultiplier)
}

func (r *Resource) FlattenedSubResources() []*Resource {
	subResources := make([]*Resource, 0, len(r.SubResources))
	for _, subResource := range r.SubResources {
		subResources = append(subResources, subResource)
		if len(subResource.SubResources) > 0 {
			subResources = append(subResources, subResource.FlattenedSubResources()...)
		}
	}
	return subResources
}

func setCostComponentPrice(resource *Resource, costComponent *CostComponent, result gjson.Result) {
	var priceVal decimal.Decimal

	products := result.Get("data.products").Array()
	if len(products) == 0 {
		log.Warnf("No products found for %s %s, using 0.00", resource.Name(), costComponent.Name())
		priceVal = decimal.Zero
		costComponent.SetPrice(priceVal)
		return
	}
	if len(products) > 1 {
		log.Warnf("Multiple products found for %s %s, using the first product", resource.Name(), costComponent.Name())
	}

	prices := products[0].Get("prices").Array()
	if len(prices) == 0 {
		log.Warnf("No prices found for %s %s, using 0.00", resource.Name(), costComponent.Name())
		priceVal = decimal.Zero
		costComponent.SetPrice(priceVal)
		return
	}
	if len(prices) > 1 {
		log.Warnf("Multiple prices found for %s %s, using the first price", resource.Name(), costComponent.Name())
	}

	priceVal, _ = decimal.NewFromString(prices[0].Get("USD").String())
	costComponent.SetPrice(priceVal)
	costComponent.SetPriceHash(prices[0].Get("priceHash").String())
}
