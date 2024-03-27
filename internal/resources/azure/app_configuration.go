package azure

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// AppConfiguration struct represents an Azure App Configuration. App
// Configuration is a managed service that helps developers centralize their
// application configurations. It provides a service to store, manage, and access
// application configuration settings.
//
// Resource information: https://azure.microsoft.com/en-us/products/app-configuration/
// Pricing information: https://azure.microsoft.com/en-us/pricing/details/app-configuration/
type AppConfiguration struct {
	Address  string
	Region   string
	SKU      string
	Replicas int64

	MonthlyAdditionalRequests *int64 `infracost_usage:"monthly_additional_requests"`
}

// CoreType returns the name of this resource type
func (r *AppConfiguration) CoreType() string {
	return "AppConfiguration"
}

// UsageSchema defines a list which represents the usage schema of AppConfiguration.
func (r *AppConfiguration) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_additional_requests", ValueType: schema.Int64, DefaultValue: 0},
	}
}

// PopulateUsage parses the u schema.UsageData into the AppConfiguration.
// It uses the `infracost_usage` struct tags to populate data into the AppConfiguration.
func (r *AppConfiguration) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid AppConfiguration struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
//
// BuildResource only returns cost components if the sku is not "free".
// "Standard" App Configuration instances are charged per instance and replica
// and per 10k requests over a daily 200k limit. However, we cannot compute the
// request count from the IaC code, so we rely on the user to provide the request
// count as a usage parameter. This usage parameter defines all total request
// made to the App Configuration instance and it's replicas in a month.
func (r *AppConfiguration) BuildResource() *schema.Resource {
	if strings.ToLower(r.SKU) == "free" {
		return &schema.Resource{
			Name:      r.Address,
			NoPrice:   true,
			IsSkipped: true,
		}
	}

	titledSku := cases.Title(language.English).String(r.SKU)
	components := []*schema.CostComponent{
		r.instanceCostComponent(titledSku),
		r.requestCostComponent(titledSku),
	}
	costComponents := components

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *AppConfiguration) requestCostComponent(titledSku string) *schema.CostComponent {
	var requestQuantity *decimal.Decimal
	if r.MonthlyAdditionalRequests != nil {
		requestQuantity = decimalPtr(decimal.NewFromInt(*r.MonthlyAdditionalRequests).Div(decimal.NewFromInt(10000)))
	}

	requestComponent := &schema.CostComponent{
		Name:            "Requests (over 200k/day per replica)",
		Unit:            "10k requests",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: requestQuantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("App Configuration"),
			ProductFamily: strPtr("Developer Tools"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", Value: strPtr(titledSku + " Overage Operations")},
			},
		},
		UsageBased: true,
	}
	return requestComponent
}

func (r *AppConfiguration) instanceCostComponent(sku string) *schema.CostComponent {
	instanceCount := decimal.NewFromInt(1)
	if r.Replicas > 0 {

		instanceCount = instanceCount.Add(decimal.NewFromInt(r.Replicas))
	}
	name := "Instance"
	if r.Replicas > 0 {
		desc := "replica"
		if r.Replicas > 1 {
			desc = "replicas"
		}

		name = fmt.Sprintf("Instance (%d %s)", r.Replicas, desc)
	}

	return &schema.CostComponent{
		Name:            name,
		Unit:            "days",
		UnitMultiplier:  instanceCount,
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(30).Mul(instanceCount)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("App Configuration"),
			ProductFamily: strPtr("Developer Tools"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr(sku)},
				{Key: "meterName", Value: strPtr(sku + " Instance")},
			},
		},
	}
}
