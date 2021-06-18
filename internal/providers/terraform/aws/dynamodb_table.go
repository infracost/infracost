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
	region := d.Get("region").String()

	billingMode := d.Get("billing_mode").String()

	var readCapacity int64
	if d.Get("read_capacity").Exists() {
		readCapacity = d.Get("read_capacity").Int()
	}

	var writeCapacity int64
	if d.Get("write_capacity").Exists() {
		writeCapacity = d.Get("write_capacity").Int()
	}

	replicaRegions := []string{}
	if d.Get("replica").Exists() {
		for _, data := range d.Get("replica").Array() {
			replicaRegions = append(replicaRegions, data.Get("region_name").String())
		}
	}

	args := &aws.DynamoDbTableArguments{
		Address:        d.Address,
		Region:         region,
		BillingMode:    billingMode,
		WriteCapacity:  writeCapacity,
		ReadCapacity:   readCapacity,
		ReplicaRegions: replicaRegions,
	}
	args.PopulateUsage(u)

	return aws.NewDynamoDBTable(args)
}
