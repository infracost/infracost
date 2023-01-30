# EC2 Complete
output "ec2_complete_id" {
  description = "The ID of the instance"
  value       = module.ec2_complete.id
}

output "ec2_complete_arn" {
  description = "The ARN of the instance"
  value       = module.ec2_complete.arn
}

output "ec2_complete_capacity_reservation_specification" {
  description = "Capacity reservation specification of the instance"
  value       = module.ec2_complete.capacity_reservation_specification
}

output "ec2_complete_instance_state" {
  description = "The state of the instance. One of: `pending`, `running`, `shutting-down`, `terminated`, `stopping`, `stopped`"
  value       = module.ec2_complete.instance_state
}

output "ec2_complete_primary_network_interface_id" {
  description = "The ID of the instance's primary network interface"
  value       = module.ec2_complete.primary_network_interface_id
}

output "ec2_complete_private_dns" {
  description = "The private DNS name assigned to the instance. Can only be used inside the Amazon EC2, and only available if you've enabled DNS hostnames for your VPC"
  value       = module.ec2_complete.private_dns
}

output "ec2_complete_public_dns" {
  description = "The public DNS name assigned to the instance. For EC2-VPC, this is only available if you've enabled DNS hostnames for your VPC"
  value       = module.ec2_complete.public_dns
}

output "ec2_complete_public_ip" {
  description = "The public IP address assigned to the instance, if applicable. NOTE: If you are using an aws_eip with your instance, you should refer to the EIP's address directly and not use `public_ip` as this field will change after the EIP is attached"
  value       = module.ec2_complete.public_ip
}

output "ec2_complete_tags_all" {
  description = "A map of tags assigned to the resource, including those inherited from the provider default_tags configuration block"
  value       = module.ec2_complete.tags_all
}

# EC2 T2 Unlimited
output "ec2_t2_unlimited_id" {
  description = "The ID of the instance"
  value       = module.ec2_t2_unlimited.id
}

output "ec2_t2_unlimited_arn" {
  description = "The ARN of the instance"
  value       = module.ec2_t2_unlimited.arn
}

output "ec2_t2_unlimited_capacity_reservation_specification" {
  description = "Capacity reservation specification of the instance"
  value       = module.ec2_t2_unlimited.capacity_reservation_specification
}

output "ec2_t2_unlimited_instance_state" {
  description = "The state of the instance. One of: `pending`, `running`, `shutting-down`, `terminated`, `stopping`, `stopped`"
  value       = module.ec2_t2_unlimited.instance_state
}

output "ec2_t2_unlimited_primary_network_interface_id" {
  description = "The ID of the instance's primary network interface"
  value       = module.ec2_t2_unlimited.primary_network_interface_id
}

output "ec2_t2_unlimited_private_dns" {
  description = "The private DNS name assigned to the instance. Can only be used inside the Amazon EC2, and only available if you've enabled DNS hostnames for your VPC"
  value       = module.ec2_t2_unlimited.private_dns
}

output "ec2_t2_unlimited_public_dns" {
  description = "The public DNS name assigned to the instance. For EC2-VPC, this is only available if you've enabled DNS hostnames for your VPC"
  value       = module.ec2_t2_unlimited.public_dns
}

output "ec2_t2_unlimited_public_ip" {
  description = "The public IP address assigned to the instance, if applicable. NOTE: If you are using an aws_eip with your instance, you should refer to the EIP's address directly and not use `public_ip` as this field will change after the EIP is attached"
  value       = module.ec2_t2_unlimited.public_ip
}

output "ec2_t2_unlimited_tags_all" {
  description = "A map of tags assigned to the resource, including those inherited from the provider default_tags configuration block"
  value       = module.ec2_t2_unlimited.tags_all
}

# EC2 T3 Unlimited
output "ec2_t3_unlimited_id" {
  description = "The ID of the instance"
  value       = module.ec2_t3_unlimited.id
}

output "ec2_t3_unlimited_arn" {
  description = "The ARN of the instance"
  value       = module.ec2_t3_unlimited.arn
}

