package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getSpotInstanceRequestRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name: "aws_spot_instance_request",
		Notes: []string{
			"Notes",
		},
		RFunc: newSpotInstanceRequest,
	}
}

func newSpotInstanceRequest(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	var instanceType, ami string

	ami = d.GetStringOrDefault("ami", ami)
	instanceType = d.GetStringOrDefault("instance_type", instanceType)

	r := &aws.Instance{
		Address:        d.Address,
		Region:         region,
		PurchaseOption: "spot",
		InstanceType:   instanceType,
		AMI:            ami,
	}
	r.PopulateUsage(u)

	return r.BuildResource()
}
