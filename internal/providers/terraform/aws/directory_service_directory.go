package aws

import (
	"regexp"
	"strings"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

var adReg = regexp.MustCompile(`(AD)`)

func getDirectoryServiceDirectory() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_directory_service_directory",
		CoreRFunc: newDirectoryServiceDirectory,
	}
}

func newDirectoryServiceDirectory(d *schema.ResourceData) schema.CoreResource {
	region := d.Get("region").String()
	regionName, ok := aws.RegionMapping[region]
	if !ok {
		logging.Logger.Warn().Msgf("Could not find mapping for resource %s region %s", d.Address, region)
	}

	a := &aws.DirectoryServiceDirectory{
		Address:    d.Address,
		Region:     region,
		RegionName: regionName,
		Type:       getType(d.Get("type").String()),
		Edition:    d.Get("edition").String(),
		Size:       d.Get("size").String(),
	}

	return a
}

// getType returns the terraform directory type with AD spaced, e.g:
// MicrosoftAD => Microsoft AD
func getType(t string) string {
	return strings.TrimSpace(adReg.ReplaceAllString(t, " AD "))
}
