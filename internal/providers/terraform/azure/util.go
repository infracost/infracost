package azure

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/schema"
)

func strPtr(s string) *string {
	return &s
}

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

func regexPtr(regex string) *string {
	return strPtr(fmt.Sprintf("/%s/i", regex))
}

var sReg = regexp.MustCompile(`\s+`)

func toAzureCLIName(location string) string {
	return strings.ToLower(sReg.ReplaceAllString(location, ""))
}

var lookupRegionFunctions = map[string]func(d *schema.ResourceData) string{
	"azurerm_mssql_elasticpool": func(d *schema.ResourceData) string {
		return lookupRegion(d, []string{"server_name", "resource_group_name"})
	},
	"azurerm_lb_rule": func(d *schema.ResourceData) string {
		return lookupRegion(d, []string{"loadbalancer_id"})
	},
	"azurerm_data_factory_integration_runtime_azure": func(d *schema.ResourceData) string {
		return lookupRegion(d, []string{"resource_group_name", "data_factory_id", "data_factory_name"})
	},
	"azurerm_traffic_manager_external_endpoint": func(d *schema.ResourceData) string {
		if len(d.References("profile_id")) > 0 {
			profile := d.References("profile_id")[0]
			return lookupRegion(profile, []string{"resource_group_name"})
		}

		return lookupRegion(d, []string{"resource_group_name"})
	},
	"azurerm_storage_queue": func(d *schema.ResourceData) string {
		return lookupRegion(d, []string{"storage_account_name"})
	},
	"azurerm_data_factory_integration_runtime_azure_ssis": func(d *schema.ResourceData) string {
		return lookupRegion(d, []string{"resource_group_name", "data_factory_id", "data_factory_name"})
	},
	"azurerm_app_service_certificate_binding": func(d *schema.ResourceData) string {
		return lookupRegion(d, []string{"certificate_id"})
	},
	"azurerm_cosmosdb_cassandra_keyspace": func(d *schema.ResourceData) string {
		if len(d.References("account_name")) > 0 {
			account := d.References("account_name")[0]
			return lookupRegion(account, []string{"account_name", "resource_group_name"})
		}

		return lookupRegion(d, []string{"resource_group_name"})
	},
	"azurerm_monitor_diagnostic_setting": func(d *schema.ResourceData) string {
		return lookupRegion(d, []string{"target_resource_id"})
	},
	"azurerm_sql_elasticpool": func(d *schema.ResourceData) string {
		return lookupRegion(d, []string{"server_name", "resource_group_name"})
	},
	"azurerm_virtual_network_peering": func(d *schema.ResourceData) string {
		return lookupRegion(d, []string{"virtual_network_name"})
	},
	"azurerm_kubernetes_cluster_node_pool": func(d *schema.ResourceData) string {
		return lookupRegion(d, []string{"kubernetes_cluster_id"})
	},
	"azurerm_key_vault_key": func(d *schema.ResourceData) string {
		return lookupRegion(d, []string{"key_vault_id"})
	},
	"azurerm_data_factory_integration_runtime_self_hosted": func(d *schema.ResourceData) string {
		return lookupRegion(d, []string{"resource_group_name", "data_factory_id", "data_factory_name"})
	},
	"azurerm_cognitive_deployment": func(d *schema.ResourceData) string {
		return lookupRegion(d, []string{"cognitive_account_id"})
	},
	"azurerm_traffic_manager_nested_endpoint": func(d *schema.ResourceData) string {
		if len(d.References("profile_id")) > 0 {
			profile := d.References("profile_id")[0]
			return lookupRegion(profile, []string{"resource_group_name"})
		}

		return lookupRegion(d, []string{"resource_group_name"})
	},
	"azurerm_lb_outbound_rule": func(d *schema.ResourceData) string {
		return lookupRegion(d, []string{"loadbalancer_id", "resource_group_name"})
	},
	"azurerm_synapse_sql_pool": func(d *schema.ResourceData) string {
		return lookupRegion(d, []string{"synapse_workspace_id"})
	},
	"azurerm_dns_ns_record": func(d *schema.ResourceData) string {
		return lookupRegion(d, []string{"resource_group_name"})
	},
	"azurerm_hdinsight_hadoop_cluster": func(d *schema.ResourceData) string { //nolint:misspell
		return lookupRegion(d, []string{})
	},
	"azurerm_traffic_manager_azure_endpoint": func(d *schema.ResourceData) string {
		if len(d.References("profile_id")) > 0 {
			profile := d.References("profile_id")[0]
			return lookupRegion(profile, []string{"resource_group_name"})
		}

		return lookupRegion(d, []string{"resource_group_name"})
	},
	"azurerm_key_vault_certificate": func(d *schema.ResourceData) string {
		return lookupRegion(d, []string{"key_vault_id"})
	},
	"azurerm_mariadb_server": func(d *schema.ResourceData) string {
		return lookupRegion(d, []string{})
	},
	"azurerm_postgresql_server": func(d *schema.ResourceData) string {
		return lookupRegion(d, []string{})
	},
	"azurerm_data_factory_integration_runtime_managed": func(d *schema.ResourceData) string {
		return lookupRegion(d, []string{"resource_group_name", "data_factory_id", "data_factory_name"})
	},
	"azurerm_storage_share": func(d *schema.ResourceData) string {
		return lookupRegion(d, []string{"storage_account_name"})
	},
	"azurerm_windows_virtual_machine": func(d *schema.ResourceData) string {
		return lookupRegion(d, []string{})
	},
	"azurerm_integration_service_environment": func(d *schema.ResourceData) string {
		return lookupRegion(d, []string{})
	},
}

