//nolint:deadcode,unused
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

func DynamoDBGetStorageBytes(ctx context.Context, region string, table string) (int64, error) {
	client, err := dynamodbNewClient(ctx, region)
	if err != nil {
		return 0, err
	}
	result, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{TableName: strPtr(table)})
	if err != nil {
		return 0, err
	}
	return result.Table.TableSizeBytes, nil
}
