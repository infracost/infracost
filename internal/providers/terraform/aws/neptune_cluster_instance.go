package aws

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetNeptuneClusterInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_neptune_cluster_instance",
		RFunc: NewNeptuneClusterInstance,
	}
}

func NewNeptuneClusterInstance(ctx *config.ProjectContext, d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var monthlyCPUCreditHrs *decimal.Decimal
	region := d.Get("region").String()
	instanceClass := d.Get("instance_class").String()
	hourlyQuantity := 1
	if d.Get("count").Type != gjson.Null {
		hourlyQuantity = int(d.Get("count").Int())
	}

	if u != nil && u.Get("monthly_cpu_credit_hrs").Type != gjson.Null {
		monthlyCPUCreditHrs = decimalPtr(decimal.NewFromInt(u.Get("monthly_cpu_credit_hrs").Int()))
	}

	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, neptuneClusterDbInstanceCostComponent(instanceClass, region, instanceClass, hourlyQuantity))

	if strings.HasPrefix(strings.ToLower(instanceClass), "db.t3.") {
		costComponents = append(costComponents, neptuneClusterCPUInstanceCostComponent(monthlyCPUCreditHrs))
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func neptuneClusterDbInstanceCostComponent(name, region, instanceType string, quantity int) *schema.CostComponent {
	return &schema.CostComponent{

		Name:           fmt.Sprintf("Database instance (on-demand, %s)", instanceType),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(int64(quantity))),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr("AmazonNeptune"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "instanceType", ValueRegex: strPtr(fmt.Sprintf("/^%s$/i", instanceType))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}

func neptuneClusterCPUInstanceCostComponent(quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		// AWS mentions that CPU Credit pricing is the same for T3 instance across all regions, but they only return prices for Hong Kong and Sao Paulo so we hard-code APE1.
		Name:           "CPU credits",
		Unit:           "vCPU-hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Service:    strPtr("AmazonNeptune"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", Value: strPtr("APE1-CPUCredits:db.t3")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}
