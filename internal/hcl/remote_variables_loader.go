package hcl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/shurcooL/graphql"
	"github.com/spacelift-io/spacectl/client"
	spaceliftSession "github.com/spacelift-io/spacectl/client/session"
	"github.com/spacelift-io/spacectl/client/structs"
	"github.com/zclconf/go-cty/cty"
	ctyJson "github.com/zclconf/go-cty/cty/json"

	"github.com/infracost/infracost/internal/extclient"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/vcs"
)

// RemoteVariableLoader is an interface for loading remote variables from a remote service.
type RemoteVariableLoader interface {
	// Load fetches remote variables from a remote service.
	Load(options RemoteVarLoaderOptions) (map[string]cty.Value, error)
}

// TFCRemoteVariablesLoader handles loading remote variables from Terraform Cloud.
type TFCRemoteVariablesLoader struct {
	client         *extclient.AuthedAPIClient
	localWorkspace string
	remoteConfig   *TFCRemoteConfig
	logger         zerolog.Logger
}

// TFCRemoteVariablesLoaderOption defines a function that can set properties on an TFCRemoteVariablesLoader.
type TFCRemoteVariablesLoaderOption func(r *TFCRemoteVariablesLoader)

type tfcWorkspaceResponse struct {
	Data struct {
		ID         string `json:"id"`
		Attributes struct {
			ExecutionMode string `json:"execution-mode"`
		} `json:"attributes"`
	} `json:"data"`
}

type tfcVarset struct {
	Attributes struct {
		Name   string `json:"name"`
		Global bool   `json:"global"`
	} `json:"attributes"`
	Relationships struct {
		Vars struct {
			Data []struct {
				ID string
			} `json:"data"`
		} `json:"vars"`
	} `json:"relationships"`
}

type tfcVar struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	Sensitive bool   `json:"sensitive"`
	Category  string `json:"category"`
	HCL       bool   `json:"hcl"`
}

type tfcVarsetResponse struct {
	Data     []tfcVarset `json:"data"`
	Included []struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes tfcVar `json:"attributes"`
	} `json:"included"`
	Meta struct {
		Pagination struct {
			NextPage int `json:"next-page"`
		} `json:"pagination"`
	} `json:"meta"`
}

type tfcVarResponse struct {
	Data []struct {
		Attributes tfcVar `json:"attributes"`
	} `json:"data"`
}

// RemoteVariablesLoaderWithRemoteConfig sets a user defined configuration for
// the TFCRemoteVariablesLoader. This is normally done to override the configuration
// detected from the HCL blocks.
func RemoteVariablesLoaderWithRemoteConfig(config TFCRemoteConfig) TFCRemoteVariablesLoaderOption {
	return func(r *TFCRemoteVariablesLoader) {
		r.remoteConfig = &config
	}
}

