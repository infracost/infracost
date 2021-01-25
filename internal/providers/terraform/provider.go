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

	"github.com/fatih/color"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/events"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/spin"
	"github.com/kballard/go-shellquote"
	"github.com/pkg/errors"
	"golang.org/x/mod/semver"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var minTerraformVer = "v0.12"

type terraformProvider struct {
	dir       string
	jsonFile  string
	planFile  string
	useState  bool
	planFlags string
	usageFile string
}

// New returns new Terraform Provider.
func New() schema.Provider {
	return &terraformProvider{}
}

func (p *terraformProvider) ProcessArgs(c *cli.Context) error {
	p.dir = c.String("tfdir")
	p.jsonFile = c.String("tfjson")
	p.planFile = c.String("tfplan")
	p.useState = c.Bool("use-tfstate")
	p.planFlags = c.String("tfflags")
	p.usageFile = c.String("usage-file")

	if p.jsonFile != "" && p.planFile != "" {
		return errors.New("Please provide either a Terraform Plan JSON file (tfjson) or a Terraform Plan file (tfplan)")
	}

	return nil
}

func (p *terraformProvider) LoadResources(usage map[string]*schema.UsageData) ([]*schema.Resource, error) {
	var resources []*schema.Resource

	var err error
	var j []byte

	if p.useState {
		j, err = p.generateStateJSON()
	} else {
		j, err = p.loadPlanJSON()
	}
	if err != nil {
		return []*schema.Resource{}, err
	}

	resources, err = parseJSON(j, usage)
	if err != nil {
		return resources, errors.Wrap(err, "Error parsing Terraform JSON")
	}

	return resources, nil
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

	return runShow(opts, "")
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
		p.planFile, planJSON, err = runPlan(opts, p.planFlags)
		defer os.Remove(p.planFile)

		if err != nil {
			return []byte{}, err
		}

		if len(planJSON) > 0 {
			return planJSON, nil
		}
	}

	return runShow(opts, p.planFile)
}

func (p *terraformProvider) setupOpts() (*CmdOptions, error) {
	opts := &CmdOptions{
		TerraformDir: p.dir,
	}

	configFile, err := CreateConfigFile(p.dir)
	if err != nil {
		return opts, err
	}

	opts.TerraformConfigFile = configFile

	return opts, nil
}

func (p *terraformProvider) terraformPreChecks() error {
	if p.jsonFile == "" {
		_, err := exec.LookPath(config.Environment.TerraformBinary)
		if err != nil {
			msg := fmt.Sprintf("Terraform binary \"%s\" could not be found.\nSet a custom Terraform binary using the environment variable TERRAFORM_BINARY.", config.Environment.TerraformBinary)
			return events.NewError(errors.Errorf(msg), "Terraform binary could not be found")
		}

		if v, ok := checkVersion(); !ok {
			return errors.Errorf("Terraform %s is not supported. Please use Terraform version >= %s.", v, minTerraformVer)
		}

		if !p.inTerraformDir() {
			msg := fmt.Sprintf("Directory \"%s\" does not have any Terraform files.\nSet the Terraform directory path using the --tfdir option.", p.dir)
			return events.NewError(errors.Errorf(msg), "Directory does not have any Terraform files")

		}
	}
	return nil
}

