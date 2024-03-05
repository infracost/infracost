package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getAppAutoscalingTargetRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_appautoscaling_target",
		CoreRFunc: NewAppAutoscalingTargetResource,
		// This reference is used by other resources (e.g. DynamoDBTable) to generate
		// a reverse reference
		ReferenceAttributes: []string{"resource_id"},
	}
}

func NewAppAutoscalingTargetResource(d *schema.ResourceData) schema.CoreResource {
	return newAppAutoscalingTarget(d, nil)

}

func newAppAutoscalingTarget(d *schema.ResourceData, u *schema.UsageData) *aws.AppAutoscalingTarget {
	r := &aws.AppAutoscalingTarget{
		Address:           d.Address,
		Region:            d.Get("region").String(),
		ResourceID:        d.Get("resource_id").String(),
		ScalableDimension: d.Get("scalable_dimension").String(),
		MinCapacity:       d.Get("min_capacity").Int(),
		MaxCapacity:       d.Get("max_capacity").Int(),
	}

	r.PopulateUsage(u)

	return r
}
