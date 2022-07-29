package ibm

import "github.com/infracost/infracost/internal/schema"

var ResourceRegistry []*schema.RegistryItem = []*schema.RegistryItem{
	getIsInstanceRegistryItem(),
	getIbmIsVpcRegistryItem(),
	getIbmCosBucketRegistryItem(),
	getIsFloatingIpRegistryItem(),
	getIsFlowLogRegistryItem(),
	getContainerVpcWorkerPoolRegistryItem(),
	getContainerVpcClusterRegistryItem(),
	getResourceInstanceRegistryItem(),
	getIsVolumeRegistryItem(),
	getIsVpnGatewayRegistryItem(),
	getTgGatewayRegistryItem(),
}

// FreeResources grouped alphabetically
var FreeResources = []string{
	"ibm_atracker_route",
	"ibm_atracker_target",
	"ibm_iam_access_group",
	"ibm_iam_access_group_dynamic_rule",
	"ibm_iam_access_group_members",
	"ibm_iam_access_group_policy",
	"ibm_iam_account_settings",
	"ibm_iam_authorization_policy",
	"ibm_is_network_acl",
	"ibm_is_security_group",
	"ibm_is_security_group_rule",
	"ibm_is_ssh_key",
	"ibm_is_subnet",
	"ibm_is_subnet_reserved_ip",
	"ibm_is_virtual_endpoint_gateway",
	"ibm_is_virtual_endpoint_gateway_ip",
	"ibm_is_vpc_address_prefix",
	"ibm_is_vpn_gateway_connection",
	"ibm_kms_key",
	"ibm_kms_key_rings",
	"ibm_resource_group",
	"ibm_resource_key",
	"ibm_tg_connection",
}

var UsageOnlyResources = []string{
	"",
}
