package aws_test

import (
	"testing"

	"infracost/internal/providers/terraform/tftest"
	"infracost/pkg/schema"
	"infracost/pkg/testutil"

	"github.com/google/go-cmp/cmp"
	"github.com/shopspring/decimal"
)

func TestAwsInstance(t *testing.T) {
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

	resources, err := tftest.RunCostCalculation(tf)
	if err != nil {
		t.Error(err)
	}

	expectedPriceHashes := [][]string{
		{"aws_instance.instance1", "Compute (m3.medium)", "666e02bbe686f6950fd8a47a55e83a75-d2c98780d7b6e36641b521f1f8145c6f"},
		{"root_block_device", "Storage", "efa8e70ebe004d2e9527fd30d50d09b2-ee3dd7e4624338037ca6fea0933a662f"},
		{"ebs_block_device[0]", "Storage", "efa8e70ebe004d2e9527fd30d50d09b2-ee3dd7e4624338037ca6fea0933a662f"},
		{"ebs_block_device[1]", "Storage", "0ed17ed1777b7be91f5b5ce79916d8d8-ee3dd7e4624338037ca6fea0933a662f"},
		{"ebs_block_device[2]", "Storage", "3122df29367c2460c76537cccf0eadb5-ee3dd7e4624338037ca6fea0933a662f"},
		{"ebs_block_device[3]", "Storage", "99450513de8c131ee2151e1b319d8143-ee3dd7e4624338037ca6fea0933a662f"},
		{"ebs_block_device[3]", "Storage IOPS", "d5c5e1fb9b8ded55c336f6ae87aa2c3b-9c483347596633f8cf3ab7fdd5502b78"},
	}

	priceHashResults := testutil.ExtractPriceHashes(resources)

	if !cmp.Equal(priceHashResults, expectedPriceHashes, testutil.PriceHashResultSort) {
		t.Error("Got unexpected price hashes", priceHashResults)
	}

	var costComponent *schema.CostComponent

	costComponent = testutil.FindCostComponent(resources, "aws_instance.instance1", "Compute (m3.medium)")
	testutil.CheckCost(t, "aws_instance.instance1", costComponent, "hourly", costComponent.Price())

	costComponent = testutil.FindCostComponent(resources, "root_block_device", "Storage")
	testutil.CheckCost(t, "root_block_device", costComponent, "monthly", costComponent.Price().Mul(decimal.NewFromInt(10)))

	costComponent = testutil.FindCostComponent(resources, "ebs_block_device[0]", "Storage")
	testutil.CheckCost(t, "ebs_block_device[0]", costComponent, "monthly", costComponent.Price().Mul(decimal.NewFromInt(10)))

	costComponent = testutil.FindCostComponent(resources, "ebs_block_device[1]", "Storage")
	testutil.CheckCost(t, "ebs_block_device[1]", costComponent, "monthly", costComponent.Price().Mul(decimal.NewFromInt(20)))

	costComponent = testutil.FindCostComponent(resources, "ebs_block_device[2]", "Storage")
	testutil.CheckCost(t, "ebs_block_device[2]", costComponent, "monthly", costComponent.Price().Mul(decimal.NewFromInt(30)))

	costComponent = testutil.FindCostComponent(resources, "ebs_block_device[3]", "Storage")
	testutil.CheckCost(t, "ebs_block_device[3]", costComponent, "monthly", costComponent.Price().Mul(decimal.NewFromInt(40)))

	costComponent = testutil.FindCostComponent(resources, "ebs_block_device[3]", "Storage IOPS")
	testutil.CheckCost(t, "ebs_block_device[3]", costComponent, "monthly", costComponent.Price().Mul(decimal.NewFromInt(1000)))
}
