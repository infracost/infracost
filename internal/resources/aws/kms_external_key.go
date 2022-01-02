package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

type KmsExternalKey struct {
	Address *string
	Region  *string
}

var KmsExternalKeyUsageSchema = []*schema.UsageItem{}

func (r *KmsExternalKey) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *KmsExternalKey) BuildResource() *schema.Resource {

	region := *r.Region

	costComponents := []*schema.CostComponent{
		CustomerMasterKeyCostComponent(region),
	}

	return &schema.Resource{
		Name:           *r.Address,
		CostComponents: costComponents, UsageSchema: KmsExternalKeyUsageSchema,
	}
}
