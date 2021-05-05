package apiclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/infracost/infracost/internal/clierror"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/ui"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func ReportCLIError(cfg *config.Config, cliErr error) error {
	if cfg.IsTelemetryDisabled() {
		log.Debug("Skipping reporting CLI error for self-hosted Infracost")
		return nil
	}

	errMsg := ui.StripColor(cliErr.Error())
	var sanitizedErr *clierror.SanitizedError
	if errors.As(cliErr, &sanitizedErr) {
		errMsg = ui.StripColor(sanitizedErr.SanitizedError())
	}

	url := fmt.Sprintf("%s/cli-error", cfg.DashboardAPIEndpoint)

	j := make(map[string]interface{})
	j["error"] = errMsg
	j["environment"] = cfg.Environment

	body, err := json.Marshal(j)
	if err != nil {
		return errors.Wrap(err, "Error generating request body")
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return errors.Wrap(err, "Error generating request")
	}

	config.AddAuthHeaders(cfg.APIKey, req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "Error sending API request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return &GraphQLError{err, "Invalid API response"}
	}

	return nil
}
