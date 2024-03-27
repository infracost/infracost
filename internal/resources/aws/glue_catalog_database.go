package aws

import (
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// GlueCatalogDatabase struct represents a serverless AWS Glue catalog. A Glue catalog is a database designed to store
// raw data fetched from Glue crawlers before the data is cleaned and transformed by a Glue job.
//
// GlueCatalogDatabase is just one resource of the wider AWS Glue service, which provides a number of different serverless services
// to build a robust data analytics pipeline.
//
// Resource information: https://aws.amazon.com/glue/
// Pricing information: https://aws.amazon.com/glue/pricing/
type GlueCatalogDatabase struct {
	Address string
	Region  string

	MonthlyObjects  *float64 `infracost_usage:"monthly_objects"`
	MonthlyRequests *float64 `infracost_usage:"monthly_requests"`
}

func (r *GlueCatalogDatabase) CoreType() string {
	return "GlueCatalogDatabase"
}

func (r *GlueCatalogDatabase) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_objects", DefaultValue: 0, ValueType: schema.Float64},
		{Key: "monthly_requests", DefaultValue: 0, ValueType: schema.Float64},
	}
}

// PopulateUsage parses the u schema.UsageData into the GlueCatalogDatabase.
// It uses the `infracost_usage` struct tags to populate data into the GlueCatalogDatabase.
func (r *GlueCatalogDatabase) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid GlueCatalogDatabase struct. GlueCatalogDatabase has the following
// schema.CostComponents associated with it:
//
//  1. Storage - charged for every 100,000 objects stored above 1M, per month.
//  2. MonthlyAdditionalRequests - charged per million requests above 1M in a month.
//
// This method is called after the resource is initialised by an IaC provider. See providers folder for more information.
func (r *GlueCatalogDatabase) BuildResource() *schema.Resource {
	return &schema.Resource{
		Name:        r.Address,
		UsageSchema: r.UsageSchema(),
		CostComponents: []*schema.CostComponent{
			r.storageObjectsCostComponent(),
			r.requestsCostComponent(),
		},
	}
}

func (r *GlueCatalogDatabase) storageObjectsCostComponent() *schema.CostComponent {
	var quantity *decimal.Decimal
	var limit float64 = 100000

	if r.MonthlyObjects != nil {
		objects := *r.MonthlyObjects

		if objects < limit {
			objects = 0
		}

		quantity = decimalPtr(decimal.NewFromFloat(objects))
	}

	return &schema.CostComponent{
		Name:            "Storage",
		Unit:            "100k objects",
		UnitMultiplier:  decimal.NewFromFloat(limit),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    vendorName,
			Region:        strPtr(r.Region),
			Service:       strPtr("AWSGlue"),
			ProductFamily: strPtr("AWS Glue"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "group", ValueRegex: strPtr("/^data catalog storage$/i")},
			},
		},
		UsageBased: true,
	}
}

func (r *GlueCatalogDatabase) requestsCostComponent() *schema.CostComponent {
	var quantity *decimal.Decimal
	var limit float64 = 1000000
	if r.MonthlyRequests != nil {
		requests := *r.MonthlyRequests

		if requests < limit {
			requests = 0
		}

		quantity = decimalPtr(decimal.NewFromFloat(requests))
	}

	return &schema.CostComponent{
		Name:            "Requests",
		Unit:            "1M requests",
		UnitMultiplier:  decimal.NewFromFloat(limit),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    vendorName,
			Region:        strPtr(r.Region),
			Service:       strPtr("AWSGlue"),
			ProductFamily: strPtr("AWS Glue"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "group", ValueRegex: strPtr("/^data catalog requests$/i")},
			},
		},
		UsageBased: true,
	}
}
