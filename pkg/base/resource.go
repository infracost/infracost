package base

import (
	"fmt"
	"reflect"

	"github.com/shopspring/decimal"
)

type Resource interface {
	Address() string
	RawValues() map[string]interface{}
	References() map[string]Resource
	GetFilters() []Filter
	AddReferences(address string, reference Resource)
	AddSubResources()
	SubResources() []Resource
	PriceComponents() []PriceComponent
	AdjustCost(cost decimal.Decimal) decimal.Decimal
	NonCostable() bool
}

type BaseResource struct {
	address         string
	rawValues       map[string]interface{}
	references      map[string]Resource
	resourceMapping *ResourceMapping
	subResources    []Resource
	priceComponents []PriceComponent
	providerFilters []Filter
}

func NewBaseResource(address string, rawValues map[string]interface{}, resourceMapping *ResourceMapping, providerFilters []Filter) *BaseResource {
	r := &BaseResource{
		address:         address,
		rawValues:       rawValues,
		resourceMapping: resourceMapping,
		references:      map[string]Resource{},
		providerFilters: providerFilters,
	}

	priceComponents := make([]PriceComponent, 0, len(resourceMapping.PriceMappings))
	for name, priceMapping := range resourceMapping.PriceMappings {
		priceComponents = append(priceComponents, NewBasePriceComponent(name, r, priceMapping))
	}
	r.priceComponents = priceComponents
	return r
}

func (r *BaseResource) Address() string {
	return r.address
}

func (r *BaseResource) RawValues() map[string]interface{} {
	return r.rawValues
}

func (r *BaseResource) References() map[string]Resource {
	return r.references
}

func (r *BaseResource) GetFilters() []Filter {
	return r.providerFilters
}

func (r *BaseResource) AddReferences(address string, reference Resource) {
	r.references[address] = reference
}

func (r *BaseResource) SubResources() []Resource {
	return r.subResources
}

func (r *BaseResource) AddSubResources() {
	subResources := make([]Resource, 0, len(r.resourceMapping.SubResourceMappings))

	overriddenSubResourceRawValues := make(map[string][]interface{})
	if r.resourceMapping.OverrideSubResourceRawValues != nil {
		overriddenSubResourceRawValues = r.resourceMapping.OverrideSubResourceRawValues(r)
	}

	for name, subResourceMapping := range r.resourceMapping.SubResourceMappings {

		var subResourceRawValues []interface{}
		subResourceRawValues = overriddenSubResourceRawValues[name]
		if subResourceRawValues == nil && r.rawValues[name] != nil {
			if reflect.TypeOf(r.rawValues[name]).Kind() == reflect.Slice {
				subResourceRawValues = r.rawValues[name].([]interface{})
			} else {
				subResourceRawValues = make([]interface{}, 1)
				subResourceRawValues = append(subResourceRawValues, r.rawValues[name])
			}
		}
		for i, sRawValues := range subResourceRawValues {
			subResourceAddress := fmt.Sprintf("%s.%s[%d]", r.Address(), name, i)
			subResource := NewBaseResource(
				subResourceAddress,
				sRawValues.(map[string]interface{}),
				subResourceMapping,
				r.providerFilters,
			)
			subResources = append(subResources, subResource)
		}
	}
	r.subResources = subResources
}

func (r *BaseResource) PriceComponents() []PriceComponent {
	return r.priceComponents
}

func (r *BaseResource) AdjustCost(cost decimal.Decimal) decimal.Decimal {
	if r.resourceMapping.AdjustCost != nil {
		return r.resourceMapping.AdjustCost(r, cost)
	}
	return cost
}

func (r *BaseResource) NonCostable() bool {
	return r.resourceMapping.NonCostable
}
