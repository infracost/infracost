package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestLambdaFunction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_lambda_function" "lambda" {
			function_name = "lambda_function_name"
			role          = "arn:aws:lambda:us-east-1:account-id:resource-id"
			handler       = "exports.test"
			runtime       = "nodejs12.x"
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_lambda_function.lambda",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Requests",
					PriceHash:       "134034e58c7ef3bbaf513831c3a0161b-4a9dfd3965ffcbab75845ead7a27fd47",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.Zero),
				},
				{
					Name:            "Duration",
					PriceHash:       "a562fdf216894a62109f5b642a702f37-1786dd5ddb52682e127baa00bfaa4c48",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.Zero),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, resourceChecks)
}

func TestLambdaFunction_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_lambda_function" "lambda" {
			function_name = "lambda_function_name"
			role          = "arn:aws:lambda:us-east-1:account-id:resource-id"
			handler       = "exports.test"
			runtime       = "nodejs12.x"
		}

		resource "aws_lambda_function" "lambda_512_mem" {
			function_name = "lambda_function_name"
			role          = "arn:aws:lambda:us-east-1:account-id:resource-id"
			handler       = "exports.test"
			runtime       = "nodejs12.x"
			memory_size   = 512
		}

		data "infracost_aws_lambda_function" "lambda" {
		  resources = [aws_lambda_function.lambda.id, aws_lambda_function.lambda_512_mem.id]

			monthly_requests {
				value = 100000
			}

			average_request_duration {
				value = 350
			}
		}`

	requestCheck := testutil.CostComponentCheck{
		Name:            "Requests",
		PriceHash:       "134034e58c7ef3bbaf513831c3a0161b-4a9dfd3965ffcbab75845ead7a27fd47",
		HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100000)),
	}

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_lambda_function.lambda",
			CostComponentChecks: []testutil.CostComponentCheck{
				requestCheck,
				{
					Name:            "Duration",
					PriceHash:       "a562fdf216894a62109f5b642a702f37-1786dd5ddb52682e127baa00bfaa4c48",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(100000.0 * (128.0 / 1024.0) * 0.4)),
				},
			},
		},
		{
			Name: "aws_lambda_function.lambda_512_mem",
			CostComponentChecks: []testutil.CostComponentCheck{
				requestCheck,
				{
					Name:            "Duration",
					PriceHash:       "a562fdf216894a62109f5b642a702f37-1786dd5ddb52682e127baa00bfaa4c48",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(100000.0 * (512.0 / 1024.0) * 0.4)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, resourceChecks)
}
