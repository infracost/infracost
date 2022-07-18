package ibm

import "github.com/infracost/infracost/internal/schema"

var ResourceRegistry []*schema.RegistryItem = []*schema.RegistryItem{
	getIsInstanceRegistryItem(),
	getIbmIsVpcRegistryItem(),
	getContainerVpcWorkerPoolRegistryItem(),
	getContainerVpcClusterRegistryItem(),
}

// FreeResources grouped alphabetically
var FreeResources = []string{
	"ibm_resource_group",
	"ibm_iam_authorization_policy",
	"ibm_is_vpc_address_prefix",
	"ibm_is_security_group",
	"ibm_is_security_group_rule",
	"ibm_is_network_acl",
	"ibm_is_subnet",
	"ibm_is_ssh_key",
}

var UsageOnlyResources = []string{
	"",
}
