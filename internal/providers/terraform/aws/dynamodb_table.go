package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"

	"github.com/infracost/infracost/internal/schema"
)

func GetDynamoDBTableRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name: "aws_dynamodb_table",
		Notes: []string{
			"DAX is not yet supported.",
		},
		RFunc: NewDynamoDBTable,
	}
}

func NewDynamoDBTable(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	replicaRegions := []string{}
	if d.Get("replica").Exists() {
		for _, data := range d.Get("replica").Array() {
			replicaRegions = append(replicaRegions, data.Get("region_name").String())
		}
	}

	args := &aws.DynamoDbTableArguments{
		Address:        d.Address,
		Region:         d.Get("region").String(),
		BillingMode:    d.Get("billing_mode").String(),
		WriteCapacity:  intPtr(d.Get("write_capacity").Int()),
		ReadCapacity:   intPtr(d.Get("read_capacity").Int()),
		ReplicaRegions: replicaRegions,
	}
	keysToSkipSync := []string{"region", "billing_mode", "write_capacity", "read_capacity", "replica_regions"}

	return aws.NewDynamoDBTable(args, u, keysToSkipSync)
}
