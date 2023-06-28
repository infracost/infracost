package apiclient

import (
	"encoding/json"
	"fmt"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/output"
	log "github.com/sirupsen/logrus"
)

type TagPolicyAPIClient struct {
	APIClient
}

func NewTagPolicyAPIClient(ctx *config.RunContext) *TagPolicyAPIClient {
	return &TagPolicyAPIClient{
		APIClient: APIClient{
			endpoint: ctx.Config.TagPolicyAPIEndpoint,
			apiKey:   ctx.Config.APIKey,
			uuid:     ctx.UUID(),
		},
	}
}

type TagPolicyQuery struct {
	TagPolicies []output.TagPolicy `json:"tagPolicies"`
}

func (c *TagPolicyAPIClient) CheckTagPolicies(ctx *config.RunContext, out output.Root) ([]output.TagPolicy, error) {
	ri, err := newRunInput(ctx, out)
	if err != nil {
		return nil, err
	}

	q := `
		query($run: RunInput!) {
			evaluateTagPolicies(run: $run) {
				name
				tagPolicyId
				message
				prComment
				blockPr
				resources {
					address
					resourceType
					path
					line
					projectNames
					missingMandatoryTags
					invalidTags {
						key
						validValues
						validRegex
					}
				}
			}
		}
	`

	v := map[string]interface{}{
		"run": *ri,
	}
	results, err := c.doQueries([]GraphQLQuery{{q, v}})
	if err != nil {
		return nil, fmt.Errorf("query failed when checking tag policies %w", err)
	}

	if len(results) > 0 {
		if results[0].Get("errors").Exists() {
			return nil, fmt.Errorf("query failed when checking tag policies, received graphql error: %s", results[0].Get("errors").String())
		}
	}

	data := results[0].Get("data")

	var tagPolicies = struct {
		TagPolicies []output.TagPolicy `json:"evaluateTagPolicies"`
	}{}

	err = json.Unmarshal([]byte(data.Raw), &tagPolicies)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal tag policies %w", err)
	}

	if len(tagPolicies.TagPolicies) > 0 {
		checkedStr := "tag policy"
		if len(tagPolicies.TagPolicies) > 1 {
			checkedStr = "tag policies"
		}
		tagPolicyMsg := fmt.Sprintf(`%d %s checked`, len(tagPolicies.TagPolicies), checkedStr)
		if ctx.Config.IsLogging() {
			log.Info(tagPolicyMsg)
		} else {
			fmt.Fprintf(ctx.ErrWriter, "%s\n", tagPolicyMsg)
		}
	}

	return tagPolicies.TagPolicies, nil
}
