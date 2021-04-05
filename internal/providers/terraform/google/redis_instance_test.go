package google_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestRedisInstance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "google_redis_instance" "basic_m1" {
			name           = "memory-cache"
			memory_size_gb = 1
		}
		resource "google_redis_instance" "basic_m2" {
			name           = "memory-cache"
			memory_size_gb = 5
		}
		resource "google_redis_instance" "basic_m3" {
			name           = "memory-cache"
			memory_size_gb = 25
		}
		resource "google_redis_instance" "basic_m4" {
			name           = "memory-cache"
			memory_size_gb = 45
		}
		resource "google_redis_instance" "basic_m5" {
			name           = "memory-cache"
			memory_size_gb = 105
		}
		resource "google_redis_instance" "standard_m1" {
			name           = "memory-cache"
			memory_size_gb = 1
			tier           = "STANDARD_HA"
		}
		resource "google_redis_instance" "standard_m2" {
			name           = "memory-cache"
			memory_size_gb = 5
			tier           = "STANDARD_HA"
		}
		resource "google_redis_instance" "standard_m3" {
			name           = "memory-cache"
			memory_size_gb = 25
			tier           = "STANDARD_HA"
		}
		resource "google_redis_instance" "standard_m4" {
			name           = "memory-cache"
			memory_size_gb = 45
			tier           = "STANDARD_HA"
		}
		resource "google_redis_instance" "standard_m5" {
			name           = "memory-cache"
			memory_size_gb = 105
			tier           = "STANDARD_HA"
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_redis_instance.basic_m1",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Redis instance (basic, M1)",
					PriceHash:        "5c7ff4d6f6712e1e460103e0cb9467d1-e400b4debea1ba77ad9bec422eeaf576",
					MonthlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
		{
			Name: "google_redis_instance.basic_m2",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Redis instance (basic, M2)",
					PriceHash:        "bed41390d56b4d59fe5cbc2b6051371a-e400b4debea1ba77ad9bec422eeaf576",
					MonthlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(5)),
				},
			},
		},
		{
			Name: "google_redis_instance.basic_m3",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Redis instance (basic, M3)",
					PriceHash:        "28740eefcf78c4ce991763088cf3de93-e400b4debea1ba77ad9bec422eeaf576",
					MonthlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(25)),
				},
			},
		},
		{
			Name: "google_redis_instance.basic_m4",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Redis instance (basic, M4)",
					PriceHash:        "3e0d21d45d10db22a6009467b753c4df-e400b4debea1ba77ad9bec422eeaf576",
					MonthlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(45)),
				},
			},
		},
		{
			Name: "google_redis_instance.basic_m5",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Redis instance (basic, M5)",
					PriceHash:        "fda7bbba7cc4bd925b2a06d2e300e418-e400b4debea1ba77ad9bec422eeaf576",
					MonthlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(105)),
				},
			},
		},
		{
			Name: "google_redis_instance.standard_m1",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Redis instance (standard, M1)",
					PriceHash:        "3c7262dff6304711007dcad84ac1b239-e400b4debea1ba77ad9bec422eeaf576",
					MonthlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
		{
			Name: "google_redis_instance.standard_m2",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Redis instance (standard, M2)",
					PriceHash:        "cd37c015d896f8dafa4e8646fe757e14-e400b4debea1ba77ad9bec422eeaf576",
					MonthlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(5)),
				},
			},
		},
		{
			Name: "google_redis_instance.standard_m3",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Redis instance (standard, M3)",
					PriceHash:        "1869a4d41adeb44fe54ba09c8876da85-e400b4debea1ba77ad9bec422eeaf576",
					MonthlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(25)),
				},
			},
		},
		{
			Name: "google_redis_instance.standard_m4",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Redis instance (standard, M4)",
					PriceHash:        "d3585ce32c14a89ea37b61e29142f180-e400b4debea1ba77ad9bec422eeaf576",
					MonthlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(45)),
				},
			},
		},
		{
			Name: "google_redis_instance.standard_m5",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Redis instance (standard, M5)",
					PriceHash:        "74db16786f927d999730e152f6320d7d-e400b4debea1ba77ad9bec422eeaf576",
					MonthlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(105)),
				},
			},
		},
	}
	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
