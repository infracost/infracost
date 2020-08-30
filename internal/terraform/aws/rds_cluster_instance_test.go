package aws_test

import "testing"

func TestRDSClusterInstanceIntegration(t *testing.T) {
	NewTestIntegration(t, "aws_rds_cluster_instance", "cluster_instances", "dbf119ea9e222f1fa7ba244500eb005b-d2c98780d7b6e36641b521f1f8145c6f",
		`resource "aws_rds_cluster_instance" "cluster_instances" {
			identifier         = "aurora-cluster-demo"
			cluster_identifier = aws_rds_cluster.default.id
			instance_class     = "db.r4.large"
			engine             = aws_rds_cluster.default.engine
			engine_version     = aws_rds_cluster.default.engine_version
		}
	  
		resource "aws_rds_cluster" "default" {
			cluster_identifier = "aurora-cluster-demo"
			availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]
			database_name      = "mydb"
			master_username    = "foo"
			master_password    = "barbut8chars"
		}`)
}
