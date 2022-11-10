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

// RecommendationClient wraps the base http.Client with common handling patterns for the
// Infracost Cloud recommendations API.
type RecommendationClient struct {
	client  *http.Client
	baseURL string
	apiKey  string
	logger  *logrus.Entry
}

// NewRecommendationClient returns safely initialised RecommendationClient.
func NewRecommendationClient(config *config.Config, logger *logrus.Entry) RecommendationClient {
	return RecommendationClient{
		client:  &http.Client{Timeout: time.Second * 5},
		baseURL: config.RecommendationAPIEndpoint,
		apiKey:  config.APIKey,
		logger:  logger,
	}
}

// GetRecommendations fetches cost optimization recommendations from Infracost Cloud.
func (r RecommendationClient) GetRecommendations(plan []byte) ([]Recommendation, error) {
	buf := bytes.NewBuffer(plan)

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/recommend", r.baseURL), buf)
	if err != nil {
		return nil, fmt.Errorf("failed to build request to recommendations API %w", err)
	}
	req.Header.Set("X-Api-Key", r.apiKey)

	res, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to recommendations API %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("received non 200 status code from recommendations API %s", b)
	}

	var result RecommendDecisionResponse
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode respoinse from recommendations API %w", err)
	}

	if len(result.Result) == 0 {
		r.logger.Debug("request to recommendations API returned nil results")
		return nil, nil
	}

	return result.Result, nil
}

type RecommendDecisionResponse struct {
	Result []Recommendation `json:"result"`
}

type Recommendation struct {
	ID                 string          `json:"id"`
	Title              string          `json:"title"`
	Description        string          `json:"description"`
	ResourceType       string          `json:"resource_type"`
	ResourceAttributes json.RawMessage `json:"resource_attributes"`
	Address            string          `json:"address"`
	Suggested          string          `json:"suggested"`
	NoCost             bool            `json:"no_cost"`
}
