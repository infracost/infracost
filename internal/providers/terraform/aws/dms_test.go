package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestNewNewDMSReplicationInstanceSingleLowStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "aws_dms_replication_instance" "my_dms_replication_instance" {
		allocated_storage            = 20
		apply_immediately            = true
		auto_minor_version_upgrade   = true
		availability_zone            = "us-east-1"
		engine_version               = "3.1.4"
		kms_key_arn                  = "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"
		multi_az                     = false
		preferred_maintenance_window = "sun:10:30-sun:14:30"
		publicly_accessible          = true
		replication_instance_class   = "dms.t2.micro"
		replication_instance_id      = "test-dms-replication-instance-tf"

		tags = {
		  Name = "test"
		}

		vpc_security_group_ids = [
		  "sg-12345678",
		]
	  }
	  `

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_dms_replication_instance.my_dms_replication_instance",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "General purpose storage (gp2)",
					PriceHash:       "ed71530b11f81a93bc9331f239089033-ee3dd7e4624338037ca6fea0933a662f",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(0)),
				},
				{
					Name:            "Instance (t2.micro)",
					PriceHash:       "83043eacffd1189707a3535ff5443948-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestNewNewDMSReplicationInstanceMultiHighStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "aws_dms_replication_instance" "my_dms_replication_instance" {
		allocated_storage            = 70
		apply_immediately            = true
		auto_minor_version_upgrade   = true
		availability_zone            = "us-east-1"
		engine_version               = "3.1.4"
		kms_key_arn                  = "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"
		multi_az                     = true
		preferred_maintenance_window = "sun:10:30-sun:14:30"
		publicly_accessible          = true
		replication_instance_class   = "dms.t2.micro"
		replication_instance_id      = "test-dms-replication-instance-tf"

		tags = {
		  Name = "test"
		}

		vpc_security_group_ids = [
		  "sg-12345678",
		]
	  }
	  `

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_dms_replication_instance.my_dms_replication_instance",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "General purpose storage (gp2)",
					PriceHash:       "309671f1b8cc2b6de57e782b60c79453-ee3dd7e4624338037ca6fea0933a662f",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(20)),
				},
				{
					Name:            "Instance (t2.micro)",
					PriceHash:       "b9210f23fc2812812995a2d8f64ef6b2-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
