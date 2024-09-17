variable "types" {
  type = list(string)
}

variable "amount" {
  type    = number
  default = 2
}

resource "aws_autoscaling_group" "test" {
  count                = var.amount
  desired_capacity     = 2
  max_size             = 3
  min_size             = 1
  launch_configuration = aws_launch_configuration.test.*.id[count.index]
}

resource "aws_launch_configuration" "test" {
  count         = var.amount
  image_id      = "fake_ami"
  instance_type = var.types[count.index]
}
