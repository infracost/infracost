package aws

import (
	"github.com/infracost/infracost/internal/schema"
)

func GetEC2TransitGatewayPeeringAttachmentRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_ec2_transit_gateway_peering_attachment",
		RFunc: NewEC2TransitGatewayPeeringAttachment,
		ReferenceAttributes: []string{
			"transit_gateway_id",
		},
	}
}

func NewEC2TransitGatewayPeeringAttachment(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	transitGatewayRefs := d.References("transit_gateway_id")
	if len(transitGatewayRefs) > 0 {
		region = transitGatewayRefs[0].Get("region").String()
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			transitGatewayAttachmentCostComponent(region, "TransitGatewayPeering"),
		},
	}
}
