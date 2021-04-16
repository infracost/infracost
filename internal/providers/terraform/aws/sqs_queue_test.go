package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestSQSQueueFunction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
        resource "aws_sqs_queue" "standard" {
            name = "my-standard-queue"
            fifo_queue = false
        }`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_sqs_queue.standard",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Requests",
					PriceHash:        "4544b62fe649690f32a140f29a64d503-4a9dfd3965ffcbab75845ead7a27fd47",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestSQSQueue_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_sqs_queue" "fifo" {
			name = "my.fifo"
			fifo_queue = true
		}

		resource "aws_sqs_queue" "standard" {
			name = "my-standard-queue"
			fifo_queue = false
		}`

	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_sqs_queue.fifo": map[string]interface{}{
			"monthly_requests": 1000000,
			"request_size":     63,
		},
		"aws_sqs_queue.standard": map[string]interface{}{
			"monthly_requests": 1000000,
			"request_size":     63,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_sqs_queue.standard",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Requests",
					PriceHash:       "4544b62fe649690f32a140f29a64d503-4a9dfd3965ffcbab75845ead7a27fd47",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(1000000)),
				},
			},
		},
		{
			Name: "aws_sqs_queue.fifo",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Requests",
					PriceHash:       "95a3926f56d0b124e2fd52c64b00924e-4a9dfd3965ffcbab75845ead7a27fd47",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(1000000)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
