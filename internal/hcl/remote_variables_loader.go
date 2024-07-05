package hcl

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/pkg/errors"
	"github.com/zclconf/go-cty/cty"

	"github.com/infracost/infracost/internal/extclient"
	"github.com/infracost/infracost/internal/logging"
)

// RemoteVariablesLoader handles loading remote variables from Terraform Cloud.
type RemoteVariablesLoader struct {
	client         *extclient.AuthedAPIClient
	localWorkspace string
	remoteConfig   *TFCRemoteConfig
}

// RemoteVariablesLoaderOption defines a function that can set properties on an RemoteVariablesLoader.
type RemoteVariablesLoaderOption func(r *RemoteVariablesLoader)

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
// the RemoteVariablesLoader. This is normally done to override the configuration
// detected from the HCL blocks.
func RemoteVariablesLoaderWithRemoteConfig(config TFCRemoteConfig) RemoteVariablesLoaderOption {
	return func(r *RemoteVariablesLoader) {
		r.remoteConfig = &config
	}
}

// NewRemoteVariablesLoader constructs a new loader for fetching remote variables.
func NewRemoteVariablesLoader(client *extclient.AuthedAPIClient, localWorkspace string, opts ...RemoteVariablesLoaderOption) *RemoteVariablesLoader {
	if localWorkspace == "" {
		localWorkspace = os.Getenv("TF_WORKSPACE")
	}

	r := &RemoteVariablesLoader{
		client:         client,
		localWorkspace: localWorkspace,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// Load fetches remote variables if terraform block contains organization and
// workspace name.
func (r *RemoteVariablesLoader) Load(blocks Blocks) (map[string]cty.Value, error) {
	logging.Logger.Trace().Msg("Downloading Terraform remote variables")
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
				return vars, err
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
		logging.Logger.Debug().Err(err).Msgf("could not request Terraform workspace: %s for organization: %s", config.Workspace, config.Organization)
		return vars, nil
	}

	var workspaceResponse tfcWorkspaceResponse
	if json.Unmarshal(body, &workspaceResponse) != nil {
		logging.Logger.Debug().Err(err).Msgf("malformed Terraform API response using workspace: %s organization: %s", config.Workspace, config.Organization)
		return vars, nil
	}

	if workspaceResponse.Data.Attributes.ExecutionMode == "local" {
		logging.Logger.Trace().Msgf("Terraform workspace %s does use local execution, skipping downloading remote variables", config.Workspace)
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
				val := getVarValue(vv)
				if !val.IsNull() && val.IsKnown() {
					vars[vv.Key] = val
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
		val := getVarValue(v.Attributes)
		if !val.IsNull() {
			vars[v.Attributes.Key] = val
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

func (r *RemoteVariablesLoader) getCloudOrganizationWorkspace(blocks Blocks) TFCRemoteConfig {
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

func (r *RemoteVariablesLoader) getBackendOrganizationWorkspace(blocks Blocks) (TFCRemoteConfig, error) {
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

func getVarValue(variable tfcVar) cty.Value {
	if variable.Sensitive || variable.Category != "terraform" || variable.Value == "" {
		return cty.DynamicVal
	}

	return cty.StringVal(variable.Value)
}
