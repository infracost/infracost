package events

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/infracost/infracost/internal/config"
	log "github.com/sirupsen/logrus"
)

func SendReport(cfg *config.Config, key string, data interface{}) {
	if cfg.PricingAPIEndpoint != cfg.DefaultPricingAPIEndpoint && config.IsFalsy(os.Getenv("INFRACOST_SELF_HOSTED_TELEMETRY")) {
		return
	}

	url := fmt.Sprintf("%s/report", cfg.DefaultPricingAPIEndpoint)

	j := make(map[string]interface{})
	j[key] = data
	j["environment"] = cfg.Environment

	body, err := json.Marshal(j)
	if err != nil {
		log.Debugf("Unable to generate event: %v", err)
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Debugf("Unable to generate event: %v", err)
		return
	}

	config.AddAuthHeaders(cfg.APIKey, req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Debugf("Unable to send event: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Debugf("Unexpected response sending event: %d", resp.StatusCode)
	}
}
