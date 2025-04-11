package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getEC2TransitGatewayPeeringAttachmentRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_ec2_transit_gateway_peering_attachment",
		RFunc: NewEC2TransitGatewayPeeringAttachment,
		ReferenceAttributes: []string{
			"transit_gateway_id",
		},
		GetRegion: func(defaultRegion string, d *schema.ResourceData) string {
			transitGatewayRefs := d.References("transitGatewayId")
			if len(transitGatewayRefs) > 0 {
				region := transitGatewayRefs[0].Get("region").String()
				if region != "" {
					return region
				}
			}

			return defaultRegion
		},
	}
}
func NewEC2TransitGatewayPeeringAttachment(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.EC2TransitGatewayPeeringAttachment{Address: d.Address, Region: d.Get("region").String()}

	transitGatewayRefs := d.References("transitGatewayId")
	if len(transitGatewayRefs) > 0 {
		r.TransitGatewayRegion = transitGatewayRefs[0].Get("region").String()
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
