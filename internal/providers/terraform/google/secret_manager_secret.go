package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"

	"github.com/tidwall/gjson"
)

func getSecretManagerSecretRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_secret_manager_secret",
		CoreRFunc: newSecretManagerSecret,
	}
}

func newSecretManagerSecret(d *schema.ResourceData) schema.CoreResource {
	return &google.SecretManagerSecret{
		Address:              d.Address,
		Region:               d.Get("region").String(),
		ReplicationLocations: secretManagerSecretReplicasCount(d),
	}
}

func secretManagerSecretReplicasCount(d *schema.ResourceData) int64 {
	replicasCount := 1

	replications := d.Get("replication.0.user_managed.0")
	if replications.Type != gjson.Null && len(replications.Array()) > 0 {
		replicasCount = len(replications.Get("replicas").Array())
	}

	return int64(replicasCount)
}
