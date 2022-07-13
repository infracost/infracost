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
	"github.com/kballard/go-shellquote"
	"github.com/pkg/errors"
	"golang.org/x/mod/semver"

	"github.com/infracost/infracost/internal/clierror"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/credentials"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/ui"

	log "github.com/sirupsen/logrus"
)

var minTerraformVer = "v0.12"

type DirProvider struct {
	ctx                  *config.ProjectContext
	Path                 string
	spinnerOpts          ui.SpinnerOptions
	IsTerragrunt         bool
	PlanFlags            string
	InitFlags            string
	Workspace            string
	UseState             bool
	TerraformBinary      string
	TerraformCloudHost   string
	TerraformCloudToken  string
	Env                  map[string]string
	cachedStateJSON      []byte
	cachedPlanJSON       []byte
	includePastResources bool
}

type RunShowOptions struct {
	CmdOptions *CmdOptions
}

func NewDirProvider(ctx *config.ProjectContext, includePastResources bool) schema.Provider {
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
		PlanFlags:            ctx.ProjectConfig.TerraformPlanFlags,
		InitFlags:            ctx.ProjectConfig.TerraformInitFlags,
		Workspace:            ctx.ProjectConfig.TerraformWorkspace,
		UseState:             ctx.ProjectConfig.TerraformUseState,
		TerraformBinary:      terraformBinary,
		TerraformCloudHost:   ctx.ProjectConfig.TerraformCloudHost,
		TerraformCloudToken:  ctx.ProjectConfig.TerraformCloudToken,
		Env:                  ctx.ProjectConfig.Env,
		includePastResources: includePastResources,
	}
}

func (p *DirProvider) Type() string {
	return "terraform_cli"
}

func (p *DirProvider) DisplayType() string {
	return "Terraform CLI"
}

func (p *DirProvider) checks() error {
	binary := p.TerraformBinary

	p.ctx.SetContextValue("terraformBinary", binary)

	_, err := exec.LookPath(binary)
	if err != nil {
		msg := fmt.Sprintf("Terraform binary '%s' could not be found. You have two options:\n", binary)
		msg += "1. Set a custom Terraform binary using the environment variable INFRACOST_TERRAFORM_BINARY.\n\n"
		msg += fmt.Sprintf("2. Set --path to a Terraform plan JSON file. See %s for how to generate this.", ui.LinkString("https://infracost.io/troubleshoot"))
		return clierror.NewCLIError(errors.Errorf(msg), "Terraform binary could not be found")
	}

	out, err := exec.Command(binary, "-version").Output()
	if err != nil {
		msg := fmt.Sprintf("Could not get version of Terraform binary '%s'", binary)
		return clierror.NewCLIError(errors.Errorf(msg), "Could not get version of Terraform binary")
	}

	fullVersion := strings.SplitN(string(out), "\n", 2)[0]
	version := shortTerraformVersion(fullVersion)

	p.ctx.SetContextValue("terraformFullVersion", fullVersion)
	p.ctx.SetContextValue("terraformVersion", version)

	return checkTerraformVersion(version, fullVersion)
}

func (p *DirProvider) AddMetadata(metadata *schema.ProjectMetadata) {
	basePath := p.ctx.ProjectConfig.Path
	if p.ctx.RunContext.Config.ConfigFilePath != "" {
		basePath = filepath.Dir(p.ctx.RunContext.Config.ConfigFilePath)
	}

	modulePath, err := filepath.Rel(basePath, metadata.Path)
	if err == nil && modulePath != "" && modulePath != "." {
		log.Debugf("Calculated relative terraformModulePath for %s from %s", basePath, metadata.Path)
		metadata.TerraformModulePath = modulePath
	}

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

	spinner := ui.NewSpinner("Extracting only cost-related params from terraform", ui.SpinnerOptions{
		EnableLogging: p.ctx.RunContext.Config.IsLogging(),
		NoColor:       p.ctx.RunContext.Config.NoColor,
		Indent:        "  ",
	})
	defer spinner.Fail()

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
		name := schema.GenerateProjectName(metadata, p.ctx.ProjectConfig.Name, p.ctx.RunContext.IsCloudEnabled())

		project := schema.NewProject(name, metadata)

		parser := NewParser(p.ctx, p.includePastResources)
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

	spinner.Success()
	return projects, nil
}

