package aws_test

import (
	"testing"

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
					Name:            "Requests",
					PriceHash:       "4544b62fe649690f32a140f29a64d503-4a9dfd3965ffcbab75845ead7a27fd47",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.Zero),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, resourceChecks)
}

func TestSQSQueue_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
        resource "aws_sqs_queue" "fifo" {
            name = "my-fifo-queue"
            fifo_queue = true
		}

        resource "aws_sqs_queue" "standard" {
            name = "my-standard-queue"
            fifo_queue = false
		}
		
        data "infracost_aws_sqs_queue" "queue" {
            resources = [aws_sqs_queue.fifo.id, aws_sqs_queue.standard.id]

            monthly_requests {
                value = 1000000
            }

            request_size {
                value = 63
            }
        }
		`

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

	tftest.ResourceTests(t, tf, resourceChecks)
}
