package aws

import (
	"plancosts/pkg/base"

	"github.com/shopspring/decimal"
)

var DefaultVolumeSize = 8

var regionMapping = map[string]string{
	"us-gov-west-1":  "AWS GovCloud (US)",
	"us-gov-east-1":  "AWS GovCloud (US-East)",
	"us-east-1":      "US East (N. Virginia)",
	"us-east-2":      "US East (Ohio)",
	"us-west-1":      "US West (N. California)",
	"us-west-2":      "US West (Oregon)",
	"ca-central-1":   "Canada (Central)",
	"cn-north-1":     "China (Beijing)",
	"cn-northwest-1": "China (Ningxia)",
	"eu-central-1":   "EU (Frankfurt)",
	"eu-west-1":      "EU (Ireland)",
	"eu-west-2":      "EU (London)",
	"eu-west-3":      "EU (Paris)",
	"eu-north-1":     "EU (Stockholm)",
	"ap-east-1":      "Asia Pacific (Hong Kong)",
	"ap-northeast-1": "Asia Pacific (Tokyo)",
	"ap-northeast-2": "Asia Pacific (Seoul)",
	"ap-northeast-3": "Asia Pacific (Osaka-Local)",
	"ap-southeast-1": "Asia Pacific (Singapore)",
	"ap-southeast-2": "Asia Pacific (Sydney)",
	"ap-south-1":     "Asia Pacific (Mumbai)",
	"me-south-1":     "Middle East (Bahrain)",
	"sa-east-1":      "South America (Sao Paulo)",
	"af-south-1":     "Africa (Cape Town)",
}

type AwsResource interface {
	base.Resource
	Region() string
	RawValues() map[string]interface{}
}

type AwsPriceComponent interface {
	base.PriceComponent
	AwsResource() AwsResource
	TimeUnit() string
}

type BaseAwsPriceComponent struct {
	name           string
	resource       AwsResource
	timeUnit       string
	regionFilters  []base.Filter
	defaultFilters []base.Filter
	valueMappings  []base.ValueMapping
	price          decimal.Decimal
}

func NewBaseAwsPriceComponent(name string, resource AwsResource, timeUnit string) *BaseAwsPriceComponent {
	c := &BaseAwsPriceComponent{
		name:     name,
		resource: resource,
		timeUnit: timeUnit,
		regionFilters: []base.Filter{
			{Key: "locationType", Value: "AWS Region"},
			{Key: "location", Value: regionMapping[resource.Region()]},
		},
		defaultFilters: []base.Filter{},
		valueMappings:  []base.ValueMapping{},
		price:          decimal.Zero,
	}

	return c
}

func (c *BaseAwsPriceComponent) AwsResource() AwsResource {
	return c.resource
}

func (c *BaseAwsPriceComponent) TimeUnit() string {
	return c.timeUnit
}

func (c *BaseAwsPriceComponent) Name() string {
	return c.name
}

func (c *BaseAwsPriceComponent) Resource() base.Resource {
	return c.resource
}

func (c *BaseAwsPriceComponent) Filters() []base.Filter {
	filters := base.MapFilters(c.valueMappings, c.resource.RawValues())
	return base.MergeFilters(c.regionFilters, c.defaultFilters, filters)
}

func (c *BaseAwsPriceComponent) SetPrice(price decimal.Decimal) {
	c.price = price
}

func (c *BaseAwsPriceComponent) HourlyCost() decimal.Decimal {
	timeUnitSecs := map[string]decimal.Decimal{
		"hour":  decimal.NewFromInt(int64(60 * 60)),
		"month": decimal.NewFromInt(int64(60 * 60 * 730)),
	}
	timeUnitMultiplier := timeUnitSecs["hour"].Div(timeUnitSecs[c.timeUnit])
	return c.price.Mul(timeUnitMultiplier)
}

func (c *BaseAwsPriceComponent) SkipQuery() bool {
	return false
}

type BaseAwsResource struct {
	address         string
	rawValues       map[string]interface{}
	region          string
	references      map[string]base.Resource
	subResources    []base.Resource
	priceComponents []base.PriceComponent
}

func NewBaseAwsResource(address string, region string, rawValues map[string]interface{}) *BaseAwsResource {
	r := &BaseAwsResource{
		address:    address,
		region:     region,
		rawValues:  rawValues,
		references: map[string]base.Resource{},
	}

	return r
}

func (r *BaseAwsResource) Address() string {
	return r.address
}

func (r *BaseAwsResource) Region() string {
	return r.region
}

func (r *BaseAwsResource) RawValues() map[string]interface{} {
	return r.rawValues
}

func (r *BaseAwsResource) SubResources() []base.Resource {
	return r.subResources
}

func (r *BaseAwsResource) PriceComponents() []base.PriceComponent {
	return r.priceComponents
}

func (r *BaseAwsResource) References() map[string]base.Resource {
	return r.references
}

func (r *BaseAwsResource) AddReference(name string, resource base.Resource) {
	r.references[name] = resource
}

func (r *BaseAwsResource) HasCost() bool {
	return true
}