func lookupRegion(d *schema.ResourceData, parentResourceKeys []string) string {
	// First check for a location set directly on a resource
	location := d.Get("location").String()
	if location != "" && !strings.Contains(location, "mock") {
		return toAzureCLIName(location)
	}

	// Then check for any parent resources with a location
	for _, k := range parentResourceKeys {
		parents := d.References(k)
		for _, p := range parents {
			location := p.Get("location").String()
			if location != "" && !strings.Contains(location, "mock") {
				return toAzureCLIName(location)
			}
		}
	}

	// When all else fails use the default region
	defaultRegion := toAzureCLIName(d.Get("region").String())
	logging.Logger.Debug().Msgf("Using %s for resource %s as its 'location' property could not be found.", defaultRegion, d.Address)
	return defaultRegion
}

func convertRegion(region string) string {
	if strings.Contains(strings.ToLower(region), "usgov") {
		return "US Gov"
	} else if strings.Contains(strings.ToLower(region), "china") {
		return "Сhina"
	} else {
		return "Global"
	}
}

// GetResourceRegion returns the region for the given azure resource data. By default,
// this uses the "location" property, but if that is not set it will traverse the
// references to find a location. The default references used are the resource
// group name. However, some resources specify different resources to get the
// location from, in which case a custom function is used that specifies which
// references to use.
func GetResourceRegion(d *schema.ResourceData) string {
	if d == nil {
		return ""
	}

	if v, ok := lookupRegionFunctions[d.Type]; ok {
		return v(d)
	}

	return lookupRegion(d, []string{"resource_group_name"})
}

// locationNameMapping returns a display name for a given location name.
// Up-to-date mapping can be found by running the following command:
//
//	az account list-locations -o json | jq '.[] | .name + " " + .displayName'
func locationNameMapping(l string) string {
	name := map[string]string{
		"eastus":              "East US",
		"eastus2":             "East US 2",
		"southcentralus":      "South Central US",
		"westus2":             "West US 2",
		"westus3":             "West US 3",
		"australiaeast":       "Australia East",
		"southeastasia":       "Southeast Asia",
		"northeurope":         "North Europe",
		"swedencentral":       "Sweden Central",
		"uksouth":             "UK South",
		"westeurope":          "West Europe",
		"centralus":           "Central US",
		"southafricanorth":    "South Africa North",
		"centralindia":        "Central India",
		"eastasia":            "East Asia",
		"japaneast":           "Japan East",
		"koreacentral":        "Korea Central",
		"canadacentral":       "Canada Central",
		"francecentral":       "France Central",
		"germanywestcentral":  "Germany West Central",
		"italynorth":          "Italy North",
		"norwayeast":          "Norway East",
		"polandcentral":       "Poland Central",
		"switzerlandnorth":    "Switzerland North",
		"uaenorth":            "UAE North",
		"brazilsouth":         "Brazil South",
		"centraluseuap":       "Central US EUAP",
		"israelcentral":       "Israel Central",
		"qatarcentral":        "Qatar Central",
		"centralusstage":      "Central US (Stage)",
		"eastusstage":         "East US (Stage)",
		"eastus2stage":        "East US 2 (Stage)",
		"northcentralusstage": "North Central US (Stage)",
		"southcentralusstage": "South Central US (Stage)",
		"westusstage":         "West US (Stage)",
		"westus2stage":        "West US 2 (Stage)",
		"asia":                "Asia",
		"asiapacific":         "Asia Pacific",
		"australia":           "Australia",
		"brazil":              "Brazil",
		"canada":              "Canada",
		"europe":              "Europe",
		"france":              "France",
		"germany":             "Germany",
		"global":              "Global",
		"india":               "India",
		"japan":               "Japan",
		"korea":               "Korea",
		"norway":              "Norway",
		"singapore":           "Singapore",
		"southafrica":         "South Africa",
		"sweden":              "Sweden",
		"switzerland":         "Switzerland",
		"uae":                 "United Arab Emirates",
		"uk":                  "United Kingdom",
		"unitedstates":        "United States",
		"unitedstateseuap":    "United States EUAP",
		"eastasiastage":       "East Asia (Stage)",
		"southeastasiastage":  "Southeast Asia (Stage)",
		"brazilus":            "Brazil US",
		"eastusstg":           "East US STG",
		"northcentralus":      "North Central US",
		"westus":              "West US",
		"japanwest":           "Japan West",
		"jioindiawest":        "Jio India West",
		"eastus2euap":         "East US 2 EUAP",
		"westcentralus":       "West Central US",
		"southafricawest":     "South Africa West",
		"australiacentral":    "Australia Central",
		"australiacentral2":   "Australia Central 2",
		"australiasoutheast":  "Australia Southeast",
		"jioindiacentral":     "Jio India Central",
		"koreasouth":          "Korea South",
		"southindia":          "South India",
		"westindia":           "West India",
		"canadaeast":          "Canada East",
		"francesouth":         "France South",
		"germanynorth":        "Germany North",
		"norwaywest":          "Norway West",
		"switzerlandwest":     "Switzerland West",
		"ukwest":              "UK West",
		"uaecentral":          "UAE Central",
		"brazilsoutheast":     "Brazil Southeast",
		"usgovvirginia":       "US Gov Virginia",
		"usgovarizona":        "US Gov Arizona",
		"usgovtexas":          "US Gov Texas",
	}[l]

	if name == "" {
		return l
	}

	return name
}

