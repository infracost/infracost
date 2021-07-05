package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"strings"

	"github.com/shopspring/decimal"
)

func GetEC2TransitGatewayVpcAttachmentRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_ec2_transit_gateway_vpc_attachment",
		RFunc: NewEC2TransitGatewayVpcAttachment,
		ReferenceAttributes: []string{
			"transit_gateway_id",
			"vpc_id",
		},
	}
}

func NewEC2TransitGatewayVpcAttachment(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	// Try to get the region from the VPC
	vpcRefs := d.References("vpc_id")
	var vpcRef *schema.ResourceData

	for _, ref := range vpcRefs {
		// the VPC ref can also be for the aws_subnet_ids resource which we don't want to consider
		if strings.ToLower(ref.Type) == "aws_default_vpc" || strings.ToLower(ref.Type) == "aws_vpc" {
			vpcRef = ref
			break
		}
	}
	if vpcRef != nil {
		region = vpcRef.Get("region").String()
	}

	// Try to get the region from the transit gateway
	transitGatewayRefs := d.References("transit_gateway_id")
	if len(transitGatewayRefs) > 0 {
		region = transitGatewayRefs[0].Get("region").String()
	}

	var gbDataProcessed *decimal.Decimal

	if u != nil && u.Get("monthly_data_processed_gb").Exists() {
		gbDataProcessed = decimalPtr(decimal.NewFromFloat(u.Get("monthly_data_processed_gb").Float()))
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			transitGatewayAttachmentCostComponent(region, "TransitGatewayVPC"),
			transitGatewayDataProcessingCostComponent(region, "TransitGatewayVPC", gbDataProcessed),
		},
	}
}

func transitGatewayAttachmentCostComponent(region string, operation string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "Transit gateway attachment",
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr("AmazonVPC"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/TransitGateway-Hours/")},
				{Key: "operation", Value: strPtr(operation)},
			},
		},
	}
}

func transitGatewayDataProcessingCostComponent(region string, operation string, gbDataProcessed *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Data processed",
		Unit:            "GB",
		UnitMultiplier:  1,
		MonthlyQuantity: gbDataProcessed,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr("AmazonVPC"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/TransitGateway-Bytes/")},
				{Key: "operation", Value: strPtr(operation)},
			},
		},
	}
}
