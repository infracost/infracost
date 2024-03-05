package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

type KMSExternalKey struct {
	Address string
	Region  string
}

func (r *KMSExternalKey) CoreType() string {
	return "KMSExternalKey"
}

func (r *KMSExternalKey) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *KMSExternalKey) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *KMSExternalKey) BuildResource() *schema.Resource {
	kmsKey := &KMSKey{
		Region: r.Region,
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: []*schema.CostComponent{kmsKey.customerMasterKeyCostComponent()},
		UsageSchema:    r.UsageSchema(),
	}
}
