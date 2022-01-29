package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getLoggingFolderSinkRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_logging_folder_sink",
		RFunc: NewLoggingFolderSink,
	}
}

func NewLoggingFolderSink(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &google.Logging{
		Address: d.Address,
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
