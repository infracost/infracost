package aws_test

import (
	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
	"testing"
)

func TestNewEFSFileSystemStandardStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_efs_file_system" "standard_storage" {}
	`
	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_efs_file_system.standard_storage": map[string]interface{}{
			"storage_gb": 230,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_efs_file_system.standard_storage",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Storage (standard)",
					PriceHash:        "032f12a61bc85ad77fd0e9fd5c1e6e3c-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(230)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}

func TestNewEFSFileSystemIAStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_efs_file_system" "ia_storage" {
			lifecycle_policy {
				transition_to_ia = "AFTER_7_DAYS"
			}
		}
	`
	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_efs_file_system.ia_storage": map[string]interface{}{
			"storage_gb":                    230,
			"infrequent_access_storage_gb":  100,
			"infrequent_access_requests_gb": 100,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_efs_file_system.ia_storage",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Storage (standard)",
					PriceHash:        "032f12a61bc85ad77fd0e9fd5c1e6e3c-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(230)),
				},
				{
					Name:             "Storage (infrequent access)",
					PriceHash:        "05a47659d1d78c2d60de3f9165b10a42-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100)),
				},
				{
					Name:             "Requests (infrequent access)",
					PriceHash:        "86d91c810b14b77cac328f9fcd26459b-b1ae3861dc57e2db217fa83a7420374f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}

func TestNewEFSFileSystemProvisionedThroughput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_efs_file_system" "provisioned" {
			provisioned_throughput_in_mibps = 100
			throughput_mode = "provisioned"
		}
	`
	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_efs_file_system.provisioned": map[string]interface{}{
			"storage_gb": 230,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_efs_file_system.provisioned",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Storage (standard)",
					PriceHash:        "032f12a61bc85ad77fd0e9fd5c1e6e3c-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(230)),
				},
				{
					Name:             "Provisioned throughput",
					PriceHash:        "41da10cb418495b3e0848200857f2d4d-8191dc82cee9b89717087e447a40abbd",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat((100*730 - (230.00*730)/20*1) / 730)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
