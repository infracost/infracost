package aws

import (
	"fmt"
	"plancosts/pkg/base"
	"strings"

	"github.com/shopspring/decimal"
)

type ScaledResource interface {
	AwsResource
	Count() int
}

type WrappedPriceComponent struct {
	*BaseAwsPriceComponent
	scaledResource        ScaledResource
	wrappedPriceComponent base.PriceComponent
}

func NewWrappedPriceComponent(scaledResource ScaledResource, wrappedPriceComponent AwsPriceComponent) *WrappedPriceComponent {
	c := &WrappedPriceComponent{
		BaseAwsPriceComponent: NewBaseAwsPriceComponent(
			wrappedPriceComponent.Name(),
			wrappedPriceComponent.AwsResource(),
			wrappedPriceComponent.TimeUnit(),
		),
		scaledResource:        scaledResource,
		wrappedPriceComponent: wrappedPriceComponent,
	}
	return c
}

func (c *WrappedPriceComponent) HourlyCost() decimal.Decimal {
	return c.wrappedPriceComponent.HourlyCost().Mul(decimal.NewFromInt(int64(c.scaledResource.Count())))
}

func (c *WrappedPriceComponent) Filters() []base.Filter {
	return c.wrappedPriceComponent.Filters()
}

type AutoscaledResource struct {
	*BaseAwsResource
	scaledResource  ScaledResource
	wrappedResource AwsResource
}

func NewWrappedResource(address string, scaledResource ScaledResource, wrappedResource AwsResource) *AutoscaledResource {
	r := &AutoscaledResource{
		BaseAwsResource: NewBaseAwsResource(address, wrappedResource.Region(), wrappedResource.RawValues()),
		scaledResource:  scaledResource,
		wrappedResource: wrappedResource,
	}

	wrappedPriceComponents := make([]base.PriceComponent, 0, len(wrappedResource.PriceComponents()))
	for _, priceComponent := range wrappedResource.PriceComponents() {
		wrappedPriceComponents = append(wrappedPriceComponents, NewWrappedPriceComponent(scaledResource, priceComponent.(AwsPriceComponent)))
	}
	r.BaseAwsResource.priceComponents = wrappedPriceComponents

	return r
}

func (r *AutoscaledResource) SubResources() []base.Resource {
	subResources := make([]base.Resource, 0)
	for _, subResource := range r.wrappedResource.SubResources() {
		address := fmt.Sprintf("%s%s", r.Address(), strings.TrimPrefix(subResource.Address(), r.wrappedResource.Address()))
		subResources = append(subResources, NewWrappedResource(address, r.scaledResource, subResource.(AwsResource)))
	}
	return subResources
}

func (r *AutoscaledResource) HasCost() bool {
	return false
}

type Ec2AutoscalingGroup struct {
	*BaseAwsResource
	wrappedResource *AutoscaledResource
}

func NewEc2AutoscalingGroup(address string, region string, rawValues map[string]interface{}) *Ec2AutoscalingGroup {
	r := &Ec2AutoscalingGroup{
		BaseAwsResource: NewBaseAwsResource(address, region, rawValues),
		wrappedResource: nil,
	}
	return r
}

func (r *Ec2AutoscalingGroup) AddReference(name string, resource base.Resource) {
	r.BaseAwsResource.AddReference(name, resource)
	if name == "launch_configuration" {
		r.wrappedResource = NewWrappedResource(r.address, r, resource.(*Ec2LaunchConfiguration))
	} else if name == "launch_template" {
		r.wrappedResource = NewWrappedResource(r.address, r, resource.(*Ec2LaunchTemplate))
	}
}

func (r *Ec2AutoscalingGroup) Count() int {
	return int(r.RawValues()["desired_capacity"].(float64))
}

func (r *Ec2AutoscalingGroup) PriceComponents() []base.PriceComponent {
	return r.wrappedResource.PriceComponents()
}

func (r *Ec2AutoscalingGroup) SubResources() []base.Resource {
	return r.wrappedResource.SubResources()
}
