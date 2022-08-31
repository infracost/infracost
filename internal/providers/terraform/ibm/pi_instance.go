package ibm

import (
	"strings"

	"github.com/infracost/infracost/internal/resources/ibm"
	"github.com/infracost/infracost/internal/schema"
)

func getPiInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "ibm_pi_instance",
		RFunc:               newPiInstance,
		ReferenceAttributes: []string{"pi_image_id"},
	}
}

func isNetweaverImage(imageName string) bool {
	return strings.Contains(imageName, "NETWEAVER")
}

func identifyOperatingSystem(imageName string) int64 {
	splittedImageName := strings.Split(imageName, "-")

	truncatedImageName := splittedImageName[0]

	if truncatedImageName == "7100" || truncatedImageName == "7200" || truncatedImageName == "7300" {
		return ibm.AIX
	}

	if truncatedImageName == "IBMi" {
		return ibm.IBMI
	}

	if truncatedImageName == "CentOS" || truncatedImageName == "Linux" || truncatedImageName == "RHEL8" {
		return ibm.RHEL
	}

	if truncatedImageName == "SLES15" {
		return ibm.SLES
	}

	return -1
}

func isIBMiVersionLegacy(imageName string) bool {
	splittedImageName := strings.Split(imageName, "-")

	if len(splittedImageName) == 1 {
		return false
	}

	version := splittedImageName[1]

	if version == "71" || version == "72" {
		return true
	}

	return false
}

func newPiInstance(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	refs := d.References("pi_image_id")

	imageName := ""

	if len(refs) > 0 {
		imageName = refs[0].Get("pi_image_name").String()
	}

	region := d.Get("region").String()
	systemType := d.Get("pi_sys_type").String()
	processorMode := d.Get("pi_proc_type").String()
	cpus := d.Get("pi_processors").Float()
	memory := d.Get("pi_memory").Float()
	storageType := d.Get("pi_storage_type").String()

	var os int64 = -1
	var isLegacyIBMiImageVersion bool
	var netweaverImage bool

	if len(imageName) > 0 {
		os = identifyOperatingSystem(imageName)
		isLegacyIBMiImageVersion = isIBMiVersionLegacy(imageName)
		netweaverImage = isNetweaverImage(imageName)
	}

	r := &ibm.PiInstance{
		Address:                d.Address,
		Region:                 region,
		SystemType:             systemType,
		ProcessorMode:          processorMode,
		Cpus:                   cpus,
		Memory:                 memory,
		StorageType:            storageType,
		OperatingSystem:        os,
		LegacyIBMiImageVersion: isLegacyIBMiImageVersion,
		NetweaverImage:         netweaverImage,
	}
	r.PopulateUsage(u)

	configuration := make(map[string]any)
	configuration["region"] = region
	configuration["systemType"] = systemType
	configuration["processorMode"] = processorMode
	configuration["cpus"] = cpus
	configuration["memory"] = memory
	configuration["storageType"] = storageType

	SetCatalogMetadata(d, d.Type, configuration)

	return r.BuildResource()
}
