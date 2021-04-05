package aws

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
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
		endpointType = strings.Replace(u.Get("endpoint_type").String(), "_", "-", 1)
		if strings.ToLower(endpointType) == "non-aws" {
			usageAmount = "0"
		}
	}

	costComponents = append(costComponents, &schema.CostComponent{
		Name:            "Health check",
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
	optionalHealthCheckCount := decimal.Zero

	if strings.HasPrefix(healthCheckType, "HTTPS") {
		optionalHealthCheckCount = optionalHealthCheckCount.Add(decimal.NewFromInt(1))
	}

	if d.Get("request_interval").String() == "10" {
		optionalHealthCheckCount = optionalHealthCheckCount.Add(decimal.NewFromInt(1))
	}

	if d.Get("measure_latency").Bool() {
		optionalHealthCheckCount = optionalHealthCheckCount.Add(decimal.NewFromInt(1))
	}

	if strings.HasSuffix(healthCheckType, "STR_MATCH") {
		optionalHealthCheckCount = optionalHealthCheckCount.Add(decimal.NewFromInt(1))
	}

	if optionalHealthCheckCount.GreaterThan(decimal.NewFromInt(0)) {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Optional features",
			Unit:            "months",
			UnitMultiplier:  1,
			MonthlyQuantity: decimalPtr(optionalHealthCheckCount),
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
		})
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
