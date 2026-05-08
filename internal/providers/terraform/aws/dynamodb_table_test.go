package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestDynamoDBTableGoldenFile(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	// IgnoreCLI: AWS provider v6 calls DynamoDB:DescribeTable during plan,
	// which fails with mock credentials.
	tftest.GoldenFileResourceTestsWithOpts(t, "dynamodb_table_test", &tftest.GoldenFileOptions{
		IgnoreCLI: true,
	})
}

func TestDynamoDBTableChinaGoldenFile(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tftest.GoldenFileResourceTestsWithOpts(t, "dynamodb_table_china_test", &tftest.GoldenFileOptions{
		Currency:  "CNY",
		IgnoreCLI: true,
	})
}
