package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

type Ec2TransitGatewayPeeringAttachment struct {
	Address              string
	Region               string
	TransitGatewayRegion string
}

var Ec2TransitGatewayPeeringAttachmentUsageSchema = []*schema.UsageItem{}

func (r *Ec2TransitGatewayPeeringAttachment) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *Ec2TransitGatewayPeeringAttachment) BuildResource() *schema.Resource {
	region := r.Region
	if r.TransitGatewayRegion != "" {
		region = r.TransitGatewayRegion
	}

	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			transitGatewayAttachmentCostComponent(region, "TransitGatewayPeering"),
		}, UsageSchema: Ec2TransitGatewayPeeringAttachmentUsageSchema,
	}
}