// regionToVNETZone returns the VNET zone for a given region.
//
// Mapped based on the values here: https://azure.microsoft.com/en-us/pricing/details/virtual-network/#faq
func regionToVNETZone(region string) string {
	return map[string]string{
		"eastus":              "Zone 1",
		"eastus2":             "Zone 1",
		"southcentralus":      "Zone 1",
		"westus2":             "Zone 1",
		"westus3":             "Zone 1",
		"australiaeast":       "Zone 2",
		"southeastasia":       "Zone 2",
		"northeurope":         "Zone 1",
		"swedencentral":       "Zone 1",
		"uksouth":             "Zone 1",
		"westeurope":          "Zone 1",
		"centralus":           "Zone 1",
		"southafricanorth":    "Zone 3",
		"centralindia":        "Zone 2",
		"eastasia":            "Zone 2",
		"japaneast":           "Zone 2",
		"koreacentral":        "Zone 2",
		"canadacentral":       "Zone 1",
		"francecentral":       "Zone 1",
		"germanywestcentral":  "Zone 1",
		"italynorth":          "Zone 1",
		"norwayeast":          "Zone 1",
		"polandcentral":       "Zone 1",
		"switzerlandnorth":    "Zone 1",
		"uaenorth":            "Zone 3",
		"brazilsouth":         "Zone 3",
		"centraluseuap":       "Zone 1",
		"israelcentral":       "Zone 1",
		"qatarcentral":        "Zone 1",
		"centralusstage":      "Zone 1",
		"eastusstage":         "Zone 1",
		"eastus2stage":        "Zone 1",
		"northcentralusstage": "Zone 1",
		"southcentralusstage": "Zone 1",
		"westusstage":         "Zone 1",
		"westus2stage":        "Zone 1",
		"asia":                "Zone 1",
		"asiapacific":         "Zone 1",
		"australia":           "Zone 1",
		"brazil":              "Zone 3",
		"canada":              "Zone 1",
		"europe":              "Zone 1",
		"france":              "Zone 1",
		"germany":             "Zone 1",
		"india":               "Zone 2",
		"japan":               "Zone 2",
		"korea":               "Zone 2",
		"norway":              "Zone 1",
		"singapore":           "Zone 1",
		"southafrica":         "Zone 3",
		"sweden":              "Zone 1",
		"switzerland":         "Zone 1",
		"uae":                 "Zone 3",
		"uk":                  "Zone 1",
		"unitedstates":        "Zone 1",
		"unitedstateseuap":    "Zone 1",
		"eastasiastage":       "Zone 2",
		"southeastasiastage":  "Zone 2",
		"brazilus":            "Zone 1",
		"eastusstg":           "Zone 1",
		"northcentralus":      "Zone 1",
		"westus":              "Zone 1",
		"japanwest":           "Zone 2",
		"jioindiawest":        "Zone 1",
		"eastus2euap":         "Zone 1",
		"westcentralus":       "Zone 1",
		"southafricawest":     "Zone 3",
		"australiacentral":    "Zone 1",
		"australiacentral2":   "Zone 1",
		"australiasoutheast":  "Zone 2",
		"jioindiacentral":     "Zone 2",
		"koreasouth":          "Zone 2",
		"southindia":          "Zone 2",
		"westindia":           "Zone 2",
		"canadaeast":          "Zone 1",
		"francesouth":         "Zone 1",
		"germanynorth":        "Zone 1",
		"norwaywest":          "Zone 1",
		"switzerlandwest":     "Zone 1",
		"ukwest":              "Zone 1",
		"uaecentral":          "Zone 3",
		"brazilsoutheast":     "Zone 3",
		"usgovvirginia":       "US Gov Zone 1",
		"usgovarizona":        "US Gov Zone 1",
		"usgovtexas":          "US Gov Zone 1",
	}[region]
}

