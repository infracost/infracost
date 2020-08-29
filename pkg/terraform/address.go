package terraform

import (
	"regexp"
	"strings"
)

func addressResourcePart(address string) string {
	addressParts := strings.Split(address, ".")
	resourceParts := addressParts[len(addressParts)-2:]
	return strings.Join(resourceParts, ".")
}

func addressModulePart(address string) string {
	addressParts := strings.Split(address, ".")
	moduleParts := addressParts[:len(addressParts)-2]
	return strings.Join(moduleParts, ".")
}

func addressModuleNames(address string) []string {
	r := regexp.MustCompile(`module\.([^\[]*)`)
	matches := r.FindAllStringSubmatch(addressModulePart(address), -1)

	moduleNames := make([]string, 0, len(matches))
	for _, match := range matches {
		moduleNames = append(moduleNames, match[1])
	}

	return moduleNames
}

func stripAddressArray(address string) string {
	addressParts := strings.Split(address, ".")
	resourceParts := addressParts[len(addressParts)-2:]

	r := regexp.MustCompile(`([^\[]+)`)
	match := r.FindStringSubmatch(strings.Join(resourceParts, "."))
	return match[1]
}
