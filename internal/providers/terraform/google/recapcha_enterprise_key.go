package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getRecaptchaEnterpriseKeyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_recaptcha_enterprise_key",
		CoreRFunc: newRecaptchaEnterpriseKey,
		GetRegion: func(defaultRegion string, d *schema.ResourceData) string {
			return "global"
		},
	}
}

func newRecaptchaEnterpriseKey(d *schema.ResourceData) schema.CoreResource {
	return &google.RecaptchaEnterpriseKey{
		Address: d.Address,
		Region:  "global", 
	}
}
