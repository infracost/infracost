package util

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"

	"github.com/infracost/infracost/internal/schema"
)

func StrPtr(s string) *string {
	return &s
}

func DecimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

func RegexPtr(regex string) *string {
	return StrPtr(fmt.Sprintf("/%s/i", regex))
}

var SReg = regexp.MustCompile(`\s+`)

func ToAzureCLIName(location string) string {
	return strings.ToLower(SReg.ReplaceAllString(location, ""))
}

func LookupRegion(d *schema.ResourceData, parentResourceKeys []string) string {
	// First check for a location set directly on a resource
	location := d.Get("location").String()
	if location != "" && !strings.Contains(location, "mock") {
		return ToAzureCLIName(location)
	}

	// Then check for any parent resources with a location
	for _, k := range parentResourceKeys {
		parents := d.References(k)
		for _, p := range parents {
			location := p.Get("location").String()
			if location != "" && !strings.Contains(location, "mock") {
				return ToAzureCLIName(location)
			}
		}
	}

	// When all else fails use the default region
	defaultRegion := ToAzureCLIName(d.Get("region").String())
	log.Warnf("Using %s for resource %s as its 'location' property could not be found.", defaultRegion, d.Address)
	return defaultRegion
}

func ConvertRegion(region string) string {
	if strings.Contains(strings.ToLower(region), "usgov") {
		return "US Gov"
	} else if strings.Contains(strings.ToLower(region), "china") {
		return "Сhina"
	} else {
		return "Global"
	}
}

func LocationNameMapping(l string) string {
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

func IntPtr(i int64) *int64 {
	return &i
}

func Contains(arr []string, e string) bool {
	for _, a := range arr {
		if a == e {
			return true
		}
	}
	return false
}