// https://learn.microsoft.com/en-us/azure/cdn/cdn-billing#what-is-a-billing-region
func regionToCDNZone(region string) string {
	return map[string]string{
		"eastus":              "Zone 1",
		"eastus2":             "Zone 1",
		"southcentralus":      "Zone 1",
		"westus2":             "Zone 1",
		"westus3":             "Zone 1",
		"australiaeast":       "Zone 4",
		"southeastasia":       "Zone 2",
		"northeurope":         "Zone 1",
		"swedencentral":       "Zone 1",
		"uksouth":             "Zone 1",
		"westeurope":          "Zone 1",
		"centralus":           "Zone 1",
		"southafricanorth":    "Zone 1",
		"centralindia":        "Zone 5",
		"eastasia":            "Zone 2",
		"japaneast":           "Zone 2",
		"koreacentral":        "Zone 2",
		"canadacentral":       "Zone 1",
		"francecentral":       "Zone 1",
		"germanywestcentral":  "Zone 1",
		"italynorth":          "Zone 1",
		"norwayeast":          "Zone 1",
		"polandcentral":       "Zone 1",
		"switzerlandnorth":    "Zone 1",
		"uaenorth":            "Zone 1",
		"brazilsouth":         "Zone 3",
		"centraluseuap":       "Zone 1",
		"israelcentral":       "Zone 1",
		"qatarcentral":        "Zone 1",
		"centralusstage":      "Zone 1",
		"eastusstage":         "Zone 1",
		"eastus2stage":        "Zone 1",
		"northcentralusstage": "Zone 1",
		"southcentralusstage": "Zone 1",
		"westusstage":         "Zone 1",
		"westus2stage":        "Zone 1",
		"asia":                "Zone 2",
		"asiapacific":         "Zone 2",
		"australia":           "Zone 4",
		"brazil":              "Zone 3",
		"canada":              "Zone 1",
		"europe":              "Zone 1",
		"france":              "Zone 1",
		"germany":             "Zone 1",
		"india":               "Zone 5",
		"japan":               "Zone 2",
		"korea":               "Zone 2",
		"norway":              "Zone 1",
		"singapore":           "Zone 2",
		"southafrica":         "Zone 1",
		"sweden":              "Zone 1",
		"switzerland":         "Zone 1",
		"uae":                 "Zone 1",
		"uk":                  "Zone 1",
		"unitedstates":        "Zone 1",
		"unitedstateseuap":    "Zone 1",
		"eastasiastage":       "Zone 2",
		"southeastasiastage":  "Zone 2",
		"brazilus":            "Zone 3",
		"eastusstg":           "Zone 1",
		"northcentralus":      "Zone 1",
		"westus":              "Zone 1",
		"japanwest":           "Zone 2",
		"jioindiawest":        "Zone 5",
		"eastus2euap":         "Zone 1",
		"westcentralus":       "Zone 1",
		"southafricawest":     "Zone 1",
		"australiacentral":    "Zone 4",
		"australiacentral2":   "Zone 4",
		"australiasoutheast":  "Zone 4",
		"jioindiacentral":     "Zone 5",
		"koreasouth":          "Zone 2",
		"southindia":          "Zone 5",
		"westindia":           "Zone 5",
		"canadaeast":          "Zone 1",
		"francesouth":         "Zone 1",
		"germanynorth":        "Zone 1",
		"norwaywest":          "Zone 1",
		"switzerlandwest":     "Zone 1",
		"ukwest":              "Zone 1",
		"uaecentral":          "Zone 1",
		"brazilsoutheast":     "Zone 3",
		"usgovvirginia":       "US Gov Zone 1",
		"usgovarizona":        "US Gov Zone 1",
		"usgovtexas":          "US Gov Zone 1",
	}[region]
}

func intPtr(i int64) *int64 {
	return &i
}

func contains(arr []string, e string) bool {
	for _, a := range arr {
		if a == e {
			return true
		}
	}
	return false
}
