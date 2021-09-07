package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func dynamodbNewClient(ctx context.Context, region string) (*dynamodb.Client, error) {
	cfg, err := getConfig(ctx, region)
	if err != nil {
		return nil, err
	}
	return dynamodb.NewFromConfig(cfg), nil
}

func dynamodbGetStorageBytes(ctx context.Context, region string, table string) float64 {
	client, err := dynamodbNewClient(ctx, region)
	if err != nil {
		sdkWarn("DynamoDB", "storage_gb", table, err)
		return 0
	}
	result, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{TableName: strPtr(table)})
	if err != nil {
		sdkWarn("DynamoDB", "storage_gb", table, err)
		return 0
	}
	return float64(result.Table.TableSizeBytes)
}
