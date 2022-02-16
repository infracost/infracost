package azure

import (
	"regexp"
	"strings"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"

	"github.com/infracost/infracost/internal/schema"
)

func strPtr(s string) *string {
	return &s
}

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

var sReg = regexp.MustCompile(`\s+`)

func toAzureCLIName(location string) string {
	return strings.ToLower(sReg.ReplaceAllString(location, ""))
}

func lookupRegion(d *schema.ResourceData, parentResourceKeys []string) string {
	// First check for a location set directly on a resource
	if d.Get("location").String() != "" {
		return toAzureCLIName(d.Get("location").String())
	}

	// Then check for any parent resources with a location
	for _, k := range parentResourceKeys {
		parents := d.References(k)
		for _, p := range parents {
			if p.Get("location").String() != "" {
				return toAzureCLIName(p.Get("location").String())
			}
		}
	}

	// When all else fails use the default region
	defaultRegion := toAzureCLIName(d.Get("region").String())
	log.Warnf("Using %s for resource %s as its 'location' property could not be found.", defaultRegion, d.Address)
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

func locationNameMapping(l string) string {
	name := map[string]string{
		"westus":             "West US",
		"westus2":            "West US 2",
		"eastus":             "East US",
		"centralus":          "Central US",
		"centraluseuap":      "Central US EUAP",
		"southcentralus":     "South Central US",
		"northcentralus":     "North Central US",
		"westcentralus":      "West Central US",
		"eastus2":            "East US 2",
		"eastus2euap":        "East US 2 EUAP",
		"brazilsouth":        "Brazil South",
		"brazilus":           "Brazil US",
		"northeurope":        "North Europe",
		"westeurope":         "West Europe",
		"eastasia":           "East Asia",
		"southeastasia":      "Southeast Asia",
		"japanwest":          "Japan West",
		"japaneast":          "Japan East",
		"koreacentral":       "Korea Central",
		"koreasouth":         "Korea South",
		"southindia":         "South India",
		"westindia":          "West India",
		"centralindia":       "Central India",
		"australiaeast":      "Australia East",
		"australiasoutheast": "Australia Southeast",
		"canadacentral":      "Canada Central",
		"canadaeast":         "Canada East",
		"uksouth":            "UK South",
		"ukwest":             "UK West",
		"francecentral":      "France Central",
		"francesouth":        "France South",
		"australiacentral":   "Australia Central",
		"australiacentral2":  "Australia Central 2",
		"uaecentral":         "UAE Central",
		"uaenorth":           "UAE North",
		"southafricanorth":   "South Africa North",
		"southafricawest":    "South Africa West",
		"switzerlandnorth":   "Switzerland North",
		"switzerlandwest":    "Switzerland West",
		"germanynorth":       "Germany North",
		"germanywestcentral": "Germany West Central",
		"norwayeast":         "Norway East",
		"norwaywest":         "Norway West",
		"brazilsoutheast":    "Brazil Southeast",
		"westus3":            "West US 3",
		"eastusslv":          "East US SLV",
		"swedencentral":      "Sweden Central",
		"swedensouth":        "Sweden South",
	}[l]

	return name
}

func intPtr(i int64) *int64 {
	return &i
}
