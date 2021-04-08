package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestCloudwatchLogGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
        resource "aws_cloudwatch_log_group" "logs" {
          name              = "log-group"
        }`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_cloudwatch_log_group.logs",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Data ingested",
					PriceHash:        "4c00b8e26729863d2cc1f1a2d824dcf0-b1ae3861dc57e2db217fa83a7420374f",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
				{
					Name:             "Archival Storage",
					PriceHash:        "af1a1c7a3c3f5fc6e72de0ba26dcf55e-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
				{
					Name:             "Insights queries data scanned",
					PriceHash:        "e4d44a4a02daffd13cd87e63d67f30a5-b1ae3861dc57e2db217fa83a7420374f",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestCloudwatchLogGroup_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "aws_cloudwatch_log_group" "my_log_group" {
		name  = "log-group"
	}
	resource "aws_cloudwatch_log_group" "logs" {
		count = 3
		name  = "log-group${count.index}"
	}`

	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_cloudwatch_log_group.my_log_group": map[string]interface{}{
			"monthly_data_ingested_gb": 1000,
			"storage_gb":               500,
			"monthly_data_scanned_gb":  250,
		},
		"aws_cloudwatch_log_group.logs[*]": map[string]interface{}{
			"monthly_data_ingested_gb": 1000,
			"storage_gb":               500,
			"monthly_data_scanned_gb":  250,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_cloudwatch_log_group.my_log_group",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Data ingested",
					PriceHash:        "4c00b8e26729863d2cc1f1a2d824dcf0-b1ae3861dc57e2db217fa83a7420374f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(1000)),
				},
				{
					Name:             "Archival Storage",
					PriceHash:        "af1a1c7a3c3f5fc6e72de0ba26dcf55e-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(500)),
				},
				{
					Name:             "Insights queries data scanned",
					PriceHash:        "e4d44a4a02daffd13cd87e63d67f30a5-b1ae3861dc57e2db217fa83a7420374f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(250)),
				},
			},
		},
		{
			Name: "aws_cloudwatch_log_group.logs[0]",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Data ingested",
					PriceHash:        "4c00b8e26729863d2cc1f1a2d824dcf0-b1ae3861dc57e2db217fa83a7420374f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(1000)),
				},
				{
					Name:             "Archival Storage",
					PriceHash:        "af1a1c7a3c3f5fc6e72de0ba26dcf55e-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(500)),
				},
				{
					Name:             "Insights queries data scanned",
					PriceHash:        "e4d44a4a02daffd13cd87e63d67f30a5-b1ae3861dc57e2db217fa83a7420374f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(250)),
				},
			},
		},
		{
			Name: "aws_cloudwatch_log_group.logs[1]",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Data ingested",
					PriceHash:        "4c00b8e26729863d2cc1f1a2d824dcf0-b1ae3861dc57e2db217fa83a7420374f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(1000)),
				},
				{
					Name:             "Archival Storage",
					PriceHash:        "af1a1c7a3c3f5fc6e72de0ba26dcf55e-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(500)),
				},
				{
					Name:             "Insights queries data scanned",
					PriceHash:        "e4d44a4a02daffd13cd87e63d67f30a5-b1ae3861dc57e2db217fa83a7420374f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(250)),
				},
			},
		},
		{
			Name: "aws_cloudwatch_log_group.logs[2]",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Data ingested",
					PriceHash:        "4c00b8e26729863d2cc1f1a2d824dcf0-b1ae3861dc57e2db217fa83a7420374f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(1000)),
				},
				{
					Name:             "Archival Storage",
					PriceHash:        "af1a1c7a3c3f5fc6e72de0ba26dcf55e-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(500)),
				},
				{
					Name:             "Insights queries data scanned",
					PriceHash:        "e4d44a4a02daffd13cd87e63d67f30a5-b1ae3861dc57e2db217fa83a7420374f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(250)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
