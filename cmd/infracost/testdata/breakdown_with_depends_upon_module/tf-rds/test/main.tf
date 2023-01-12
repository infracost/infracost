resource "aws_eip" "test" {
  count = 1
}

resource "aws_eip" "count_zero" {
  count = 0
}

