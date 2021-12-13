package aws

import (
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetELBRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_elb",
		RFunc: NewELB,
	}
}

func NewELB(ctx *config.RunContext, d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	productFamily := "Load Balancer"
	costComponentName := "Classic load balancer"

	var dataProcessed *decimal.Decimal
	if u != nil && u.Get("monthly_data_processed_gb").Exists() {
		dataProcessed = decimalPtr(decimal.NewFromInt(u.Get("monthly_data_processed_gb").Int()))
	}

	var maxLCU *decimal.Decimal

	return newLBResource(d, productFamily, costComponentName, dataProcessed, maxLCU)
}
