package apiclient

import (
	"github.com/infracost/infracost/internal/clierror"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/ui"
	"github.com/pkg/errors"
)

func ReportCLIError(ctx *config.RunContext, cliErr error) error {
	errMsg := ui.StripColor(cliErr.Error())
	var sanitizedErr *clierror.SanitizedError
	if errors.As(cliErr, &sanitizedErr) {
		errMsg = ui.StripColor(sanitizedErr.SanitizedError())
	}

	d := ctx.Metadata()
	d["error"] = errMsg

	c := NewPricingAPIClient(ctx)
	return c.AddEvent("infracost-error", d)
}
