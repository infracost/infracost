package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getNewKMSExternalKeyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_kms_external_key",
		RFunc: NewKmsExternalKey,
	}
}
func NewKmsExternalKey(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.KmsExternalKey{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
