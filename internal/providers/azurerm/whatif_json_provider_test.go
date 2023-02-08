package azurerm

import (
	"path/filepath"
	"testing"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/usage"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestWhatIfJsonProvider(t *testing.T) {
	ctx := config.NewProjectContext(config.EmptyRunContext(), &config.Project{}, log.Fields{})

	ctx.ProjectConfig.Path = filepath.Join("testdata", "what_if.json")

	provider := NewWhatifJsonProvider(ctx, true)
	usage := usage.NewBlankUsageFile().ToUsageDataMap()

	project, err := provider.LoadResources(usage)
	if err != nil {
		t.Fatalf("Error loading resources: " + err.Error())
	}

	// Ensure all resources in the whatif are returned from the provider
	assert.Equal(t, 3, len(project[0].PartialResources))
	assert.Equal(t, 0, len(project[0].PartialPastResources))
}
