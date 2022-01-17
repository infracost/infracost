package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"strings"

	"github.com/shopspring/decimal"
)

type LB struct {
	Address           *string
	LoadBalancerType  *string
	Region            *string
	RuleEvaluations   *int64   `infracost_usage:"rule_evaluations"`
	NewConnections    *int64   `infracost_usage:"new_connections"`
	ActiveConnections *int64   `infracost_usage:"active_connections"`
	ProcessedBytesGb  *float64 `infracost_usage:"processed_bytes_gb"`
}

var LBUsageSchema = []*schema.UsageItem{{Key: "rule_evaluations", ValueType: schema.Int64, DefaultValue: 0}, {Key: "new_connections", ValueType: schema.Int64, DefaultValue: 0}, {Key: "active_connections", ValueType: schema.Int64, DefaultValue: 0}, {Key: "processed_bytes_gb", ValueType: schema.Float64, DefaultValue: 0}}

func (r *LB) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *LB) BuildResource() *schema.Resource {
	var maxLCU *decimal.Decimal

	var newConnectionsLCU *decimal.Decimal
	if r.NewConnections != nil {
		newConnections := decimal.NewFromInt(*r.NewConnections)
		newConnectionsLCU = decimalPtr(newConnections.Div(decimal.NewFromInt(100)))
		maxLCU = newConnectionsLCU
	}

	var activeConnectionsLCU *decimal.Decimal
	if r.ActiveConnections != nil {
		activeConnections := decimal.NewFromInt(*r.ActiveConnections)
		activeConnectionsLCU = decimalPtr(activeConnections.Div(decimal.NewFromInt(3000)))
		maxLCU = decimalPtr(decimal.Max(*maxLCU, *activeConnectionsLCU))
	}

	var processedBytesLCU *decimal.Decimal
	if r.ProcessedBytesGb != nil {
		processedBytes := decimal.NewFromFloat(*r.ProcessedBytesGb)
		processedBytesLCU = decimalPtr(processedBytes.Div(decimal.NewFromInt(1)))
		maxLCU = decimalPtr(decimal.Max(*maxLCU, *processedBytesLCU))
	}

	if strings.ToLower(*r.LoadBalancerType) == "application" {
		costComponentName := "Application load balancer"
		productFamily := "Load Balancer-Application"

		var ruleEvaluationsLCU *decimal.Decimal
		if r.RuleEvaluations != nil {
			ruleEvaluations := decimal.NewFromInt(*r.RuleEvaluations)
			ruleEvaluationsLCU = decimalPtr(ruleEvaluations.Div(decimal.NewFromInt(1000)))
			maxLCU = decimalPtr(decimal.Max(*maxLCU, *ruleEvaluationsLCU))
		}

		return newLBResource(r.Region, r.Address, productFamily, costComponentName, &decimal.Zero, maxLCU)
	}

	costComponentName := "Network load balancer"
	productFamily := "Load Balancer-Network"

	return newLBResource(r.Region, r.Address, productFamily, costComponentName, &decimal.Zero, maxLCU)
}

func newLBResource(rRegion, rAddress *string, productFamily string, costComponentName string, dataProcessed *decimal.Decimal, maxLCU *decimal.Decimal) *schema.Resource {
	region := *rRegion

	costComponents := []*schema.CostComponent{
		{
			Name:           costComponentName,
			Unit:           "hours",
			HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			UnitMultiplier: decimal.NewFromInt(1),
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
			UnitMultiplier:  decimal.NewFromInt(1),
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
		Name:           *rAddress,
		CostComponents: costComponents,
	}
}
