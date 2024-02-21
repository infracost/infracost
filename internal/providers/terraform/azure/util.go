package azure

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

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
	log.Warn().Msgf("Using %s for resource %s as its 'location' property could not be found.", defaultRegion, d.Address)
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
