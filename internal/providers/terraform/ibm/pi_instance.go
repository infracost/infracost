package ibm

import (
	"strings"

	"github.com/infracost/infracost/internal/resources/ibm"
	"github.com/infracost/infracost/internal/schema"
)

// Operating System
const (
	AIX int64 = iota
	IBMI
	RHEL
	SLES
)

func getPiInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "ibm_pi_instance",
		RFunc:               newPiInstance,
		ReferenceAttributes: []string{"pi_image_id"},
	}
}

func identifyOperatingSystem(imageName string) int64 {
	splittedImageName := strings.Split(imageName, "-")

	if len(splittedImageName) == 0 {
		return -1
	}

	truncatedImageName := splittedImageName[0]

	if truncatedImageName == "7100" || truncatedImageName == "7200" || truncatedImageName == "7300" {
		return AIX
	}

	if truncatedImageName == "IBMi" {
		return IBMI
	}

	if truncatedImageName == "CentOS" || truncatedImageName == "Linux" || truncatedImageName == "RHEL8" {
		return RHEL
	}

	if truncatedImageName == "SLES15" {
		return SLES
	}

	return -1
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
