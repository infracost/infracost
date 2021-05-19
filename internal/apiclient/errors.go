package apiclient

import (
	"github.com/infracost/infracost/internal/clierror"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/ui"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func ReportCLIError(ctx *config.RunContext, cliErr error) error {
	if ctx.Config.IsTelemetryDisabled() {
		log.Debug("Skipping reporting CLI error for self-hosted Infracost")
		return nil
	}

	errMsg := ui.StripColor(cliErr.Error())
	var sanitizedErr *clierror.SanitizedError
	if errors.As(cliErr, &sanitizedErr) {
		errMsg = ui.StripColor(sanitizedErr.SanitizedError())
	}

	d := ctx.EventEnv()
	d["error"] = errMsg

	c := NewDashboardAPIClient(ctx)
	return c.AddEvent("infracost-error", d)
}
