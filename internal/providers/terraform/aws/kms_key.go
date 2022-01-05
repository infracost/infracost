package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getNewKMSKeyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_kms_key",
		RFunc: NewKmsKey,
	}
}
func NewKmsKey(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.KmsKey{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String()), CustomerMasterKeySpec: strPtr(d.Get("customer_master_key_spec").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
