package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getAppAutoscalingTargetRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_appautoscaling_target",
		RFunc: NewAppAutoscalingTargetResource,
		// This reference is used by other resources (e.g. DynamoDBTable) to generate
		// a reverse reference
		ReferenceAttributes: []string{"resource_id"},
	}
}

func NewAppAutoscalingTargetResource(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := newAppAutoscalingTarget(d, u)
	return r.BuildResource()
}

func newAppAutoscalingTarget(d *schema.ResourceData, u *schema.UsageData) *aws.AppAutoscalingTarget {
	r := &aws.AppAutoscalingTarget{
		Address:           d.Address,
		Region:            d.Get("region").String(),
		ResourceID:        d.Get("resourceId").String(),
		ScalableDimension: d.Get("scalableDimension").String(),
		MinCapacity:       d.Get("minCapacity").Int(),
		MaxCapacity:       d.Get("maxCapacity").Int(),
	}

	r.PopulateUsage(u)

	return r
}