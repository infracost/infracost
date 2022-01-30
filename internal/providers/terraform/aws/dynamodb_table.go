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
		RFunc: NewDynamoDBTableResource,
	}
}

func NewDynamoDBTableResource(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	replicaRegions := []string{}
	if d.Get("replica").Exists() {
		for _, data := range d.Get("replica").Array() {
			replicaRegions = append(replicaRegions, data.Get("region_name").String())
		}
	}

	a := &aws.DynamoDBTable{
		Address:        d.Address,
		Region:         d.Get("region").String(),
		Name:           d.Get("name").String(),
		BillingMode:    d.Get("billing_mode").String(),
		WriteCapacity:  intPtr(d.Get("write_capacity").Int()),
		ReadCapacity:   intPtr(d.Get("read_capacity").Int()),
		ReplicaRegions: replicaRegions,
	}
	a.PopulateUsage(u)

	return a.BuildResource()
}
