package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getKMSCryptoKeyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_kms_crypto_key",
		RFunc: NewKMSCryptoKey,
	}
}
func NewKMSCryptoKey(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &google.KMSCryptoKey{Address: d.Address, Region: d.Get("region").String(), VersionTemplate0Algorithm: d.Get("version_template.0.algorithm").String(), VersionTemplate0ProtectionLevel: d.Get("version_template.0.protection_level").String()}
	if !d.IsEmpty("rotation_period") {
		r.RotationPeriod = d.Get("rotation_period").String()
	}
	if !d.IsEmpty("version_template") {
		r.VersionTemplate = d.Get("version_template").String()
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
