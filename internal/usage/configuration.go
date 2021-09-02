package usage

import (
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage/aws"
)

// ConfigureEstimation caches cloud-vendor SDK configuration in the project
// context which can be used during sync-usage to query cloud APIs.
func ConfigureEstimation(ctx *config.ProjectContext, p *schema.Project) error {
	// TODO: store in context a map of Terraform provider nickname to AWS profile (or other options)
	//   - needed for multi-profile support in complex monorepos
	return aws.ConfigureEstimation(ctx)
}
