package aws_test

tf := `
resource "aws_rds_cluster_instance" "cluster_instances" {
	count              = 2
	identifier         = "aurora-cluster-demo-${count.index}"
	cluster_identifier = aws_rds_cluster.default.id
	instance_class     = "db.r4.large"
	engine             = aws_rds_cluster.default.engine
	engine_version     = aws_rds_cluster.default.engine_version
}`

TestRDSClusterInstanceIntegration := TestIntegration("aws_rds_cluster_instance", "cluster_instances", tf, "6e137a9da0718f0ec80fb60866730ba9-d2c98780d7b6e36641b521f1f8145c6f")