package terraform

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/events"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/ui"
	"github.com/kballard/go-shellquote"
	"github.com/pkg/errors"
	"golang.org/x/mod/semver"

	log "github.com/sirupsen/logrus"
)

var minTerraformVer = "v0.12"

type terraformProvider struct {
	env                 *config.Environment
	spinnerOpts         ui.SpinnerOptions
	binary              string
	dir                 string
	projectName         string
	workspace           string
	jsonFile            string
	planFile            string
	planFlags           string
	useState            bool
	terraformCloudHost  string
	terraformCloudToken string
}

// New returns new Terraform Provider.
func New(cfg *config.Config, projectCfg *config.TerraformProject) schema.Provider {
	binary := projectCfg.Binary
	if binary == "" {
		binary = defaultTerraformBinary
	}

	dir := projectCfg.Dir
	if dir == "" {
		dir = "."
	}

	return &terraformProvider{
		env: cfg.Environment,
		spinnerOpts: ui.SpinnerOptions{
			EnableLogging: cfg.IsLogging(),
			NoColor:       cfg.NoColor,
			Indent:        "  ",
		},
		binary:              binary,
		dir:                 dir,
		projectName:         projectCfg.DisplayName(),
		workspace:           projectCfg.Workspace,
		jsonFile:            projectCfg.JSONFile,
		planFile:            projectCfg.PlanFile,
		planFlags:           projectCfg.PlanFlags,
		useState:            projectCfg.UseState,
		terraformCloudHost:  projectCfg.TerraformCloudHost,
		terraformCloudToken: projectCfg.TerraformCloudToken,
	}
}

func (p *terraformProvider) LoadResources(usage map[string]*schema.UsageData) (*schema.Project, error) {
	var project *schema.Project = schema.NewProject(p.projectName)

	var err error
	var j []byte

	if p.useState {
		j, err = p.generateStateJSON()
	} else {
		j, err = p.loadPlanJSON()
	}
	if err != nil {
		return project, err
	}

	parser := NewParser(p.env)
	pastResources, resources, err := parser.parseJSON(j, usage)
	if err != nil {
		return project, errors.Wrap(err, "Error parsing Terraform JSON")
	}

	project.HasDiff = !p.useState
	if project.HasDiff {
		project.PastResources = pastResources
	}
	project.Resources = resources

	return project, nil
}

func (p *terraformProvider) loadPlanJSON() ([]byte, error) {
	if p.jsonFile == "" {
		return p.generatePlanJSON()
	}

	out, err := ioutil.ReadFile(p.jsonFile)
	if err != nil {
		return []byte{}, errors.Wrap(err, "Error reading Terraform plan file")
	}

	return out, nil
}

func (p *terraformProvider) generateStateJSON() ([]byte, error) {
	err := p.terraformPreChecks()
	if err != nil {
		return []byte{}, err
	}

	opts, err := p.setupOpts()
	if err != nil {
		return []byte{}, err
	}
	if opts.TerraformConfigFile != "" {
		defer os.Remove(opts.TerraformConfigFile)
	}

	return p.runShow(opts, "")
}

func (p *terraformProvider) generatePlanJSON() ([]byte, error) {
	err := p.terraformPreChecks()
	if err != nil {
		return []byte{}, err
	}

	opts, err := p.setupOpts()
	if err != nil {
		return []byte{}, err
	}
	if opts.TerraformConfigFile != "" {
		defer os.Remove(opts.TerraformConfigFile)
	}

	if p.planFile == "" {

		var planJSON []byte
		p.planFile, planJSON, err = p.runPlan(opts, p.planFlags, true)
		defer os.Remove(p.planFile)

		if err != nil {
			return []byte{}, err
		}

		if len(planJSON) > 0 {
			return planJSON, nil
		}
	}

	return p.runShow(opts, p.planFile)
}

func (p *terraformProvider) setupOpts() (*CmdOptions, error) {
	opts := &CmdOptions{
		TerraformBinary:    p.binary,
		TerraformWorkspace: p.workspace,
		TerraformDir:       p.dir,
	}

	cfgFile, err := CreateConfigFile(p.dir, p.terraformCloudHost, p.terraformCloudToken)
	if err != nil {
		return opts, err
	}

	opts.TerraformConfigFile = cfgFile

	return opts, nil
}

func (p *terraformProvider) terraformPreChecks() error {
	if p.jsonFile == "" {
		_, err := exec.LookPath(p.binary)
		if err != nil {
			msg := fmt.Sprintf("Terraform binary \"%s\" could not be found.\nSet a custom Terraform binary in your Infracost config or using the environment variable TERRAFORM_BINARY.", p.binary)
			return events.NewError(errors.Errorf(msg), "Terraform binary could not be found")
		}

		if v, ok := checkVersion(p.env); !ok {
			return errors.Errorf("Terraform %s is not supported. Please use Terraform version >= %s.", v, minTerraformVer)
		}

		if !p.inTerraformDir() {
			msg := fmt.Sprintf("Directory \"%s\" does not have any Terraform files.\nSet the Terraform directory path using the --terraform-dir option.", p.dir)
			return events.NewError(errors.Errorf(msg), "Directory does not have any Terraform files")

		}
	}
	return nil
}

