package base

type Resource interface {
	Address() string
	RawValues() map[string]interface{}
	References() map[string]*Resource
	GetFilters() []Filter
	AddReferences(address string, reference *Resource)
	PriceComponents() []PriceComponent
}

type BaseResource struct {
	address         string
	rawValues       map[string]interface{}
	references      map[string]*Resource
	resourceMapping *ResourceMapping
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

func (r *BaseResource) PriceComponents() []PriceComponent {
	return r.priceComponents
}
