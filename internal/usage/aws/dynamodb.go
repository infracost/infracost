//nolint:deadcode,unused
package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/infracost/infracost/internal/logging"
)

func dynamodbNewClient(ctx context.Context, region string) (*dynamodb.Client, error) {
	cfg, err := getConfig(ctx, region)
	if err != nil {
		return nil, err
	}
	return dynamodb.NewFromConfig(cfg), nil
}

func dynamodbGetRequests(ctx context.Context, region string, table string, metric string) (float64, error) {
	logging.Logger.Debug().Msgf("Querying AWS CloudWatch: AWS/DynamoDB %s (region: %s, TableName: %s)", metric, region, table)
	stats, err := cloudwatchGetMonthlyStats(ctx, statsRequest{
		region:     region,
		namespace:  "AWS/DynamoDB",
		metric:     metric,
		dimensions: map[string]string{"TableName": table},
		statistic:  statSum,
		unit:       unitCount,
	})
	if err != nil {
		return 0, err
	}
	if len(stats.Datapoints) == 0 {
		return 0, nil
	}
	return *stats.Datapoints[0].Sum, nil
}

func DynamoDBGetStorageBytes(ctx context.Context, region string, table string) (int64, error) {
	client, err := dynamodbNewClient(ctx, region)
	if err != nil {
		return 0, err
	}
	logging.Logger.Debug().Msgf("Querying AWS DynamoDB API: DescribeTable(region: %s, table: %s)", region, table)
	result, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{TableName: strPtr(table)})
	if err != nil {
		return 0, err
	}

	if result.Table == nil {
		return 0, nil
	}

	if result.Table.TableSizeBytes == nil {
		return 0, nil
	}

	return *result.Table.TableSizeBytes, nil
}

func DynamoDBGetRRU(ctx context.Context, region string, table string) (float64, error) {
	return dynamodbGetRequests(ctx, region, table, "ConsumedReadCapacityUnits")
}

func DynamoDBGetWRU(ctx context.Context, region string, table string) (float64, error) {
	return dynamodbGetRequests(ctx, region, table, "ConsumedWriteCapacityUnits")
}
