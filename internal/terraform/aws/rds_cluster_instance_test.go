package aws_test

import "testing"

func TestRDSClusterInstanceIntegration(t *testing.T) {
	NewTestIntegration(t, "aws_rds_cluster_instance", "cluster_instances", "2d53444cfb77f6baa54be3ac2cc3cb90-d2c98780d7b6e36641b521f1f8145c6f",
		`resource "aws_rds_cluster_instance" "cluster_instances" {
		count              = 2
		identifier         = "aurora-cluster-demo-${count.index}"
		cluster_identifier = "mock_id"
		instance_class     = "db.r4.large"
		engine             = aws_rds_cluster.default.engine
		engine_version     = aws_rds_cluster.default.engine_version
	}`)
}
