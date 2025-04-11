package aws

import (
	"strings"

	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getGlueJobRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_glue_job",
		RFunc: newGlueJob,
	}
}

func newGlueJob(d *schema.ResourceData) schema.CoreResource {
	region := d.Get("region").String()
	dpus := d.GetFloat64OrDefault("maxCapacity", 1)

	if !d.IsEmpty("number_of_workers") {
		var dpuPerWorker float64 = 1
		if strings.ToLower(d.Get("workerType").String()) == "g.2x" {
			dpuPerWorker = 2
		}

		dpus = d.Get("numberOfWorkers").Float() * dpuPerWorker
	}

	r := &aws.GlueJob{
		Address: d.Address,
		Region:  region,
		DPUs:    dpus,
	}

	return r
}
