package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/usage"
)

type VPCEndpoint struct {
	Address                string
	Region                 string
	Type                   string
	Interfaces             *int64
	MonthlyDataProcessedGb *float64 `infracost_usage:"monthly_data_processed_gb"`
}

func (r *VPCEndpoint) CoreType() string {
	return "VPCEndpoint"
}

func (r *VPCEndpoint) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_data_processed_gb", ValueType: schema.Float64, DefaultValue: 0},
	}
}

func (r *VPCEndpoint) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *VPCEndpoint) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{}

	vpcEndpointType := r.Type
	if vpcEndpointType == "" {
		vpcEndpointType = "Gateway"
	}

	vpcEndpointInterfaceCount := int64(1)
	if r.Interfaces != nil {
		vpcEndpointInterfaceCount = *r.Interfaces
	}

	if strings.ToLower(vpcEndpointType) == "gateway" {
		return &schema.Resource{
			Name:        r.Address,
			NoPrice:     true,
			IsSkipped:   true,
			UsageSchema: r.UsageSchema(),
		}
	}

	var dataProcessedGB *decimal.Decimal
	if r.MonthlyDataProcessedGb != nil {
		dataProcessedGB = decimalPtr(decimal.NewFromFloat(*r.MonthlyDataProcessedGb))
	}

	var endpointHours, endpointBytes string

	if strings.ToLower(vpcEndpointType) == "interface" {
		endpointHours = "VpcEndpoint-Hours"
		endpointBytes = "VpcEndpoint-Bytes"
		if dataProcessedGB != nil {
			gbLimits := []int{1000000, 4000000}
			tiers := usage.CalculateTierBuckets(*dataProcessedGB, gbLimits)

			if tiers[0].GreaterThan(decimal.NewFromInt(0)) {
				costComponents = append(costComponents, r.dataProcessedCostComponent(endpointBytes, "Data processed (first 1PB)", "0", &tiers[0]))
			}
			if tiers[1].GreaterThan(decimal.NewFromInt(0)) {
				costComponents = append(costComponents, r.dataProcessedCostComponent(endpointBytes, "Data processed (next 4PB)", "1048576", &tiers[1]))
			}
			if tiers[2].GreaterThan(decimal.NewFromInt(0)) {
				costComponents = append(costComponents, r.dataProcessedCostComponent(endpointBytes, "Data processed (over 5PB)", "5242880", &tiers[2]))
			}
		} else {
			costComponents = append(costComponents, r.dataProcessedCostComponent(endpointBytes, "Data processed (first 1PB)", "0", dataProcessedGB))
		}
	} else if strings.ToLower(vpcEndpointType) == "gatewayloadbalancer" {
		endpointHours = "VpcEndpoint-GWLBE-Hours"
		endpointBytes = "VpcEndpoint-GWLBE-Bytes"
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Data processed",
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: dataProcessedGB,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(r.Region),
				Service:       strPtr("AmazonVPC"),
				ProductFamily: strPtr("VpcEndpoint"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/i", endpointBytes))},
				},
			},
			UsageBased: true,
		})
	}

	costComponents = append(costComponents, &schema.CostComponent{
		Name:           fmt.Sprintf("Endpoint (%s)", vpcEndpointType),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(int64(vpcEndpointInterfaceCount))),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonVPC"),
			ProductFamily: strPtr("VpcEndpoint"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/i", endpointHours))},
			},
		},
	})

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *VPCEndpoint) dataProcessedCostComponent(endpointBytes string, displayName string, usageTier string, dataProcessedGB *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            displayName,
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: dataProcessedGB,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonVPC"),
			ProductFamily: strPtr("VpcEndpoint"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/i", endpointBytes))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(usageTier),
		},
		UsageBased: true,
	}
}
