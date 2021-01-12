package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestNewECRRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
        resource "aws_ecr_repository" "repo" {
            name = "my-ecr-repo"
		}
		`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_ecr_repository.repo",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Storage",
					PriceHash:       "905a9ae7cd5892895ec711d45290c5a0-ee3dd7e4624338037ca6fea0933a662f",
					HourlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
