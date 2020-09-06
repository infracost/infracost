resource "aws_instance" "web_app" {
  ami           = var.aws_ami_id
  instance_type = "t3.micro"

  root_block_device {
    volume_size = 15
  }
}

resource "aws_ebs_volume" "storage_option_1" {
  availability_zone = var.availability_zone_names[0]
  type              = "io1"
  size              = 15
  iops              = 100
}

resource "aws_ebs_volume" "storage_option_2" {
  availability_zone = var.availability_zone_names[0]
  type              = "standard"
  size              = 15
}

resource "aws_eip" "nat_eip" {
  vpc = true
}

resource "aws_nat_gateway" "nat" {
  allocation_id = aws_eip.nat_eip.id
  subnet_id     = var.aws_subnet_ids[0]
}

# TODO Enable once we have released terraform-provider-infracost
# data "infracost_aws_nat_gateway" "nat" {
#   resources = [aws_nat_gateway.nat.id]

#   monthly_gb_data_processed {
#     value = 100
#   }
# }

resource "aws_lambda_function" "lambda" {
  function_name = "lambda_function_name"
  role          = "arn:aws:lambda:us-east-1:account-id:resource-id"
  handler       = "exports.test"
  runtime       = "nodejs12.x"
  memory_size   = 512
}

# TODO Enable once we have released terraform-provider-infracost
# data "infracost_aws_lambda_function" "lambda" {
#   resources = [aws_lambda_function.lambda.id]

#   monthly_requests {
#     value = 100000
#   }

#   average_request_duration {
#     value = 350
#   }
# }
