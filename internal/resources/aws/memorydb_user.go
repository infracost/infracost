package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// MemoryDBUser represents an AWS MemoryDB user
//
// Resource information: https://docs.aws.amazon.com/memorydb/latest/devguide/users-and-roles.html
// Pricing information: Free
type MemoryDBUser struct {
	Address string
	Region  string
}

func (r *MemoryDBUser) CoreType() string {
	return "MemoryDBUser"
}

func (r *MemoryDBUser) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *MemoryDBUser) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *MemoryDBUser) BuildResource() *schema.Resource {
	return &schema.Resource{
		Name:        r.Address,
		NoPrice:     true,
		IsSkipped:   true,
		SkipMessage: "Free resource.",
		UsageSchema: r.UsageSchema(),
	}
}
