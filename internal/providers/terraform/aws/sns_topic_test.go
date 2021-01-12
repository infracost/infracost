package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestSQSTopicFunction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
        resource "aws_sns_topic" "topic" {
            name = "my-standard-queue"
        }`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_sns_topic.topic",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Requests",
					PriceHash:        "33cb108100e772bb071322e9b4736e98-4a9dfd3965ffcbab75845ead7a27fd47",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
