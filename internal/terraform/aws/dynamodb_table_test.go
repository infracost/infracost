package aws_test

import (
	"infracost/internal/testutil"
	"infracost/pkg/costs"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/shopspring/decimal"
)

func TestDynamoDBTableIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "aws_dynamodb_table" "dynamodb-table" {
		name           = "GameScores"
		billing_mode   = "PROVISIONED"
		read_capacity  = 30
		write_capacity = 20
		hash_key       = "UserId"
		range_key      = "GameTitle"
	  
		attribute {
		  name = "UserId"
		  type = "S"
		}
	  
		attribute {
		  name = "GameTitle"
		  type = "S"
		}
	  
		replica {
		  region_name = "us-east-2"
		}
	  
		replica {
		  region_name = "us-west-1"
		}
	  }`

	resourceCostBreakdowns, err := testutil.RunTFCostBreakdown(tf)
	if err != nil {
		t.Error(err)
	}

	expectedPriceHashes := [][]string{
		{"aws_dynamodb_table.dynamodb-table", "Read capacity unit (RCU)", "30812d4142a0a73eb1efbd902581679f-bd107312a4bed8ba719b7dc8dcfdaf95"},
		{"aws_dynamodb_table.dynamodb-table", "Write capacity unit (WCU)", "b90795c897109784ce65409754460c41-8931e75640eb28f75b8eeb7989b3629d"},
		{"aws_dynamodb_table.dynamodb-table.global_table.us-east-2", "Replicated write capacity unit (rWCU)", "95e8dec74ece19d8d6b9c3ff60ef881b-af782957bf62d705bf1e97f981caeab1"},
		{"aws_dynamodb_table.dynamodb-table.global_table.us-west-1", "Replicated write capacity unit (rWCU)", "f472a25828ce71ef30b1aa898b7349ac-af782957bf62d705bf1e97f981caeab1"},
	}

	priceHashResults := testutil.ExtractPriceHashes(resourceCostBreakdowns)

	if !cmp.Equal(priceHashResults, expectedPriceHashes, testutil.PriceHashResultSort) {
		t.Error("got unexpected price hashes", priceHashResults)
	}

	var priceComponentCost *costs.PriceComponentCost
	var actual, expected decimal.Decimal

	priceComponentCost = testutil.PriceComponentCostFor(resourceCostBreakdowns, "aws_dynamodb_table.dynamodb-table", "Read capacity unit (RCU)")
	actual = priceComponentCost.HourlyCost
	expected = priceComponentCost.PriceComponent.Price().Mul(decimal.NewFromInt(int64(30)))
	if !cmp.Equal(actual, expected) {
		t.Error("got unexpected cost", "aws_dynamodb_table.dynamodb-table", "Read capacity unit (RCU)", actual, expected)
	}

	priceComponentCost = testutil.PriceComponentCostFor(resourceCostBreakdowns, "aws_dynamodb_table.dynamodb-table", "Write capacity unit (WCU)")
	actual = priceComponentCost.HourlyCost
	expected = priceComponentCost.PriceComponent.Price().Mul(decimal.NewFromInt(int64(20)))
	if !cmp.Equal(actual, expected) {
		t.Error("got unexpected cost", "aws_dynamodb_table.dynamodb-table", "Write capacity unit (WCU)", actual, expected)
	}

	priceComponentCost = testutil.PriceComponentCostFor(resourceCostBreakdowns, "aws_dynamodb_table.dynamodb-table.global_table.us-east-2", "Replicated write capacity unit (rWCU)")
	actual = priceComponentCost.HourlyCost
	expected = priceComponentCost.PriceComponent.Price().Mul(decimal.NewFromInt(int64(20)))
	if !cmp.Equal(actual, expected) {
		t.Error("got unexpected cost", "aws_dynamodb_table.dynamodb-table.global_table.us-east-2", "Replicated write capacity unit (rWCU)", actual, expected)
	}

	priceComponentCost = testutil.PriceComponentCostFor(resourceCostBreakdowns, "aws_dynamodb_table.dynamodb-table.global_table.us-west-1", "Replicated write capacity unit (rWCU)")
	actual = priceComponentCost.HourlyCost
	expected = priceComponentCost.PriceComponent.Price().Mul(decimal.NewFromInt(int64(20)))
	if !cmp.Equal(actual, expected) {
		t.Error("got unexpected cost", "aws_dynamodb_table.dynamodb-table.global_table.us-west-1", "Replicated write capacity unit (rWCU)", actual, expected)
	}

}
