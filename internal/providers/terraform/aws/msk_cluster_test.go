package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestMSKCluster(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
       resource "aws_msk_cluster" "cluster-2-nodes" {
           cluster_name = "cluster-2-nodes"
           kafka_version = "2.4.1"
           number_of_broker_nodes = 2
           broker_node_group_info {
             client_subnets = []
             ebs_volume_size = 500
             instance_type = "kafka.t3.small"
             security_groups = []
           }
       }

       resource "aws_msk_cluster" "cluster-4-nodes" {
           cluster_name = "cluster-4-nodes"
           kafka_version = "2.4.1"
           number_of_broker_nodes = 4
           broker_node_group_info {
             client_subnets = []
             ebs_volume_size = 1000
             instance_type = "kafka.m5.24xlarge"
             security_groups = []
           }
       }
`
	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_msk_cluster.cluster-2-nodes",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Instance (kafka.t3.small)",
					PriceHash:        "7384559214ca7695916562cfcdf52adf-1fb365d8a0bc1f462690ec9d444f380c",
					MonthlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(2)),
				},
				{
					Name:             "Storage",
					PriceHash:        "c051f24392b78290de9613ef486ba264-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1000)),
				},
			},
		}, {
			Name: "aws_msk_cluster.cluster-4-nodes",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Instance (kafka.m5.24xlarge)",
					PriceHash:        "9cc7cb79c34e4c7812d98da6dcdc8411-1fb365d8a0bc1f462690ec9d444f380c",
					MonthlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(4)),
				},
				{
					Name:             "Storage",
					PriceHash:        "c051f24392b78290de9613ef486ba264-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(4000)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
