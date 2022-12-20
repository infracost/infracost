package apiclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/infracost/infracost/internal/config"
)

// PolicyClient wraps the base http.Client with common handling patterns for the
// Infracost Cloud policy API.
type PolicyClient struct {
	client  *http.Client
	baseURL string
	apiKey  string
	logger  *logrus.Entry
}

// NewPolicyClient returns safely initialised PolicyClient.
func NewPolicyClient(config *config.Config, logger *logrus.Entry) PolicyClient {
	return PolicyClient{
		client:  &http.Client{Timeout: time.Second * 5},
		baseURL: config.PolicyAPIEndpoint,
		apiKey:  config.APIKey,
		logger:  logger,
	}
}

// GetPolicies fetches cost optimization policy from Infracost Cloud.
func (r PolicyClient) GetPolicies(plan []byte) ([]Policy, error) {
	buf := bytes.NewBuffer(plan)

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/policy", r.baseURL), buf)
	if err != nil {
		return nil, fmt.Errorf("failed to build request to policy API %w", err)
	}
	req.Header.Set("X-Api-Key", r.apiKey)

	res, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to policy API %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("received non 200 status code from policy API %s", b)
	}

	var result PolicyDecisionResponse
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode respoinse from policy API %w", err)
	}

	if len(result.Result) == 0 {
		r.logger.Debug("request to policy API returned nil results")
		return nil, nil
	}

	return result.Result, nil
}

type PolicyDecisionResponse struct {
	Result []Policy `json:"result"`
}

type Policy struct {
	ID                 string          `json:"id"`
	Title              string          `json:"title"`
	Description        string          `json:"description"`
	ResourceType       string          `json:"resource_type"`
	ResourceAttributes json.RawMessage `json:"resource_attributes"`
	Address            string          `json:"address"`
	Suggested          string          `json:"suggested"`
	NoCost             bool            `json:"no_cost"`
}
