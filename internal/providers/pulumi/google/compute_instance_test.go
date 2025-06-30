package google_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/pulumi/putest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
)

func TestComputeInstance(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	// Example Pulumi preview JSON with a Compute Instance
	pulumiJSON := `{
		"steps": [{
			"resource": {
				"type": "gcp:compute/instance:Instance",
				"name": "my-instance",
				"urn": "urn:pulumi:dev::test::gcp:compute/instance:Instance::my-instance",
				"properties": {
					"name": "my-instance",
					"machineType": "n1-standard-2",
					"zone": "us-central1-a",
					"bootDisk": {
						"initializeParams": {
							"image": "debian-cloud/debian-9",
							"size": 50,
							"type": "pd-ssd"
						}
					},
					"networkInterfaces": [{
						"network": "default",
						"accessConfigs": [{
							"natIp": "",
							"networkTier": "PREMIUM"
						}]
					}],
					"scheduling": {
						"preemptible": false
					}
				}
			}
		}]
	}`

	usage := schema.NewUsageMap(map[string]interface{}{
		"my-instance": map[string]interface{}{
			"monthly_hours": 730,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_compute_instance.my-instance",
			CostComponentChecks: []testutil.CostComponentCheck{
				{Name: "Instance usage (Linux/UNIX, on-demand, n1-standard-2)", MonthlyQuantity: testutil.FloatPtr(730)},
				{Name: "Storage (SSD, us-central1)", MonthlyQuantity: testutil.FloatPtr(50)},
			},
		},
	}

	putest.ResourceTests(t, pulumiJSON, usage, resourceChecks)
}