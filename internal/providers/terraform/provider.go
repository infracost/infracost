package terraform

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/spin"
	"github.com/kballard/go-shellquote"
	"github.com/pkg/errors"
	"golang.org/x/mod/semver"

	"github.com/urfave/cli/v2"
)

var minTerraformVer = "v0.12"

type terraformProvider struct {
	dir       string
	jsonFile  string
	planFile  string
	useState  bool
	planFlags string
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

	if p.jsonFile != "" && p.planFile != "" {
		return errors.New("Please provide either a Terraform Plan JSON file (tfjson) or a Terraform Plan file (tfplan)")
	}

	return nil
}

func (p *terraformProvider) LoadResources() ([]*schema.Resource, error) {
	var resources []*schema.Resource

	var j []byte
	var err error

	if p.useState {
		j, err = p.generateStateJSON()
	} else {
		j, err = p.loadPlanJSON()
	}
	if err != nil {
		return []*schema.Resource{}, err
	}

	resources, err = parseJSON(j)
	if err != nil {
		return resources, errors.Wrap(err, "Error parsing Terraform JSON")
	}

	return resources, nil
}

func (p *terraformProvider) loadPlanJSON() ([]byte, error) {
	if p.jsonFile == "" {
		return p.generatePlanJSON()
	}

	f, err := os.Open(p.jsonFile)
	if err != nil {
		return []byte{}, errors.Wrapf(err, "Error reading Terraform plan file")
	}
	defer f.Close()

	out, err := ioutil.ReadAll(f)
	if err != nil {
		return []byte{}, errors.Wrapf(err, "Error reading Terraform plan file")
	}

	return out, nil
}

func (p *terraformProvider) generateStateJSON() ([]byte, error) {
	err := p.terraformPreChecks()
	if err != nil {
		return []byte{}, err
	}

	opts := &CmdOptions{
		TerraformDir: p.dir,
	}

	spinner := spin.NewSpinner("Running terraform show")
	out, err := Cmd(opts, "show", "-no-color", "-json")
	if err != nil {
		spinner.Fail()
		terraformError(err)
		return []byte{}, errors.Wrap(err, "Error running terraform show")
	}
	spinner.Success()

	return out, nil
}

func (p *terraformProvider) generatePlanJSON() ([]byte, error) {
	err := p.terraformPreChecks()
	if err != nil {
		return []byte{}, err
	}

	opts := &CmdOptions{
		TerraformDir: p.dir,
	}

	var spinner *spin.Spinner

	if p.planFile == "" {
		spinner = spin.NewSpinner("Running terraform init")
		_, err := Cmd(opts, "init", "-no-color")
		if err != nil {
			spinner.Fail()
			terraformError(err)
			return []byte{}, errors.Wrap(err, "Error running terraform init")
		}
		spinner.Success()

		spinner = spin.NewSpinner("Running terraform plan")
		f, err := ioutil.TempFile(os.TempDir(), "tfplan")
		if err != nil {
			spinner.Fail()
			return []byte{}, errors.Wrap(err, "Error creating temporary file 'tfplan'")
		}
		defer os.Remove(f.Name())

		flags, err := shellquote.Split(p.planFlags)
		if err != nil {
			return []byte{}, errors.Wrap(err, "Error parsing terraform plan flags")
		}
		args := []string{"plan", "-input=false", "-lock=false", "-no-color"}
		args = append(args, flags...)
		args = append(args, fmt.Sprintf("-out=%s", f.Name()))
		_, err = Cmd(opts, args...)
		if err != nil {
			spinner.Fail()
			terraformError(err)
			return []byte{}, errors.Wrap(err, "Error running terraform plan")
		}
		spinner.Success()

		p.planFile = f.Name()
	}

	spinner = spin.NewSpinner("Running terraform show")
	out, err := Cmd(opts, "show", "-no-color", "-json", p.planFile)
	if err != nil {
		spinner.Fail()
		terraformError(err)
		return []byte{}, errors.Wrap(err, "Error running terraform show")
	}
	spinner.Success()

	return out, nil
}

func (p *terraformProvider) terraformPreChecks() error {
	if p.jsonFile == "" {
		_, err := exec.LookPath(terraformBinary())
		if err != nil {
			return errors.Errorf("Terraform binary \"%s\" could not be found.\nSet a custom Terraform binary using the environment variable TERRAFORM_BINARY.", terraformBinary())
		}

		if v, ok := checkVersion(); !ok {
			return errors.Errorf("Terraform %s is not supported. Please use Terraform version >= %s.", v, minTerraformVer)
		}

		if !p.inTerraformDir() {
			return errors.Errorf("Directory \"%s\" does not have any .tf files.\nSet the Terraform directory path using the --tfdir option.", p.dir)
		}
	}
	return nil
}

func checkVersion() (string, bool) {
	out, err := Version()
	if err != nil {
		// If we encounter any errors here we just return true
		// since it might be caused by a custom Terraform binary
		return "", true
	}
	p := strings.Split(out, " ")
	v := p[len(p)-1]

	// Allow any non-terraform binaries, e.g. terragrunt
	if !strings.HasPrefix(out, "Terraform ") {
		return v, true
	}

	return v, semver.Compare(v, minTerraformVer) >= 0
}

func (p *terraformProvider) inTerraformDir() bool {
	matches, err := filepath.Glob(filepath.Join(p.dir, "*.tf"))
	return matches != nil && err == nil
}

func terraformError(err error) {
	if e, ok := err.(*CmdError); ok {
		stderr := stripBlankLines(string(e.Stderr))

		msg := fmt.Sprintf("\n  Terraform command failed with:\n%s\n", indent(stderr, "    "))

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
