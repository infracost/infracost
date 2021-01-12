package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestRDSClusterInstance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_rds_cluster" "default" {
			cluster_identifier = "aurora-cluster-demo"
			availability_zones = ["us-east-1a", "us-east-1b", "us-east-1c"]
			database_name      = "mydb"
			master_username    = "foo"
			master_password    = "barbut8chars"
		}

		resource "aws_rds_cluster_instance" "cluster_instance" {
			identifier         = "aurora-cluster-demo"
			cluster_identifier = aws_rds_cluster.default.id
			instance_class     = "db.r4.large"
			engine             = aws_rds_cluster.default.engine
			engine_version     = aws_rds_cluster.default.engine_version
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_rds_cluster_instance.cluster_instance",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Database instance",
					PriceHash:       "dbf119ea9e222f1fa7ba244500eb005b-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
