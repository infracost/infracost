package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestNewDocDBClusterSnapshot(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_docdb_cluster_snapshot" "example" {
		db_cluster_identifier          = "fake"
		db_cluster_snapshot_identifier = "resourcetestsnapshot1234"
	  }`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_docdb_cluster_snapshot.example",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Backup storage",
					PriceHash:        "b508a58e978730edb23511dd40ad77d6-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestNewDocDBClusterSnapshot_usage(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_docdb_cluster_snapshot" "example" {
		db_cluster_identifier          = "fake"
		db_cluster_snapshot_identifier = "resourcetestsnapshot1234"
	  }`

	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_docdb_cluster_snapshot.example": map[string]interface{}{
			"backup_storage_gb": 1000,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_docdb_cluster_snapshot.example",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Backup storage",
					PriceHash:        "b508a58e978730edb23511dd40ad77d6-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(1000)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