output "ec2_t3_unlimited_capacity_reservation_specification" {
  description = "Capacity reservation specification of the instance"
  value       = module.ec2_t3_unlimited.capacity_reservation_specification
}

output "ec2_t3_unlimited_instance_state" {
  description = "The state of the instance. One of: `pending`, `running`, `shutting-down`, `terminated`, `stopping`, `stopped`"
  value       = module.ec2_t3_unlimited.instance_state
}

output "ec2_t3_unlimited_primary_network_interface_id" {
  description = "The ID of the instance's primary network interface"
  value       = module.ec2_t3_unlimited.primary_network_interface_id
}

output "ec2_t3_unlimited_private_dns" {
  description = "The private DNS name assigned to the instance. Can only be used inside the Amazon EC2, and only available if you've enabled DNS hostnames for your VPC"
  value       = module.ec2_t3_unlimited.private_dns
}

output "ec2_t3_unlimited_public_dns" {
  description = "The public DNS name assigned to the instance. For EC2-VPC, this is only available if you've enabled DNS hostnames for your VPC"
  value       = module.ec2_t3_unlimited.public_dns
}

output "ec2_t3_unlimited_public_ip" {
  description = "The public IP address assigned to the instance, if applicable. NOTE: If you are using an aws_eip with your instance, you should refer to the EIP's address directly and not use `public_ip` as this field will change after the EIP is attached"
  value       = module.ec2_t3_unlimited.public_ip
}

output "ec2_t3_unlimited_tags_all" {
  description = "A map of tags assigned to the resource, including those inherited from the provider default_tags configuration block"
  value       = module.ec2_t3_unlimited.tags_all
}

# EC2 Multiple
output "ec2_multiple" {
  description = "The full output of the `ec2_module` module"
  value       = module.ec2_multiple
}

# EC2 Spot Instance
output "ec2_spot_instance_id" {
  description = "The ID of the instance"
  value       = module.ec2_spot_instance.id
}

output "ec2_spot_instance_arn" {
  description = "The ARN of the instance"
  value       = module.ec2_spot_instance.arn
}

output "ec2_spot_instance_capacity_reservation_specification" {
  description = "Capacity reservation specification of the instance"
  value       = module.ec2_spot_instance.capacity_reservation_specification
}

output "ec2_spot_instance_instance_state" {
  description = "The state of the instance. One of: `pending`, `running`, `shutting-down`, `terminated`, `stopping`, `stopped`"
  value       = module.ec2_spot_instance.instance_state
}

output "ec2_spot_instance_primary_network_interface_id" {
  description = "The ID of the instance's primary network interface"
  value       = module.ec2_spot_instance.primary_network_interface_id
}

output "ec2_spot_instance_private_dns" {
  description = "The private DNS name assigned to the instance. Can only be used inside the Amazon EC2, and only available if you've enabled DNS hostnames for your VPC"
  value       = module.ec2_spot_instance.private_dns
}

output "ec2_spot_instance_public_dns" {
  description = "The public DNS name assigned to the instance. For EC2-VPC, this is only available if you've enabled DNS hostnames for your VPC"
  value       = module.ec2_spot_instance.public_dns
}

output "ec2_spot_instance_public_ip" {
  description = "The public IP address assigned to the instance, if applicable. NOTE: If you are using an aws_eip with your instance, you should refer to the EIP's address directly and not use `public_ip` as this field will change after the EIP is attached"
  value       = module.ec2_spot_instance.public_ip
}

output "ec2_spot_instance_tags_all" {
  description = "A map of tags assigned to the resource, including those inherited from the provider default_tags configuration block"
  value       = module.ec2_spot_instance.tags_all
}

output "spot_bid_status" {
  description = "The current bid status of the Spot Instance Request"
  value       = module.ec2_spot_instance.spot_bid_status
}

output "spot_request_state" {
  description = "The current request state of the Spot Instance Request"
  value       = module.ec2_spot_instance.spot_request_state
}

output "spot_instance_id" {
  description = "The Instance ID (if any) that is currently fulfilling the Spot Instance request"
  value       = module.ec2_spot_instance.spot_instance_id
}
