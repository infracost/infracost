package schema

import (
	// nolint:gosec

	"crypto/md5" // nolint:gosec
	"encoding/base32"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/util"
	"github.com/infracost/infracost/internal/vcs"
)

const (
	// Diags for local module issues
	diagJSONParsingFailure                = 101
	diagModuleEvaluationFailure           = 102
	diagTerragruntEvaluationFailure       = 103
	diagTerragruntModuleEvaluationFailure = 104
	diagMissingVars                       = 105
	diagEmptyPathType                     = 106

	// Diags for git module issues
	diagPrivateModuleDownloadFailure = 201

	// Diags for registry module issues
	diagPrivateRegistryModuleDownloadFailure = 301

	// Diags for infracost cloud issues
	diagRunQuotaExceeded = 401
)

// ProjectDiag holds information about all diagnostics associated with a project.
// This can be both critical or warnings.
type ProjectDiag struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	IsError bool   `json:"isError"`

	// FriendlyMessage should be used to display a readable message to the CLI user.
	FriendlyMessage string `json:"-"`
}

// IsEmptyPathTypeError checks if the error is a diag for an empty path type.
func IsEmptyPathTypeError(err error) bool {
	if err == nil {
		return false
	}

	var diag *ProjectDiag
	return errors.As(err, &diag) && diag.Code == diagEmptyPathType
}

// NewEmptyPathTypeError returns a project diag to indicate that a path type
// cannot be detected.
func NewEmptyPathTypeError(err error) *ProjectDiag {
	return newDiag(diagEmptyPathType, err.Error(), true, nil, err)
}

// NewDiagRunQuotaExceeded returns a project diag for a run quota exceeded error.
func NewDiagRunQuotaExceeded(err error) *ProjectDiag {
	return &ProjectDiag{
		Code:    diagRunQuotaExceeded,
		Message: err.Error(),
		IsError: true,
	}
}

// NewDiagTerragruntModuleEvaluationFailure returns a project diag for a
// terragrunt module evaluation failure. This is used when a Terraform module
// which a Terragrunt configuration file references fails to evaluate.
func NewDiagTerragruntModuleEvaluationFailure(err error) *ProjectDiag {
	return newDiag(diagTerragruntModuleEvaluationFailure, err.Error(), true, nil, err)
}

// NewDiagTerragruntEvaluationFailure returns a project diag for a Terragrunt
// evaluation failure. This is used when a Terragrunt fails to run/evaluate in
// most cases this means that the entire project is not evaluated. It is
// considered a critical error.
func NewDiagTerragruntEvaluationFailure(err error) *ProjectDiag {
	return newDiag(diagTerragruntEvaluationFailure, err.Error(), true, nil, err)
}

// NewDiagModuleEvaluationFailure returns a project diag for a module evaluation
// failure. This is used when a Terraform module fails to evaluate. This can
// either be a root module of a Terraform project or a child module.
func NewDiagModuleEvaluationFailure(err error) *ProjectDiag {
	return newDiag(diagModuleEvaluationFailure, err.Error(), true, nil, err)
}

// NewDiagJSONParsingFailure returns a project diag for a JSON parsing failure in
// the intermediary JSON that the HCL provider generates. This is considered a
// critical error as a project will not have any costs if this happens.
func NewDiagJSONParsingFailure(err error) *ProjectDiag {
	return newDiag(diagJSONParsingFailure, err.Error(), true, nil, err)
}

// NewPrivateRegistryDiag returns a project diag for a private registry module
// download failure. This contains additional information about the module source
// and the discovered location returned by the registry.
func NewPrivateRegistryDiag(source string, moduleLocation *string, err error) *ProjectDiag {
	source = util.RedactUrl(source)
	moduleLocation = util.RedactUrlPtr(moduleLocation)
	errorString := util.RedactUrl(err.Error())

	return newDiag(
		diagPrivateRegistryModuleDownloadFailure,
		fmt.Sprintf("Failed to lookup module %q - %s", source, errorString),
		true,
		map[string]any{
			"moduleLocation": moduleLocation,
			"moduleSource":   source,
		},
		err,
	)
}

