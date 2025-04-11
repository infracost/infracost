package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"strings"

	"github.com/shopspring/decimal"
)

type LB struct {
	Address           string
	LoadBalancerType  string
	Region            string
	RuleEvaluations   *int64   `infracost_usage:"rule_evaluations"`
	NewConnections    *int64   `infracost_usage:"new_connections"`
	ActiveConnections *int64   `infracost_usage:"active_connections"`
	ProcessedBytesGB  *float64 `infracost_usage:"processed_bytes_gb"`
}

var LBUsageSchema = []*schema.UsageItem{
	{Key: "rule_evaluations", ValueType: schema.Int64, DefaultValue: 0},
	{Key: "new_connections", ValueType: schema.Int64, DefaultValue: 0},
	{Key: "active_connections", ValueType: schema.Int64, DefaultValue: 0},
	{Key: "processed_bytes_gb", ValueType: schema.Float64, DefaultValue: 0},
}

func (r *LB) CoreType() string {
	return "LB"
}

func (r *LB) UsageSchema() []*schema.UsageItem {
	return LBUsageSchema
}

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

		if maxLCU == nil {
			maxLCU = activeConnectionsLCU
		} else {
			maxLCU = decimalPtr(decimal.Max(*maxLCU, *activeConnectionsLCU))
		}
	}

	var processedBytesLCU *decimal.Decimal
	if r.ProcessedBytesGB != nil {
		processedBytes := decimal.NewFromFloat(*r.ProcessedBytesGB)
		processedBytesLCU = decimalPtr(processedBytes.Div(decimal.NewFromInt(1)))

		if maxLCU == nil {
			maxLCU = processedBytesLCU
		} else {
			maxLCU = decimalPtr(decimal.Max(*maxLCU, *processedBytesLCU))
		}
	}

	var costComponents []*schema.CostComponent

	if strings.ToLower(r.LoadBalancerType) == "application" {
		var ruleEvaluationsLCU decimal.Decimal
		if r.RuleEvaluations != nil && maxLCU != nil {
			ruleEvaluations := decimal.NewFromInt(*r.RuleEvaluations)
			ruleEvaluationsLCU = ruleEvaluations.Div(decimal.NewFromInt(1000))

			if maxLCU == nil {
				maxLCU = &ruleEvaluationsLCU
			} else {
				maxLCU = decimalPtr(decimal.Max(*maxLCU, ruleEvaluationsLCU))
			}
		}

		costComponents = r.applicationLBCostComponents(maxLCU)
	} else {
		costComponents = r.networkLBCostComponents(maxLCU)
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *LB) applicationLBCostComponents(maxLCU *decimal.Decimal) []*schema.CostComponent {
	productFamily := "Load Balancer-Application"

	return []*schema.CostComponent{
		{
			Name:           "Application load balancer",
			Unit:           "hours",
			HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			UnitMultiplier: decimal.NewFromInt(1),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(r.Region),
				Service:       strPtr("AWSELB"),
				ProductFamily: strPtr("Load Balancer-Application"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "locationType", Value: strPtr("AWS Region")},
					{Key: "usagetype", ValueRegex: regexPtr("^([A-Z]{3}\\d-|Global-|EU-)?LoadBalancerUsage$")},
				},
			},
		},
		r.capacityUnitsCostComponent(productFamily, maxLCU),
	}
}

func (r *LB) networkLBCostComponents(maxLCU *decimal.Decimal) []*schema.CostComponent {
	productFamily := "Load Balancer-Network"

	return []*schema.CostComponent{
		{
			Name:           "Network load balancer",
			Unit:           "hours",
			HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			UnitMultiplier: decimal.NewFromInt(1),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(r.Region),
				Service:       strPtr("AWSELB"),
				ProductFamily: strPtr("Load Balancer-Network"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "locationType", Value: strPtr("AWS Region")},
					{Key: "usagetype", ValueRegex: strPtr("/LoadBalancerUsage/")},
				},
			},
		},
		r.capacityUnitsCostComponent(productFamily, maxLCU),
	}
}

func (r *LB) capacityUnitsCostComponent(productFamily string, maxLCU *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Load balancer capacity units",
		Unit:            "LCU",
		UnitMultiplier:  schema.HourToMonthUnitMultiplier,
		MonthlyQuantity: maxLCU,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AWSELB"),
			ProductFamily: strPtr(productFamily),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "locationType", Value: strPtr("AWS Region")},
				{Key: "usagetype", ValueRegex: strPtr("/^([A-Z]{3}\\d-|Global-|EU-)?LCUUsage/")},
			},
		},
		UsageBased: true,
	}
}
