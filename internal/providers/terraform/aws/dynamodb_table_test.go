package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/testutil"

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
		monthly_gb_continuous_backup_storage {
			value = 2300
		}
		monthly_gb_on_demand_backup_storage {
			value = 460
		}
		monthly_gb_restore {
			value = 230
		}
		monthly_streams_read_request_units {
			value = 2000000
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
				{
					Name:            "Continuous backup storage (PITR)",
					PriceHash:       "b4ed90c18b808ffff191ffbc16090c8e-ee3dd7e4624338037ca6fea0933a662f",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(2300)),
				},
				{
					Name:            "On-demand backup storage",
					PriceHash:       "0e228653f3f9c663398e91a605c911bd-8753f776c1e737f1a5548191571abc76",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(460)),
				},
				{
					Name:            "Restore data size",
					PriceHash:       "38fc5fdbec6f4ef5e3bdf6967dbe1cb2-b1ae3861dc57e2db217fa83a7420374f",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(230)),
				},
				{
					Name:            "Streams read request unit (sRRU)",
					PriceHash:       "dd063861f705295d00a801050a700b3e-4a9dfd3965ffcbab75845ead7a27fd47",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(2000000)),
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name: "Global table (us-east-2)",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:            "Replicated write request unit (rWRU)",
							PriceHash:       "bd1c30b527edcc061037142f79c06955-cf867fc796b8147fa126205baed2922c",
							HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(3000000)),
						},
					},
				},
				{
					Name: "Global table (us-west-1)",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:            "Replicated write request unit (rWRU)",
							PriceHash:       "67f1a3e0472747acf74cd5e925422fbb-cf867fc796b8147fa126205baed2922c",
							HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(3000000)),
						},
					},
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
				{
					Name:      "Data storage",
					SkipCheck: true,
				},
				{
					Name:      "Continuous backup storage (PITR)",
					SkipCheck: true,
				},
				{
					Name:      "On-demand backup storage",
					SkipCheck: true,
				},
				{
					Name:      "Restore data size",
					SkipCheck: true,
				},
				{
					Name:      "Streams read request unit (sRRU)",
					SkipCheck: true,
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name: "Global table (us-east-2)",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:            "Replicated write capacity unit (rWCU)",
							PriceHash:       "95e8dec74ece19d8d6b9c3ff60ef881b-af782957bf62d705bf1e97f981caeab1",
							HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(20)),
						},
					},
				},
				{
					Name: "Global table (us-west-1)",
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
