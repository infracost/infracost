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
		ReferenceAttributes: []string{"awsAppautoscalingTarget.resourceId"},
		RFunc:               NewDynamoDBTableResource,
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

func NewDynamoDBTableResource(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	replicaRegions := []string{}
	if d.Get("replica").Exists() {
		for _, data := range d.Get("replica").Array() {
			replicaRegions = append(replicaRegions, data.Get("regionName").String())
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
		BillingMode:          d.GetStringOrDefault("billingMode", "PROVISIONED"),
		WriteCapacity:        intPtr(d.Get("writeCapacity").Int()),
		ReadCapacity:         intPtr(d.Get("readCapacity").Int()),
		ReplicaRegions:       replicaRegions,
		AppAutoscalingTarget: targets,
	}
	a.PopulateUsage(u)

	return a.BuildResource()
}
