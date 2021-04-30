package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestDocDBClusterInstanceGoldenFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tftest.GoldenFileResourceTests(t, "docdb_cluster_instance_test")
}

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
					Name:            "Database instance (on-demand, db.t3.medium)",
					PriceHash:       "b21c3c7708229fb149bff23b4cfe6833-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Storage",
					PriceHash:        "856b9e5bd87c953bffd1df698a6a1b3d-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
				{
					Name:             "I/O",
					PriceHash:        "c6f1f44a4f05ef22044c5af6490b6808-5be345988e7c9a0759c5cf8365868ee4",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
				{
					Name:             "CPU credits",
					PriceHash:        "f6d2bda62e25c6eb08020075859e5a97-e8e892be2fbd1c8f42fd6761ad8977d8",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestNewDocDBClusterInstance_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_docdb_cluster_instance" "medium" {
			cluster_identifier = "fake123"
			instance_class     = "db.t3.medium"
		}

		resource "aws_docdb_cluster_instance" "large" {
			cluster_identifier = "fake123"
			instance_class     = "db.r5.4xlarge"
		}`

	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_docdb_cluster_instance.medium": map[string]interface{}{
			"data_storage_gb":        1000,
			"monthly_io_request":     10000000,
			"monthly_cpu_credit_hrs": 10,
		},
		"aws_docdb_cluster_instance.large": map[string]interface{}{
			"data_storage_gb":    1000,
			"monthly_io_request": 10000000,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_docdb_cluster_instance.medium",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Database instance (on-demand, db.t3.medium)",
					PriceHash:       "b21c3c7708229fb149bff23b4cfe6833-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Storage",
					PriceHash:        "856b9e5bd87c953bffd1df698a6a1b3d-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(1000)),
				},
				{
					Name:             "I/O",
					PriceHash:        "c6f1f44a4f05ef22044c5af6490b6808-5be345988e7c9a0759c5cf8365868ee4",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(10000000)),
				},
				{
					Name:             "CPU credits",
					PriceHash:        "f6d2bda62e25c6eb08020075859e5a97-e8e892be2fbd1c8f42fd6761ad8977d8",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(10)),
				},
			},
		},
		{
			Name: "aws_docdb_cluster_instance.large",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Database instance (on-demand, db.r5.4xlarge)",
					PriceHash:       "5a03b2fda8befe3597ab04a003578d38-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Storage",
					PriceHash:        "856b9e5bd87c953bffd1df698a6a1b3d-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(1000)),
				},
				{
					Name:             "I/O",
					PriceHash:        "c6f1f44a4f05ef22044c5af6490b6808-5be345988e7c9a0759c5cf8365868ee4",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(10000000)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
