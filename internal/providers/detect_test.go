package providers

import (
	"testing"

	"github.com/infracost/infracost/internal/config"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestDetectAzureRmWhatif(t *testing.T) {
	ctx := config.NewProjectContext(config.EmptyRunContext(), &config.Project{}, log.Fields{})

	ctx.ProjectConfig.Path = "../../examples/azurerm/web_app/what_if.json"

	res, err := Detect(ctx, true)

	if err != nil {
		t.Fatal("Detect threw an error: " + err.Error())
	}

	assert.Equal(t, "azurerm_whatif_json", res.Type())
}
