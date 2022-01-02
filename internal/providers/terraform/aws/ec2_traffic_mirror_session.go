package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getEC2TrafficMirroSessionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_ec2_traffic_mirror_session",
		RFunc: NewEc2TrafficMirrorSession,
	}
}
func NewEc2TrafficMirrorSession(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.Ec2TrafficMirrorSession{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
