package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestElastiCacheReplicationGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
    resource "aws_elasticache_replication_group" "cluster" {
        replication_group_description = "This Replication Group"
        replication_group_id = "tf-rep-group-1"
        automatic_failover_enabled = true
        node_type = "cache.m4.large"

        engine = "redis"

        cluster_mode {
          num_node_groups = 4
          replicas_per_node_group = 3
        }
    }

    resource "aws_elasticache_replication_group" "non-cluster" {
        replication_group_description = "This Replication Group"
        replication_group_id = "tf-rep-group-2"

        engine = "redis"

        node_type = "cache.r5.4xlarge"
        number_cache_clusters = 3
    }

    resource "aws_elasticache_replication_group" "non-cluster-snapshot" {
        replication_group_description = "This Replication Group"
        replication_group_id = "tf-rep-group-3"
        snapshot_retention_limit = 2

        engine = "redis"

        node_type = "cache.m6g.12xlarge"
        number_cache_clusters = 3
    }
	`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_elasticache_replication_group.cluster",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Elasticache (on-demand, cache.m4.large)",
					PriceHash:        "e117895199efffd8123c0683af6e5334-d2c98780d7b6e36641b521f1f8145c6f",
					MonthlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(16)),
				},
			},
		},
		{
			Name: "aws_elasticache_replication_group.non-cluster",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Elasticache (on-demand, cache.r5.4xlarge)",
					PriceHash:        "a11f8e86a1c660d8018fb539356c84e8-d2c98780d7b6e36641b521f1f8145c6f",
					MonthlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(3)),
				},
			},
		},
		{
			Name: "aws_elasticache_replication_group.non-cluster-snapshot",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Elasticache (on-demand, cache.m6g.12xlarge)",
					PriceHash:        "3f3987c22a11fe709398881c9a36dc2a-d2c98780d7b6e36641b521f1f8145c6f",
					MonthlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(3)),
				},
				{
					Name:             "Backup storage",
					PriceHash:        "5a1365e07213003f7a7b9deaa791b017-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(0)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
