package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetNewKMSKeyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "states.aws.kms.key.present",
		RFunc: NewKMSKey,
	}
}

func NewKMSKey(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.KMSKey{
		Address:               d.Address,
		Region:                d.Get("region").String(),
		CustomerMasterKeySpec: d.Get("key_spec").String(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
