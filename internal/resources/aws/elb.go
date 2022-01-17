package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type Elb struct {
	Address                *string
	Region                 *string
	MonthlyDataProcessedGb *float64 `infracost_usage:"monthly_data_processed_gb"`
}

var ElbUsageSchema = []*schema.UsageItem{{Key: "monthly_data_processed_gb", ValueType: schema.Float64, DefaultValue: 0}}

func (r *Elb) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *Elb) BuildResource() *schema.Resource {
	productFamily := "Load Balancer"
	costComponentName := "Classic load balancer"

	var dataProcessed *decimal.Decimal
	if r.MonthlyDataProcessedGb != nil {
		dataProcessed = decimalPtr(decimal.NewFromFloat(*r.MonthlyDataProcessedGb))
	}

	var maxLCU *decimal.Decimal

	return newLBResource(r.Region, r.Address, productFamily, costComponentName, dataProcessed, maxLCU)
}