// NewFailedDownloadDiagnostic returns a project diag for a failed module download.
// This contains additional information about the module source and the type of source.
func NewFailedDownloadDiagnostic(source string, err error) *ProjectDiag {
	sourceType := "https"
	if strings.Contains(err.Error(), "ssh://") {
		sourceType = "ssh"
	}

	source = util.RedactUrl(source)
	errorString := util.RedactUrl(err.Error())
	return newDiag(
		diagPrivateModuleDownloadFailure,
		fmt.Sprintf("Failed to download module %q - %s", source, errorString),
		true,
		map[string]any{
			"source":       sourceType,
			"moduleSource": source,
		},
		err,
	)
}

func newDiag(code int, message string, isError bool, data any, err error) *ProjectDiag {
	// if the error is already a ProjectDiag, return it rather than creating a new
	// one. This is to avoid collision of diagnostics.
	var diag *ProjectDiag
	if errors.As(err, &diag) {
		return diag
	}

	return &ProjectDiag{
		Code:    code,
		Message: message,
		IsError: isError,
		Data:    data,
	}
}

// NewDiagMissingVars returns a ProjectDiag for missing Terraform vars. This is
// considered a non-critical error and is used to notify the user.
func NewDiagMissingVars(vars ...string) *ProjectDiag {
	return &ProjectDiag{
		Code:    diagMissingVars,
		Message: "Missing Terraform vars",
		Data:    vars,
		FriendlyMessage: fmt.Sprintf(
			"Input values were not provided for following Terraform variables: %s. %s",
			joinQuotes(vars),
			"Use --terraform-var-file or --terraform-var to specify them.",
		),
	}
}

func joinQuotes(elems []string) string {

	quoted := make([]string, len(elems))
	for i, elem := range elems {
		quoted[i] = fmt.Sprintf("%q", elem)
	}

	return strings.Join(quoted, ", ")
}

func (p *ProjectDiag) Error() string {
	if p == nil {
		return ""
	}

	return p.Message
}

type ProjectMetadata struct {
	Path                string             `json:"path"`
	Type                string             `json:"type"`
	ConfigSha           string             `json:"configSha,omitempty"`
	PolicySha           string             `json:"policySha,omitempty"`
	PastPolicySha       string             `json:"pastPolicySha,omitempty"`
	TerraformModulePath string             `json:"terraformModulePath,omitempty"`
	TerraformWorkspace  string             `json:"terraformWorkspace,omitempty"`
	VCSSubPath          string             `json:"vcsSubPath,omitempty"`
	VCSCodeChanged      *bool              `json:"vcsCodeChanged,omitempty"`
	Errors              []*ProjectDiag     `json:"errors,omitempty"` // contains merged current and past errors
	CurrentErrors       []*ProjectDiag     `json:"currentErrors,omitempty"`
	PastErrors          []*ProjectDiag     `json:"pastErrors,omitempty"`
	Warnings            []*ProjectDiag     `json:"warnings,omitempty"`
	Policies            Policies           `json:"policies,omitempty"`
	Providers           []ProviderMetadata `json:"providers,omitempty"`
	RemoteModuleCalls   []string           `json:"remoteModuleCalls,omitempty"`
}

// DetectProjectMetadata returns a new ProjectMetadata struct initialized
// from environment variables and the provided path.
func DetectProjectMetadata(path string) *ProjectMetadata {
	vcsSubPath := os.Getenv("INFRACOST_VCS_SUB_PATH")
	terraformWorkspace := os.Getenv("INFRACOST_TERRAFORM_WORKSPACE")

	if vcsSubPath == "" {
		vcsSubPath = gitSubPath(path)
	}

	return &ProjectMetadata{
		Path:               path,
		TerraformWorkspace: terraformWorkspace,
		VCSSubPath:         vcsSubPath,
	}
}

