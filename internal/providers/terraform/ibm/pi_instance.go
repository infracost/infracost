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

func identifyOperatingSystem(imageName string) string {
	splittedImageName := strings.Split(imageName, "-")[0]

	if splittedImageName == "7100" || splittedImageName == "7200" || splittedImageName == "7300" {
		return "aix"
	}

	if splittedImageName == "IBMi" {
		return "ibmi"
	}

	if splittedImageName == "CentOS" || splittedImageName == "Linux" || splittedImageName == "RHEL8" {
		return "rhel"
	}

	if splittedImageName == "SLES15" {
		return "sles"
	}

	return ""
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
	os := identifyOperatingSystem(imageName)

	r := &ibm.PiInstance{
		Address:         d.Address,
		Region:          region,
		SystemType:      systemType,
		ProcessorMode:   processorMode,
		Cpus:            cpus,
		Memory:          memory,
		StorageType:     storageType,
		OperatingSystem: os,
	}
	r.PopulateUsage(u)

	return r.BuildResource()
}
