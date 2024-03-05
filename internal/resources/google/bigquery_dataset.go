package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"

	"github.com/shopspring/decimal"
)

type BigQueryDataset struct {
	Address          string
	Region           string
	MonthlyQueriesTB *float64 `infracost_usage:"monthly_queries_tb"`
}

func (r *BigQueryDataset) CoreType() string {
	return "BigQueryDataset"
}

func (r *BigQueryDataset) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_queries_tb", ValueType: schema.Float64, DefaultValue: 0},
	}
}

func (r *BigQueryDataset) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *BigQueryDataset) BuildResource() *schema.Resource {
	var queriesTB *decimal.Decimal
	if r.MonthlyQueriesTB != nil {
		queriesTB = decimalPtr(decimal.NewFromFloat(*r.MonthlyQueriesTB))
	}

	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Queries (on-demand)",
				Unit:            "TB",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: queriesTB,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("gcp"),
					Region:        strPtr(r.Region),
					Service:       strPtr("BigQuery"),
					ProductFamily: strPtr("ApplicationServices"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "description", Value: strPtr(fmt.Sprintf("Analysis (%s)", r.Region))},
					},
				},
				PriceFilter: &schema.PriceFilter{
					StartUsageAmount: strPtr("1"),
				},
				UsageBased: true,
			},
		},
		UsageSchema: r.UsageSchema(),
	}
}
