package aws

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
	"github.com/shopspring/decimal"
)

func GetVpcEndpointRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_vpc_endpoint",
		RFunc: NewVpcEndpoint,
	}
}

func NewVpcEndpoint(ctx *config.ProjectContext, d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	costComponents := []*schema.CostComponent{}
	vpcEndpointType := "Gateway"
	vpcEndpointInterfaces := 1
	var endpointHours, endpointBytes string
	var gbDataProcessed *decimal.Decimal

	if d.Get("vpc_endpoint_type").Exists() {
		vpcEndpointType = d.Get("vpc_endpoint_type").String()
	}

	if len(d.Get("subnet_ids").Array()) > 1 {
		vpcEndpointInterfaces = len(d.Get("subnet_ids").Array())
	}

	// Gateway endpoints don't have a cost associated with them
	if strings.ToLower(vpcEndpointType) == "gateway" {
		return &schema.Resource{
			Name:      d.Address,
			NoPrice:   true,
			IsSkipped: true,
		}
	}

	if u != nil && u.Get("monthly_data_processed_gb").Exists() {
		gbDataProcessed = decimalPtr(decimal.NewFromFloat(u.Get("monthly_data_processed_gb").Float()))
	}

	if strings.ToLower(vpcEndpointType) == "interface" {
		endpointHours = "VpcEndpoint-Hours"
		endpointBytes = "VpcEndpoint-Bytes"
		if gbDataProcessed != nil {
			gbLimits := []int{1000, 4000}
			tiers := usage.CalculateTierBuckets(*gbDataProcessed, gbLimits)

			if tiers[0].GreaterThan(decimal.NewFromInt(0)) {
				costComponents = append(costComponents, vpcEndpointDataProcessedCostComponent(region, endpointBytes, "Data processed (first 1PB)", "0", &tiers[0]))
			}
			if tiers[1].GreaterThan(decimal.NewFromInt(0)) {
				costComponents = append(costComponents, vpcEndpointDataProcessedCostComponent(region, endpointBytes, "Data processed (next 4PB)", "1048576", &tiers[1]))
			}
			if tiers[2].GreaterThan(decimal.NewFromInt(0)) {
				costComponents = append(costComponents, vpcEndpointDataProcessedCostComponent(region, endpointBytes, "Data processed (over 5PB)", "5242880", &tiers[2]))
			}
		} else {
			costComponents = append(costComponents, vpcEndpointDataProcessedCostComponent(region, endpointBytes, "Data processed (first 1PB)", "0", gbDataProcessed))
		}
	} else if strings.ToLower(vpcEndpointType) == "gatewayloadbalancer" {
		endpointHours = "VpcEndpoint-GWLBE-Hours"
		endpointBytes = "VpcEndpoint-GWLBE-Bytes"
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Data processed",
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: gbDataProcessed,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonVPC"),
				ProductFamily: strPtr("VpcEndpoint"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/i", endpointBytes))},
				},
			},
		})
	}

	costComponents = append(costComponents, &schema.CostComponent{
		Name:           fmt.Sprintf("Endpoint (%s)", vpcEndpointType),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(int64(vpcEndpointInterfaces))),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonVPC"),
			ProductFamily: strPtr("VpcEndpoint"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/i", endpointHours))},
			},
		},
	})

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func vpcEndpointDataProcessedCostComponent(region string, endpointBytes string, displayName string, usageTier string, gbDataProcessed *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            displayName,
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: gbDataProcessed,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonVPC"),
			ProductFamily: strPtr("VpcEndpoint"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/i", endpointBytes))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(usageTier),
		},
	}
}
