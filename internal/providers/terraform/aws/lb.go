package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"strings"

	"github.com/shopspring/decimal"
)

func GetLBRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_lb",
		RFunc: NewLB,
	}
}
func GetALBRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_alb",
		RFunc: NewLB,
	}
}

func NewLB(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var maxLCU *decimal.Decimal

	var newConnectionsLCU *decimal.Decimal
	if u != nil && u.Get("new_connections").Exists() {
		newConnections := decimal.NewFromInt(u.Get("new_connections").Int())
		newConnectionsLCU = decimalPtr(newConnections.Div(decimal.NewFromInt(100)))
		maxLCU = newConnectionsLCU
	}

	var activeConnectionsLCU *decimal.Decimal
	if u != nil && u.Get("active_connections").Exists() {
		activeConnections := decimal.NewFromInt(u.Get("active_connections").Int())
		activeConnectionsLCU = decimalPtr(activeConnections.Div(decimal.NewFromInt(3000)))
		maxLCU = decimalPtr(decimal.Max(*maxLCU, *activeConnectionsLCU))
	}

	var processedBytesLCU *decimal.Decimal
	if u != nil && u.Get("processed_bytes_gb").Exists() {
		processedBytes := decimal.NewFromInt(u.Get("processed_bytes_gb").Int())
		processedBytesLCU = decimalPtr(processedBytes.Div(decimal.NewFromInt(1)))
		maxLCU = decimalPtr(decimal.Max(*maxLCU, *processedBytesLCU))
	}

	if strings.ToLower(d.Get("load_balancer_type").String()) == "application" {
		costComponentName := "Application load balancer"
		productFamily := "Load Balancer-Application"

		var ruleEvaluationsLCU *decimal.Decimal
		if u != nil && u.Get("rule_evaluations").Exists() {
			ruleEvaluations := decimal.NewFromInt(u.Get("rule_evaluations").Int())
			ruleEvaluationsLCU = decimalPtr(ruleEvaluations.Div(decimal.NewFromInt(1000)))
			maxLCU = decimalPtr(decimal.Max(*maxLCU, *ruleEvaluationsLCU))
		}

		return newLBResource(d, productFamily, costComponentName, &decimal.Zero, maxLCU)
	}

	costComponentName := "Network load balancer"
	productFamily := "Load Balancer-Network"

	return newLBResource(d, productFamily, costComponentName, &decimal.Zero, maxLCU)
}

func newLBResource(d *schema.ResourceData, productFamily string, costComponentName string, dataProcessed *decimal.Decimal, maxLCU *decimal.Decimal) *schema.Resource {
	region := d.Get("region").String()

	costComponents := []*schema.CostComponent{
		{
			Name:           costComponentName,
			Unit:           "hours",
			HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			UnitMultiplier: 1,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AWSELB"),
				ProductFamily: strPtr(productFamily),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "locationType", Value: strPtr("AWS Region")},
					{Key: "usagetype", ValueRegex: strPtr("/LoadBalancerUsage/")},
				},
			},
		},
	}

	if strings.ToLower(productFamily) == "load balancer" {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Data processed",
			Unit:            "GB",
			UnitMultiplier:  1,
			MonthlyQuantity: dataProcessed,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AWSELB"),
				ProductFamily: strPtr(productFamily),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/DataProcessing-Bytes/")},
				},
			},
		})
	}

	if strings.ToLower(productFamily) == "load balancer-application" || strings.ToLower(productFamily) == "load balancer-network" {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Load balancer capacity units",
			Unit:            "LCU",
			UnitMultiplier:  schema.HourToMonthUnitMultiplier,
			MonthlyQuantity: maxLCU,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AWSELB"),
				ProductFamily: strPtr(productFamily),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "locationType", Value: strPtr("AWS Region")},
					{Key: "usagetype", ValueRegex: strPtr("/LCUUsage/")},
				},
			},
		})
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
