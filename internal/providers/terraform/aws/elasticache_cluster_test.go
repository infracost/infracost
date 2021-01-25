package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestElastiCacheCluster(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
    resource "aws_elasticache_cluster" "memcached" {
        cluster_id = "cluster-example"
        engine = "memcached"
        node_type = "cache.m4.large"
        num_cache_nodes = 2
        parameter_group_name = "default.redis3.2"
    }

    resource "aws_elasticache_cluster" "redis" {
        cluster_id = "cluster-example"
        engine = "redis"
        node_type = "cache.m6g.12xlarge"
        num_cache_nodes = 1
        parameter_group_name = "default.redis3.2"
    }

    resource "aws_elasticache_cluster" "redis_snapshot" {
        cluster_id = "cluster-example"
        engine = "redis"
        node_type = "cache.m6g.12xlarge"
        num_cache_nodes = 1
        parameter_group_name = "default.redis3.2"
        snapshot_retention_limit = 2
    }
	`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_elasticache_cluster.memcached",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Elasticache (on-demand, cache.m4.large)",
					PriceHash:        "cb9f01e4a8f7b42c8b0d0a6ca3d0ca73-d2c98780d7b6e36641b521f1f8145c6f",
					MonthlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(2)),
				},
			},
		},
		{
			Name: "aws_elasticache_cluster.redis",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Elasticache (on-demand, cache.m6g.12xlarge)",
					PriceHash:        "3f3987c22a11fe709398881c9a36dc2a-d2c98780d7b6e36641b521f1f8145c6f",
					MonthlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		}, {
			Name: "aws_elasticache_cluster.redis_snapshot",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Elasticache (on-demand, cache.m6g.12xlarge)",
					PriceHash:        "3f3987c22a11fe709398881c9a36dc2a-d2c98780d7b6e36641b521f1f8145c6f",
					MonthlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
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
