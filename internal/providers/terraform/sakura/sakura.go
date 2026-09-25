package sakura

import "github.com/infracost/infracost/internal/schema"

var DefaultProviderRegion = "is1a"

func GetDefaultRefIDFunc(d *schema.ResourceData) []string {
	return []string{d.Get("id").String()}
}

func DefaultCloudResourceIDFunc(d *schema.ResourceData) []string {
	id := d.Get("id").String()
	if id == "" {
		return []string{}
	}
	return []string{id}
}

func GetResourceRegion(d *schema.ResourceData) string {
	return d.Get("zone").String()
}
