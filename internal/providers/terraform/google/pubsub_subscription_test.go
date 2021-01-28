package google_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestPubSubSubscription(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "google_pubsub_subscription" "my_subscription" {
			name  = "example-subscription"
			topic = "my_topic"
		}
	`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_pubsub_subscription.my_subscription",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Message delivery data",
					PriceHash:        "cd1eba29e63e40fed467ba3eb7c5a6e6-a17de3a1a22680d865c82275b34c9d21",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
				{
					Name:             "Retained acknowledged message storage",
					PriceHash:        "2f872d150153ed13eee94c2bbcc2b69f-57bc5d148491a8381abaccb21ca6b4e9",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
				{
					Name:             "Snapshot message backlog storage",
					PriceHash:        "c70f49a01dcd455baf39b999e45961d3-57bc5d148491a8381abaccb21ca6b4e9",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}
	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestPubSubSubscription_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "google_pubsub_subscription" "my_subscription" {
			name  = "example-subscription"
			topic = "my_topic"
		}
	`

	usage := schema.NewUsageMap(map[string]interface{}{
		"google_pubsub_subscription.my_subscription": map[string]interface{}{
			"monthly_message_data_tb": 10,
			"storage_gb":              20,
			"snapshot_storage_gb":     30,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_pubsub_subscription.my_subscription",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Message delivery data",
					PriceHash:        "cd1eba29e63e40fed467ba3eb7c5a6e6-a17de3a1a22680d865c82275b34c9d21",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(10)),
				},
				{
					Name:             "Retained acknowledged message storage",
					PriceHash:        "2f872d150153ed13eee94c2bbcc2b69f-57bc5d148491a8381abaccb21ca6b4e9",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(20)),
				},
				{
					Name:             "Snapshot message backlog storage",
					PriceHash:        "c70f49a01dcd455baf39b999e45961d3-57bc5d148491a8381abaccb21ca6b4e9",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(30)),
				},
			},
		},
	}
	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
