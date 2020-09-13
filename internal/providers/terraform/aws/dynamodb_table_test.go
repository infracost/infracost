package aws_test

import (
	"testing"

	"github.com/infracost/infracost/pkg/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestNewDynamoDBTableOnDemand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "aws_dynamodb_table" "my_dynamodb_table" {
		name           = "GameScores"
		billing_mode   = "PAY_PER_REQUEST"
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
	}
	  
	data "infracost_aws_dynamodb_table" "my_dynamodb_table" {
		resources = list(aws_dynamodb_table.my_dynamodb_table.id,)
	
		monthly_write_request_units {
			value = 3000000
		}
		monthly_read_request_units {
			value = 8000000
		}
		monthly_gb_data_storage {
		 	value = 230
		}
	}
	  `

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_dynamodb_table.my_dynamodb_table",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Write request unit (WRU)",
					PriceHash:       "075760076246f7bf5a2b46546e49cb31-418b228ac00af0f32e1843fecbc3d141",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(3000000)),
				},
				{
					Name:            "Read request unit (RRU)",
					PriceHash:       "641aa07510d472901906f3e97cee96c4-668942c2f9f9b475e74de593d4c32257",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(8000000)),
				},
				{
					Name:            "Data storage",
					PriceHash:       "a9781acb5ee117e6c50ab836dd7285b5-ee3dd7e4624338037ca6fea0933a662f",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(230)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, resourceChecks)

}

func TestNewDynamoDBTableProvisioned(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "aws_dynamodb_table" "my_dynamodb_table" {
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

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_dynamodb_table.my_dynamodb_table",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Write capacity unit (WCU)",
					PriceHash:       "b90795c897109784ce65409754460c41-8931e75640eb28f75b8eeb7989b3629d",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(20)),
				},
				{
					Name:            "Read capacity unit (RCU)",
					PriceHash:       "30812d4142a0a73eb1efbd902581679f-bd107312a4bed8ba719b7dc8dcfdaf95",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(30)),
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name: "us-east-2",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:            "Replicated write capacity unit (rWCU)",
							PriceHash:       "95e8dec74ece19d8d6b9c3ff60ef881b-af782957bf62d705bf1e97f981caeab1",
							HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(20)),
						},
					},
				},
				{
					Name: "us-west-1",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:            "Replicated write capacity unit (rWCU)",
							PriceHash:       "f472a25828ce71ef30b1aa898b7349ac-af782957bf62d705bf1e97f981caeab1",
							HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(20)),
						},
					},
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, resourceChecks)

}
