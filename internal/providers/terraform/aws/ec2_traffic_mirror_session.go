package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getEC2TrafficMirrorSessionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_ec2_traffic_mirror_session",
		RFunc: NewEC2TrafficMirrorSession,
	}
}
func NewEC2TrafficMirrorSession(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.EC2TrafficMirrorSession{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
