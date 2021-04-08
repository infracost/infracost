package google_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestContainerRegistry(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "google_container_registry" "my_registry" {
		project  = "my-project"
		location = "EU"
	  }
	`

	usage := schema.NewUsageMap(map[string]interface{}{
		"google_container_registry.my_registry": map[string]interface{}{
			"storage_gb":                 150,
			"monthly_class_a_operations": 40000,
			"monthly_class_b_operations": 20000,
			"monthly_data_retrieval_gb":  500,
			"monthly_egress_data_transfer_gb": map[string]interface{}{
				"same_continent": 550,
				"worldwide":      12500,
				"asia":           1500,
				"china":          50,
				"australia":      250,
			},
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_container_registry.my_registry",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Storage (standard)",
					PriceHash:       "fc9e1d9f7ff70a2a143b33dd97962bc6-57bc5d148491a8381abaccb21ca6b4e9",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(150)),
				},
				{
					Name:            "Object adds, bucket/object list (class A)",
					PriceHash:       "90f2b2a3b4c22f37d28a13253ae8cd24-a478a7b16172e9b2ff0541b2fc37320e",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(40000)),
				},
				{
					Name:            "Object gets, retrieve bucket/object metadata (class B)",
					PriceHash:       "cab7649d29acc0f74a8d7fc6ed7d420c-8b1be436ac01a15b378be3c4d0590e97",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(20000)),
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name: "Network egress",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:            "Data transfer in same continent",
							PriceHash:       "99caf41700f8e761f8ab246b426edbf2-8012a4febcd0213911ed09e53341a976",
							HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(550)),
						},
						{
							Name:            "Data transfer to worldwide excluding Asia, Australia (first 1TB)",
							PriceHash:       "fa69ceb2a41a4b9cda9222f96d0e32f1-0c23081f8c5fa7d720ec507ecfd47cf6",
							HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1024)),
						},
						{
							Name:            "Data transfer to worldwide excluding Asia, Australia (next 9TB)",
							PriceHash:       "fa69ceb2a41a4b9cda9222f96d0e32f1-1b0e0067a261ee1db4b1b62b351927dc",
							HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(9216)),
						},
						{
							Name:            "Data transfer to worldwide excluding Asia, Australia (over 10TB)",
							PriceHash:       "fa69ceb2a41a4b9cda9222f96d0e32f1-4d6929fe300ded2d5807f08cac9b0ca0",
							HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(2260)),
						},
						{
							Name:            "Data transfer to Asia excluding China, but including Hong Kong (first 1TB)",
							PriceHash:       "d63ba0daedaf0de514cdd32537310c00-0c23081f8c5fa7d720ec507ecfd47cf6",
							HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1024)),
						},
						{
							Name:            "Data transfer to Asia excluding China, but including Hong Kong (next 9TB)",
							PriceHash:       "d63ba0daedaf0de514cdd32537310c00-1b0e0067a261ee1db4b1b62b351927dc",
							HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(476)),
						},
						{
							Name:            "Data transfer to China excluding Hong Kong (first 1TB)",
							PriceHash:       "237057d62af52bee885b9f353bab90e2-a62ab44470fc752864d0f5c5534f3d33",
							HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(50)),
						},
						{
							Name:            "Data transfer to Australia (first 1TB)",
							PriceHash:       "a3e569b71cd1e9d2294629e1b995c1f6-a62ab44470fc752864d0f5c5534f3d33",
							HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(250)),
						},
					},
				},
			},
		},
	}
	tftest.ResourceTests(t, tf, usage, resourceChecks)
}

func TestContainerRegistry_EuMulti(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "google_container_registry" "my_registry" {
		project  = "my-project"
		location = "EU"
	  }
	`
	usage := schema.NewUsageMap(map[string]interface{}{
		"google_container_registry.my_registry": map[string]interface{}{
			"storage_gb":                 150,
			"monthly_class_a_operations": 40000,
			"monthly_class_b_operations": 20000,
			"monthly_data_retrieval_gb":  500,
			"monthly_egress_data_transfer_gb": map[string]interface{}{
				"same_continent": 550,
				"worldwide":      12500,
				"asia":           1500,
				"china":          50,
				"australia":      250,
			},
		},
	})
	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_container_registry.my_registry",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Storage (standard)",
					PriceHash:       "fc9e1d9f7ff70a2a143b33dd97962bc6-57bc5d148491a8381abaccb21ca6b4e9",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(150)),
				},
				{
					Name:      "Object adds, bucket/object list (class A)",
					SkipCheck: true,
				},
				{
					Name:      "Object gets, retrieve bucket/object metadata (class B)",
					SkipCheck: true,
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name:      "Network egress",
					SkipCheck: true,
				},
			},
		},
	}
	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
