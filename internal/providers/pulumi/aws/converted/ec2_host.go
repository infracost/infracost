package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getEC2HostRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_ec2_host",
		RFunc: newEC2Host,
	}
}

func newEC2Host(d *schema.ResourceData) schema.CoreResource {
	region := d.Get("region").String()
	r := &aws.EC2Host{
		Address:        d.Address,
		Region:         region,
		InstanceType:   d.Get("instanceType").String(),
		InstanceFamily: d.Get("instanceFamily").String(),
	}
	return r
}
