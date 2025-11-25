package aws_test

import (
	"testing"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestDataTransferGoldenFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	opts := tftest.DefaultGoldenFileOptions()
	opts.CaptureLogs = true
	tftest.GoldenFileResourceTestsWithOpts(t, "data_transfer_test", opts)
}

func TestChinaDataTransfer(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := ``

	usage := schema.NewUsageMapFromInterface(map[string]any{
		"aws_data_transfer.cn-north-1": map[string]any{
			"region":                            "cn-north-1",
			"monthly_intra_region_gb":           10,
			"monthly_outbound_other_regions_gb": 20,
			"monthly_outbound_internet_gb":      30,
		},
		"aws_data_transfer.cn-northwest-1": map[string]any{
			"region":                            "cn-northwest-1",
			"monthly_intra_region_gb":           10,
			"monthly_outbound_other_regions_gb": 20,
			"monthly_outbound_internet_gb":      30,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_data_transfer.cn-north-1",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Intra-region data transfer",
					PriceHash:       "db9b51870d0a4f81bf8126e9bde3565d-b1ae3861dc57e2db217fa83a7420374f",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(20)),
				},
				{
					Name:             "Outbound data transfer to Internet",
					PriceHash:        "2cc1913e1158dd7df40e5fe4de21eeaa-b1ae3861dc57e2db217fa83a7420374f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(30)),
				},
				{
					Name:             "Outbound data transfer to other regions",
					PriceHash:        "3700bb72dc52aa255c23d186418a8ee5-b1ae3861dc57e2db217fa83a7420374f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(20)),
				},
			},
		},
		{
			Name: "aws_data_transfer.cn-northwest-1",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Intra-region data transfer",
					PriceHash:       "2d407f24b5e7092c3d678616ac4fe7af-b1ae3861dc57e2db217fa83a7420374f",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(20)),
				},
				{
					Name:             "Outbound data transfer to Internet",
					PriceHash:        "284a7eac6795c354fccbb1aad5b06704-b1ae3861dc57e2db217fa83a7420374f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(30)),
				},
				{
					Name:             "Outbound data transfer to other regions",
					PriceHash:        "c1bbbe9eb53baf0f1ca39c03ffcf5c80-b1ae3861dc57e2db217fa83a7420374f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(20)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
