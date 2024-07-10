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
		// this is a reverse reference, it depends on the aws_appautoscaling_target RegistryItem
		// defining "resource_id" as a ReferenceAttribute
		ReferenceAttributes: []string{"aws_appautoscaling_target.resource_id", "replica.0.propagate_tags"},
		CoreRFunc:           NewDynamoDBTableResource,
		CustomRefIDFunc: func(d *schema.ResourceData) []string {
			// returns a table name that will match the custom format used by aws_appautoscaling_target.resource_id
			name := d.Get("name").String()
			if name != "" {
				return []string{"table/" + name}
			}
			return nil
		},
	}
}

func NewDynamoDBTableResource(d *schema.ResourceData) schema.CoreResource {
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

	var pitrEnabled bool
	if d.Get("point_in_time_recovery").Exists() &&
		len(d.Get("point_in_time_recovery").Array()) > 0 {
		pitrEnabled = d.Get("point_in_time_recovery").Array()[0].Get("enabled").Bool()
	}

	a := &aws.DynamoDBTable{
		Address:                    d.Address,
		Region:                     d.Get("region").String(),
		Name:                       d.Get("name").String(),
		BillingMode:                d.GetStringOrDefault("billing_mode", "PROVISIONED"),
		WriteCapacity:              intPtr(d.Get("write_capacity").Int()),
		ReadCapacity:               intPtr(d.Get("read_capacity").Int()),
		ReplicaRegions:             replicaRegions,
		AppAutoscalingTarget:       targets,
		PointInTimeRecoveryEnabled: pitrEnabled,
	}
	return a
}
