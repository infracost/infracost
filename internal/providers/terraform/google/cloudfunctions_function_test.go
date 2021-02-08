package google_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
)

func TestCloudFunctions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "google_cloudfunctions_function" "function" {
			name        = "function-test"
			description = "My function"
			runtime     = "nodejs10"
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_cloudfunctions_function.function",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:      "CPU Time",
					PriceHash: "49ab2e8c10b6537ef08737f3a7ae9d8d-1b88db789f176f93c9bfe3a2b0616ddd",
				},
				{
					Name:      "Memory Time",
					PriceHash: "f75b0be4dd48eb627f5b50332da5dae6-54bcde0ad0d3ad3534913736ebf89cd7",
				},
				{
					Name:      "Invocations",
					PriceHash: "fcfa5cdd12792fc691843752c9bdfb37-d1cef0b8c8b290c869ada456280c923d",
				},
				{
					Name:      "Outbound Data",
					PriceHash: "c8a783f8b465c7522c5d2fd48953e4cf-679daa94b986938e2c556da1f34c6f86",
				},
			},
		},
	}
	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
