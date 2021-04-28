package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestInstance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_instance" "instance1" {
			ami           = "fake_ami"
			instance_type = "m3.medium"

			root_block_device {
				volume_size = 10
			}

			ebs_block_device {
				device_name = "xvdf"
				volume_size = 10
			}

			ebs_block_device {
				device_name = "xvdg"
				volume_type = "standard"
				volume_size = 20
			}

			ebs_block_device {
				device_name = "xvdh"
				volume_type = "sc1"
				volume_size = 30
			}

			ebs_block_device {
				device_name = "xvdi"
				volume_type = "io1"
				volume_size = 40
				iops        = 1000
			}

			ebs_block_device {
				device_name = "xvdj"
				volume_type = "gp3"
				volume_size = 20
			}
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_instance.instance1",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (Linux/UNIX, on-demand, m3.medium)",
					PriceHash:       "666e02bbe686f6950fd8a47a55e83a75-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name: "root_block_device",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:            "Storage (general purpose SSD, gp2)",
							PriceHash:       "efa8e70ebe004d2e9527fd30d50d09b2-ee3dd7e4624338037ca6fea0933a662f",
							HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(10)),
						},
					},
				},
				{
					Name: "ebs_block_device[0]",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:            "Storage (general purpose SSD, gp2)",
							PriceHash:       "efa8e70ebe004d2e9527fd30d50d09b2-ee3dd7e4624338037ca6fea0933a662f",
							HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(10)),
						},
					},
				},
				{
					Name: "ebs_block_device[1]",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:            "Storage (magnetic)",
							PriceHash:       "0ed17ed1777b7be91f5b5ce79916d8d8-ee3dd7e4624338037ca6fea0933a662f",
							HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(20)),
						},
						{
							Name:             "I/O requests",
							PriceHash:        "3085cb7cbdb1e1f570812e7400f8dbc6-5be345988e7c9a0759c5cf8365868ee4",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
					},
				},
				{
					Name: "ebs_block_device[2]",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:            "Storage (cold HDD, sc1)",
							PriceHash:       "3122df29367c2460c76537cccf0eadb5-ee3dd7e4624338037ca6fea0933a662f",
							HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(30)),
						},
					},
				},
				{
					Name: "ebs_block_device[3]",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:            "Storage (provisioned IOPS SSD, io1)",
							PriceHash:       "99450513de8c131ee2151e1b319d8143-ee3dd7e4624338037ca6fea0933a662f",
							HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(40)),
						},
						{
							Name:            "Provisioned IOPS",
							PriceHash:       "d5c5e1fb9b8ded55c336f6ae87aa2c3b-9c483347596633f8cf3ab7fdd5502b78",
							HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1000)),
						},
					},
				},
				{
					Name: "ebs_block_device[4]",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:            "Storage (general purpose SSD, gp3)",
							PriceHash:       "b7a83d535d47fcfd1be68ec37f046b3d-ee3dd7e4624338037ca6fea0933a662f",
							HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(20)),
						},
					},
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestInstance_ebsOptimized(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_instance" "instance1" {
			ami           = "fake_ami"
			instance_type = "m3.large"
			ebs_optimized = true
		}

		resource "aws_instance" "instance2" {
			ami           = "fake_ami"
			instance_type = "r3.xlarge"
			ebs_optimized = true
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_instance.instance1",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (Linux/UNIX, on-demand, m3.large)",
					PriceHash:       "1abac89a8296443758727a2728579a2a-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name:      "root_block_device",
					SkipCheck: true,
				},
			},
		},
		{
			Name: "aws_instance.instance2",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (Linux/UNIX, on-demand, r3.xlarge)",
					PriceHash:       "5fc0daede99fac3cce64d575979d7233-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:            "EBS-optimized usage",
					PriceHash:       "7f4fb9da921a628aedfbe150d930e255-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name:      "root_block_device",
					SkipCheck: true,
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestInstance_hostTenancy(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_instance" "instance1" {
			ami           = "fake_ami"
			instance_type = "m3.medium"
			tenancy       = "host"
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_instance.instance1",
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestInstance_cpuCredits(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_instance.t3_default": map[string]interface{}{
			"cpu_credit_hrs":    0,
			"virtual_cpu_count": 2,
		},
		"aws_instance.t3_unlimited": map[string]interface{}{
			"cpu_credit_hrs":    730,
			"virtual_cpu_count": 2,
		},
		"aws_instance.t2_unlimited": map[string]interface{}{
			"cpu_credit_hrs":    300,
			"virtual_cpu_count": 2,
		},
	})

	tf := `
		resource "aws_instance" "t3_default" {
			ami           = "fake_ami"
			instance_type = "t3.medium"
		}

		resource "aws_instance" "t3_unlimited" {
			ami           = "fake_ami"
			instance_type = "t3.medium"
			credit_specification {
				cpu_credits = "unlimited"
			}
		}

		resource "aws_instance" "t3_standard" {
			ami           = "fake_ami"
			instance_type = "t3.medium"
			credit_specification {
				cpu_credits = "standard"
			}
		}

		resource "aws_instance" "t2_default" {
			ami           = "fake_ami"
			instance_type = "t2.medium"
		}

		resource "aws_instance" "t2_unlimited" {
			ami           = "fake_ami"
			instance_type = "t2.medium"
			credit_specification {
				cpu_credits = "unlimited"
			}
		}

		resource "aws_instance" "t2_standard" {
			ami           = "fake_ami"
			instance_type = "t2.medium"
			credit_specification {
				cpu_credits = "standard"
			}
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_instance.t3_default",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:      "Instance usage (Linux/UNIX, on-demand, t3.medium)",
					SkipCheck: true,
				},
				{
					Name:             "CPU credits",
					PriceHash:        "ccdf11d8e4c0267d78a19b6663a566c1-e8e892be2fbd1c8f42fd6761ad8977d8",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.Zero),
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name:      "root_block_device",
					SkipCheck: true,
				},
			},
		},
		{
			Name: "aws_instance.t3_unlimited",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:      "Instance usage (Linux/UNIX, on-demand, t3.medium)",
					SkipCheck: true,
				},
				{
					Name:             "CPU credits",
					PriceHash:        "ccdf11d8e4c0267d78a19b6663a566c1-e8e892be2fbd1c8f42fd6761ad8977d8",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(730 * 2)),
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name:      "root_block_device",
					SkipCheck: true,
				},
			},
		},
		{
			Name: "aws_instance.t3_standard",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:      "Instance usage (Linux/UNIX, on-demand, t3.medium)",
					SkipCheck: true,
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name:      "root_block_device",
					SkipCheck: true,
				},
			},
		},
		{
			Name: "aws_instance.t2_default",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:      "Instance usage (Linux/UNIX, on-demand, t2.medium)",
					SkipCheck: true,
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name:      "root_block_device",
					SkipCheck: true,
				},
			},
		},
		{
			Name: "aws_instance.t2_unlimited",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:      "Instance usage (Linux/UNIX, on-demand, t2.medium)",
					SkipCheck: true,
				},
				{
					Name:             "CPU credits",
					PriceHash:        "4aaa3d22a88b57f7997e91888f867be9-e8e892be2fbd1c8f42fd6761ad8977d8",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(300 * 2)),
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name:      "root_block_device",
					SkipCheck: true,
				},
			},
		},
		{
			Name: "aws_instance.t2_standard",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:      "Instance usage (Linux/UNIX, on-demand, t2.medium)",
					SkipCheck: true,
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name:      "root_block_device",
					SkipCheck: true,
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
func TestInstance_ec2DetailedMonitoring(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_instance" "instance1" {
			ami           = "fake_ami"
			instance_type = "m3.large"
			ebs_optimized = true
			monitoring    = true
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_instance.instance1",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:      "Instance usage (Linux/UNIX, on-demand, m3.large)",
					SkipCheck: true,
				},
				{
					Name:            "EC2 detailed monitoring",
					PriceHash:       "df2e2141bd6d5e2b758fa0617157ff46-fd21869c4f4d79599eea951b2b7353e6",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(7)),
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name:      "root_block_device",
					SkipCheck: true,
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestInstance_RIPrices(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_instance" "std_1yr_no_upfront" {
			ami           = "fake_ami"
			instance_type = "t3.medium"
		}
		resource "aws_instance" "std_3yr_no_upfront" {
			ami           = "fake_ami"
			instance_type = "t3.medium"
		}
		resource "aws_instance" "std_1yr_partial_upfront" {
			ami           = "fake_ami"
			instance_type = "t3.medium"
		}
		resource "aws_instance" "std_3yr_partial_upfront" {
			ami           = "fake_ami"
			instance_type = "t3.medium"
		}
		resource "aws_instance" "std_1yr_all_upfront" {
			ami           = "fake_ami"
			instance_type = "t3.medium"
		}
		resource "aws_instance" "std_3yr_all_upfront" {
			ami           = "fake_ami"
			instance_type = "t3.medium"
		}

		resource "aws_instance" "cnvr_1yr_no_upfront" {
			ami           = "fake_ami"
			instance_type = "t3.medium"
		}
		resource "aws_instance" "cnvr_3yr_no_upfront" {
			ami           = "fake_ami"
			instance_type = "t3.medium"
		}
		resource "aws_instance" "cnvr_1yr_partial_upfront" {
			ami           = "fake_ami"
			instance_type = "t3.medium"
		}
		resource "aws_instance" "cnvr_3yr_partial_upfront" {
			ami           = "fake_ami"
			instance_type = "t3.medium"
		}
		resource "aws_instance" "cnvr_1yr_all_upfront" {
			ami           = "fake_ami"
			instance_type = "t3.medium"
		}
		resource "aws_instance" "cnvr_3yr_all_upfront" {
			ami           = "fake_ami"
			instance_type = "t3.medium"
		}`

	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_instance.std_1yr_no_upfront": map[string]interface{}{
			"reserved_instance_type":           "standard",
			"reserved_instance_term":           "1_year",
			"reserved_instance_payment_option": "no_upfront",
		},
		"aws_instance.std_3yr_no_upfront": map[string]interface{}{
			"reserved_instance_type":           "standard",
			"reserved_instance_term":           "3_year",
			"reserved_instance_payment_option": "no_upfront",
		},
		"aws_instance.std_1yr_partial_upfront": map[string]interface{}{
			"reserved_instance_type":           "standard",
			"reserved_instance_term":           "1_year",
			"reserved_instance_payment_option": "partial_upfront",
		},
		"aws_instance.std_3yr_partial_upfront": map[string]interface{}{
			"reserved_instance_type":           "standard",
			"reserved_instance_term":           "3_year",
			"reserved_instance_payment_option": "partial_upfront",
		},
		"aws_instance.std_1yr_all_upfront": map[string]interface{}{
			"reserved_instance_type":           "standard",
			"reserved_instance_term":           "1_year",
			"reserved_instance_payment_option": "all_upfront",
		},
		"aws_instance.std_3yr_all_upfront": map[string]interface{}{
			"reserved_instance_type":           "standard",
			"reserved_instance_term":           "3_year",
			"reserved_instance_payment_option": "all_upfront",
		},
		"aws_instance.cnvr_1yr_no_upfront": map[string]interface{}{
			"reserved_instance_type":           "convertible",
			"reserved_instance_term":           "1_year",
			"reserved_instance_payment_option": "no_upfront",
		},
		"aws_instance.cnvr_3yr_no_upfront": map[string]interface{}{
			"reserved_instance_type":           "convertible",
			"reserved_instance_term":           "3_year",
			"reserved_instance_payment_option": "no_upfront",
		},
		"aws_instance.cnvr_1yr_partial_upfront": map[string]interface{}{
			"reserved_instance_type":           "convertible",
			"reserved_instance_term":           "1_year",
			"reserved_instance_payment_option": "partial_upfront",
		},
		"aws_instance.cnvr_3yr_partial_upfront": map[string]interface{}{
			"reserved_instance_type":           "convertible",
			"reserved_instance_term":           "3_year",
			"reserved_instance_payment_option": "partial_upfront",
		},
		"aws_instance.cnvr_1yr_all_upfront": map[string]interface{}{
			"reserved_instance_type":           "convertible",
			"reserved_instance_term":           "1_year",
			"reserved_instance_payment_option": "all_upfront",
		},
		"aws_instance.cnvr_3yr_all_upfront": map[string]interface{}{
			"reserved_instance_type":           "convertible",
			"reserved_instance_term":           "3_year",
			"reserved_instance_payment_option": "all_upfront",
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_instance.std_1yr_no_upfront",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (Linux/UNIX, reserved, t3.medium)",
					PriceHash:       "c8faba8210cd512ccab6b71ca400f4de-354de5028123250997d97c05d011fe1c",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:      "CPU credits",
					SkipCheck: true,
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name:      "root_block_device",
					SkipCheck: true,
				},
			},
		},
		{
			Name: "aws_instance.std_3yr_no_upfront",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (Linux/UNIX, reserved, t3.medium)",
					PriceHash:       "c8faba8210cd512ccab6b71ca400f4de-eacbcf31b049c055c292e5f56fbe6f38",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:      "CPU credits",
					SkipCheck: true,
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name:      "root_block_device",
					SkipCheck: true,
				},
			},
		},
		{
			Name: "aws_instance.std_1yr_partial_upfront",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (Linux/UNIX, reserved, t3.medium)",
					PriceHash:       "c8faba8210cd512ccab6b71ca400f4de-1b51d9b46826b8797099f7cfdfcdf299",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:      "CPU credits",
					SkipCheck: true,
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name:      "root_block_device",
					SkipCheck: true,
				},
			},
		},
		{
			Name: "aws_instance.std_3yr_partial_upfront",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (Linux/UNIX, reserved, t3.medium)",
					PriceHash:       "c8faba8210cd512ccab6b71ca400f4de-bf59b46c8f98c6a49405f768bfa8b60a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:      "CPU credits",
					SkipCheck: true,
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name:      "root_block_device",
					SkipCheck: true,
				},
			},
		},
		{
			Name: "aws_instance.std_1yr_all_upfront",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (Linux/UNIX, reserved, t3.medium)",
					PriceHash:       "c8faba8210cd512ccab6b71ca400f4de-0b517bfa356310108e91658d6759b4d5",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:      "CPU credits",
					SkipCheck: true,
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name:      "root_block_device",
					SkipCheck: true,
				},
			},
		},
		{
			Name: "aws_instance.std_3yr_all_upfront",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (Linux/UNIX, reserved, t3.medium)",
					PriceHash:       "c8faba8210cd512ccab6b71ca400f4de-4c69aedc693029aad69299aaef81901a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:      "CPU credits",
					SkipCheck: true,
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name:      "root_block_device",
					SkipCheck: true,
				},
			},
		},
		{
			Name: "aws_instance.cnvr_1yr_no_upfront",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (Linux/UNIX, reserved, t3.medium)",
					PriceHash:       "c8faba8210cd512ccab6b71ca400f4de-db82ffe7b4996cd80e13db57284de443",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:      "CPU credits",
					SkipCheck: true,
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name:      "root_block_device",
					SkipCheck: true,
				},
			},
		},
		{
			Name: "aws_instance.cnvr_3yr_no_upfront",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (Linux/UNIX, reserved, t3.medium)",
					PriceHash:       "c8faba8210cd512ccab6b71ca400f4de-9252443b383d5d512783a5b68e9a901f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:      "CPU credits",
					SkipCheck: true,
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name:      "root_block_device",
					SkipCheck: true,
				},
			},
		},
		{
			Name: "aws_instance.cnvr_1yr_partial_upfront",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (Linux/UNIX, reserved, t3.medium)",
					PriceHash:       "c8faba8210cd512ccab6b71ca400f4de-39830253d678995796c122c70b428e1b",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:      "CPU credits",
					SkipCheck: true,
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name:      "root_block_device",
					SkipCheck: true,
				},
			},
		},
		{
			Name: "aws_instance.cnvr_3yr_partial_upfront",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (Linux/UNIX, reserved, t3.medium)",
					PriceHash:       "c8faba8210cd512ccab6b71ca400f4de-99d3f32d59d2381ad0a77075299b58e6",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:      "CPU credits",
					SkipCheck: true,
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name:      "root_block_device",
					SkipCheck: true,
				},
			},
		},
		{
			Name: "aws_instance.cnvr_1yr_all_upfront",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (Linux/UNIX, reserved, t3.medium)",
					PriceHash:       "c8faba8210cd512ccab6b71ca400f4de-a0fcfdebef129d176f42bb38df23dad9",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:      "CPU credits",
					SkipCheck: true,
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name:      "root_block_device",
					SkipCheck: true,
				},
			},
		},
		{
			Name: "aws_instance.cnvr_3yr_all_upfront",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (Linux/UNIX, reserved, t3.medium)",
					PriceHash:       "c8faba8210cd512ccab6b71ca400f4de-4a6714faa3f991aee5551b21d785b47c",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:      "CPU credits",
					SkipCheck: true,
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name:      "root_block_device",
					SkipCheck: true,
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
