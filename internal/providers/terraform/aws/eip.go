package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getEIPRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_eip",
		RFunc: NewEip,
	}
}
func NewEip(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.Eip{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	if !d.IsEmpty("network_interface") {
		r.NetworkInterface = strPtr(d.Get("network_interface").String())
	}
	if !d.IsEmpty("customer_owned_ipv4_pool") {
		r.CustomerOwnedIpv4Pool = strPtr(d.Get("customer_owned_ipv4_pool").String())
	}
	if !d.IsEmpty("instance") {
		r.Instance = strPtr(d.Get("instance").String())
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
