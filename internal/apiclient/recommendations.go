package apiclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/infracost/infracost/internal/config"
)

type RecommendationClient struct {
	client  *http.Client
	baseURL string
	apiKey  string
	logger  *logrus.Entry
}

func NewRecommendationClient(config *config.Config, logger *logrus.Entry) RecommendationClient {
	return RecommendationClient{
		client:  &http.Client{Timeout: time.Second * 5},
		baseURL: config.RecommendationAPIEndpoint,
		apiKey:  config.APIKey,
		logger:  logger,
	}
}

func (r RecommendationClient) GetSuggestions(plan []byte) ([]Suggestion, error) {
	buf := bytes.NewBuffer(plan)

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/recommend", r.baseURL), buf)
	if err != nil {
		return nil, fmt.Errorf("failed to build request to suggestions API %w", err)
	}

	res, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to suggestions API %w", err)
	}

	var result RecommendDecisionResponse
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode respoinse from suggestions API %w", err)
	}

	if len(result.Result) == 0 {
		r.logger.Debug("request to suggestions API returned nil results")
		return nil, nil
	}

	return result.Result, nil
}

type RecommendDecisionResponse struct {
	Result []Suggestion `json:"result"`
}

type Suggestion struct {
	ID                 string          `json:"id"`
	Title              string          `json:"title"`
	Description        string          `json:"description"`
	ResourceType       string          `json:"resourceType"`
	ResourceAttributes json.RawMessage `json:"resourceAttributes"`
	Address            string          `json:"address"`
	Suggested          string          `json:"suggested"`
	NoCost             bool            `json:"no_cost"`
}
