package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"

	"github.com/infracost/infracost/internal/schema"
)

func getDynamoDBTableRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name: "aws_dynamodb_table",
		Notes: []string{
			"DAX is not yet supported.",
		},
		ReferenceAttributes: []string{"aws_appautoscaling_target.resource_id"},
		RFunc:               NewDynamoDBTableResource,
		CustomRefIDFunc:     customRefIDs,
	}
}

// customRefIDs returns a table name that will match the custom format used by aws_appautoscaling_target.resource_id
func customRefIDs(d *schema.ResourceData) []string {
	name := d.Get("name").String()
	if name != "" {
		return []string{"table/" + name}
	}
	return nil
}

func NewDynamoDBTableResource(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	replicaRegions := []string{}
	if d.Get("replica").Exists() {
		for _, data := range d.Get("replica").Array() {
			replicaRegions = append(replicaRegions, data.Get("region_name").String())
		}
	}

	targets := []*aws.AppAutoscalingTarget{}
	for _, ref := range d.References("aws_appautoscaling_target.resource_id") {
		targets = append(targets, newAppAutoscalingTarget(ref, ref.UsageData))
	}

	a := &aws.DynamoDBTable{
		Address:              d.Address,
		Region:               d.Get("region").String(),
		Name:                 d.Get("name").String(),
		BillingMode:          d.Get("billing_mode").String(),
		WriteCapacity:        intPtr(d.Get("write_capacity").Int()),
		ReadCapacity:         intPtr(d.Get("read_capacity").Int()),
		ReplicaRegions:       replicaRegions,
		AppAutoscalingTarget: targets,
	}
	a.PopulateUsage(u)

	return a.BuildResource()
}
