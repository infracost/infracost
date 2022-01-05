package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"

	"github.com/shopspring/decimal"
)

type BigqueryDataset struct {
	Address          *string
	Region           *string
	MonthlyQueriesTb *float64 `infracost_usage:"monthly_queries_tb"`
}

var BigqueryDatasetUsageSchema = []*schema.UsageItem{{Key: "monthly_queries_tb", ValueType: schema.Float64, DefaultValue: 0}}

func (r *BigqueryDataset) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *BigqueryDataset) BuildResource() *schema.Resource {
	region := *r.Region

	var queriesTB *decimal.Decimal
	if r.MonthlyQueriesTb != nil {
		queriesTB = decimalPtr(decimal.NewFromFloat(*r.MonthlyQueriesTb))
	}

	return &schema.Resource{
		Name: *r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Queries (on-demand)",
				Unit:            "TB",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: queriesTB,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("gcp"),
					Region:        strPtr(region),
					Service:       strPtr("BigQuery"),
					ProductFamily: strPtr("ApplicationServices"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "description", Value: strPtr(fmt.Sprintf("Analysis (%s)", region))},
					},
				},
				PriceFilter: &schema.PriceFilter{
					StartUsageAmount: strPtr("1"),
				},
			},
		}, UsageSchema: BigqueryDatasetUsageSchema,
	}
}
