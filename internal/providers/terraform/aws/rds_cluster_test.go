package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestRDSClusterGoldenFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tftest.GoldenFileResourceTests(t, "rds_cluster_test")
}

func TestRDSAuroraServerlessCluster(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_rds_cluster" "postgres" {
			cluster_identifier = "aurora-serverless"
			engine = "aurora-postgresql"
			engine_mode = "serverless"
			master_username    = "foo"
			master_password    = "barbut8chars"
		}

		resource "aws_rds_cluster" "my_sql" {
			cluster_identifier = "aurora-serverless"
			engine = "aurora-mysql"
			engine_mode = "serverless"
			master_username    = "foo"
			master_password    = "barbut8chars"
		}
`

	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_rds_cluster.postgres": map[string]interface{}{
			"capacity_units_per_hr":  1000,
			"storage_gb":             100,
			"write_requests_per_sec": 10,
			"read_requests_per_sec":  10,
		},
		"aws_rds_cluster.my_sql": map[string]interface{}{
			"capacity_units_per_hr":  1000,
			"storage_gb":             100,
			"write_requests_per_sec": 10,
			"read_requests_per_sec":  10,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_rds_cluster.postgres",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Aurora serverless",
					PriceHash:       "2f7af5f2d0e381575c55d86307fac661-3f912465de6efe7a7216c8be67acddd5",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1000)),
				},
				{
					Name:             "Storage rate",
					PriceHash:        "9248555ff8796324d893f7bbd87b7277-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(100)),
				},
				{
					Name:             "I/O rate",
					PriceHash:        "51e320a6840e4018a0da4a84663eb222-5be345988e7c9a0759c5cf8365868ee4",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(20).Mul(decimal.NewFromInt(730)).Mul(decimal.NewFromInt(60)).Mul(decimal.NewFromInt(60))),
				},
			},
		},
		{
			Name: "aws_rds_cluster.my_sql",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Aurora serverless",
					PriceHash:       "d05dcffc9dd2984d2f1b2efe4d814a01-3f912465de6efe7a7216c8be67acddd5",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1000)),
				},
				{
					Name:             "Storage rate",
					PriceHash:        "2719bd917d46a9a24d7f4415b92a16f9-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(100)),
				},
				{
					Name:             "I/O rate",
					PriceHash:        "87cf7c84d9f0be12d710cb03cb93ba90-5be345988e7c9a0759c5cf8365868ee4",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(20).Mul(decimal.NewFromInt(730)).Mul(decimal.NewFromInt(60)).Mul(decimal.NewFromInt(60))),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}

func TestRDSAuroraServerlessClusterWithBackup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_rds_cluster" "postgres" {
			cluster_identifier = "aurora-serverless"
			engine = "aurora-postgresql"
			engine_mode = "serverless"
			backup_retention_period = 5
			master_username    = "foo"
			master_password    = "barbut8chars"
		}
`

	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_rds_cluster.postgres": map[string]interface{}{
			"capacity_units_per_hr":   1000,
			"storage_gb":              100,
			"write_requests_per_sec":  10,
			"read_requests_per_sec":   10,
			"backup_snapshot_size_gb": 100,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_rds_cluster.postgres",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Aurora serverless",
					PriceHash:       "2f7af5f2d0e381575c55d86307fac661-3f912465de6efe7a7216c8be67acddd5",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1000)),
				},
				{
					Name:             "Storage rate",
					PriceHash:        "9248555ff8796324d893f7bbd87b7277-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(100)),
				},
				{
					Name:             "I/O rate",
					PriceHash:        "51e320a6840e4018a0da4a84663eb222-5be345988e7c9a0759c5cf8365868ee4",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(20).Mul(decimal.NewFromInt(730)).Mul(decimal.NewFromInt(60)).Mul(decimal.NewFromInt(60))),
				},
				{
					Name:             "Backup storage",
					PriceHash:        "861c69e8602bd946dabae7ca550851be-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(100.00).Mul(decimal.NewFromInt(5)).Sub(decimal.NewFromFloat(100.00))),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}

func TestRDSAuroraServerlessClusterWithExport(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_rds_cluster" "postgres" {
			cluster_identifier = "aurora-serverless"
			engine = "aurora-postgresql"
			engine_mode = "serverless"
			backup_retention_period = 5
			master_username    = "foo"
			master_password    = "barbut8chars"
		}
`

	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_rds_cluster.postgres": map[string]interface{}{
			"capacity_units_per_hr":        1000,
			"storage_gb":                   100,
			"write_requests_per_sec":       10,
			"read_requests_per_sec":        10,
			"backup_snapshot_size_gb":      100,
			"average_statements_per_hr":    10000000,
			"change_records_per_statement": 0.38,
			"backtrack_window_hrs":         24,
			"snapshot_export_size_gb":      200,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_rds_cluster.postgres",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Aurora serverless",
					PriceHash:       "2f7af5f2d0e381575c55d86307fac661-3f912465de6efe7a7216c8be67acddd5",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1000)),
				},
				{
					Name:             "Storage rate",
					PriceHash:        "9248555ff8796324d893f7bbd87b7277-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(100)),
				},
				{
					Name:             "I/O rate",
					PriceHash:        "51e320a6840e4018a0da4a84663eb222-5be345988e7c9a0759c5cf8365868ee4",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(20).Mul(decimal.NewFromInt(730)).Mul(decimal.NewFromInt(60)).Mul(decimal.NewFromInt(60))),
				},
				{
					Name:             "Backup storage",
					PriceHash:        "861c69e8602bd946dabae7ca550851be-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(100.00).Mul(decimal.NewFromInt(5)).Sub(decimal.NewFromFloat(100.00))),
				},
				{
					Name:             "Snapshot export",
					PriceHash:        "87bb7827d9684b87e0a421f624e7142d-b1ae3861dc57e2db217fa83a7420374f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(200)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}

func TestRDSAuroraClusterBacktrack(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_rds_cluster" "mysql" {
			cluster_identifier = "aurora-mysql"
			engine = "aurora-mysql"
			backup_retention_period = 5
			master_username    = "foo"
			master_password    = "barbut8chars"
		}
`

	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_rds_cluster.mysql": map[string]interface{}{
			"storage_gb":                   100,
			"write_requests_per_sec":       10,
			"read_requests_per_sec":        10,
			"backup_snapshot_size_gb":      100,
			"average_statements_per_hr":    10000000,
			"change_records_per_statement": 0.38,
			"backtrack_window_hrs":         24,
			"snapshot_export_size_gb":      200,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_rds_cluster.mysql",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Storage rate",
					PriceHash:        "2719bd917d46a9a24d7f4415b92a16f9-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(100)),
				},
				{
					Name:             "I/O rate",
					PriceHash:        "87cf7c84d9f0be12d710cb03cb93ba90-5be345988e7c9a0759c5cf8365868ee4",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(20).Mul(decimal.NewFromInt(730)).Mul(decimal.NewFromInt(60)).Mul(decimal.NewFromInt(60))),
				},
				{
					Name:             "Backup storage",
					PriceHash:        "0fb26ac3fb816a8cd4c13a64fc3166fe-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(100.00).Mul(decimal.NewFromInt(5)).Sub(decimal.NewFromFloat(100.00))),
				},
				{
					Name:             "Backtrack",
					PriceHash:        "0dd6d3925e3e60292490015f5e904d76-5617335f1d1e54cf41c65ca230dc1725",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(10000000).Mul(decimal.NewFromInt(730)).Mul(decimal.NewFromFloat(0.38).Mul(decimal.NewFromInt(24)))),
				},
				{
					Name:             "Snapshot export",
					PriceHash:        "4ff9640fb2c41ced13784dc45deb035b-b1ae3861dc57e2db217fa83a7420374f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(200)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
