package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestEBSVolume(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_ebs_volume" "gp2" {
			availability_zone = "us-east-1a"
			size              = 10
		}

		resource "aws_ebs_volume" "standard" {
			availability_zone = "us-east-1a"
			size              = 20
			type              = "standard"
		}

		resource "aws_ebs_volume" "io1" {
			availability_zone = "us-east-1a"
			type              = "io1"
			size              = 30
			iops              = 300
		}

		resource "aws_ebs_volume" "io2" {
			availability_zone = "us-east-1a"
			type              = "io2"
			size              = 30
			iops              = 300
		}

		resource "aws_ebs_volume" "st1" {
			availability_zone = "us-east-1a"
			size              = 40
			type              = "st1"
		}

		resource "aws_ebs_volume" "sc1" {
			availability_zone = "us-east-1a"
			size              = 50
			type              = "sc1"
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_ebs_volume.gp2",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Storage (general purpose SSD, gp2)",
					PriceHash:        "efa8e70ebe004d2e9527fd30d50d09b2-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(10)),
				},
			},
		},
		{
			Name: "aws_ebs_volume.standard",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Storage (magnetic)",
					PriceHash:        "0ed17ed1777b7be91f5b5ce79916d8d8-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(20)),
				},
				{
					Name:             "I/O requests",
					PriceHash:        "3085cb7cbdb1e1f570812e7400f8dbc6-5be345988e7c9a0759c5cf8365868ee4",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
		{
			Name: "aws_ebs_volume.io1",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Storage (provisioned IOPS SSD, io1)",
					PriceHash:        "99450513de8c131ee2151e1b319d8143-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(30)),
				},
				{
					Name:             "Provisioned IOPS",
					PriceHash:        "d5c5e1fb9b8ded55c336f6ae87aa2c3b-9c483347596633f8cf3ab7fdd5502b78",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(300)),
				},
			},
		},
		{
			Name: "aws_ebs_volume.io2",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Storage (provisioned IOPS SSD, io2)",
					PriceHash:        "9e420d5e498eddb54a405d09b89a668e-c86ea75c5a17b237464f7c8cc81c1ab8",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(30)),
				},
				{
					Name:             "Provisioned IOPS",
					PriceHash:        "9cff4839500aaabec26f2bace787491b-9c483347596633f8cf3ab7fdd5502b78",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(300)),
				},
			},
		},
		{
			Name: "aws_ebs_volume.st1",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Storage (throughput optimized HDD, st1)",
					PriceHash:        "eea972b50a795c92487cbcb96e8fdc29-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(40)),
				},
			},
		},
		{
			Name: "aws_ebs_volume.sc1",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Storage (cold HDD, sc1)",
					PriceHash:        "3122df29367c2460c76537cccf0eadb5-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(50)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestEBSVolume_GP3(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "aws_ebs_volume" "gp3" {
		availability_zone = "us-west-2a"
		size              = 40
		type = "gp3"
		iops = 4000
		throughput = 130
		
		tags = {
			Name = "HelloWorld"
		}
	}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_ebs_volume.gp3",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Storage (general purpose SSD, gp3)",
					PriceHash:        "b7a83d535d47fcfd1be68ec37f046b3d-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(40)),
				},
				{
					Name:             "Provisioned throughput",
					PriceHash:        "9775542e08e79c6f6dff9b6537f84c04-8191dc82cee9b89717087e447a40abbd",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(5)),
				},
				{
					Name:             "Provisioned IOPS",
					PriceHash:        "4299307b615dc9d0e4e8aa362c6b4c6f-9c483347596633f8cf3ab7fdd5502b78",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1000)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestEBSStandardVolume_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "aws_ebs_volume" "standard" {
		availability_zone = "us-east-1a"
		size              = 20
		type              = "standard"
	}`

	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_ebs_volume.standard": map[string]interface{}{
			"monthly_standard_io_requests": 1000000,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_ebs_volume.standard",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Storage (magnetic)",
					PriceHash:        "0ed17ed1777b7be91f5b5ce79916d8d8-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(20)),
				},
				{
					Name:             "I/O requests",
					PriceHash:        "3085cb7cbdb1e1f570812e7400f8dbc6-5be345988e7c9a0759c5cf8365868ee4",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(1000000)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
