package aws_test

import (
	"infracost/internal/testutil"
	"infracost/pkg/costs"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/shopspring/decimal"
)

func TestEc2InstanceIntegration(t *testing.T) {
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
}`

	resourceCostBreakdowns, err := testutil.RunTFCostBreakdown(tf)
	if err != nil {
		t.Error(err)
	}

	expectedPriceHashes := [][]string{
		{"aws_instance.instance1", "instance hours (m3.medium)", "666e02bbe686f6950fd8a47a55e83a75-d2c98780d7b6e36641b521f1f8145c6f"},
		{"aws_instance.instance1.root_block_device", "GB", "efa8e70ebe004d2e9527fd30d50d09b2-ee3dd7e4624338037ca6fea0933a662f"},
		{"aws_instance.instance1.ebs_block_device[0]", "GB", "efa8e70ebe004d2e9527fd30d50d09b2-ee3dd7e4624338037ca6fea0933a662f"},
		{"aws_instance.instance1.ebs_block_device[1]", "GB", "0ed17ed1777b7be91f5b5ce79916d8d8-ee3dd7e4624338037ca6fea0933a662f"},
		{"aws_instance.instance1.ebs_block_device[2]", "GB", "3122df29367c2460c76537cccf0eadb5-ee3dd7e4624338037ca6fea0933a662f"},
		{"aws_instance.instance1.ebs_block_device[3]", "GB", "99450513de8c131ee2151e1b319d8143-ee3dd7e4624338037ca6fea0933a662f"},
		{"aws_instance.instance1.ebs_block_device[3]", "IOPS", "d5c5e1fb9b8ded55c336f6ae87aa2c3b-9c483347596633f8cf3ab7fdd5502b78"},
	}

	priceHashResults := testutil.ExtractPriceHashes(resourceCostBreakdowns)

	if !cmp.Equal(priceHashResults, expectedPriceHashes, testutil.PriceHashResultSort) {
		t.Error("got unexpected price hashes", priceHashResults)
	}

	var priceComponentCost *costs.PriceComponentCost

	priceComponentCost = testutil.PriceComponentCostFor(resourceCostBreakdowns, "aws_instance.instance1", "instance hours (m3.medium)")
	if !cmp.Equal(priceComponentCost.HourlyCost, priceComponentCost.PriceComponent.Price()) {
		t.Error("got unexpected cost", "aws_instance.instance1", "instance hours (m3.medium)")
	}

	priceComponentCost = testutil.PriceComponentCostFor(resourceCostBreakdowns, "aws_instance.instance1.root_block_device", "GB")
	if !cmp.Equal(priceComponentCost.MonthlyCost, priceComponentCost.PriceComponent.Price().Mul(decimal.NewFromInt(int64(10)))) {
		t.Error("got unexpected cost", "aws_instance.instance1.root_block_device", "GB")
	}

	priceComponentCost = testutil.PriceComponentCostFor(resourceCostBreakdowns, "aws_instance.instance1.ebs_block_device[0]", "GB")
	if !cmp.Equal(priceComponentCost.MonthlyCost, priceComponentCost.PriceComponent.Price().Mul(decimal.NewFromInt(int64(10)))) {
		t.Error("got unexpected cost", "aws_instance.instance1.ebs_block_device[0]", "GB")
	}

	priceComponentCost = testutil.PriceComponentCostFor(resourceCostBreakdowns, "aws_instance.instance1.ebs_block_device[1]", "GB")
	if !cmp.Equal(priceComponentCost.MonthlyCost, priceComponentCost.PriceComponent.Price().Mul(decimal.NewFromInt(int64(20)))) {
		t.Error("got unexpected cost", "aws_instance.instance1.ebs_block_device[1]", "GB")
	}

	priceComponentCost = testutil.PriceComponentCostFor(resourceCostBreakdowns, "aws_instance.instance1.ebs_block_device[2]", "GB")
	if !cmp.Equal(priceComponentCost.MonthlyCost, priceComponentCost.PriceComponent.Price().Mul(decimal.NewFromInt(int64(30)))) {
		t.Error("got unexpected cost", "aws_instance.instance1.ebs_block_device[2]", "GB")
	}

	priceComponentCost = testutil.PriceComponentCostFor(resourceCostBreakdowns, "aws_instance.instance1.ebs_block_device[3]", "GB")
	if !cmp.Equal(priceComponentCost.MonthlyCost, priceComponentCost.PriceComponent.Price().Mul(decimal.NewFromInt(int64(40)))) {
		t.Error("got unexpected cost", "aws_instance.instance1.ebs_block_device[3]", "GB")
	}

	priceComponentCost = testutil.PriceComponentCostFor(resourceCostBreakdowns, "aws_instance.instance1.ebs_block_device[3]", "IOPS")
	if !cmp.Equal(priceComponentCost.MonthlyCost, priceComponentCost.PriceComponent.Price().Mul(decimal.NewFromInt(int64(1000)))) {
		t.Error("got unexpected cost", "aws_instance.instance1.ebs_block_device[3]", "IOPS")
	}
}
