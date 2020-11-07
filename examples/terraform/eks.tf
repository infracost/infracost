resource "aws_eks_cluster" "example" {
  name     = "example"
  role_arn = "arn:aws:iam::123456789012:role/Example"

  vpc_config {
    subnet_ids      = ["subnet_id"]
  }
}

resource "aws_eks_fargate_profile" "example" {
  cluster_name           = aws_eks_cluster.example.name
  fargate_profile_name   = "example"
  pod_execution_role_arn = "arn:aws:iam::123456789012:role/Example"
  subnet_ids      = ["subnet_id"]

  selector {
    namespace = "example"
  }
}

resource "aws_eks_node_group" "example" {
  cluster_name    = "test aws_eks_node_group"
  node_group_name = "example"
  node_role_arn   = "node_role_arn"
  subnet_ids      = ["subnet_id"]

  scaling_config {
    desired_size = 1
    max_size     = 1
    min_size     = 1
  }
}
