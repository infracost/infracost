package aws

import (
	"fmt"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

func GetVpcEndpointRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_vpc_endpoint",
		RFunc: NewVpcEndpoint,
	}
}

func NewVpcEndpoint(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	region := d.Get("region").String()

	vpcEndpointType := "Gateway"

	var endpointHours string
	var endpointBytes string

	if d.Get("vpc_endpoint_type").Exists() {
		vpcEndpointType = d.Get("vpc_endpoint_type").String()
	}

	// Gateway endpoints don't have a cost associated with them
	if vpcEndpointType == "Gateway" {
		return nil
	}

	switch vpcEndpointType {
	case "Interface":
		endpointHours = "VpcEndpoint-Hours"
		endpointBytes = "VpcEndpoint-Bytes"
	case "GatewayLoadBalancer":
		endpointHours = "VpcEndpoint-GWLBE-Hours"
		endpointBytes = "VpcEndpoint-GWLBE-Bytes"
	}

	var gbDataProcessed *decimal.Decimal
	if u != nil && u.Get("monthly_gb_data_processed.0.value").Exists() {
		gbDataProcessed = decimalPtr(decimal.NewFromFloat(u.Get("monthly_gb_data_processed.0.value").Float()))
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:           fmt.Sprintf("VPC %s endpoint", vpcEndpointType),
				Unit:           "hours",
				UnitMultiplier: 1,
				HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonVPC"),
					ProductFamily: strPtr("VpcEndpoint"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/", endpointHours))},
					},
				},
			},
			{
				Name:            "Data processed",
				Unit:            "GB",
				UnitMultiplier:  1,
				MonthlyQuantity: gbDataProcessed,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonVPC"),
					ProductFamily: strPtr("VpcEndpoint"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/", endpointBytes))},
					},
				},
			},
		},
	}
}
