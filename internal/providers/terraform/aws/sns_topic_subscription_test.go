package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestSNSTopicSubscriptionFunction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
        resource "aws_sns_topic_subscription" "http" {
          endpoint = "some-dummy-endpoint"
          protocol = "http"
          topic_arn = "aws_sns_topic.topic.arn"
        }
`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_sns_topic_subscription.http",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "HTTP notifications",
					PriceHash:        "0b289a170f868cdf934546d365df8097-3a73d6a2c60c01675dc5432bc383db67",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
