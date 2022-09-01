package schema

import (
	// nolint:gosec

	"crypto/md5" // nolint:gosec
	"encoding/base32"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/infracost/infracost/internal/logging"
)

type ProjectMetadata struct {
	Path                string `json:"path"`
	Type                string `json:"type"`
	TerraformModulePath string `json:"terraformModulePath,omitempty"`
	TerraformWorkspace  string `json:"terraformWorkspace,omitempty"`
	VCSSubPath          string `json:"vcsSubPath,omitempty"`
}

func (m *ProjectMetadata) WorkspaceLabel() string {
	if m.TerraformWorkspace == "default" {
		return ""
	}

	return m.TerraformWorkspace
}

func (m *ProjectMetadata) GenerateProjectName(repoURL string, dashboardEnabled bool) string {
	// If the VCS repo is set, create the name from that
	if repoURL != "" {
		n := nameFromRepoURL(repoURL)

		if m.VCSSubPath != "" {
			n += "/" + m.VCSSubPath
		}

		return n
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
	Name          string
	Metadata      *ProjectMetadata
	PastResources []*Resource
	Resources     []*Resource
	Diff          []*Resource
	HasDiff       bool
}

func NewProject(name string, metadata *ProjectMetadata) *Project {
	return &Project{
		Name:     name,
		Metadata: metadata,
		HasDiff:  true,
	}
}

// AllResources returns a pointer list of all resources of the state.
func (p *Project) AllResources() []*Resource {
	var resources []*Resource
	resources = append(resources, p.PastResources...)
	resources = append(resources, p.Resources...)
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

// Parses the "org/repo" from the git URL if possible.
// Otherwise it just returns the URL.
func nameFromRepoURL(rawURL string) string {
	escaped, err := url.QueryUnescape(rawURL)
	if err == nil {
		rawURL = escaped
	}

	// Removes schema and/or credentials
	r := regexp.MustCompile(`^.*@|http(?:s)?:\/\/`)
	rawURL = r.ReplaceAllString(rawURL, "")

	// Removes extension if present
	r = regexp.MustCompile(`\.git$`)
	rawURL = r.ReplaceAllString(rawURL, "")

	// Removes port if present
	r = regexp.MustCompile(`:\d+/`)
	rawURL = r.ReplaceAllString(rawURL, ":")

	// Delimits domain from repo name with `:`
	if !strings.Contains(rawURL, ":") {
		rawURL = strings.Replace(rawURL, "/", ":", 1)
	}

	parts := strings.Split(rawURL, ":")

	if len(parts) == 2 {
		if parts[0] == "dev.azure.com" || parts[0] == "ssh.dev.azure.com" {
			name := parseAzureDevOpsRepoPath(parts[1])

			return name
		}

		return parts[1]
	}

	return rawURL
}

func parseAzureDevOpsRepoPath(path string) string {
	r := regexp.MustCompile(`(?:(.+)(?:\/(.+)\/_git\/)(.+))`)
	m := r.FindStringSubmatch(path)

	if len(m) > 3 {
		return fmt.Sprintf("%s/%s/%s", m[1], m[2], m[3])
	}

	r = regexp.MustCompile(`(?:(?:v3\/)(.+)(?:\/(.+)\/)(.+))`)
	m = r.FindStringSubmatch(path)

	if len(m) > 3 {
		return fmt.Sprintf("%s/%s/%s", m[1], m[2], m[3])
	}

	return path
}

// Returns a lowercase truncated hash of length l
func shortHash(s string, l int) string {
	sum := md5.Sum([]byte(s)) //nolint:gosec
	var b = sum[:]
	h := base32.StdEncoding.EncodeToString(b)

	return strings.ToLower(h)[:l]
}
