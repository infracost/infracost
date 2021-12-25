package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetEC2TrafficMirroSessionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_ec2_traffic_mirror_session",
		RFunc: NewEC2TrafficMirroSession,
	}
}
func NewEC2TrafficMirroSession(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.EC2TrafficMirroSession{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
