output "id" {
  description = "The ID of the instance"
  value       = try(aws_instance.this[0].id, aws_spot_instance_request.this[0].id, "")
}

output "arn" {
  description = "The ARN of the instance"
  value       = try(aws_instance.this[0].arn, aws_spot_instance_request.this[0].arn, "")
}

output "capacity_reservation_specification" {
  description = "Capacity reservation specification of the instance"
  value       = try(aws_instance.this[0].capacity_reservation_specification, aws_spot_instance_request.this[0].capacity_reservation_specification, "")
}

output "instance_state" {
  description = "The state of the instance. One of: `pending`, `running`, `shutting-down`, `terminated`, `stopping`, `stopped`"
  value       = try(aws_instance.this[0].instance_state, aws_spot_instance_request.this[0].instance_state, "")
}

output "outpost_arn" {
  description = "The ARN of the Outpost the instance is assigned to"
  value       = try(aws_instance.this[0].outpost_arn, aws_spot_instance_request.this[0].outpost_arn, "")
}

output "password_data" {
  description = "Base-64 encoded encrypted password data for the instance. Useful for getting the administrator password for instances running Microsoft Windows. This attribute is only exported if `get_password_data` is true"
  value       = try(aws_instance.this[0].password_data, aws_spot_instance_request.this[0].password_data, "")
}

output "primary_network_interface_id" {
  description = "The ID of the instance's primary network interface"
  value       = try(aws_instance.this[0].primary_network_interface_id, aws_spot_instance_request.this[0].primary_network_interface_id, "")
}

output "private_dns" {
  description = "The private DNS name assigned to the instance. Can only be used inside the Amazon EC2, and only available if you've enabled DNS hostnames for your VPC"
  value       = try(aws_instance.this[0].private_dns, aws_spot_instance_request.this[0].private_dns, "")
}

output "public_dns" {
  description = "The public DNS name assigned to the instance. For EC2-VPC, this is only available if you've enabled DNS hostnames for your VPC"
  value       = try(aws_instance.this[0].public_dns, aws_spot_instance_request.this[0].public_dns, "")
}

output "public_ip" {
  description = "The public IP address assigned to the instance, if applicable. NOTE: If you are using an aws_eip with your instance, you should refer to the EIP's address directly and not use `public_ip` as this field will change after the EIP is attached"
  value       = try(aws_instance.this[0].public_ip, aws_spot_instance_request.this[0].public_ip, "")
}

output "private_ip" {
  description = "The private IP address assigned to the instance."
  value       = try(aws_instance.this[0].private_ip, aws_spot_instance_request.this[0].private_ip, "")
}

output "ipv6_addresses" {
  description = "The IPv6 address assigned to the instance, if applicable."
  value       = try(aws_instance.this[0].ipv6_addresses, [])
}

output "tags_all" {
  description = "A map of tags assigned to the resource, including those inherited from the provider default_tags configuration block"
  value       = try(aws_instance.this[0].tags_all, aws_spot_instance_request.this[0].tags_all, {})
}

output "spot_bid_status" {
  description = "The current bid status of the Spot Instance Request"
  value       = try(aws_spot_instance_request.this[0].spot_bid_status, "")
}

output "spot_request_state" {
  description = "The current request state of the Spot Instance Request"
  value       = try(aws_spot_instance_request.this[0].spot_request_state, "")
}

output "spot_instance_id" {
  description = "The Instance ID (if any) that is currently fulfilling the Spot Instance request"
  value       = try(aws_spot_instance_request.this[0].spot_instance_id, "")
}
