package terraform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/infracost/infracost/internal/clierror"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/ui"
	"github.com/kballard/go-shellquote"
	"github.com/pkg/errors"
	"golang.org/x/mod/semver"

	log "github.com/sirupsen/logrus"
)

var minTerraformVer = "v0.12"

type DirProvider struct {
	ctx                 *config.ProjectContext
	Path                string
	spinnerOpts         ui.SpinnerOptions
	IsTerragrunt        bool
	PlanFlags           string
	Workspace           string
	UseState            bool
	TerraformBinary     string
	TerraformCloudHost  string
	TerraformCloudToken string
}

type RunShowOptions struct {
	CmdOptions *CmdOptions
}

func NewDirProvider(ctx *config.ProjectContext) schema.Provider {
	terraformBinary := ctx.ProjectConfig.TerraformBinary
	if terraformBinary == "" {
		terraformBinary = defaultTerraformBinary
	}

	return &DirProvider{
		ctx:  ctx,
		Path: ctx.ProjectConfig.Path,
		spinnerOpts: ui.SpinnerOptions{
			EnableLogging: ctx.RunContext.Config.IsLogging(),
			NoColor:       ctx.RunContext.Config.NoColor,
			Indent:        "  ",
		},
		PlanFlags:           ctx.ProjectConfig.TerraformPlanFlags,
		Workspace:           ctx.ProjectConfig.TerraformWorkspace,
		UseState:            ctx.ProjectConfig.TerraformUseState,
		TerraformBinary:     terraformBinary,
		TerraformCloudHost:  ctx.ProjectConfig.TerraformCloudHost,
		TerraformCloudToken: ctx.ProjectConfig.TerraformCloudToken,
	}
}

func (p *DirProvider) Type() string {
	return "terraform_dir"
}

func (p *DirProvider) DisplayType() string {
	return "Terraform directory"
}

func (p *DirProvider) checks() error {
	binary := p.TerraformBinary

	p.ctx.SetContextValue("terraformBinary", binary)

	_, err := exec.LookPath(binary)
	if err != nil {
		msg := fmt.Sprintf("Terraform binary \"%s\" could not be found.\nSet a custom Terraform binary in your Infracost config or using the environment variable INFRACOST_TERRAFORM_BINARY.", binary)
		return clierror.NewSanitizedError(errors.Errorf(msg), "Terraform binary could not be found")
	}

	out, err := exec.Command(binary, "-version").Output()
	if err != nil {
		msg := fmt.Sprintf("Could not get version of Terraform binary \"%s\"", binary)
		return clierror.NewSanitizedError(errors.Errorf(msg), "Could not get version of Terraform binary")
	}

	fullVersion := strings.SplitN(string(out), "\n", 2)[0]
	version := shortTerraformVersion(fullVersion)

	p.ctx.SetContextValue("terraformFullVersion", fullVersion)
	p.ctx.SetContextValue("terraformVersion", version)

	return checkTerraformVersion(version, fullVersion)
}

func (p *DirProvider) AddMetadata(metadata *schema.ProjectMetadata) {
	terraformWorkspace := p.Workspace

	if terraformWorkspace == "" {
		binary := p.TerraformBinary
		cmd := exec.Command(binary, "workspace", "show")
		cmd.Dir = p.Path

		out, err := cmd.Output()
		if err != nil {
			log.Debugf("Could not detect Terraform workspace for %s", p.Path)
		}
		terraformWorkspace = strings.Split(string(out), "\n")[0]
	}

	metadata.TerraformWorkspace = terraformWorkspace
}

func (p *DirProvider) LoadResources(usage map[string]*schema.UsageData) ([]*schema.Project, error) {
	projects := make([]*schema.Project, 0)
	var out []byte
	var err error

	if p.UseState {
		out, err = p.generateStateJSON()
	} else {
		out, err = p.generatePlanJSON()
	}
	if err != nil {
		return projects, err
	}

	jsons := [][]byte{out}
	if p.IsTerragrunt {
		jsons = bytes.Split(out, []byte{'\n'})
		if len(jsons) > 1 {
			jsons = jsons[:len(jsons)-1]
		}
	}

	for _, j := range jsons {
		metadata := config.DetectProjectMetadata(p.ctx.ProjectConfig.Path)
		metadata.Type = p.Type()
		p.AddMetadata(metadata)
		name := schema.GenerateProjectName(metadata, p.ctx.RunContext.Config.EnableDashboard)

		project := schema.NewProject(name, metadata)

		parser := NewParser(p.ctx)
		pastResources, resources, err := parser.parseJSON(j, usage)
		if err != nil {
			return projects, errors.Wrap(err, "Error parsing Terraform JSON")
		}

		project.HasDiff = !p.UseState
		if project.HasDiff {
			project.PastResources = pastResources
		}
		project.Resources = resources

		projects = append(projects, project)
	}

	return projects, nil
}

