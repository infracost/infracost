package sakura

import "github.com/infracost/infracost/internal/schema"

var ResourceRegistry []*schema.RegistryItem = []*schema.RegistryItem{
	getApprunSharedRegistryItem(),
	getDiskRegistryItem(),
	getInternetRegistryItem(),
	getPrivateHostRegistryItem(),
	getServerRegistryItem(),
}

// FreeResources lists sakura provider resources that have no cost.
var FreeResources = []string{
	"sakura_apprun_dedicated_cluster",
	"sakura_auto_backup",
	"sakura_bridge",
	"sakura_icon",
	"sakura_ipv4_ptr",
	"sakura_packet_filter",
	"sakura_packet_filter_rules",
	"sakura_script",
	"sakura_ssh_key",
	"sakura_switch",
	"sakura_vswitch",
	"sakura_zone",
}

var UsageOnlyResources = []string{}
