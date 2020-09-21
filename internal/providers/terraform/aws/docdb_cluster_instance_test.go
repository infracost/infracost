package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/pkg/testutil"
	"github.com/shopspring/decimal"
)

func TestNewDocDBClusterInstance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_docdb_cluster_instance" "db" {
			cluster_identifier = "fake123"
			instance_class     = "db.t3.medium"
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_docdb_cluster_instance.db",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Database Instance (on_demand, db.t3.medium)",
					PriceHash:       "b21c3c7708229fb149bff23b4cfe6833-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:            "Storage",
					PriceHash:       "856b9e5bd87c953bffd1df698a6a1b3d-ee3dd7e4624338037ca6fea0933a662f",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.Zero),
				},
				{
					Name:            "I/O",
					PriceHash:       "c6f1f44a4f05ef22044c5af6490b6808-5be345988e7c9a0759c5cf8365868ee4",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.Zero),
				},
				{
					Name:            "Backup Storage",
					PriceHash:       "b508a58e978730edb23511dd40ad77d6-ee3dd7e4624338037ca6fea0933a662f",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.Zero),
				},
				{
					Name:            "CPU Credits",
					PriceHash:       "f6d2bda62e25c6eb08020075859e5a97-e8e892be2fbd1c8f42fd6761ad8977d8",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.Zero),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, resourceChecks)
}
