package apiclient

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/hashicorp/go-retryablehttp"
	json "github.com/json-iterator/go"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/schema"
)

var jsonSorted = json.Config{SortMapKeys: true}.Froze()

type PolicyAPIClient struct {
	APIClient

	allowLists    map[string]allowList
	allowListErr  error
	allowListOnce sync.Once
}

// NewPolicyAPIClient retrieves resource allow-list info from Infracost Cloud and returns a new policy client
func NewPolicyAPIClient(ctx *config.RunContext) (*PolicyAPIClient, error) {
	client := retryablehttp.NewClient()
	client.Logger = &LeveledLogger{Logger: logging.Logger.With().Str("library", "retryablehttp").Logger()}
	c := PolicyAPIClient{
		APIClient: APIClient{
			httpClient: client.StandardClient(),
			endpoint:   ctx.Config.PolicyV2APIEndpoint,
			apiKey:     ctx.Config.APIKey,
			uuid:       ctx.UUID(),
		},
	}

	return &c, nil
}

// UploadPolicyData sends a filtered set of a project's resource information to Infracost Cloud and
// potentially adds PolicySha and PastPolicySha to the project's metadata.
func (c *PolicyAPIClient) UploadPolicyData(project *schema.Project, rds, pastRds []*schema.ResourceData) error {
	if project.Metadata == nil {
		project.Metadata = &schema.ProjectMetadata{}
	}

	// remove .git suffix from the module path
	project.Metadata.TerraformModulePath = strings.ReplaceAll(project.Metadata.TerraformModulePath, ".git/", "/")

	err := c.fetchAllowList()
	if err != nil {
		return err
	}

	filteredResources := c.filterResources(rds)
	if len(filteredResources) > 0 {
		sha, err := c.uploadProjectPolicyData(filteredResources)
		if err != nil {
			return fmt.Errorf("failed to upload filtered partial resources %w", err)
		}
		project.Metadata.PolicySha = sha
	} else {
		project.Metadata.PolicySha = "0" // set a fake sha so we can tell policy checks actually ran
	}

	filteredPastResources := c.filterResources(pastRds)
	if len(filteredPastResources) > 0 {
		sha, err := c.uploadProjectPolicyData(filteredPastResources)
		if err != nil {
			return fmt.Errorf("failed to upload filtered past partial resources %w", err)
		}
		project.Metadata.PastPolicySha = sha
	} else {
		project.Metadata.PastPolicySha = "0" // set a fake sha so we can tell policy checks actually ran
	}

	return nil
}

