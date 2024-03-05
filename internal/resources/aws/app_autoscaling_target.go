package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

type AppAutoscalingTarget struct {
	Address string
	Region  string

	ResourceID        string
	ScalableDimension string

	MinCapacity int64
	MaxCapacity int64

	// "usage" args
	Capacity *int64 `infracost_usage:"capacity"`
}

func (r *AppAutoscalingTarget) CoreType() string {
	return "AppAutoscalingTarget"
}

func (r *AppAutoscalingTarget) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "capacity", ValueType: schema.Int64, DefaultValue: 0},
	}
}

func (r *AppAutoscalingTarget) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *AppAutoscalingTarget) BuildResource() *schema.Resource {
	return &schema.Resource{
		Name:        r.Address,
		UsageSchema: r.UsageSchema(),
	}
}
