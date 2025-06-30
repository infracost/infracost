package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/pulumi/putest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
)

func TestLambdaFunction(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	// Example Pulumi preview JSON with a Lambda function
	pulumiJSON := `{
		"steps": [{
			"resource": {
				"type": "aws:lambda/function:Function",
				"name": "my-function",
				"urn": "urn:pulumi:dev::test::aws:lambda/function:Function::my-function",
				"properties": {
					"name": "my-function",
					"memorySize": 512,
					"timeout": 30,
					"runtime": "nodejs14.x",
					"handler": "index.handler",
					"role": "arn:aws:iam::123456789012:role/lambda-role",
					"region": "us-east-1"
				}
			}
		}]
	}`

	usage := schema.NewUsageMap(map[string]interface{}{
		"my-function": map[string]interface{}{
			"monthly_requests":     10000000,
			"request_duration_ms":  350,
			"architecture":         "arm64",
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_lambda_function.my-function",
			CostComponentChecks: []testutil.CostComponentCheck{
				{Name: "Requests", MonthlyQuantity: testutil.FloatPtr(10)},
				{Name: "Duration (ARM)", MonthlyQuantity: testutil.FloatPtr(0.5)},
			},
		},
	}

	putest.ResourceTests(t, pulumiJSON, usage, resourceChecks)
}