package aws_test

import (
	"testing"

	"github.com/infracost/infracost/pkg/testutil"

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
					Name:            "Storage",
					PriceHash:       "efa8e70ebe004d2e9527fd30d50d09b2-ee3dd7e4624338037ca6fea0933a662f",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(10)),
				},
			},
		},
		{
			Name: "aws_ebs_volume.standard",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Storage",
					PriceHash:       "0ed17ed1777b7be91f5b5ce79916d8d8-ee3dd7e4624338037ca6fea0933a662f",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(20)),
				},
			},
		},
		{
			Name: "aws_ebs_volume.io1",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Storage",
					PriceHash:       "99450513de8c131ee2151e1b319d8143-ee3dd7e4624338037ca6fea0933a662f",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(30)),
				},
				{
					Name:            "Storage IOPS",
					PriceHash:       "d5c5e1fb9b8ded55c336f6ae87aa2c3b-9c483347596633f8cf3ab7fdd5502b78",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(300)),
				},
			},
		},
		{
			Name: "aws_ebs_volume.st1",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Storage",
					PriceHash:       "eea972b50a795c92487cbcb96e8fdc29-ee3dd7e4624338037ca6fea0933a662f",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(40)),
				},
			},
		},
		{
			Name: "aws_ebs_volume.sc1",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Storage",
					PriceHash:       "3122df29367c2460c76537cccf0eadb5-ee3dd7e4624338037ca6fea0933a662f",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(50)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, resourceChecks)
}
