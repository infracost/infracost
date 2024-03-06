package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type Route53HealthCheck struct {
	Address         string
	RequestInterval string
	MeasureLatency  bool
	Type            string
	EndpointType    *string `infracost_usage:"endpoint_type"`
}

func (r *Route53HealthCheck) CoreType() string {
	return "Route53HealthCheck"
}

func (r *Route53HealthCheck) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "endpoint_type", ValueType: schema.String, DefaultValue: "aws"},
	}
}

func (r *Route53HealthCheck) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *Route53HealthCheck) BuildResource() *schema.Resource {
	costComponents := make([]*schema.CostComponent, 0)

	endpointType := "aws"
	usageAmount := "50"
	if r.EndpointType != nil {
		endpointType = strings.Replace(*r.EndpointType, "_", "-", 1)
		if strings.ToLower(endpointType) == "non-aws" {
			usageAmount = "0"
		}
	}

	costComponents = append(costComponents, &schema.CostComponent{
		Name:            "Health check",
		Unit:            "months",
		UnitMultiplier:  decimal.NewFromInt(1),
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

	healthCheckType := r.Type
	optionalHealthCheckCount := decimal.Zero

	if strings.HasPrefix(healthCheckType, "HTTPS") {
		optionalHealthCheckCount = optionalHealthCheckCount.Add(decimal.NewFromInt(1))
	}

	if r.RequestInterval == "10" {
		optionalHealthCheckCount = optionalHealthCheckCount.Add(decimal.NewFromInt(1))
	}

	if r.MeasureLatency {
		optionalHealthCheckCount = optionalHealthCheckCount.Add(decimal.NewFromInt(1))
	}

	if strings.HasSuffix(healthCheckType, "STR_MATCH") {
		optionalHealthCheckCount = optionalHealthCheckCount.Add(decimal.NewFromInt(1))
	}

	if optionalHealthCheckCount.GreaterThan(decimal.NewFromInt(0)) {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Optional features",
			Unit:            "months",
			UnitMultiplier:  decimal.NewFromInt(1),
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
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}
