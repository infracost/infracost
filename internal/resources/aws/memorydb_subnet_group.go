package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// MemoryDBSubnetGroup represents an AWS MemoryDB subnet group
//
// Resource information: https://docs.aws.amazon.com/memorydb/latest/devguide/subnet-groups.html
// Pricing information: Free
type MemoryDBSubnetGroup struct {
	Address string
	Region  string
}

func (r *MemoryDBSubnetGroup) CoreType() string {
	return "MemoryDBSubnetGroup"
}

func (r *MemoryDBSubnetGroup) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *MemoryDBSubnetGroup) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *MemoryDBSubnetGroup) BuildResource() *schema.Resource {
	return &schema.Resource{
		Name:        r.Address,
		NoPrice:     true,
		IsSkipped:   true,
		SkipMessage: "Free resource.",
		UsageSchema: r.UsageSchema(),
	}
}
