package apiclient

import (
	"github.com/pkg/errors"

	"github.com/infracost/infracost/internal/clierror"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/ui"
)

func ReportCLIError(ctx *config.RunContext, cliErr error) error {
	errMsg := ui.StripColor(cliErr.Error())
	var sanitizedErr *clierror.SanitizedError
	if errors.As(cliErr, &sanitizedErr) {
		errMsg = ui.StripColor(sanitizedErr.SanitizedError())
	}

	d := ctx.EventEnv()
	d["error"] = errMsg

	c := NewPricingAPIClient(ctx)
	return c.AddEvent("infracost-error", d)
}
