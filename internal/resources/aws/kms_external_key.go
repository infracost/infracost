package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

type NewKMSExternalKey struct {
	Address *string
	Region  *string
}

var NewKMSExternalKeyUsageSchema = []*schema.UsageItem{}

func (r *NewKMSExternalKey) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *NewKMSExternalKey) BuildResource() *schema.Resource {

	region := *r.Region

	costComponents := []*schema.CostComponent{
		CustomerMasterKeyCostComponent(region),
	}

	return &schema.Resource{
		Name:           *r.Address,
		CostComponents: costComponents, UsageSchema: NewKMSExternalKeyUsageSchema,
	}
}