func (p *DirProvider) generatePlanJSON() ([]byte, error) {
	err := p.checks()
	if err != nil {
		return []byte{}, err
	}

	opts, err := p.buildCommandOpts(p.Path)
	if err != nil {
		return []byte{}, err
	}
	if opts.TerraformConfigFile != "" {
		defer os.Remove(opts.TerraformConfigFile)
	}

	spinner := ui.NewSpinner("Running terraform plan", p.spinnerOpts)
	planFile, planJSON, err := p.runPlan(opts, spinner, true)
	defer os.Remove(planFile)

	if err != nil {
		return []byte{}, err
	}

	if len(planJSON) > 0 {
		return planJSON, nil
	}

	spinner = ui.NewSpinner("Running terraform show", p.spinnerOpts)
	return p.runShow(opts, spinner, planFile)
}

func (p *DirProvider) generateStateJSON() ([]byte, error) {
	err := p.checks()
	if err != nil {
		return []byte{}, err
	}

	opts, err := p.buildCommandOpts(p.Path)
	if err != nil {
		return []byte{}, err
	}
	if opts.TerraformConfigFile != "" {
		defer os.Remove(opts.TerraformConfigFile)
	}

	spinner := ui.NewSpinner("Running terraform show", p.spinnerOpts)
	return p.runShow(opts, spinner, "")
}

func (p *DirProvider) buildCommandOpts(path string) (*CmdOptions, error) {
	opts := &CmdOptions{
		TerraformBinary:    p.TerraformBinary,
		TerraformWorkspace: p.Workspace,
		Dir:                path,
	}

	cfgFile, err := CreateConfigFile(p.Path, p.TerraformCloudHost, p.TerraformCloudToken)
	if err != nil {
		return opts, err
	}

	opts.TerraformConfigFile = cfgFile

	return opts, nil
}

func (p *DirProvider) runPlan(opts *CmdOptions, spinner *ui.Spinner, initOnFail bool) (string, []byte, error) {
	var planJSON []byte

	fileName := ".tfplan-" + uuid.New().String()
	// For Terragrunt we need a relative path
	if !p.IsTerragrunt {
		fileName = filepath.Join(os.TempDir(), fileName)
	}

	flags, err := shellquote.Split(p.PlanFlags)
	if err != nil {
		return "", planJSON, errors.Wrap(err, "Error parsing terraform plan flags")
	}

	args := []string{}
	if p.IsTerragrunt {
		args = append(args, "run-all", "--terragrunt-ignore-external-dependencies")
	}

	args = append(args, "plan", "-input=false", "-lock=false", "-no-color")
	args = append(args, flags...)
	_, err = Cmd(opts, append(args, fmt.Sprintf("-out=%s", fileName))...)

	// Check if the error requires a remote run or an init
	if err != nil {
		extractedErr := extractStderr(err)

		// If the plan returns this error then Terraform is configured with remote execution mode
		if strings.HasPrefix(extractedErr, "Error: Saving a generated plan is currently not supported") {
			log.Info("Continuing with Terraform Remote Execution Mode")
			p.ctx.SetContextValue("terraformRemoteExecutionModeEnabled", true)
			planJSON, err = p.runRemotePlan(opts, args)
		} else if initOnFail && (strings.Contains(extractedErr, "Error: Could not load plugin") ||
			strings.Contains(extractedErr, "Error: Initialization required") ||
			strings.Contains(extractedErr, "Error: Backend initialization required") ||
			strings.Contains(extractedErr, "Error: Provider requirements cannot be satisfied by locked dependencies") ||
			strings.Contains(extractedErr, "Error: Module not installed")) {
			spinner.Stop()
			err = p.runInit(opts, ui.NewSpinner("Running terraform init", p.spinnerOpts))
			if err != nil {
				return "", planJSON, err
			}
			return p.runPlan(opts, spinner, false)
		}
	}

	if err != nil {
		spinner.Fail()

		if errors.Is(err, ErrMissingCloudToken) {
			msg := "Please set your TERRAFORM_CLOUD_TOKEN environment variable.\n"
			msg += "It seems like Terraform Cloud's Remote Execution Mode is being used.\n"
			msg += "Create a Team or User API Token in the Terraform Cloud dashboard and set this environment variable."
			fmt.Fprintln(os.Stderr, msg)
		} else if errors.Is(err, ErrInvalidCloudToken) {
			msg := "Please set your TERRAFORM_CLOUD_TOKEN environment variable.\n"
			msg += "It seems like Terraform Cloud's Remote Execution Mode is being used.\n"
			msg += "Create a Team or User API Token in the Terraform Cloud dashboard and set this environment variable."
			fmt.Fprintln(os.Stderr, msg)
		} else {
			printTerraformErr(err)
		}
		return "", planJSON, errors.Wrap(err, "Error running terraform plan")
	}

	spinner.Success()

	return fileName, planJSON, nil
}

func (p *DirProvider) runInit(opts *CmdOptions, spinner *ui.Spinner) error {
	args := []string{}
	if p.IsTerragrunt {
		args = append(args, "run-all", "--terragrunt-ignore-external-dependencies")
	}
	args = append(args, "init", "-input=false", "-no-color")

	_, err := Cmd(opts, args...)
	if err != nil {
		spinner.Fail()
		printTerraformErr(err)
		return errors.Wrap(err, "Error running terraform init")
	}

	spinner.Success()
	return nil
}

