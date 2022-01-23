package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/usage"
	"github.com/shopspring/decimal"
)

type VpcEndpoint struct {
	Address                *string
	Region                 *string
	VpcEndpointType        *string
	VpcEndpointInterfaces  *int64
	MonthlyDataProcessedGb *float64 `infracost_usage:"monthly_data_processed_gb"`
}

var VpcEndpointUsageSchema = []*schema.UsageItem{{Key: "monthly_data_processed_gb", ValueType: schema.Float64, DefaultValue: 0}}

func (r *VpcEndpoint) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *VpcEndpoint) BuildResource() *schema.Resource {
	region := *r.Region
	costComponents := []*schema.CostComponent{}
	vpcEndpointType := "Gateway"
	vpcEndpointInterfaceCount := int64(1)
	var endpointHours, endpointBytes string
	var gbDataProcessed *decimal.Decimal

	if r.VpcEndpointType != nil {
		vpcEndpointType = *r.VpcEndpointType
	}

	if r.VpcEndpointInterfaces != nil {
		vpcEndpointInterfaceCount = *r.VpcEndpointInterfaces
	}

	if strings.ToLower(vpcEndpointType) == "gateway" {
		return &schema.Resource{
			Name:      *r.Address,
			NoPrice:   true,
			IsSkipped: true, UsageSchema: VpcEndpointUsageSchema,
		}
	}

	if r.MonthlyDataProcessedGb != nil {
		gbDataProcessed = decimalPtr(decimal.NewFromFloat(*r.MonthlyDataProcessedGb))
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
		HourlyQuantity: decimalPtr(decimal.NewFromInt(int64(vpcEndpointInterfaceCount))),
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
		Name:           *r.Address,
		CostComponents: costComponents, UsageSchema: VpcEndpointUsageSchema,
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
