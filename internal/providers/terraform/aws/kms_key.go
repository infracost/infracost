package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetNewKMSKeyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_kms_key",
		RFunc: NewNewKMSKey,
	}
}
func NewNewKMSKey(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.NewKMSKey{Address: strPtr(d.Address), CustomerMasterKeySpec: strPtr(d.Get("customer_master_key_spec").String()), Region: strPtr(d.Get("region").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
