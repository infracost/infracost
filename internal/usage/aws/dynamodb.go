package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func dynamodbNewClient(region string) (*dynamodb.Client, error) {
	config, err := sdkNewConfig(region)
	if err != nil {
		return nil, err
	}
	return dynamodb.NewFromConfig(config), nil
}

func dynamodbGetStorageBytes(region string, table string) float64 {
	client, err := dynamodbNewClient(region)
	if err != nil {
		sdkWarn("DynamoDB", "storage_gb", table, err)
		return 0
	}
	result, err := client.DescribeTable(context.TODO(), &dynamodb.DescribeTableInput{TableName: strPtr(table)})
	if err != nil {
		sdkWarn("DynamoDB", "storage_gb", table, err)
		return 0
	}
	return float64(result.Table.TableSizeBytes)
}