func checkVersion() (string, bool) {
	v := config.Environment.TerraformVersion

	// Allow any non-terraform binaries, e.g. terragrunt
	if !strings.HasPrefix(config.Environment.TerraformFullVersion, "Terraform ") {
		return v, true
	}

	return v, semver.Compare(v, minTerraformVer) >= 0
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

func runInit(opts *CmdOptions) error {
	spinner := spin.NewSpinner("Running terraform init")

	_, err := Cmd(opts, "init", "-input=false", "-no-color")
	if err != nil {
		spinner.Fail()
		terraformError(err)
		return errors.Wrap(err, "Error running terraform init")
	}

	spinner.Success()
	return nil
}

func runPlan(opts *CmdOptions, planFlags string) (string, []byte, error) {
	spinner := spin.NewSpinner("Running terraform plan")
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

	if err != nil {
		extractedErr := extractStderr(err)

		// If the plan returns this error then Terraform is configured with remote execution mode
		if strings.HasPrefix(extractedErr, "Error: Saving a generated plan is currently not supported") {
			log.Info("Continuing with Terraform Remote Execution Mode")
			config.Environment.TerraformRemoteExecutionModeEnabled = true
			planJSON, err = runRemotePlan(opts, args)
		} else if strings.Contains(extractedErr, "Error: Could not load plugin") ||
			strings.Contains(extractedErr, "Error: Initialization required") {
			spinner.Stop()
			err = runInit(opts)
			if err != nil {
				return "", planJSON, err
			}
			return runPlan(opts, planFlags)
		}

		spinner.Fail()

		red := color.New(color.FgHiRed)
		bold := color.New(color.Bold, color.FgHiWhite)

		if errors.Is(err, ErrMissingCloudToken) {
			msg := fmt.Sprintf("\n%s %s %s\n%s\n%s\n",
				red.Sprint("Please set your"),
				bold.Sprint("TERRAFORM_CLOUD_TOKEN"),
				red.Sprint("environment variable."),
				"It seems like Terraform Cloud's Remote Execution Mode is being used.",
				"Create a Team or User API Token in the Terraform Cloud dashboard and set this environment variable.",
			)
			fmt.Fprintln(os.Stderr, msg)
		} else if errors.Is(err, ErrInvalidCloudToken) {
			msg := fmt.Sprintf("\n%s %s %s\n%s\n%s\n",
				red.Sprint("Please check your"),
				bold.Sprint("TERRAFORM_CLOUD_TOKEN"),
				red.Sprint("environment variable."),
				"It seems like Terraform Cloud's Remote Execution Mode is being used.",
				"Create a Team or User API Token in the Terraform Cloud dashboard and set this environment variable.",
			)
			fmt.Fprintln(os.Stderr, msg)
		} else {
			terraformError(err)
		}
		return "", planJSON, errors.Wrap(err, "Error running terraform plan")
	}
	spinner.Success()

	return f.Name(), planJSON, nil
}

func runRemotePlan(opts *CmdOptions, args []string) ([]byte, error) {
	if !checkConfigSet() {
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

	token := cloudToken(host)
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

func runShow(opts *CmdOptions, planFile string) ([]byte, error) {
	spinner := spin.NewSpinner("Running terraform show")

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

func terraformError(err error) {
	stderr := extractStderr(err)
	if stderr == "" {
		return
	}

	msg := fmt.Sprintf("\n  Terraform command failed with:\n%s\n", indent(stderr, "    "))

	if strings.HasPrefix(stderr, "Error: Failed to select workspace") {
		msg += "\nRun `terraform workspace select your_workspace` first or set the TF_WORKSPACE environment variable.\n"
	}
	if strings.HasPrefix(stderr, "Error: Required token could not be found") {
		msg += "\nRun `terraform login` first or set the TF_CLI_CONFIG_FILE environment variable to the ABSOLUTE path.\n"
	}
	if strings.HasPrefix(stderr, "Error: No value for required variable") {
		msg += "\nPass Terraform flags using the --tfflags option.\n"
		msg += "For example: infracost --tfdir=path/to/terraform --tfflags=\"-var-file=myvars.tfvars\"\n"
	}
	if strings.HasPrefix(stderr, "Error: Failed to read variables file") {
		msg += "\nSpecify the -var-file flag as a path relative to your Terraform directory.\n"
		msg += "For example: infracost --tfdir=path/to/terraform --tfflags=\"-var-file=myvars.tfvars\"\n"
	}
	if strings.HasPrefix(stderr, "Terraform couldn't read the given file as a state or plan file.") {
		msg += "\nSpecify the --tfplan flag as a path relative to your Terraform directory.\n"
		msg += "For example: infracost --tfdir=path/to/terraform --tfplan=plan.save\n"
	}

	fmt.Fprintln(os.Stderr, color.HiRedString(msg))
}

func extractStderr(err error) string {
	if e, ok := err.(*CmdError); ok {
		return stripBlankLines(string(e.Stderr))
	}
	return ""
}

func indent(s, indent string) string {
	lines := make([]string, 0)
	for _, j := range strings.Split(s, "\n") {
		lines = append(lines, indent+j)
	}
	return strings.Join(lines, "\n")
}

func stripBlankLines(s string) string {
	return regexp.MustCompile(`[\t\r\n]+`).ReplaceAllString(strings.TrimSpace(s), "\n")
}
