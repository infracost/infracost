package aws

import (
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// GlueCrawler struct represents a serverless AWS Glue crawler. A Glue crawler crawls defined data sources and sends them
// into a Glue data catalog, ready for a Glue job to transform and merge into a main dataset/lake.
//
// GlueCrawler is just one resource of the wider AWS Glue service, which provides a number of different serverless services
// to build a robust data analytics pipeline.
//
// Resource information: https://aws.amazon.com/glue/
// Pricing information: https://aws.amazon.com/glue/pricing/
type GlueCrawler struct {
	Address string
	Region  string

	MonthlyHours *float64 `infracost_usage:"monthly_hours"`
}

func (r *GlueCrawler) CoreType() string {
	return "GlueCrawler"
}

func (r *GlueCrawler) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_hours", DefaultValue: 0, ValueType: schema.Float64},
	}
}

// PopulateUsage parses the u schema.UsageData into the GlueCrawler.
// It uses the `infracost_usage` struct tags to populate data into the GlueCrawler.
func (r *GlueCrawler) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid GlueCrawler struct. GlueCrawler has just one schema.CostComponent
// associated with it:
//
//  1. Hours - GlueCrawler is charged per hour that the crawler is run.
//
// This method is called after the resource is initialised by an IaC provider. See providers folder for more information.
func (r *GlueCrawler) BuildResource() *schema.Resource {
	var quantity *decimal.Decimal
	if r.MonthlyHours != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*r.MonthlyHours))
	}

	return &schema.Resource{
		Name:        r.Address,
		UsageSchema: r.UsageSchema(),
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Duration",
				Unit:            "hours",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: quantity,
				ProductFilter: &schema.ProductFilter{
					VendorName:    vendorName,
					Region:        strPtr(r.Region),
					Service:       strPtr("AWSGlue"),
					ProductFamily: strPtr("AWS Glue"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "operation", ValueRegex: strPtr("/^crawlerrun$/i")},
					},
				},
				UsageBased: true,
			},
		},
	}
}
