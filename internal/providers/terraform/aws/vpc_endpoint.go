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

func NewVpcEndpoint(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	vpcEndpointType := "Gateway"

	vpcEndpointInterfaces := 1

	var endpointHours string
	var endpointBytes string

	if d.Get("vpc_endpoint_type").Exists() {
		vpcEndpointType = d.Get("vpc_endpoint_type").String()
	}

	if len(d.Get("subnet_ids").Array()) > 1 {
		vpcEndpointInterfaces = len(d.Get("subnet_ids").Array())
	}

	// Gateway endpoints don't have a cost associated with them
	if vpcEndpointType == "Gateway" {
		return &schema.Resource{
			NoPrice:   true,
			IsSkipped: true,
		}
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
	if u != nil && u.Get("monthly_data_processed_gb").Exists() {
		gbDataProcessed = decimalPtr(decimal.NewFromFloat(u.Get("monthly_data_processed_gb").Float()))
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:           fmt.Sprintf("Endpoint (%s)", vpcEndpointType),
				Unit:           "hours",
				UnitMultiplier: 1,
				HourlyQuantity: decimalPtr(decimal.NewFromInt(int64(vpcEndpointInterfaces))),
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
