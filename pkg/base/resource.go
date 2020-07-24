package base

import (
	"encoding/json"
	"sort"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

var timeUnitSecs = map[string]decimal.Decimal{
	"hour":  decimal.NewFromInt(int64(60 * 60)),
	"month": decimal.NewFromInt(int64(60 * 60 * 730)),
}

type PriceComponent interface {
	Name() string
	Unit() string
	Filters() []Filter
	Quantity() decimal.Decimal
	Price() decimal.Decimal
	SetPrice(price decimal.Decimal)
	HourlyCost() decimal.Decimal
}

type Resource interface {
	Address() string
	RawValues() map[string]interface{}
	SubResources() []Resource
	AddSubResource(Resource)
	PriceComponents() []PriceComponent
	AddPriceComponent(PriceComponent)
	References() map[string]Resource
	AddReference(name string, resource Resource)
	ResourceCount() int
	SetResourceCount(count int)
	HasCost() bool
}

func ToGJSON(values map[string]interface{}) gjson.Result {
	jsonVal, _ := json.Marshal(values)
	return gjson.ParseBytes(jsonVal)
}

func FlattenSubResources(resource Resource) []Resource {
	subResources := make([]Resource, 0, len(resource.SubResources()))
	for _, subResource := range resource.SubResources() {
		subResources = append(subResources, subResource)
		if len(subResource.SubResources()) > 0 {
			subResources = append(subResources, FlattenSubResources(subResource)...)
		}
	}
	return subResources
}

type BasePriceComponent struct {
	name                   string
	resource               Resource
	timeUnit               string
	unit                   string
	filters                []Filter
	quantityMultiplierFunc func(resource Resource) decimal.Decimal
	price                  decimal.Decimal
}

func NewBasePriceComponent(name string, resource Resource, unit string, timeUnit string) *BasePriceComponent {
	return &BasePriceComponent{
		name:     name,
		resource: resource,
		timeUnit: timeUnit,
		unit:     unit,
		filters:  []Filter{},
		price:    decimal.Zero,
	}
}

func (c *BasePriceComponent) Name() string {
	return c.name
}

func (c *BasePriceComponent) Unit() string {
	return c.unit
}

func (c *BasePriceComponent) Filters() []Filter {
	return c.filters
}

func (c *BasePriceComponent) AddFilters(filters []Filter) {
	c.filters = append(c.filters, filters...)
}

func (c *BasePriceComponent) Quantity() decimal.Decimal {
	quantity := decimal.NewFromInt(int64(1))
	if c.quantityMultiplierFunc != nil {
		quantity = quantity.Mul(c.quantityMultiplierFunc(c.resource))
	}

	timeUnitMultiplier := timeUnitSecs["month"].Div(timeUnitSecs[c.timeUnit])
	resourceCount := decimal.NewFromInt(int64(c.resource.ResourceCount()))

	return quantity.Mul(timeUnitMultiplier).Mul(resourceCount)
}

func (c *BasePriceComponent) SetQuantityMultiplierFunc(f func(resource Resource) decimal.Decimal) {
	c.quantityMultiplierFunc = f
}

func (c *BasePriceComponent) Price() decimal.Decimal {
	return c.price
}

func (c *BasePriceComponent) SetPrice(price decimal.Decimal) {
	c.price = price
}

func (c *BasePriceComponent) HourlyCost() decimal.Decimal {
	monthToHourMultiplier := timeUnitSecs["hour"].Div(timeUnitSecs["month"])
	return c.price.Mul(c.Quantity()).Mul(monthToHourMultiplier)
}

type BaseResource struct {
	address         string
	rawValues       map[string]interface{}
	hasCost         bool
	references      map[string]Resource
	resourceCount   int
	subResources    []Resource
	priceComponents []PriceComponent
}

func NewBaseResource(address string, rawValues map[string]interface{}, hasCost bool) *BaseResource {
	return &BaseResource{
		address:       address,
		rawValues:     rawValues,
		hasCost:       hasCost,
		references:    map[string]Resource{},
		resourceCount: 1,
	}
}

func (r *BaseResource) Address() string {
	return r.address
}

func (r *BaseResource) RawValues() map[string]interface{} {
	return r.rawValues
}

func (r *BaseResource) SubResources() []Resource {
	sort.Slice(r.subResources, func(i, j int) bool {
		return r.subResources[i].Address() < r.subResources[j].Address()
	})
	return r.subResources
}

func (r *BaseResource) AddSubResource(subResource Resource) {
	r.subResources = append(r.subResources, subResource)
}

func (r *BaseResource) PriceComponents() []PriceComponent {
	sort.Slice(r.priceComponents, func(i, j int) bool {
		return r.priceComponents[i].Name() < r.priceComponents[j].Name()
	})
	return r.priceComponents
}

func (r *BaseResource) AddPriceComponent(priceComponent PriceComponent) {
	r.priceComponents = append(r.priceComponents, priceComponent)
}

func (r *BaseResource) References() map[string]Resource {
	return r.references
}

func (r *BaseResource) AddReference(name string, resource Resource) {
	r.references[name] = resource
}

func (r *BaseResource) ResourceCount() int {
	return r.resourceCount
}

func (r *BaseResource) SetResourceCount(resourceCount int) {
	r.resourceCount = resourceCount
}

func (r *BaseResource) HasCost() bool {
	return r.hasCost
}