func (p *DirProvider) generatePlanJSON() ([]byte, error) {
	if p.cachedPlanJSON != nil {
		return p.cachedPlanJSON, nil
	}

	if UsePlanCache(p) {
		spinner := ui.NewSpinner("Checking for cached plan...", p.spinnerOpts)
		defer spinner.Fail()

		cached, err := ReadPlanCache(p)
		if err != nil {
			spinner.SuccessWithMessage(fmt.Sprintf("Checking for cached plan... %v", err.Error()))
		} else {
			p.cachedPlanJSON = cached
			spinner.SuccessWithMessage("Checking for cached plan... found")
			return p.cachedPlanJSON, nil
		}
	}

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
	defer spinner.Fail()

	planFile, planJSON, err := p.runPlan(opts, spinner, true)
	defer os.Remove(planFile)

	if err != nil {
		return []byte{}, err
	}

	if len(planJSON) > 0 {
		return planJSON, nil
	}

	spinner = ui.NewSpinner("Running terraform show", p.spinnerOpts)
	j, err := p.runShow(opts, spinner, planFile)
	if err == nil {
		p.cachedPlanJSON = j
		if UsePlanCache(p) {
			// Note we check UsePlanCache again because we have discovered we're using remote execution inside p.runPlan
			WritePlanCache(p, j)
		}
	}
	return j, err
}

func (p *DirProvider) generateStateJSON() ([]byte, error) {
	if p.cachedStateJSON != nil {
		return p.cachedStateJSON, nil
	}

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
	defer spinner.Fail()

	j, err := p.runShow(opts, spinner, "")
	if err == nil {
		p.cachedStateJSON = j
	}
	return j, err
}

func (p *DirProvider) buildCommandOpts(path string) (*CmdOptions, error) {
	opts := &CmdOptions{
		TerraformBinary:    p.TerraformBinary,
		TerraformWorkspace: p.Workspace,
		Dir:                path,
		Env:                p.Env,
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
			strings.Contains(extractedErr, "Error: Required plugins are not installed") ||
			strings.Contains(extractedErr, "Error: Initialization required") ||
			strings.Contains(extractedErr, "Error: Backend initialization required") ||
			strings.Contains(extractedErr, "Error: Provider requirements cannot be satisfied by locked dependencies") ||
			strings.Contains(extractedErr, "Error: Inconsistent dependency lock file") ||
			strings.Contains(extractedErr, "Error: Module not installed") ||
			strings.Contains(extractedErr, "Error: Terraform Cloud initialization required") ||
			strings.Contains(extractedErr, "please run \"terraform init\"")) {
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
		err = p.buildTerraformErr(err, false)

		cmdName := "terraform plan"
		if p.IsTerragrunt {
			cmdName = "terragrunt run-all plan"
		}
		msg := fmt.Sprintf("%s failed", cmdName)
		return "", planJSON, clierror.NewCLIError(fmt.Errorf("%s: %s", msg, err), msg)
	}

	spinner.Success()

	return fileName, planJSON, nil
}

func (p *DirProvider) runInit(opts *CmdOptions, spinner *ui.Spinner) error {
	args := []string{}
	if p.IsTerragrunt {
		args = append(args, "run-all", "--terragrunt-ignore-external-dependencies")
	}

	flags, err := shellquote.Split(p.InitFlags)
	if err != nil {
		msg := "parsing terraform-init-flags failed"
		return clierror.NewCLIError(fmt.Errorf("%s: %s", msg, err), msg)
	}

	args = append(args, "init", "-input=false", "-no-color")
	args = append(args, flags...)

	if config.IsTest() {
		args = append(args, "-upgrade")
	}

	_, err = Cmd(opts, args...)
	if err != nil {
		spinner.Fail()
		err = p.buildTerraformErr(err, true)

		cmdName := "terraform init"
		if p.IsTerragrunt {
			cmdName = "terragrunt run-all init"
		}
		msg := fmt.Sprintf("%s failed", cmdName)
		return clierror.NewCLIError(fmt.Errorf("%s: %s", msg, err), msg)
	}

	spinner.Success()
	return nil
}

func (p *DirProvider) runRemotePlan(opts *CmdOptions, args []string) ([]byte, error) {
	if p.TerraformCloudToken == "" && !credentials.CheckCloudConfigSet() {
		return []byte{}, credentials.ErrMissingCloudToken
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
		token = credentials.FindTerraformCloudToken(host)
	}
	if token == "" {
		return []byte{}, credentials.ErrMissingCloudToken
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
		err = p.buildTerraformErr(err, false)

		cmdName := "terraform show"
		if p.IsTerragrunt {
			cmdName = "terragrunt show"
		}
		msg := fmt.Sprintf("%s failed", cmdName)
		return []byte{}, clierror.NewCLIError(fmt.Errorf("%s: %s", msg, err), msg)
	}
	spinner.Success()

	return out, nil
}