// NewTFCRemoteVariablesLoader constructs a new loader for fetching remote variables.
func NewTFCRemoteVariablesLoader(client *extclient.AuthedAPIClient, localWorkspace string, logger zerolog.Logger, opts ...TFCRemoteVariablesLoaderOption) *TFCRemoteVariablesLoader {
	if localWorkspace == "" {
		localWorkspace = os.Getenv("TF_WORKSPACE")
	}

	r := &TFCRemoteVariablesLoader{
		client:         client,
		localWorkspace: localWorkspace,
		logger:         logger,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

type RemoteVarLoaderOptions struct {
	Blocks      Blocks
	ModulePath  string
	Environment string
}

// Load fetches remote variables if terraform block contains organization and
// workspace name.
func (r *TFCRemoteVariablesLoader) Load(options RemoteVarLoaderOptions) (map[string]cty.Value, error) {
	blocks := options.Blocks

	r.logger.Debug().Msg("Downloading Terraform remote variables")
	vars := map[string]cty.Value{}

	var config TFCRemoteConfig
	if r.remoteConfig != nil {
		config = *r.remoteConfig
	} else {
		var err error
		config = r.getCloudOrganizationWorkspace(blocks)
		if !config.valid() {
			config, err = r.getBackendOrganizationWorkspace(blocks)
			if err != nil {
				r.logger.Warn().Err(err).Msg("could not detect Terraform Cloud organization and workspace")
				return vars, nil
			}

			if !config.valid() {
				return vars, nil
			}
		}
	}

	if config.Host != "" {
		r.client.SetHost(config.Host)
	}

	endpoint := fmt.Sprintf("/api/v2/organizations/%s/workspaces/%s", config.Organization, config.Workspace)
	body, err := r.client.Get(endpoint)
	if err != nil {
		r.logger.Debug().Err(err).Msgf("could not request Terraform workspace: %s for organization: %s", config.Workspace, config.Organization)
		return vars, nil
	}

	var workspaceResponse tfcWorkspaceResponse
	if json.Unmarshal(body, &workspaceResponse) != nil {
		r.logger.Debug().Err(err).Msgf("malformed Terraform API response using workspace: %s organization: %s", config.Workspace, config.Organization)
		return vars, nil
	}

	if workspaceResponse.Data.Attributes.ExecutionMode == "local" {
		r.logger.Debug().Msgf("Terraform workspace %s does use local execution, skipping downloading remote variables", config.Workspace)
		return vars, nil
	}

	workspaceID := workspaceResponse.Data.ID

	pageNumber := 1

	varsets := []tfcVarset{}
	varsMap := map[string]tfcVar{}

	for i := 0; i < 10; i++ {
		endpoint = fmt.Sprintf("/api/v2/workspaces/%s/varsets?include=vars&page[number]=%d&page[size]=50", workspaceID, pageNumber)
		body, err = r.client.Get(endpoint)
		if err != nil {
			return vars, err
		}

		var varsetsResponse tfcVarsetResponse
		if json.Unmarshal(body, &varsetsResponse) != nil {
			return vars, errors.New("unable to parse Workspace Variable Sets response")
		}

		varsets = append(varsets, varsetsResponse.Data...)
		varsetVars := varsetsResponse.Included

		for _, v := range varsetVars {
			if v.Type == "vars" {
				varsMap[v.ID] = v.Attributes
			}
		}

		if varsetsResponse.Meta.Pagination.NextPage > pageNumber {
			pageNumber = varsetsResponse.Meta.Pagination.NextPage
		} else {
			break
		}
	}

	// Sort varsets alphabetically, global varsets are lower in priority and can
	// be overridden by workspace's varsets.
	sort.Slice(varsets, func(i, j int) bool {
		if varsets[i].Attributes.Global && !varsets[j].Attributes.Global {
			return true
		}
		if !varsets[i].Attributes.Global && varsets[j].Attributes.Global {
			return false
		}

		return varsets[i].Attributes.Name > varsets[j].Attributes.Name
	})

	for _, varset := range varsets {
		for _, v := range varset.Relationships.Vars.Data {
			vv, ok := varsMap[v.ID]
			if ok {
				val := r.getVarValue(vv)
				if !val.IsNull() && val.IsKnown() {
					k := r.getVarKey(vv)
					r.logger.Debug().Msgf("adding variable %s from varset %s", k, varset.Attributes.Name)
					vars[k] = val
				}
			}
		}
	}

	endpoint = fmt.Sprintf("/api/v2/workspaces/%s/vars", workspaceID)
	body, err = r.client.Get(endpoint)
	if err != nil {
		return vars, err
	}

	var varsResponse tfcVarResponse
	if json.Unmarshal(body, &varsResponse) != nil {
		return vars, errors.New("unable to parse Workspace Variables response")
	}

	for _, v := range varsResponse.Data {
		val := r.getVarValue(v.Attributes)
		if !val.IsNull() {
			k := r.getVarKey(v.Attributes)
			r.logger.Debug().Msgf("adding variable %s from workspace", k)
			vars[k] = val
		}
	}

	return vars, nil
}

type TFCRemoteConfig struct {
	Organization string
	Workspace    string
	Host         string
}

func (c TFCRemoteConfig) valid() bool {
	return c.Organization != "" && c.Workspace != ""
}

func (r *TFCRemoteVariablesLoader) getCloudOrganizationWorkspace(blocks Blocks) TFCRemoteConfig {
	var conf TFCRemoteConfig

	for _, block := range blocks.OfType("terraform") {
		for _, c := range block.childBlocks.OfType("cloud") {
			conf.Organization = getAttribute(c, "organization")
			conf.Host = getAttribute(c, "hostname")

			for _, cc := range c.childBlocks.OfType("workspaces") {
				conf.Workspace = getAttribute(cc, "name")
				return conf
			}
		}
	}

	return conf
}

func (r *TFCRemoteVariablesLoader) getBackendOrganizationWorkspace(blocks Blocks) (TFCRemoteConfig, error) {
	var conf TFCRemoteConfig

	for _, block := range blocks.OfType("terraform") {
		for _, c := range block.childBlocks.OfType("backend") {
			if c.Label() != "remote" {
				continue
			}

			conf.Organization = getAttribute(c, "organization")
			conf.Host = getAttribute(c, "hostname")

			for _, cc := range c.childBlocks.OfType("workspaces") {
				name := getAttribute(cc, "name")

				if name != "" {
					conf.Workspace = name
					return conf, nil
				}

				namePrefix := getAttribute(cc, "prefix")

				if namePrefix != "" {
					if r.localWorkspace == "" {
						return conf, errors.Errorf("--terraform-workspace is not specified. Unable to detect organization or workspace.")
					}

					conf.Workspace = namePrefix + r.localWorkspace
					return conf, nil
				}
			}
		}
	}

	return conf, nil
}

func (r *TFCRemoteVariablesLoader) getVarKey(variable tfcVar) string {
	if variable.Category == "env" && strings.HasPrefix(variable.Key, "TF_VAR_") {
		return strings.TrimPrefix(variable.Key, "TF_VAR_")
	}

	return variable.Key
}

func (r *TFCRemoteVariablesLoader) getVarValue(variable tfcVar) cty.Value {
	if variable.Sensitive {
		r.logger.Debug().Msgf("skipping sensitive variable %s", variable.Key)
		return cty.DynamicVal
	}

	if variable.Value == "" {
		r.logger.Debug().Msgf("skipping empty variable %s", variable.Key)
		return cty.DynamicVal
	}

	if variable.Category == "env" && !strings.HasPrefix(variable.Key, "TF_VAR_") {
		r.logger.Debug().Msgf("skipping environment variable %s", variable.Key)
		return cty.DynamicVal
	}

	if variable.HCL {
		val, err := ParseVariable(variable.Value)
		if err != nil {
			r.logger.Debug().Err(err).Msgf("could not parse variable %s with HCL value", variable.Key)
			return cty.DynamicVal
		}
		return val
	}

	return cty.StringVal(variable.Value)
}

func getAttribute(block *Block, name string) string {
	if block == nil {
		return ""
	}

	attr := block.GetAttribute(name)
	if attr != nil {
		return attr.AsString()
	}

	return ""
}

// SpaceliftRemoteVariableLoader orchestrates communicating with the Spacelift API to fetch remote variables.
type SpaceliftRemoteVariableLoader struct {
	Client   client.Client
	Metadata vcs.Metadata

	cache  *sync.Map
	hitMap *sync.Map
}

// NewSpaceliftRemoteVariableLoader creates a new SpaceliftRemoteVariableLoader, this function
// expects that the required environment variables are set.
func NewSpaceliftRemoteVariableLoader(metadata vcs.Metadata, apiKeyEndpoint, apiKeyId, apiKeySecret string) (*SpaceliftRemoteVariableLoader, error) {
	httpClient := http.DefaultClient

	session, err := spaceliftSession.FromAPIKey(context.Background(), httpClient)(apiKeyEndpoint, apiKeyId, apiKeySecret)
	if err != nil {
		return nil, fmt.Errorf("could not create Spacelift session: %w", err)
	}

	return &SpaceliftRemoteVariableLoader{
		Client:   client.New(httpClient, session),
		Metadata: metadata,
		cache:    &sync.Map{},
		hitMap:   &sync.Map{},
	}, nil
}

// Load fetches remote variables from Spacelift by querying the stacks for the
// provided environment name and remote name.
func (s *SpaceliftRemoteVariableLoader) Load(options RemoteVarLoaderOptions) (map[string]cty.Value, error) {
	if options.ModulePath == "" {
		logging.Logger.Trace().Msg("no module path provided, skipping Spacelift remote variable loading")
		return nil, nil
	}

	// get the stack which matches the remote name and the environment name
	// in future we should get all stacks for the remote name and then
	// dynamically create projects out of the stacks returned.
	stacks, err := s.getStacks(context.Background(), &getStackOptions{
		count:          10, // We only want 1 stack, but want to filter by the module suffix
		repositoryName: s.Metadata.Remote.Name,
		projectRoot:    options.ModulePath,
	})
	if err != nil {
		return nil, fmt.Errorf("could not get Spacelift stacks: %w", err)
	}

	if len(stacks) == 0 {
		logging.Logger.Debug().Msgf("no Spacelift stack found for module path %q", options.ModulePath)
		return nil, nil
	}

	// If there is a module suffix, filter the stacks by it
	if options.Environment != "" {
		var filteredStacks []stack
		for _, stack := range stacks {
			if strings.HasSuffix(stack.Name, fmt.Sprintf(":%s", options.Environment)) || strings.HasPrefix(stack.Name, fmt.Sprintf("%s/", options.Environment)) {
				filteredStacks = append(filteredStacks, stack)
			}
		}
		stacks = filteredStacks
	}

	if len(stacks) == 0 {
		logging.Logger.Warn().Msgf("no Spacelift stack found for module path %q with environment %q", options.ModulePath, options.Environment)
		return nil, nil
	}

	if len(stacks) > 1 {
		logging.Logger.Warn().Msgf("found multiple Spacelift stacks for module path %q with environment %q, skipping Spacelift remote variable loading", options.ModulePath, options.Environment)
		return nil, nil
	}
	stack := stacks[0]
	prevOptions, hit := s.hitMap.Load(stack.ID)
	if hit {
		logging.Logger.Debug().Msgf("spacelift stack %q already used for options %+v will not load remote variables", stack.ID, prevOptions)
		return nil, nil
	}

	s.hitMap.Store(stack.ID, options)

	logging.Logger.Debug().Msgf("found Spacelift stack %q for module path %q with environment: %q", stack.Name, options.ModulePath, options.Environment)

	vars := map[string]cty.Value{}

	// Spacelift precedence is runtime config > config > attached contexts
	for _, env := range stacks[0].AttachedContexts {
		if strings.HasPrefix(env.ID, "TF_VAR_") {
			vars[strings.TrimPrefix(env.ID, "TF_VAR_")] = cty.StringVal(env.ContextName)
		}
	}

	for _, env := range stacks[0].Config {
		if strings.HasPrefix(env.ID, "TF_VAR_") {
			vars[strings.TrimPrefix(env.ID, "TF_VAR_")] = cty.StringVal(env.Value)
		}
	}

	for _, env := range stacks[0].RuntimeConfig {
		if strings.HasPrefix(env.Element.ID, "TF_VAR_") {
			vars[strings.TrimPrefix(env.Element.ID, "TF_VAR_")] = cty.StringVal(env.Element.Value)
		}
	}

	logging.Logger.Debug().Msgf("loaded %d Spacelift remote variables", len(vars))

	// Only marshal and log the variables if the log level is trace
	// Otherwise we want to skip the marshalling for performance reasons
	if logging.Logger.GetLevel() == zerolog.TraceLevel {
		s := ctyJson.SimpleJSONValue{Value: cty.ObjectVal(vars)}
		b, _ := s.MarshalJSON()
		logging.Logger.Trace().Msgf("Spacelift remote variables: %v", string(b))
	}

	return vars, nil
}

type stack struct {
	ID               string            `graphql:"id" json:"id,omitempty"`
	Name             string            `graphql:"name" json:"name,omitempty"`
	ProjectRoot      string            `graphql:"projectRoot" json:"projectRoot,omitempty"`
	Config           []stackConfig     `graphql:"config" json:"config,omitempty"`
	RuntimeConfig    []runtimeConfig   `graphql:"runtimeConfig" json:"runtimeConfig,omitempty"`
	AttachedContexts []attachedContext `graphql:"attachedContexts" json:"attachedContexts,omitempty"`
}

type stackConfig struct {
	ID    string `graphql:"id" json:"id,omitempty"`
	Value string `graphql:"value" json:"value,omitempty"`
}

type runtimeConfig struct {
	Element stackConfig `graphql:"element" json:"element,omitempty"`
}

type attachedContext struct {
	ID          string `graphql:"id" json:"id,omitempty"`
	ContextName string `graphql:"contextName" json:"contextName,omitempty"`
}

type getStackOptions struct {
	count int32

	repositoryName string
	projectRoot    string
}

func (s *SpaceliftRemoteVariableLoader) getStacks(ctx context.Context, p *getStackOptions) ([]stack, error) {
	if v, ok := s.cache.Load(p); ok {
		return v.([]stack), nil
	}

	var query struct {
		SearchStacksOutput struct {
			Edges []struct {
				Node stack `graphql:"node"`
			} `graphql:"edges"`
			PageInfo structs.PageInfo `graphql:"pageInfo"`
		} `graphql:"searchStacks(input: $input)"`
	}
	var conditions []structs.QueryPredicate
	if p.repositoryName != "" {
		conditions = append(conditions, structs.QueryPredicate{
			Field: graphql.String("repository"),
			Constraint: structs.QueryFieldConstraint{
				StringMatches: &[]graphql.String{graphql.String(p.repositoryName)},
			},
		})
	}
	if p.projectRoot != "" {
		conditions = append(conditions, structs.QueryPredicate{
			Field: graphql.String("projectRoot"),
			Constraint: structs.QueryFieldConstraint{
				StringMatches: &[]graphql.String{graphql.String(p.projectRoot)},
			},
		})
	}

	input := structs.SearchInput{
		First:      graphql.NewInt(graphql.Int(p.count)),
		Predicates: &conditions,
	}

	variables := map[string]interface{}{"input": input}

	if err := s.Client.Query(
		ctx,
		&query,
		variables,
	); err != nil {
		return nil, errors.Wrap(err, "failed search for stacks")
	}

	result := make([]stack, 0)
	for _, q := range query.SearchStacksOutput.Edges {
		result = append(result, q.Node)
	}

	s.cache.Store(p, result)
	return result, nil
}
