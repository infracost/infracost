package base

import (
	"fmt"
	"reflect"
)

type Resource interface {
	Address() string
	RawValues() map[string]interface{}
	References() map[string]*Resource
	GetFilters() []Filter
	AddReferences(address string, reference *Resource)
	SubResources() []Resource
	PriceComponents() []PriceComponent
}

type BaseResource struct {
	address         string
	rawValues       map[string]interface{}
	references      map[string]*Resource
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
		references:      map[string]*Resource{},
		providerFilters: providerFilters,
	}

	priceComponents := make([]PriceComponent, 0, len(resourceMapping.PriceMappings))
	for name, priceMapping := range resourceMapping.PriceMappings {
		priceComponents = append(priceComponents, NewBasePriceComponent(name, r, priceMapping))
	}
	r.priceComponents = priceComponents

	subResources := make([]Resource, 0, len(resourceMapping.SubResourceMappings))
	for name, subResourceMapping := range resourceMapping.SubResourceMappings {

		// Cast the subresource raw values to an array if isn't already
		var subResourceRawValues []interface{}
		if reflect.TypeOf(rawValues[name]).Kind() == reflect.Slice {
			subResourceRawValues = rawValues[name].([]interface{})
		} else {
			subResourceRawValues = make([]interface{}, 1)
			subResourceRawValues = append(subResourceRawValues, rawValues[name])
		}

		for i, sRawValues := range subResourceRawValues {
			subResourceAddress := fmt.Sprintf("address.%s.%d", name, i)
			subResource := NewBaseResource(
				subResourceAddress,
				sRawValues.(map[string]interface{}),
				subResourceMapping,
				providerFilters,
			)
			subResources = append(subResources, subResource)
		}
	}
	r.subResources = subResources
	return r
}

func (r *BaseResource) Address() string {
	return r.address
}

func (r *BaseResource) RawValues() map[string]interface{} {
	return r.rawValues
}

func (r *BaseResource) References() map[string]*Resource {
	return r.references
}

func (r *BaseResource) GetFilters() []Filter {
	return r.providerFilters
}

func (r *BaseResource) AddReferences(address string, reference *Resource) {
	r.references[address] = reference
}

func (r *BaseResource) SubResources() []Resource {
	return r.subResources
}

func (r *BaseResource) PriceComponents() []PriceComponent {
	return r.priceComponents
}
