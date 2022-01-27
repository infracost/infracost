package aws

import (
	"strings"

	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getEC2TransitGatewayVpcAttachmentRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_ec2_transit_gateway_vpc_attachment",
		RFunc: NewEc2TransitGatewayVpcAttachment,
		ReferenceAttributes: []string{
			"transit_gateway_id",
			"vpc_id",
		},
	}
}
func NewEc2TransitGatewayVpcAttachment(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.Ec2TransitGatewayVpcAttachment{Address: d.Address, Region: d.Get("region").String()}

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
		r.VPCRegion = vpcRef.Get("region").String()
	}

	// Try to get the region from the transit gateway
	transitGatewayRefs := d.References("transit_gateway_id")
	if len(transitGatewayRefs) > 0 {
		r.TransitGatewayRegion = transitGatewayRefs[0].Get("region").String()
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
