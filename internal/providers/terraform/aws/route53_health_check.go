package aws

import (
	"fmt"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"strings"
)

func GetRoute53HealthCheck() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_route53_health_check",
		RFunc:               NewRoute53HealthCheck,
		ReferenceAttributes: []string{"alias.0.name"},
	}
}

func NewRoute53HealthCheck(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	costComponents := make([]*schema.CostComponent, 0)

	endpointType := "aws"
	usageAmount := "50"
	if u != nil && u.Get("endpoint_type").Exists() {
		endpointType = u.Get("endpoint_type").String()
		if strings.ToLower(endpointType) == "non-aws" {
			usageAmount = "0"
		}
	}

	costComponents = append(costComponents, &schema.CostComponent{
		Name:            "Health check (basic)",
		Unit:            "months",
		UnitMultiplier:  1,
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Service:       strPtr("AmazonRoute53"),
			ProductFamily: strPtr("DNS Health Check"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/Health-Check-%s/i", endpointType))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(usageAmount),
		},
	})

	healthCheckType := d.Get("type").String()

	if strings.HasPrefix(healthCheckType, "HTTPS") {
		costComponents = append(costComponents, calcOptionalHealthChecks(endpointType, "https"))
	}

	if d.Get("request_interval").String() == "10" {
		costComponents = append(costComponents, calcOptionalHealthChecks(endpointType, "fast interval"))
	}

	if d.Get("measure_latency").Bool() {
		costComponents = append(costComponents, calcOptionalHealthChecks(endpointType, "latency measurement"))
	}

	if strings.HasSuffix(healthCheckType, "STR_MATCH") {
		costComponents = append(costComponents, calcOptionalHealthChecks(endpointType, "string matching"))
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func calcOptionalHealthChecks(endpointType string, healthCheckType string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Health check (optional, %s)", healthCheckType),
		Unit:            "months",
		UnitMultiplier:  1,
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Service:       strPtr("AmazonRoute53"),
			ProductFamily: strPtr("DNS Health Check"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/Health-Check-Option-%s/i", endpointType))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("0"),
		},
	}
}