func (p *DirProvider) runRemotePlan(opts *CmdOptions, args []string) ([]byte, error) {
	if p.TerraformCloudToken == "" && !checkCloudConfigSet() {
		return []byte{}, ErrMissingCloudToken
	}

	stdout, err := Cmd(opts, args...)
	if err != nil {
		return []byte{}, err
	}

	r := regexp.MustCompile(`To view this run in a browser, visit:\n(.*)`)
	matches := r.FindAllStringSubmatch(string(stdout), 1)
	if len(matches) == 0 || len(matches[0]) <= 1 {
		return []byte{}, errors.New("Could not parse the remote run URL")
	}

	u, err := url.Parse(matches[0][1])
	if err != nil {
		return []byte{}, err
	}
	host := u.Host
	s := strings.Split(u.Path, "/")
	runID := s[len(s)-1]

	token := p.TerraformCloudToken
	if token == "" {
		token = findCloudToken(host)
	}
	if token == "" {
		return []byte{}, ErrMissingCloudToken
	}

	body, err := cloudAPI(host, fmt.Sprintf("/api/v2/runs/%s/plan", runID), token)
	if err != nil {
		return []byte{}, err
	}

	var parsedResp struct {
		Data struct {
			Links map[string]string
		}
	}
	if json.Unmarshal(body, &parsedResp) != nil {
		return []byte{}, err
	}

	jsonPath, ok := parsedResp.Data.Links["json-output"]
	if !ok || jsonPath == "" {
		return []byte{}, errors.New("Could not parse path to plan JSON from remote")
	}
	return cloudAPI(host, jsonPath, token)
}

func (p *DirProvider) runShow(opts *CmdOptions, spinner *ui.Spinner, planFile string) ([]byte, error) {
	args := []string{"show", "-no-color", "-json"}
	if planFile != "" {
		args = append(args, planFile)
	}
	out, err := Cmd(opts, args...)
	if err != nil {
		spinner.Fail()
		printTerraformErr(err)
		return []byte{}, errors.Wrap(err, "Error running terraform show")
	}
	spinner.Success()

	return out, nil
}

func IsTerraformDir(path string) bool {
	for _, ext := range []string{"tf", "hcl", "hcl.json", "tf.json"} {
		matches, err := filepath.Glob(filepath.Join(path, fmt.Sprintf("*.%s", ext)))
		if matches != nil && err == nil {
			return true
		}
	}
	return false
}

func shortTerraformVersion(full string) string {
	p := strings.Split(full, " ")
	if len(p) > 1 {
		return p[len(p)-1]
	}

	return ""
}

func checkTerraformVersion(v string, fullV string) error {
	if strings.HasPrefix(fullV, "Terraform ") && semver.Compare(v, minTerraformVer) < 0 {
		return errors.Errorf("Terraform %s is not supported. Please use Terraform version >= %s.", v, minTerraformVer)
	}

	if strings.HasPrefix(fullV, "terragrunt") && semver.Compare(v, minTerragruntVer) < 0 {
		return errors.Errorf("Terragrunt %s is not supported. Please use Terragrunt version >= %s.", v, minTerragruntVer)
	}

	// Allow any non-terraform and non-terragrunt binaries
	return nil
}

func printTerraformErr(err error) {
	stderr := extractStderr(err)
	if stderr == "" {
		return
	}

	msg := fmt.Sprintf("\n  Terraform command failed with:\n%s\n", ui.Indent(stderr, "    "))

	if strings.HasPrefix(stderr, "Error: Failed to select workspace") {
		msg += "\nRun `terraform workspace select your_workspace` first or set the TF_WORKSPACE environment variable.\n"
	}
	if strings.HasPrefix(stderr, "Error: Required token could not be found") {
		msg += "\nRun `terraform login` first or set the TF_CLI_CONFIG_FILE environment variable to the ABSOLUTE path.\n"
	}
	if strings.HasPrefix(stderr, "Error: No value for required variable") {
		msg += "\nPass Terraform flags using the --terraform-plan-flags option.\n"
		msg += "For example: infracost --path=path/to/terraform --terraform-plan-flags=\"-var-file=my.tfvars\"\n"
	}
	if strings.HasPrefix(stderr, "Error: Failed to read variables file") {
		msg += "\nSpecify the -var-file flag as a path relative to your Terraform directory.\n"
		msg += "For example: infracost --path=path/to/terraform --terraform-plan-flags=\"-var-file=my.tfvars\"\n"
	}

	fmt.Fprintln(os.Stderr, msg)
}

func extractStderr(err error) string {
	if e, ok := err.(*CmdError); ok {
		return stripBlankLines(string(e.Stderr))
	}
	return ""
}

func stripBlankLines(s string) string {
	return regexp.MustCompile(`[\t\r\n]+`).ReplaceAllString(strings.TrimSpace(s), "\n")
}
