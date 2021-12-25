package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

func GetEIPRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_eip",
		RFunc: NewEIP,
	}
}
func NewEIP(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.EIP{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	if d.Get("customer_owned_ipv4_pool").Exists() && d.Get("customer_owned_ipv4_pool").Type != gjson.Null {
		r.CustomerOwnedIpv4Pool = strPtr(d.Get("customer_owned_ipv4_pool").String())
	}
	if d.Get("instance").Exists() && d.Get("instance").Type != gjson.Null {
		r.Instance = strPtr(d.Get("instance").String())
	}
	if d.Get("network_interface").Exists() && d.Get("network_interface").Type != gjson.Null {
		r.NetworkInterface = strPtr(d.Get("network_interface").String())
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