func (c *PolicyAPIClient) uploadProjectPolicyData(p2rs []policy2Resource) (string, error) {
	q := `
	mutation($policyResources: [PolicyResourceInput!]!) {
		storePolicyResources(policyResources: $policyResources) {
			sha
		}
	}
	`

	v := map[string]any{
		"policyResources": p2rs,
	}

	results, err := c.DoQueries([]GraphQLQuery{{q, v}})
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

type policy2Reference struct {
	Key       string   `json:"key"`
	Addresses []string `json:"addresses"`
}

type policy2Resource struct {
	ResourceType                            string                   `json:"resourceType"`
	ProviderName                            string                   `json:"providerName"`
	Address                                 string                   `json:"address"`
	Tags                                    *[]policy2Tag            `json:"tags,omitempty"`
	DefaultTags                             *[]policy2Tag            `json:"defaultTags,omitempty"`
	SupportForDefaultTags                   bool                     `json:"supportForDefaultTags"`
	TagPropagation                          *TagPropagation          `json:"tagPropagation,omitempty"`
	Values                                  json.RawMessage          `json:"values"`
	References                              []policy2Reference       `json:"references"`
	Metadata                                policy2InfracostMetadata `json:"infracostMetadata"`
	Region                                  string                   `json:"region"`
	MissingVarsCausingUnknownTagKeys        []string                 `json:"missingVarsCausingUnknownTagKeys,omitempty"`
	MissingVarsCausingUnknownDefaultTagKeys []string                 `json:"missingVarsCausingUnknownDefaultTagKeys,omitempty"`
}

type TagPropagation struct {
	To                    string        `json:"to"`
	From                  *string       `json:"from"`
	Tags                  *[]policy2Tag `json:"tags"`
	Attribute             string        `json:"attribute"`
	HasRequiredAttributes bool          `json:"hasRequiredAttributes"`
}

type policy2InfracostMetadata struct {
	Calls          []policy2InfracostMetadataCall `json:"calls"`
	Checksum       string                         `json:"checksum"`
	EndLine        int64                          `json:"endLine"`
	Filename       string                         `json:"filename"`
	StartLine      int64                          `json:"startLine"`
	ModuleFilename string                         `json:"moduleFilename,omitempty"`
}

type policy2InfracostMetadataCall struct {
	Filename  string `json:"filename"`
	BlockName string `json:"blockName"`
	StartLine int64  `json:"startLine"`
	EndLine   int64  `json:"endLine"`
}

func (c *PolicyAPIClient) filterResources(rds []*schema.ResourceData) []policy2Resource {
	var p2rs []policy2Resource
	for _, rd := range rds {
		if f, ok := c.allowLists[rd.Type]; ok {
			p2rs = append(p2rs, filterResource(rd, f))
		}
	}

	sort.Slice(p2rs, func(i, j int) bool {
		return p2rs[i].Address < p2rs[j].Address
	})

	return p2rs
}

func filterResource(rd *schema.ResourceData, al allowList) policy2Resource {
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

	var defaultTagsPtr *[]policy2Tag
	if rd.DefaultTags != nil {
		tags := make([]policy2Tag, 0, len(*rd.DefaultTags))
		for k, v := range *rd.DefaultTags {
			tags = append(tags, policy2Tag{Key: k, Value: v})
		}
		sort.Slice(tags, func(i, j int) bool {
			return tags[i].Key < tags[j].Key
		})

		defaultTagsPtr = &tags
	}

	var propagatedTagsPtr *[]policy2Tag
	if rd.TagPropagation != nil && rd.TagPropagation.Tags != nil {
		tags := make([]policy2Tag, 0, len(*rd.Tags))
		for k, v := range *rd.TagPropagation.Tags {
			tags = append(tags, policy2Tag{Key: k, Value: v})
		}
		sort.Slice(tags, func(i, j int) bool {
			return tags[i].Key < tags[j].Key
		})
		propagatedTagsPtr = &tags
	}

	// make sure the keys in the values json are sorted so we get consistent policyShas
	valuesJSON, err := jsonSorted.Marshal(filterValues(rd.RawValues, al))
	if err != nil {
		logging.Logger.Debug().Err(err).Str("address", rd.Address).Msg("Failed to marshal filtered values")
	}

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

	calls := rd.Metadata["calls"].Array()
	mdCalls := make([]policy2InfracostMetadataCall, 0, len(calls))
	for _, c := range calls {
		mdCalls = append(mdCalls, policy2InfracostMetadataCall{
			BlockName: c.Get("blockName").String(),
			EndLine:   c.Get("endLine").Int(),
			Filename:  c.Get("filename").String(),
			StartLine: c.Get("startLine").Int(),
		})
	}

	checksum := rd.Metadata["checksum"].String()

	if checksum == "" {
		// this must be a plan json run.  calculate a checksum now.
		checksum = calcChecksum(rd)
	}

	var tagPropagation *TagPropagation
	if rd.TagPropagation != nil {
		tagPropagation = &TagPropagation{
			To:                    rd.TagPropagation.To,
			From:                  rd.TagPropagation.From,
			Tags:                  propagatedTagsPtr,
			Attribute:             rd.TagPropagation.Attribute,
			HasRequiredAttributes: rd.TagPropagation.HasRequiredAttributes,
		}
	}

	metadata := policy2InfracostMetadata{
		Calls:          mdCalls,
		Checksum:       checksum,
		EndLine:        rd.Metadata["endLine"].Int(),
		Filename:       rd.Metadata["filename"].String(),
		StartLine:      rd.Metadata["startLine"].Int(),
		ModuleFilename: rd.Metadata["moduleFilename"].String(),
	}

	return policy2Resource{
		ResourceType:                            rd.Type,
		ProviderName:                            rd.ProviderName,
		Address:                                 rd.Address,
		Tags:                                    tagsPtr,
		DefaultTags:                             defaultTagsPtr,
		TagPropagation:                          tagPropagation,
		Values:                                  valuesJSON,
		References:                              references,
		Metadata:                                metadata,
		Region:                                  rd.Region,
		MissingVarsCausingUnknownTagKeys:        rd.MissingVarsCausingUnknownTagKeys,
		MissingVarsCausingUnknownDefaultTagKeys: rd.MissingVarsCausingUnknownDefaultTagKeys,
	}
}

func filterValues(rd gjson.Result, allowList map[string]gjson.Result) map[string]any {
	values := make(map[string]any, len(allowList))
	for k, v := range rd.Map() {
		if allow, ok := allowList[k]; ok {
			if allow.IsBool() {
				if allow.Bool() {
					values[k] = json.RawMessage(v.Raw)
				}
			} else if allow.IsObject() {
				nestedAllow := allow.Map()
				if v.IsArray() {
					vArray := v.Array()
					nestedVals := make([]any, 0, len(vArray))
					for _, el := range vArray {
						nestedVals = append(nestedVals, filterValues(el, nestedAllow))
					}
					values[k] = nestedVals
				} else {
					values[k] = filterValues(v, nestedAllow)
				}
			} else {
				logging.Logger.Debug().Str("Key", k).Str("Type", allow.Type.String()).Msg("Unknown allow type")
			}
		}
	}
	return values
}

func calcChecksum(rd *schema.ResourceData) string {
	h := sha256.New()
	h.Write([]byte(rd.ProviderName))
	h.Write([]byte(rd.Address))
	h.Write([]byte(rd.RawValues.Raw))

	return hex.EncodeToString(h.Sum(nil))
}

type allowList map[string]gjson.Result

func (c *PolicyAPIClient) fetchAllowList() error {
	c.allowListOnce.Do(func() {
		prw, err := c.getPolicyResourceAllowList()
		if err != nil {
			c.allowListErr = err
		}
		c.allowLists = prw
	})

	return c.allowListErr
}

func (c *PolicyAPIClient) getPolicyResourceAllowList() (map[string]allowList, error) {
	q := `
		query GetPolicyResourceAllowList{
			policyResourceAllowList {
				resourceType
                allowed
			}
		}
	`
	v := map[string]any{}

	results, err := c.DoQueries([]GraphQLQuery{{q, v}})
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
			Type    string          `json:"resourceType"`
			Allowed json.RawMessage `json:"allowed"`
		} `json:"policyResourceAllowList"`
	}

	err = json.Unmarshal([]byte(data.Raw), &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal policyResourceAllowList %w", err)
	}

	aw := map[string]allowList{}

	for _, rtf := range response.AllowLists {
		aw[rtf.Type] = gjson.ParseBytes(rtf.Allowed).Map()
	}

	return aw, nil
}
