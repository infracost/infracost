package apiclient

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/hashicorp/go-retryablehttp"
	json "github.com/json-iterator/go"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/output"
	"github.com/infracost/infracost/internal/schema"
)

type PolicyV2APIClient struct {
	APIClient
	allowList resourceAllowList
}

// NewPolicyV2APIClient retrieves resource allow-list info from Infracost Cloud and returns a new policy client
func NewPolicyV2APIClient(ctx *config.RunContext) (*PolicyV2APIClient, error) {
	client := retryablehttp.NewClient()
	client.Logger = &LeveledLogger{Logger: logging.Logger.WithField("library", "retryablehttp")}
	c := PolicyV2APIClient{
		APIClient: APIClient{
			httpClient: client.StandardClient(),
			endpoint:   ctx.Config.PolicyV2APIEndpoint,
			apiKey:     ctx.Config.APIKey,
			uuid:       ctx.UUID(),
		},
	}

	prw, err := c.getPolicyResourceAllowList()
	if err != nil {
		return nil, err
	}
	c.allowList = prw

	return &c, nil
}

type PolicyOutput struct {
	TagPolicies    []output.TagPolicy
	FinOpsPolicies []output.FinOpsPolicy
}

func (c *PolicyV2APIClient) CheckPolicies(ctx *config.RunContext, out output.Root) (*PolicyOutput, error) {
	ri, err := newRunInput(ctx, out)
	if err != nil {
		return nil, err
	}

	q := `
		query($run: RunInput!) {
			evaluatePolicies(run: $run) {
				tagPolicyResults {
					name
					tagPolicyId
					message
					prComment
					blockPr
					totalDetectedResources
					totalTaggableResources
					resources {
						address
						resourceType
						path
						line
						projectNames
						missingMandatoryTags
						invalidTags {
							key
							value
							validValues
							validRegex
						}
					}
				}
				finopsPolicyResults {
					name
					policyId
					message
					blockPr
					prComment
					totalApplicableResources
					resources {
						checksum
						address
						resourceType
						path
						startLine
						endLine
						projectNames
						issues {
							attribute
							value
							description
						}
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

	if len(results) == 0 {
		return nil, nil
	}

	if results[0].Get("errors").Exists() {
		return nil, fmt.Errorf("query failed when checking tag policies, received graphql error: %s", results[0].Get("errors").String())
	}

	data := results[0].Get("data")

	var policies = struct {
		EvaluatePolicies struct {
			TagPolicies    []output.TagPolicy    `json:"tagPolicyResults"`
			FinOpsPolicies []output.FinOpsPolicy `json:"finopsPolicyResults"`
		} `json:"evaluatePolicies"`
	}{}

	err = json.Unmarshal([]byte(data.Raw), &policies)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal tag policies %w", err)
	}

	if len(policies.EvaluatePolicies.TagPolicies) > 0 {
		checkedStr := "tag policy"
		if len(policies.EvaluatePolicies.TagPolicies) > 1 {
			checkedStr = "tag policies"
		}
		msg := fmt.Sprintf(`%d %s checked`, len(policies.EvaluatePolicies.TagPolicies), checkedStr)
		if ctx.Config.IsLogging() {
			logging.Logger.Info(msg)
		} else {
			fmt.Fprintf(ctx.ErrWriter, "%s\n", msg)
		}
	}

	if len(policies.EvaluatePolicies.FinOpsPolicies) > 0 {
		checkedStr := "finops policy"
		if len(policies.EvaluatePolicies.FinOpsPolicies) > 1 {
			checkedStr = "finops policies"
		}
		msg := fmt.Sprintf(`%d %s checked`, len(policies.EvaluatePolicies.FinOpsPolicies), checkedStr)
		if ctx.Config.IsLogging() {
			logging.Logger.Info(msg)
		} else {
			fmt.Fprintf(ctx.ErrWriter, "%s\n", msg)
		}
	}

	return &PolicyOutput{policies.EvaluatePolicies.TagPolicies, policies.EvaluatePolicies.FinOpsPolicies}, nil
}

// UploadPolicyData sends a filtered set of a project's resource information to Infracost Cloud and
// potentially adds PolicySha and PastPolicySha to the project's metadata.
func (c *PolicyV2APIClient) UploadPolicyData(project *schema.Project) error {
	if project.Metadata == nil {
		project.Metadata = &schema.ProjectMetadata{}
	}

	filteredResources := c.filterResources(project.PartialResources)
	if len(filteredResources) > 0 {
		sha, err := c.uploadProjectPolicyData(filteredResources)
		if err != nil {
			return fmt.Errorf("failed to upload filtered partial resources %w", err)
		}
		project.Metadata.PolicySha = sha
	}

	filteredPastResources := c.filterResources(project.PartialPastResources)
	if len(filteredPastResources) > 0 {
		sha, err := c.uploadProjectPolicyData(filteredPastResources)
		if err != nil {
			return fmt.Errorf("failed to upload filtered past partial resources %w", err)
		}
		project.Metadata.PastPolicySha = sha
	}

	return nil
}

func (c *PolicyV2APIClient) uploadProjectPolicyData(p2rs []policy2Resource) (string, error) {
	q := `
	mutation($policyResources: [PolicyResourceInput!]!) {
		storePolicyResources(policyResources: $policyResources) {
			sha
		}
	}
	`

	v := map[string]interface{}{
		"policyResources": p2rs,
	}

	results, err := c.doQueries([]GraphQLQuery{{q, v}})
	if err != nil {
		return "", fmt.Errorf("query storePolicyResources failed  %w", err)
	}

	if len(results) == 0 {
		return "", nil
	}

	if results[0].Get("errors").Exists() {
		return "", fmt.Errorf("query storePolicyResources failed, received graphql error: %s", results[0].Get("errors").String())
	}

	data := results[0].Get("data")

	var response struct {
		AddFilteredResourceSet struct {
			Sha string `json:"sha"`
		} `json:"storePolicyResources"`
	}

	err = json.Unmarshal([]byte(data.Raw), &response)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal storePolicyResources %w", err)
	}

	return response.AddFilteredResourceSet.Sha, nil
}

// graphql doesn't really like map/dictionary parameters, so convert tags,
// values, and refs to key/value arrays.

type policy2Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type policy2Value struct {
	Key   string          `json:"key"`
	Value json.RawMessage `json:"value"`
}

type policy2Reference struct {
	Key       string   `json:"key"`
	Addresses []string `json:"addresses"`
}

type policy2Resource struct {
	ResourceType string                   `json:"resourceType"`
	ProviderName string                   `json:"providerName"`
	Address      string                   `json:"address"`
	Tags         *[]policy2Tag            `json:"tags,omitempty"`
	Values       []policy2Value           `json:"values"`
	References   []policy2Reference       `json:"references"`
	Metadata     policy2InfracostMetadata `json:"infracostMetadata"`
}

type policy2InfracostMetadata struct {
	Calls     []policy2InfracostMetadataCall `json:"calls"`
	Checksum  string                         `json:"checksum"`
	EndLine   int64                          `json:"endLine"`
	Filename  string                         `json:"filename"`
	StartLine int64                          `json:"startLine"`
}

type policy2InfracostMetadataCall struct {
	Filename  string `json:"filename"`
	BlockName string `json:"blockName"`
	StartLine int64  `json:"startLine"`
	EndLine   int64  `json:"endLine"`
}

func (c *PolicyV2APIClient) filterResources(partials []*schema.PartialResource) []policy2Resource {
	var p2rs []policy2Resource
	for _, partial := range partials {
		if partial != nil && partial.ResourceData != nil {
			rd := partial.ResourceData
			if f, ok := c.allowList[rd.Type]; ok {
				p2rs = append(p2rs, filterResource(rd, f))
			}
		}
	}

	sort.Slice(p2rs, func(i, j int) bool {
		return p2rs[i].Address < p2rs[j].Address
	})

	return p2rs
}

func filterResource(rd *schema.ResourceData, allowedKeys map[string]bool) policy2Resource {
	var tagsPtr *[]policy2Tag
	if rd.Tags != nil {
		tags := make([]policy2Tag, 0, len(*rd.Tags))
		for k, v := range *rd.Tags {
			tags = append(tags, policy2Tag{Key: k, Value: v})
		}
		sort.Slice(tags, func(i, j int) bool {
			return tags[i].Key < tags[j].Key
		})

		tagsPtr = &tags
	}

	values := make([]policy2Value, 0, len(allowedKeys))
	for k, v := range rd.RawValues.Map() {
		if b, ok := allowedKeys[k]; ok && b {
			values = append(values, policy2Value{Key: k, Value: []byte(v.Raw)})
		}
	}
	sort.Slice(values, func(i, j int) bool {
		return values[i].Key < values[j].Key
	})

	references := make([]policy2Reference, 0, len(rd.ReferencesMap))
	for k, refRds := range rd.ReferencesMap {
		refAddresses := make([]string, 0, len(refRds))
		for _, refRd := range refRds {
			refAddresses = append(refAddresses, refRd.Address)
		}
		references = append(references, policy2Reference{Key: k, Addresses: refAddresses})
	}
	sort.Slice(references, func(i, j int) bool {
		return references[i].Key < references[j].Key
	})

	var mdCalls []policy2InfracostMetadataCall
	for _, c := range rd.Metadata["calls"].Array() {
		mdCalls = append(mdCalls, policy2InfracostMetadataCall{
			BlockName: c.Get("blockName").String(),
			EndLine:   rd.Metadata["endLine"].Int(),
			Filename:  rd.Metadata["filename"].String(),
			StartLine: rd.Metadata["startLine"].Int(),
		})
	}

	checksum := rd.Metadata["checksum"].String()

	if checksum == "" {
		// this must be a plan json run.  calculate a checksum now.
		checksum = calcChecksum(rd)
	}

	return policy2Resource{
		ResourceType: rd.Type,
		ProviderName: rd.ProviderName,
		Address:      rd.Address,
		Tags:         tagsPtr,
		Values:       values,
		References:   references,
		Metadata: policy2InfracostMetadata{
			Calls:     mdCalls,
			Checksum:  checksum,
			EndLine:   rd.Metadata["endLine"].Int(),
			Filename:  rd.Metadata["filename"].String(),
			StartLine: rd.Metadata["startLine"].Int(),
		},
	}
}

func calcChecksum(rd *schema.ResourceData) string {
	h := sha256.New()
	h.Write([]byte(rd.ProviderName))
	h.Write([]byte(rd.Address))
	h.Write([]byte(rd.RawValues.Raw))

	return hex.EncodeToString(h.Sum(nil))
}

type resourceAllowList map[string]map[string]bool

func (c *PolicyV2APIClient) getPolicyResourceAllowList() (resourceAllowList, error) {
	q := `
		query {
			policyResourceAllowList {
				resourceType
                allowedKeys
			}
		}
	`
	v := map[string]interface{}{}

	results, err := c.doQueries([]GraphQLQuery{{q, v}})
	if err != nil {
		return nil, fmt.Errorf("query policyResourceAllowList failed %w", err)
	}

	if len(results) == 0 {
		return nil, nil
	}

	if results[0].Get("errors").Exists() {
		return nil, fmt.Errorf("query policyResourceAllowList failed, received graphql error: %s", results[0].Get("errors").String())
	}

	data := results[0].Get("data")

	var response struct {
		AllowLists []struct {
			Type   string   `json:"resourceType"`
			Values []string `json:"allowedKeys"`
		} `json:"policyResourceAllowList"`
	}

	err = json.Unmarshal([]byte(data.Raw), &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal policyResourceAllowList %w", err)
	}

	aw := resourceAllowList{}

	for _, rtf := range response.AllowLists {
		rtfMap := map[string]bool{}
		for _, v := range rtf.Values {
			rtfMap[v] = true
		}
		aw[rtf.Type] = rtfMap
	}

	return aw, nil
}
