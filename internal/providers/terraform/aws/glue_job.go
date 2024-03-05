package aws

import (
	"strings"

	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getGlueJobRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_glue_job",
		CoreRFunc: newGlueJob,
	}
}

func newGlueJob(d *schema.ResourceData) schema.CoreResource {
	region := d.Get("region").String()
	dpus := d.GetFloat64OrDefault("max_capacity", 1)

	if !d.IsEmpty("number_of_workers") {
		var dpuPerWorker float64 = 1
		if strings.ToLower(d.Get("worker_type").String()) == "g.2x" {
			dpuPerWorker = 2
		}

		dpus = d.Get("number_of_workers").Float() * dpuPerWorker
	}

	r := &aws.GlueJob{
		Address: d.Address,
		Region:  region,
		DPUs:    dpus,
	}

	return r
}
