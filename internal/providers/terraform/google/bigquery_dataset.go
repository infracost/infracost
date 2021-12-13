package google

import (
	"fmt"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetBigqueryDatasetRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_bigquery_dataset",
		RFunc: NewBigqueryDataset,
	}
}

func NewBigqueryDataset(ctx *config.RunContext, d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	var queriesTB *decimal.Decimal
	if u != nil && u.Get("monthly_queries_tb").Type != gjson.Null {
		queriesTB = decimalPtr(decimal.NewFromFloat(u.Get("monthly_queries_tb").Float()))
	}

	return &schema.Resource{
		Name: d.Address,
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
		},
	}
}
