package aws_test

import (
	"github.com/shopspring/decimal"
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestRedshiftCluster(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_redshift_cluster" "ca" {
		  cluster_identifier = "tf-ca-cluster"
		  database_name      = "mydb"
		  master_username    = "foo"
		  master_password    = "Mustbe8characters"
		  node_type          = "dc2.large"
		  cluster_type       = "multi-node"
		  number_of_nodes    = "4"
		}

		resource "aws_redshift_cluster" "cb" {
		  cluster_identifier = "tf-cb-cluster"
		  database_name      = "mydb"
		  master_username    = "foo"
		  master_password    = "Mustbe8characters"
		  node_type          = "ds2.8xlarge"
		  cluster_type       = "single-node"
		}
`

	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_redshift_cluster.ca": map[string]interface{}{
			"excess_concurrency_scaling_secs": 4321,
			"spectrum_data_scanned_tb":        0.5,
			"backup_storage_gb":               612000,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_redshift_cluster.ca",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Cluster usage (on-demand, dc2.large)",
					PriceHash:       "7bcd81271b49c130729963694241ce61-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(4)),
				},
				{
					Name:             "Concurrency scaling (dc2.large)",
					PriceHash:        "ac168057c0de84c734f1e73076bf9c17-1786dd5ddb52682e127baa00bfaa4c48",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(4321)),
				},
				{
					Name:             "Spectrum",
					PriceHash:        "34795c8357acd40ec5f6308d75e733ea-665524fe1a13b69627c1f48ee6dcc45c",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(0.5)),
				},
				{
					Name:             "Backup storage (first 50 TB)",
					PriceHash:        "d991ea538483f74fd655b3233c61b427-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(51200)),
				},
				{
					Name:             "Backup storage (next 450 TB)",
					PriceHash:        "d991ea538483f74fd655b3233c61b427-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(512000)),
				},
				{
					Name:             "Backup storage (over 500 TB)",
					PriceHash:        "d991ea538483f74fd655b3233c61b427-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(48800)),
				},
			},
		},
		{
			Name: "aws_redshift_cluster.cb",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Cluster usage (on-demand, ds2.8xlarge)",
					PriceHash:       "98f7e94248d331282ac71bdefb73c1e1-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Concurrency scaling (ds2.8xlarge)",
					PriceHash:        "49ccdc0d30284e44b7245fcb6dd7af0b-1786dd5ddb52682e127baa00bfaa4c48",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
				{
					Name:             "Spectrum",
					PriceHash:        "34795c8357acd40ec5f6308d75e733ea-665524fe1a13b69627c1f48ee6dcc45c",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
				{
					Name:             "Backup storage (first 50 TB)",
					PriceHash:        "d991ea538483f74fd655b3233c61b427-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}

func TestRedshiftClusterWithManagedStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_redshift_cluster" "manageda" {
		  cluster_identifier = "tf-manageda-cluster"
		  database_name      = "mydb"
		  master_username    = "foo"
		  master_password    = "Mustbe8characters"
		  node_type          = "ra3.4xlarge"
		  cluster_type       = "multi-node"
		  number_of_nodes    = "6"
		}

		resource "aws_redshift_cluster" "managedb" {
		  cluster_identifier = "tf-managedb-cluster"
		  database_name      = "mydb"
		  master_username    = "foo"
		  master_password    = "Mustbe8characters"
		  node_type          = "ra3.16xlarge"
		  cluster_type       = "single-node"
		}
`

	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_redshift_cluster.manageda": map[string]interface{}{
			"managed_storage_gb": 321,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_redshift_cluster.manageda",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Cluster usage (on-demand, ra3.4xlarge)",
					PriceHash:       "69cd750f83010ae5166daec0615946b0-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(6)),
				},
				{
					Name:             "Managed storage (ra3.4xlarge)",
					PriceHash:        "ce02d660b0ba1d6c9e3a6c1767a99959-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(321)),
				},
				{
					Name:             "Concurrency scaling (ra3.4xlarge)",
					PriceHash:        "f3c3a2dbe50398af979ff3a5831b346d-1786dd5ddb52682e127baa00bfaa4c48",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
				{
					Name:             "Spectrum",
					PriceHash:        "34795c8357acd40ec5f6308d75e733ea-665524fe1a13b69627c1f48ee6dcc45c",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
				{
					Name:             "Backup storage (first 50 TB)",
					PriceHash:        "d991ea538483f74fd655b3233c61b427-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
		{
			Name: "aws_redshift_cluster.managedb",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Cluster usage (on-demand, ra3.16xlarge)",
					PriceHash:       "f7980b3ecb3526b919d608a6ce53e056-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Managed storage (ra3.16xlarge)",
					PriceHash:        "f5d3ad45c4ede98594971bd44109a0a9-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
				{
					Name:             "Concurrency scaling (ra3.16xlarge)",
					PriceHash:        "5b62904026ee6c37c82c3f412017d672-1786dd5ddb52682e127baa00bfaa4c48",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
				{
					Name:             "Spectrum",
					PriceHash:        "34795c8357acd40ec5f6308d75e733ea-665524fe1a13b69627c1f48ee6dcc45c",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
				{
					Name:             "Backup storage (first 50 TB)",
					PriceHash:        "d991ea538483f74fd655b3233c61b427-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
