package schema

import (
	// nolint:gosec

	"crypto/md5" // nolint:gosec
	"encoding/base32"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/vcs"
)

const (
	DiagJSONParsingFailure = iota + 1
	DiagModuleEvaluationFailure
	DiagTerragruntEvaluationFailure
	DiagTerragruntModuleEvaluationFailure
	DiagPrivateModuleDownloadFailure
	DiagPrivateRegistryModuleDownloadFailure
)

// ProjectDiag holds information about all diagnostics associated with a project.
// This can be both critical or warnings.
type ProjectDiag struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func (p *ProjectDiag) Error() string {
	if p == nil {
		return ""
	}

	return p.Message
}

type ProjectMetadata struct {
	Path                string        `json:"path"`
	Type                string        `json:"type"`
	ConfigSha           string        `json:"configSha,omitempty"`
	TerraformModulePath string        `json:"terraformModulePath,omitempty"`
	TerraformWorkspace  string        `json:"terraformWorkspace,omitempty"`
	VCSSubPath          string        `json:"vcsSubPath,omitempty"`
	VCSCodeChanged      *bool         `json:"vcsCodeChanged,omitempty"`
	Errors              []ProjectDiag `json:"errors,omitempty"`
	Warnings            []ProjectDiag `json:"warnings,omitempty"`
	Policies            Policies      `json:"policies,omitempty"`
}

// AddError pushes the provided error onto the metadata list. It does a naive conversion to ProjectDiag
// if the error provided is not already a diagnostic.
func (m *ProjectMetadata) AddError(err error) {
	var diag *ProjectDiag
	if errors.As(err, &diag) {
		m.Errors = append(m.Errors, *diag)
		return
	}

	m.Errors = append(m.Errors, ProjectDiag{Message: err.Error()})
}

// AddErrorWithCode is the same as AddError except adds a code to the fallback diagnostic.
func (m *ProjectMetadata) AddErrorWithCode(err error, code int) {
	var diag *ProjectDiag
	if errors.As(err, &diag) {
		m.Errors = append(m.Errors, *diag)
		return
	}

	m.Errors = append(m.Errors, ProjectDiag{Code: code, Message: err.Error()})
}

func (m *ProjectMetadata) HasErrors() bool {
	return len(m.Errors) > 0
}

func (m *ProjectMetadata) WorkspaceLabel() string {
	if m.TerraformWorkspace == "default" {
		return ""
	}

	return m.TerraformWorkspace
}

func (m *ProjectMetadata) GenerateProjectName(remote vcs.Remote, dashboardEnabled bool) string {
	// If the VCS repo is set, use its name
	if remote.Name != "" {
		name := remote.Name

		if m.VCSSubPath != "" {
			name += "/" + m.VCSSubPath
		}

		return name
	}

	// If not then use a hash of the absolute filepath to the project
	if dashboardEnabled {
		absPath, err := filepath.Abs(m.Path)
		if err != nil {
			logging.Logger.Debugf("Could not get absolute path for %s", m.Path)
			absPath = m.Path
		}

		return fmt.Sprintf("project_%s", shortHash(absPath, 8))
	}

	return m.Path
}

// Projects is a slice of Project that is ordered alphabetically by project name.
type Projects []*Project

func (p Projects) Len() int           { return len(p) }
func (p Projects) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p Projects) Less(i, j int) bool { return p[i].Name < p[j].Name }

// Project contains the existing, planned state of
// resources and the diff between them.
type Project struct {
	Name                 string
	Metadata             *ProjectMetadata
	PartialPastResources []*PartialResource
	PartialResources     []*PartialResource
	PastResources        []*Resource
	Resources            []*Resource
	Diff                 []*Resource
	HasDiff              bool
}

func NewProject(name string, metadata *ProjectMetadata) *Project {
	return &Project{
		Name:     name,
		Metadata: metadata,
		HasDiff:  true,
	}
}

// NameWithWorkspace returns the proect Name appended with the paranenthized workspace name
// from Metadata if one exists.
func (p *Project) NameWithWorkspace() string {
	if p.Metadata.WorkspaceLabel() == "" {
		return p.Name
	}
	return fmt.Sprintf("%s (%s)", p.Name, p.Metadata.WorkspaceLabel())
}

// AllResources returns a pointer list of all resources of the state.
func (p *Project) AllResources() []*Resource {
	m := make(map[*Resource]bool)
	for _, r := range p.PastResources {
		m[r] = true
	}

	for _, r := range p.Resources {
		if _, ok := m[r]; !ok {
			m[r] = true
		}
	}

	var resources []*Resource
	for r := range m {
		resources = append(resources, r)
	}

	return resources
}

// AllPartialResources returns a pointer list of the current and past partial resources
func (p *Project) AllPartialResources() []*PartialResource {
	m := make(map[*PartialResource]bool)
	for _, r := range p.PartialPastResources {
		m[r] = true
	}

	for _, r := range p.PartialResources {
		if _, ok := m[r]; !ok {
			m[r] = true
		}
	}

	var resources []*PartialResource
	for r := range m {
		resources = append(resources, r)
	}

	return resources
}

// BuildResources builds the resources from the partial resources
// and sets the PastResources and Resources fields.
func (p *Project) BuildResources(usageMap UsageMap) {
	pastResources := make([]*Resource, 0, len(p.PartialPastResources))
	resources := make([]*Resource, 0, len(p.PartialResources))

	seen := make(map[*PartialResource]*Resource)

	for _, p := range p.PartialPastResources {
		u := usageMap.Get(p.ResourceData.Address)
		r := BuildResource(p, u)
		seen[p] = r
		pastResources = append(pastResources, r)
	}

	for _, p := range p.PartialResources {
		r, ok := seen[p]
		if !ok {
			u := usageMap.Get(p.ResourceData.Address)
			r = BuildResource(p, u)
			seen[p] = r
		}
		resources = append(resources, r)
	}

	p.PastResources = pastResources
	p.Resources = resources
}

// CalculateDiff calculates the diff of past and current resources
func (p *Project) CalculateDiff() {
	if p.HasDiff {
		p.Diff = CalculateDiff(p.PastResources, p.Resources)
	}
}

// AllProjectResources returns the resources for all projects
func AllProjectResources(projects []*Project) []*Resource {
	resources := make([]*Resource, 0)

	for _, p := range projects {
		resources = append(resources, p.Resources...)
	}

	return resources
}

// Returns a lowercase truncated hash of length l
func shortHash(s string, l int) string {
	sum := md5.Sum([]byte(s)) //nolint:gosec
	var b = sum[:]
	h := base32.StdEncoding.EncodeToString(b)

	return strings.ToLower(h)[:l]
}
