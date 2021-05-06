package google_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestPubSubTopic(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "google_pubsub_topic" "my_topic" {
			name = "example-topic"
		}
	`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_pubsub_topic.my_topic",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Message ingestion data",
					PriceHash:        "cd1eba29e63e40fed467ba3eb7c5a6e6-a17de3a1a22680d865c82275b34c9d21",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}
	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestPubSubTopic_usage(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "google_pubsub_topic" "my_topic" {
			name = "example-topic"
		}
	`

	usage := schema.NewUsageMap(map[string]interface{}{
		"google_pubsub_topic.my_topic": map[string]interface{}{
			"monthly_message_data_tb": 10,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_pubsub_topic.my_topic",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Message ingestion data",
					PriceHash:        "cd1eba29e63e40fed467ba3eb7c5a6e6-a17de3a1a22680d865c82275b34c9d21",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(10)),
				},
			},
		},
	}
	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
