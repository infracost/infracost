package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getEIPRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_eip",
		RFunc: NewEIP,
	}
}
func NewEIP(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.EIP{
		Address:               d.Address,
		Region:                d.Get("region").String(),
		CustomerOwnedIPv4Pool: d.Get("customer_owned_ipv4_pool").String(),
		NetworkInterface:      d.Get("network_interface").String(),
		Instance:              d.Get("instance").String(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
