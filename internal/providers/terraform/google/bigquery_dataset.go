package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getBigqueryDatasetRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_bigquery_dataset",
		RFunc: NewBigqueryDataset,
	}
}
func NewBigqueryDataset(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &google.BigqueryDataset{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
