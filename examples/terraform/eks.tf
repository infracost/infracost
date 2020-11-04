resource "aws_eks_cluster" "example" {
  name     = "example"
  role_arn = "arn:aws:iam::123456789012:role/Example"

  vpc_config {
    subnet_ids      = ["subnet_id"]
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
