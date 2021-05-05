package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestNewDocDBCluster(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "aws_docdb_cluster" "docdb" {
		cluster_identifier      = "my-docdb-cluster"
		engine                  = "docdb"
		master_username         = "foo"
		master_password         = "mustbeeightchars"
		backup_retention_period = 5
		preferred_backup_window = "07:00-09:00"
		skip_final_snapshot     = true
	  }`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_docdb_cluster.docdb",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Backup storage",
					PriceHash:        "b508a58e978730edb23511dd40ad77d6-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks, tmpDir)
}

func TestNewDocDBCluster_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "aws_docdb_cluster" "docdb" {
		cluster_identifier      = "my-docdb-cluster"
		engine                  = "docdb"
		master_username         = "foo"
		master_password         = "mustbeeightchars"
		backup_retention_period = 5
		preferred_backup_window = "07:00-09:00"
		skip_final_snapshot     = true
	  }`

	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_docdb_cluster.docdb": map[string]interface{}{
			"backup_storage_gb": 1000,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_docdb_cluster.docdb",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Backup storage",
					PriceHash:        "b508a58e978730edb23511dd40ad77d6-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(1000)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks, tmpDir)
}
