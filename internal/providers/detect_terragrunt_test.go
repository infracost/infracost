package providers

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
)

// A Terragrunt unit whose config produces a decode diagnostic (here an exclude block, unknown to the
// vendored Terragrunt parser) and has no inputs attribute must not crash the Terragrunt HCL provider.
func TestTerragruntUnitWithDiagnosticAndNoInputs(t *testing.T) {
	ctx := config.EmptyRunContext()
	project := &config.Project{Path: filepath.Join("testdata", "terragrunt_exclude_no_inputs")}

	out, err := Detect(ctx, project, false)
	require.NoError(t, err)
	require.NotEmpty(t, out.Providers)

	resources := map[string]int{}
	for _, provider := range out.Providers {
		projects, err := provider.LoadResources(schema.UsageMap{})
		if !assert.NoError(t, err, "provider %s", provider.RelativePath()) {
			continue
		}

		for _, p := range projects {
			for _, diag := range p.Metadata.Errors {
				assert.Failf(t, "unexpected project error", "project %s: %s", p.Name, diag.Message)
			}
			resources[filepath.Base(p.Metadata.TerraformModulePath)] += len(p.PartialResources)
		}
	}

	assert.Equal(t, map[string]int{"unit": 1, "unit_with_inputs": 1}, resources,
		"expected the google_compute_address of each unit to be loaded")
}
