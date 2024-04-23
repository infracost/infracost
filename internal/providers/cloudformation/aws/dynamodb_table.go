package aws

import (
	"github.com/awslabs/goformation/v7/cloudformation/dynamodb"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetDynamoDBTableRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name: "AWS::DynamoDB::Table",
		Notes: []string{
			"DAX is not yet supported.",
		},
		RFunc: NewDynamoDBTable,
	}
}

func NewDynamoDBTable(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	cfr, ok := d.CFResource.(*dynamodb.Table)
	if !ok {
		logging.Logger.Debug().Msgf("Skipping resource %s as it did not have the expected type (got %T)", d.Address, d.CFResource)
		return nil
	}

	region := "us-east-1" // TODO figure out how to set region
	billingMode := cfr.BillingMode
	var readCapacity int64
	if cfr.ProvisionedThroughput != nil {
		readCapacity = int64(cfr.ProvisionedThroughput.ReadCapacityUnits)
	}
	var writeCapacity int64
	if cfr.ProvisionedThroughput != nil {
		writeCapacity = int64(cfr.ProvisionedThroughput.WriteCapacityUnits)
	}

	a := &aws.DynamoDBTable{
		Address:        d.Address,
		Region:         region,
		BillingMode:    *billingMode,
		WriteCapacity:  &writeCapacity,
		ReadCapacity:   &readCapacity,
		ReplicaRegions: []string{}, // Global Tables are defined using AWS::DynamoDB::GlobalTable
	}
	a.PopulateUsage(u)

	resource := a.BuildResource()

	return resource
}
