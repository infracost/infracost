package aws

import "github.com/infracost/infracost/internal/schema"

var (
	freeResourcesList []string = []string{
		// Hashicorp
		"null_resource",
		"local_file",
		"template_dir",
		"random_id",
		"random_integer",
		"random_password",
		"random_pet",
		"random_shuffle",
		"random_string",
		"random_uuid",
		"tls_locally_signed_cert",
		"tls_private_key",
		"tls_self_signed_cert",
		"time_offset",
		"time_rotating",
		"time_sleep",
		"time_static",

		// IAM aws_iam_* resources
		"aws_iam_access_key",
		"aws_iam_account_alias",
		"aws_iam_account_password_policy",
		"aws_iam_group",
		"aws_iam_group_membership",
		"aws_iam_group_policy",
		"aws_iam_group_policy_attachment",
		"aws_iam_instance_profile",
		"aws_iam_openid_connect_provider",
		"aws_iam_policy",
		"aws_iam_policy_attachment",
		"aws_iam_role",
		"aws_iam_role_policy",
		"aws_iam_role_policy_attachment",
		"aws_iam_saml_provider",
		"aws_iam_server_certificate",
		"aws_iam_service_linked_role",
		"aws_iam_user",
		"aws_iam_user_group_membership",
		"aws_iam_user_login_profile",
		"aws_iam_user_policy",
		"aws_iam_user_policy_attachment",
		"aws_iam_user_ssh_key",

		// IAM aws_iam_* data sources
		"aws_iam_account_alias",
		"aws_iam_group",
		"aws_iam_instance_profile",
		"aws_iam_policy",
		"aws_iam_policy_document",
		"aws_iam_role",
		"aws_iam_server_certificate",
		"aws_iam_user",

		// Route53
		"aws_route53_zone_association",

		// S3
		"aws_s3_access_point",
		"aws_s3_account_public_access_block",
		"aws_s3_bucket_metric",
		"aws_s3_bucket_notification",
		"aws_s3_bucket_object", // Costs are shown at the bucket level
		"aws_s3_bucket_ownership_controls",
		"aws_s3_bucket_policy",
		"aws_s3_bucket_public_access_block",

		// VPC
		"aws_customer_gateway",
		"aws_default_network_acl",
		"aws_default_route_table",
		"aws_default_security_group",
		"aws_default_subnet",
		"aws_default_vpc",
		"aws_default_vpc_dhcp_options",
		"aws_egress_only_internet_gateway",
		"aws_flow_log",
		"aws_internet_gateway",
		"aws_main_route_table_association",
		"aws_network_acl",
		"aws_network_acl_rule",
		"aws_network_interface",
		"aws_network_interface_attachment",
		"aws_network_interface_sg_attachment",
		"aws_route",
		"aws_route_table",
		"aws_route_table_association",
		"aws_security_group",
		"aws_security_group_rule",
		"aws_subnet",
		"aws_vpc",
		"aws_vpc_dhcp_options",
		"aws_vpc_dhcp_options_association",
		"aws_vpc_endpoint_connection_notification",
		"aws_vpc_endpoint_route_table_association",
		"aws_vpc_endpoint_service",
		"aws_vpc_endpoint_service_allowed_principal",
		"aws_vpc_endpoint_subnet_association",
		"aws_vpc_ipv4_cidr_block_association",
		"aws_vpc_peering_connection",
		"aws_vpc_peering_connection_accepter",
		"aws_vpc_peering_connection_options",
		"aws_vpn_connection_route",
		"aws_vpn_gateway",
		"aws_vpn_gateway_attachment",
		"aws_vpn_gateway_route_propagation",

		// Elastic Load Balancing
		"aws_alb_listener",
		"aws_alb_listener_certificate",
		"aws_alb_listener_rule",
		"aws_alb_target_group",
		"aws_alb_target_group_attachment",
		"aws_lb_listener",
		"aws_lb_listener_certificate",
		"aws_lb_listener_rule",
		"aws_lb_target_group",
		"aws_lb_target_group_attachment",
		"aws_app_cookie_stickiness_policy",
		"aws_elb_attachment",
		"aws_lb_cookie_stickiness_policy",
		"aws_lb_ssl_negotiation_policy",
		"aws_load_balancer_backend_server_policy",
		"aws_load_balancer_listener_policy",
		"aws_load_balancer_policy",

		// Others (sorted alphabetically)
		"aws_rds_cluster",
		"aws_ecs_cluster",
		"aws_ecs_task_definition",
		"aws_eip_association",
		"aws_launch_configuration",
		"aws_launch_template",
	}
)

func GetFreeResources() []*schema.RegistryItem {
	freeResources := make([]*schema.RegistryItem, 0)
	for _, resourceName := range freeResourcesList {
		freeResources = append(freeResources, &schema.RegistryItem{
			Name:    resourceName,
			NoPrice: true,
			Notes:   []string{"Free resource."},
		})
	}
	return freeResources
}
