package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getEC2HostRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_ec2_host",
		RFunc: newEC2Host,
	}
}

func newEC2Host(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	r := &aws.EC2Host{
		Address:        d.Address,
		Region:         region,
		InstanceType:   d.Get("instance_type").String(),
		InstanceFamily: d.Get("instance_family").String(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
