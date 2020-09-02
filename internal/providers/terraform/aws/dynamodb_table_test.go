package aws_test

import (
	"infracost/internal/providers/terraform/tftest"
	"infracost/pkg/testutil"
	"testing"

	"github.com/shopspring/decimal"
)

func TestNewDynamoDBTable(t *testing.T) {
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

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_dynamodb_table.dynamodb-table",
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