func IsTerraformDir(path string) bool {
	for _, ext := range []string{"tf", "tf.json"} {
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
	if len(v) > 0 && v[0] != 'v' {
		// The semver package requires a 'v' prefix to do a proper Compare.
		v = "v" + v
	}

	if strings.HasPrefix(fullV, "Terraform ") && semver.Compare(v, minTerraformVer) < 0 {
		return fmt.Errorf("Terraform %s is not supported. Please use Terraform version >= %s. Update it or set the environment variable INFRACOST_TERRAFORM_BINARY.", v, minTerraformVer) //nolint
	}

	if strings.HasPrefix(fullV, "terragrunt") && semver.Compare(v, minTerragruntVer) < 0 {
		return fmt.Errorf("Terragrunt %s is not supported. Please use Terragrunt version >= %s. Update it or set the environment variable INFRACOST_TERRAFORM_BINARY.", v, minTerragruntVer) //nolint
	}

	// Allow any non-terraform and non-terragrunt binaries
	return nil
}

func (p *DirProvider) buildTerraformErr(err error, isInit bool) error {
	stderr := extractStderr(err)

	binName := "Terraform"
	if p.IsTerragrunt {
		binName = "Terragrunt"
	}

	msg := ""

	if stderr != "" {
		msg += fmt.Sprintf("\n\n  %s command failed with:\n%s", binName, ui.Indent(stderr, "    "))
	}

	if strings.HasPrefix(stderr, "Error: Failed to select workspace") {
		msg += "\n\nYou have two options:\n"
		msg += "1. Run `terraform workspace select your_workspace` first or set the TF_WORKSPACE environment variable.\n\n"
		msg += fmt.Sprintf("2. Set --path to a Terraform plan JSON file. See %s for how to generate this.", ui.LinkString("https://infracost.io/troubleshoot"))
	} else if errors.Is(err, credentials.ErrMissingCloudToken) || errors.Is(err, credentials.ErrInvalidCloudToken) || strings.HasPrefix(stderr, "Error: Required token could not be found") {
		msg += "\n\nYou have two options:\n"
		msg += "1. Run `terraform login` or set the INFRACOST_TERRAFORM_CLOUD_TOKEN environment variable.'\n\n"
		msg += fmt.Sprintf("2. Set --path to a Terraform plan JSON file. See %s for how to generate this.", ui.LinkString("https://infracost.io/troubleshoot"))
	} else if strings.HasPrefix(stderr, "Error: No value for required variable") {
		msg += "\n\nYou have two options:\n"
		msg += "1. Pass the variables using the --terraform-plan-flags option,\n     e.g. --terraform-plan-flags=\"-var-file=my.tfvars\"\n\n"
		msg += fmt.Sprintf("2. Set --path to a Terraform plan JSON file. See %s for how to generate this.", ui.LinkString("https://infracost.io/troubleshoot"))
	} else if strings.HasPrefix(stderr, "Error: Failed to read variables file") {
		msg += "\n\nYou have two options:\n"
		msg += "1. Ensure the variable file is passed relative to your Terraform directory,\n     e.g. --path=path/to/code --terraform-plan-flags=\"-var-file=my.tfvars\"\n\n"
		msg += fmt.Sprintf("2. Set --path to a Terraform plan JSON file. See %s for how to generate this.", ui.LinkString("https://infracost.io/troubleshoot"))
	} else if strings.HasPrefix(stderr, "Error: error configuring Terraform AWS Provider: no valid credential sources for Terraform AWS Provider found.") {
		msg += "\n\nTerraform requires AWS credentials to be set. You have two options:\n"
		msg += fmt.Sprintf("1. See %s for details of how to set credentials.\n\n", ui.LinkString("https://registry.terraform.io/providers/hashicorp/aws/latest/docs#authentication"))
		msg += fmt.Sprintf("2. Set --path to a Terraform plan JSON file. See %s for how to generate this.", ui.LinkString("https://infracost.io/troubleshoot"))
	} else if p.IsTerragrunt {
		msg += fmt.Sprintf("\n\nSee %s for how to generate multiple Terraform plan JSON files for your Terragrunt project, and use them with Infracost.", ui.LinkString("https://infracost.io/troubleshoot"))
	} else if isInit {
		msg += fmt.Sprintf("\n\nTry using --terraform-init-flags to pass any required Terraform init flags, or skip Terraform init entirely and set the --path to a Terraform plan JSON file. See %s for how to generate this.", ui.LinkString("https://infracost.io/troubleshoot"))
	} else if strings.HasPrefix(stderr, "Error: Unsupported Terraform Core version") {
		msg += "\n\nUpdate Terraform to the required version or set a custom Terraform binary using the environment variable INFRACOST_TERRAFORM_BINARY."
	} else {
		msg += fmt.Sprintf("\n\nTry setting the --path to a Terraform plan JSON file. See %s for how to generate this.", ui.LinkString("https://infracost.io/troubleshoot"))
	}

	return fmt.Errorf("%v%s", err, msg)
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
