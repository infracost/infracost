package hcl

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/infracost/infracost/internal/extclient"
	"github.com/infracost/infracost/internal/ui"
	"github.com/pkg/errors"
	"github.com/zclconf/go-cty/cty"
)

// RemoteVariablesLoader handles loading remote variables from Terraform Cloud.
type RemoteVariablesLoader struct {
	client         extclient.AuthedAPIClient
	localWorkspace string
	newSpinner     ui.SpinnerFunc
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

// RemoteVariablesLoaderWithSpinner enables the RemoteVariablesLoader to use an ui.Spinner to
// show the progress of loading the remote variables.
func RemoteVariablesLoaderWithSpinner(f ui.SpinnerFunc) RemoteVariablesLoaderOption {
	return func(r *RemoteVariablesLoader) {
		r.newSpinner = f
	}
}

// NewRemoteVariablesLoaderLoader constructs a new loader for fetching remote
// variables.
func NewRemoteVariablesLoader(client extclient.AuthedAPIClient, localWorkspace string, opts ...RemoteVariablesLoaderOption) *RemoteVariablesLoader {
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
	spinnerMsg := "Downloading Terraform remote variables"
	vars := map[string]cty.Value{}

	var organization, workspace string

	organization, workspace, err := r.getCloudOrganizationWorkspace(blocks)
	if err != nil {
		organization, workspace, err = r.getBackendOrganizationWorkspace(blocks)

		if err != nil {
			var spinner *ui.Spinner
			if r.newSpinner != nil {
				// In case name prefix is set, but workspace flag is missing show the
				// failed spinner message. Otherwise the remote variables loading is
				// skipped entirely.
				spinner = r.newSpinner(spinnerMsg)
				spinner.Fail()
			}
			return vars, err
		}

		if organization == "" && workspace == "" {
			return vars, nil
		}
	}

	var spinner *ui.Spinner
	if r.newSpinner != nil {
		spinner = r.newSpinner(spinnerMsg)
		defer spinner.Success()
	}

	endpoint := fmt.Sprintf("/api/v2/organizations/%s/workspaces/%s", organization, workspace)
	body, err := r.client.Get(endpoint)
	if err != nil {
		if spinner != nil {
			spinner.Fail()
		}
		return vars, err
	}

	var workspaceResponse tfcWorkspaceResponse
	if json.Unmarshal(body, &workspaceResponse) != nil {
		if spinner != nil {
			spinner.Fail()
		}
		return vars, errors.New("unable to parse Workspace response")
	}

	if workspaceResponse.Data.Attributes.ExecutionMode != "remote" {
		if spinner != nil {
			spinner.Fail()
		}
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
			if spinner != nil {
				spinner.Fail()
			}
			return vars, err
		}

		var varsetsResponse tfcVarsetResponse
		if json.Unmarshal(body, &varsetsResponse) != nil {
			if spinner != nil {
				spinner.Fail()
			}
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
				if !val.IsNull() {
					vars[vv.Key] = val
				}
			}
		}
	}

	endpoint = fmt.Sprintf("/api/v2/workspaces/%s/vars", workspaceID)
	body, err = r.client.Get(endpoint)
	if err != nil {
		if spinner != nil {
			spinner.Fail()
		}
		return vars, err
	}

	var varsResponse tfcVarResponse
	if json.Unmarshal(body, &varsResponse) != nil {
		if spinner != nil {
			spinner.Fail()
		}
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

func (r *RemoteVariablesLoader) getCloudOrganizationWorkspace(blocks Blocks) (string, string, error) {
	organization := ""

	for _, block := range blocks.OfType("terraform") {
		for _, c := range block.childBlocks.OfType("cloud") {
			organization = getAttribute(c, "organization")

			for _, cc := range c.childBlocks.OfType("workspaces") {
				return organization, getAttribute(cc, "name"), nil
			}
		}
	}

	return "", "", errors.Errorf("unable to detect organization or workspace")
}

func (r *RemoteVariablesLoader) getBackendOrganizationWorkspace(blocks Blocks) (string, string, error) {
	organization := ""

	for _, block := range blocks.OfType("terraform") {
		for _, c := range block.childBlocks.OfType("backend") {
			if c.Label() != "remote" {
				continue
			}

			organization = getAttribute(c, "organization")

			for _, cc := range c.childBlocks.OfType("workspaces") {
				name := getAttribute(cc, "name")

				if name != "" {
					return organization, name, nil
				}

				namePrefix := getAttribute(cc, "prefix")

				if namePrefix != "" {
					if r.localWorkspace == "" {
						return "", "", errors.Errorf("--terraform-workspace is not specified. Unable to detect organization or workspace.")
					}

					return organization, namePrefix + r.localWorkspace, nil
				}
			}
		}
	}

	return "", "", nil
}

func getAttribute(block *Block, name string) string {
	result := ""

	if block == nil {
		return result
	}

	attr := block.GetAttribute(name)
	if attr != nil {
		val := attr.Value()
		if !val.IsNull() {
			result = val.AsString()
		}
	}

	return result
}

func getVarValue(variable tfcVar) cty.Value {
	if variable.Sensitive || variable.Category != "terraform" || variable.Value == "" {
		return cty.NilVal
	}

	return cty.StringVal(variable.Value)
}
