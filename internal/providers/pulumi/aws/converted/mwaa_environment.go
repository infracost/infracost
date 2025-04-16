package aws

import (
	"strings"

	"github.com/tidwall/gjson"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getMWAAEnvironmentRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_mwaa_environment",
		RFunc: NewMWAAEnvironment,
	}
}

func NewMWAAEnvironment(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	size := "mw1.small"
	if d.Get("environmentClass").Type != gjson.Null {
		size = d.Get("environmentClass").String()
	}

	size = strings.ToLower(size)
	size = strings.ReplaceAll(size, "mw1.", "")
	size = cases.Title(language.English).String(size)

	a := &aws.MWAAEnvironment{
		Address: d.Address,
		Region:  region,
		Size:    size,
	}

	a.PopulateUsage(u)
	return a.BuildResource()
}