func gitSubPath(path string) string {
	topLevel, err := gitToplevel(path)
	if err != nil {
		logging.Logger.Debug().Err(err).Msgf("Could not get git top level directory for %s", path)
		return ""
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		logging.Logger.Debug().Err(err).Msgf("Could not get absolute path for %s", path)
		return ""
	}

	subPath, err := filepath.Rel(topLevel, absPath)
	if err != nil {
		logging.Logger.Debug().Err(err).Msgf("Could not get relative path for %s from %s", absPath, topLevel)
		return ""
	}

	if subPath == "." {
		return ""
	}

	return subPath
}

func gitToplevel(path string) (string, error) {
	r, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return "", fmt.Errorf("failed to detect a git directory in path %s of any of its parent dirs %w", path, err)
	}
	wt, err := r.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to return worktree for path %s %w", path, err)
	}

	return wt.Filesystem.Root(), nil
}

type ProviderMetadata struct {
	Name                      string                     `json:"name,omitempty"`
	DefaultTags               map[string]string          `json:"defaultTags,omitempty"`
	Filename                  string                     `json:"filename,omitempty"`
	StartLine                 int64                      `json:"startLine,omitempty"`
	EndLine                   int64                      `json:"endLine,omitempty"`
	AttributesWithUnknownKeys []AttributeWithUnknownKeys `json:"attributesWithUnknownKeys,omitempty"`
}

type AttributeWithUnknownKeys struct {
	Attribute        string   `json:"attribute"`
	MissingVariables []string `json:"missingVariables"`
}

// AddError pushes the provided error onto the metadata list. It does a naive
// conversion to ProjectDiag if the error provided is not already a diagnostic.
func (m *ProjectMetadata) AddError(err error) {
	var diag *ProjectDiag
	if errors.As(err, &diag) {
		m.Errors = append(m.Errors, diag)
		return
	}

	m.Errors = append(m.Errors, &ProjectDiag{Message: err.Error()})
}

func (m *ProjectMetadata) HasErrors() bool {
	return len(m.Errors) > 0
}

func (m *ProjectMetadata) IsEmptyProjectError() bool {
	if len(m.Errors) == 0 || len(m.Errors) > 1 {
		return false
	}

	return m.Errors[0].Code == diagEmptyPathType
}

// IsRunQuotaExceeded checks if any of the project diags are of type "run quota
// exceeded". If found it returns the associated message with this diag. This
// should be used when in any output that notifies the user.
func (m *ProjectMetadata) IsRunQuotaExceeded() (string, bool) {
	for _, diag := range m.Errors {
		if diag.Code == diagRunQuotaExceeded {
			return diag.Message, true
		}
	}

	return "", false
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
			logging.Logger.Debug().Msgf("Could not get absolute path for %s", m.Path)
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
	DisplayName          string
}

func (p *Project) AddProviderMetadata(metadatas []ProviderMetadata) {
	if p.Metadata == nil {
		p.Metadata = &ProjectMetadata{}
	}

	p.Metadata.Providers = metadatas
}

func NewProject(name string, metadata *ProjectMetadata) *Project {
	return &Project{
		Name:     name,
		Metadata: metadata,
		HasDiff:  true,
	}
}

// NameWithWorkspace returns the project Name appended with the parenthesized workspace name
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

	resources := make([]*Resource, 0, len(m))
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

	resources := make([]*PartialResource, 0, len(m))
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
		u := usageMap.Get(p.Address)
		r := BuildResource(p, u)
		seen[p] = r
		pastResources = append(pastResources, r)
	}

	for _, p := range p.PartialResources {
		r, ok := seen[p]
		if !ok {
			u := usageMap.Get(p.Address)
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
