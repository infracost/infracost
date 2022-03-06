package apiclient

import (
	"regexp"

	"github.com/pkg/errors"

	"github.com/infracost/infracost/internal/clierror"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/ui"
)

// Regex for finding potential URLs and file paths in error messages
// so we can sanitize them. This regex might be too greedy and match
// too many things, but it's a start.
var pathRegex = regexp.MustCompile(`(\w+:)?[\.~\w]*[\/\\]+([^\s:'\"\]]+)|([\w+-]\.\w{2,})`)

func ReportCLIError(ctx *config.RunContext, cliErr error) error {
	errMsg := ui.StripColor(cliErr.Error())
	var sanitizedErr *clierror.SanitizedError

	if errors.As(cliErr, &sanitizedErr) {
		errMsg = ui.StripColor(sanitizedErr.SanitizedError())
	}

	errMsg = pathRegex.ReplaceAllString(errMsg, "REPLACED_PATH")

	d := ctx.EventEnv()
	d["error"] = errMsg

	c := NewPricingAPIClient(ctx)
	return c.AddEvent("infracost-error", d)
}
