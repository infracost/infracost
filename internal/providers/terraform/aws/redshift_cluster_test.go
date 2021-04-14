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
		  node_type          = "dc1.large"
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

	usage := schema.NewUsageMap(map[string]interface{}{})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_redshift_cluster.ca",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Redshift Cluster (dc1.large)",
					PriceHash:       "84aee42f68a79efe21f00bd6465bb922-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(4)),
				},
			},
		},
		{
			Name: "aws_redshift_cluster.cb",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Redshift Cluster (ds2.8xlarge)",
					PriceHash:       "98f7e94248d331282ac71bdefb73c1e1-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
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

	usage := schema.NewUsageMap(map[string]interface{}{})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_redshift_cluster.manageda",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Redshift Cluster (ra3.4xlarge)",
					PriceHash:       "69cd750f83010ae5166daec0615946b0-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(6)),
				},
			},
		},
		{
			Name: "aws_redshift_cluster.managedb",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Redshift Cluster (ra3.16xlarge)",
					PriceHash:       "f7980b3ecb3526b919d608a6ce53e056-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
