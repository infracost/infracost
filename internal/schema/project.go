package schema

import (
	// nolint:gosec

	"crypto/md5" // nolint:gosec
	"encoding/base32"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/vcs"
)

// Warning holds information about non-critical errors that occurred when evaluating a project.
type Warning struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type ProjectMetadata struct {
	Path                string    `json:"path"`
	Type                string    `json:"type"`
	TerraformModulePath string    `json:"terraformModulePath,omitempty"`
	TerraformWorkspace  string    `json:"terraformWorkspace,omitempty"`
	VCSSubPath          string    `json:"vcsSubPath,omitempty"`
	Warnings            []Warning `json:"warnings,omitempty"`
	Policies            Policies  `json:"policies,omitempty"`
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
	var resources []*Resource
	resources = append(resources, p.PastResources...)
	resources = append(resources, p.Resources...)
	return resources
}

// AllPartialResources returns a pointer list of the current and past partial resources
func (p *Project) AllPartialResources() []*PartialResource {
	var resources []*PartialResource
	resources = append(resources, p.PartialPastResources...)
	resources = append(resources, p.PartialResources...)
	return resources
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