func (p *terraformProvider) inTerraformDir() bool {
	for _, ext := range []string{"tf", "hcl", "hcl.json"} {
		matches, err := filepath.Glob(filepath.Join(p.dir, fmt.Sprintf("*.%s", ext)))
		if matches != nil && err == nil {
			return true
		}
	}
	return false
}

func (p *terraformProvider) runInit(opts *CmdOptions) error {
	spinner := ui.NewSpinner("Running terraform init", p.spinnerOpts)

	_, err := Cmd(opts, "init", "-input=false", "-no-color")
	if err != nil {
		spinner.Fail()
		terraformError(err)
		return errors.Wrap(err, "Error running terraform init")
	}

	spinner.Success()
	return nil
}

func (p *terraformProvider) runPlan(opts *CmdOptions, planFlags string, initOnFail bool) (string, []byte, error) {
	spinner := ui.NewSpinner("Running terraform plan", p.spinnerOpts)
	var planJSON []byte

	f, err := ioutil.TempFile(os.TempDir(), "tfplan")
	if err != nil {
		spinner.Fail()
		return "", planJSON, errors.Wrap(err, "Error creating temporary file 'tfplan'")
	}

	flags, err := shellquote.Split(planFlags)
	if err != nil {
		return "", planJSON, errors.Wrap(err, "Error parsing terraform plan flags")
	}

	args := []string{"plan", "-input=false", "-lock=false", "-no-color"}
	args = append(args, flags...)
	_, err = Cmd(opts, append(args, fmt.Sprintf("-out=%s", f.Name()))...)

	// Check if the error requires a remote run or an init
	if err != nil {
		extractedErr := extractStderr(err)

		// If the plan returns this error then Terraform is configured with remote execution mode
		if strings.HasPrefix(extractedErr, "Error: Saving a generated plan is currently not supported") {
			log.Info("Continuing with Terraform Remote Execution Mode")
			p.env.TerraformRemoteExecutionModeEnabled = true
			planJSON, err = p.runRemotePlan(opts, args)
		} else if initOnFail && (strings.Contains(extractedErr, "Error: Could not load plugin") ||
			strings.Contains(extractedErr, "Error: Initialization required") ||
			strings.Contains(extractedErr, "Error: Module not installed") ||
			strings.Contains(extractedErr, "Error: Provider requirements cannot be satisfied by locked dependencies") ||
			strings.Contains(extractedErr, "Error: Module not installed")) {
			spinner.Stop()
			err = p.runInit(opts)
			if err != nil {
				return "", planJSON, err
			}
			return p.runPlan(opts, planFlags, false)
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
			terraformError(err)
		}
		return "", planJSON, errors.Wrap(err, "Error running terraform plan")
	}

	spinner.Success()

	return f.Name(), planJSON, nil
}

func (p *terraformProvider) runRemotePlan(opts *CmdOptions, args []string) ([]byte, error) {
	if p.terraformCloudToken == "" && !checkCloudConfigSet() {
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

	token := p.terraformCloudToken
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

func (p *terraformProvider) runShow(opts *CmdOptions, planFile string) ([]byte, error) {
	spinner := ui.NewSpinner("Running terraform show", p.spinnerOpts)

	args := []string{"show", "-no-color", "-json"}
	if planFile != "" {
		args = append(args, planFile)
	}
	out, err := Cmd(opts, args...)
	if err != nil {
		spinner.Fail()
		terraformError(err)
		return []byte{}, errors.Wrap(err, "Error running terraform show")
	}
	spinner.Success()

	return out, nil
}

func checkVersion(env *config.Environment) (string, bool) {
	v := env.TerraformVersion

	// Allow any non-terraform binaries, e.g. terragrunt
	if !strings.HasPrefix(env.TerraformFullVersion, "Terraform ") {
		return v, true
	}

	return v, semver.Compare(v, minTerraformVer) >= 0
}

func terraformError(err error) {
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
		msg += "For example: infracost --terraform-dir=path/to/terraform --terraform-plan-flags=\"-var-file=myvars.tfvars\"\n"
	}
	if strings.HasPrefix(stderr, "Error: Failed to read variables file") {
		msg += "\nSpecify the -var-file flag as a path relative to your Terraform directory.\n"
		msg += "For example: infracost --terraform-dir=path/to/terraform --terraform-plan-flags=\"-var-file=myvars.tfvars\"\n"
	}
	if strings.HasPrefix(stderr, "Terraform couldn't read the given file as a project or plan file.") {
		msg += "\nSpecify the --terraform-plan-file flag as a path relative to your Terraform directory.\n"
		msg += "For example: infracost --terraform-dir=path/to/terraform --terraform-plan-file=plan.save\n"
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
