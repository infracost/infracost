package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// MemoryDBACL represents an AWS MemoryDB ACL
//
// Resource information: https://docs.aws.amazon.com/memorydb/latest/devguide/clusters.acls.html
// Pricing information: Free
type MemoryDBACL struct {
	Address string
	Region  string
}

func (r *MemoryDBACL) CoreType() string {
	return "MemoryDBACL"
}

func (r *MemoryDBACL) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *MemoryDBACL) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *MemoryDBACL) BuildResource() *schema.Resource {
	return &schema.Resource{
		Name:        r.Address,
		NoPrice:     true,
		IsSkipped:   true,
		SkipMessage: "Free resource.",
		UsageSchema: r.UsageSchema(),
	}
}
