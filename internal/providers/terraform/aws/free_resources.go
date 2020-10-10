package aws

import "github.com/infracost/infracost/internal/schema"

var (
	freeResourcesList []string = []string{
		"null_resource",

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

		// VPS aws_security_group_* resources
		"aws_security_group",
		"aws_security_group_rule",

		// Others
		"aws_launch_configuration",
		"aws_launch_template",
		"aws_ecs_cluster",
		"aws_ecs_task_definition",
		"aws_rds_cluster",
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
