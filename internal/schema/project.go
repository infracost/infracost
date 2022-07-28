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
	"time"

	log "github.com/sirupsen/logrus"
)

type ProjectMetadata struct {
	Path                string `json:"path"`
	InfracostCommand    string `json:"infracostCommand"`
	Type                string `json:"type"`
	TerraformModulePath string `json:"terraformModulePath,omitempty"`
	TerraformWorkspace  string `json:"terraformWorkspace,omitempty"`

	Branch            string    `json:"branch"`
	Commit            string    `json:"commit"`
	CommitAuthorName  string    `json:"commitAuthorName"`
	CommitAuthorEmail string    `json:"commitAuthorEmail"`
	CommitTimestamp   time.Time `json:"commitTimestamp"`
	CommitMessage     string    `json:"commitMessage"`

	VCSRepoURL           string `json:"vcsRepoUrl,omitempty"`
	VCSSubPath           string `json:"vcsSubPath,omitempty"`
	VCSProvider          string `json:"vcsProvider,omitempty"`
	VCSBaseBranch        string `json:"vcsBaseBranch,omitempty"`
	VCSPullRequestTitle  string `json:"vcsPullRequestTitle,omitempty"`
	VCSPullRequestURL    string `json:"vcsPullRequestUrl,omitempty"`
	VCSPullRequestAuthor string `json:"vcsPullRequestAuthor,omitempty"`
	VCSPipelineRunID     string `json:"vcsPipelineRunId,omitempty"`
	VCSPullRequestID     string `json:"vcsPullRequestId,omitempty"`
}

func (m *ProjectMetadata) WorkspaceLabel() string {
	if m.TerraformWorkspace == "default" {
		return ""
	}

	return m.TerraformWorkspace
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

func GenerateProjectName(metadata *ProjectMetadata, projectName string, dashboardEnabled bool) string {
	// If there is a user defined project name, use it.
	if projectName != "" {
		return projectName
	}

	// If the VCS repo is set, create the name from that
	if metadata.VCSRepoURL != "" {
		n := nameFromRepoURL(metadata.VCSRepoURL)

		if metadata.VCSSubPath != "" {
			n += "/" + metadata.VCSSubPath
		}

		return n
	}

	// If not then use a hash of the absolute filepath to the project
	if dashboardEnabled {
		absPath, err := filepath.Abs(metadata.Path)
		if err != nil {
			log.Debugf("Could not get absolute path for %s", metadata.Path)
			absPath = metadata.Path
		}

		return fmt.Sprintf("project_%s", shortHash(absPath, 8))
	}

	return metadata.Path
}

// Parses the "org/repo" from the git URL if possible.
// Otherwise it just returns the URL.
func nameFromRepoURL(rawURL string) string {
	escaped, err := url.QueryUnescape(rawURL)
	if err == nil {
		rawURL = escaped
	}

	r := regexp.MustCompile(`(?:\w+@|http(?:s)?:\/\/)(?:.*@)?([^:\/]+)[:\/]([^\.]+)`)
	m := r.FindStringSubmatch(rawURL)

	if len(m) > 2 {
		var n = m[2]

		if m[1] == "dev.azure.com" || m[1] == "ssh.dev.azure.com" {
			n = parseAzureDevOpsRepoPath(m[2])
		}

		return n
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
