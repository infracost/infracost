package resource

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

type AttributeFilter struct {
	Key        string  `json:"key"`
	Value      *string `json:"value,omitempty"`
	ValueRegex *string `json:"value_regex,omitempty"`
}

type ProductFilter struct {
	VendorName       *string            `json:"vendorName,omitempty"`
	Service          *string            `json:"service,omitempty"`
	ProductFamily    *string            `json:"productFamily,omitempty"`
	Region           *string            `json:"region,omitempty"`
	Sku              *string            `json:"sku,omitempty"`
	AttributeFilters *[]AttributeFilter `json:"attributeFilters,omitempty"`
}

type PriceFilter struct {
	PurchaseOption     *string `json:"purchaseOption,omitempty"`
	Unit               *string `json:"unit,omitempty"`
	Description        *string `json:"description,omitempty"`
	DescriptionRegex   *string `json:"descriptionRegex,omitempty"`
	TermLength         *string `json:"termLength,omitempty"`
	TermPurchaseOption *string `json:"termPurchaseOption,omitempty"`
	TermOfferingClass  *string `json:"termOfferingClass,omitempty"`
}

type PriceComponent interface {
	Name() string
	Unit() string
	ProductFilter() *ProductFilter
	PriceFilter() *PriceFilter
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
	productFilter          *ProductFilter
	priceFilter            *PriceFilter
	quantityMultiplierFunc func(resource Resource) decimal.Decimal
	price                  decimal.Decimal
}

func NewBasePriceComponent(name string, resource Resource, unit string, timeUnit string, productFilter *ProductFilter, priceFilter *PriceFilter) *BasePriceComponent {
	return &BasePriceComponent{
		name:          name,
		resource:      resource,
		timeUnit:      timeUnit,
		unit:          unit,
		productFilter: productFilter,
		priceFilter:   priceFilter,
		price:         decimal.Zero,
	}
}

func (c *BasePriceComponent) Name() string {
	return c.name
}

func (c *BasePriceComponent) Unit() string {
	return c.unit
}
func (c *BasePriceComponent) ProductFilter() *ProductFilter {
	return c.productFilter
}

func (c *BasePriceComponent) PriceFilter() *PriceFilter {
	return c.priceFilter
}

func (c *BasePriceComponent) SetProductFilter(productFilter *ProductFilter) {
	c.productFilter = productFilter
}

func (c *BasePriceComponent) SetPriceFilter(priceFilter *PriceFilter) {
	c.priceFilter = priceFilter
}

func (c *BasePriceComponent) Quantity() decimal.Decimal {
	quantity := decimal.NewFromInt(int64(1))
	if c.quantityMultiplierFunc != nil {
		quantity = quantity.Mul(c.quantityMultiplierFunc(c.resource))
	}

	timeUnitMultiplier := timeUnitSecs["month"].Div(timeUnitSecs[c.timeUnit])
	resourceCount := decimal.NewFromInt(int64(c.resource.ResourceCount()))

	return quantity.Mul(timeUnitMultiplier).Mul(resourceCount).Round(6)
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
	monthToHourDivisor := timeUnitSecs["month"].Div(timeUnitSecs["hour"])
	return c.price.Mul(c.Quantity()).Div(monthToHourDivisor)
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
	for _, subResource := range r.SubResources() {
		subResource.SetResourceCount(resourceCount)
	}
}

func (r *BaseResource) HasCost() bool {
	return r.hasCost
}
