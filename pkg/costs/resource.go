package costs

import (
	"fmt"
	"infracost/pkg/schema"

	"github.com/shopspring/decimal"
)

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
		// TODO
		fmt.Println(queryResult)
	}

	return nil
}

func (r *Resource) HourlyCost() decimal.Decimal {
	return decimal.Zero
}

func (r *Resource) MonthlyCost() decimal.Decimal {
	return decimal.Zero
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
