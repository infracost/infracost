package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getBigQueryDatasetRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_bigquery_dataset",
		CoreRFunc: NewBigQueryDataset,
	}
}

func NewBigQueryDataset(d *schema.ResourceData) schema.CoreResource {
	r := &google.BigQueryDataset{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}

	return r
}
