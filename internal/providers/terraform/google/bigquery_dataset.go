package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getBigQueryDatasetRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_bigquery_dataset",
		RFunc: NewBigQueryDataset,
	}
}

func NewBigQueryDataset(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &google.BigQueryDataset{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
