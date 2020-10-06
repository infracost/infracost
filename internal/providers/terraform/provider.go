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
	"github.com/infracost/infracost/internal/spin"
	"github.com/infracost/infracost/pkg/schema"
	"github.com/pkg/errors"

	"github.com/urfave/cli/v2"
)

type terraformProvider struct {
	jsonFile string
	planFile string
	dir      string
}

// New returns new Terraform Provider
func New() schema.Provider {
	return &terraformProvider{}
}

func (p *terraformProvider) ProcessArgs(c *cli.Context) error {
	p.jsonFile = c.String("tfjson")
	p.planFile = c.String("tfplan")
	p.dir = c.String("tfdir")

	if p.jsonFile != "" && p.planFile != "" {
		return errors.New("Please provide either a Terraform Plan JSON file (tfjson) or a Terraform Plan file (tfplan)")
	}

	if p.planFile != "" && p.dir == "" {
		return errors.New("Please provide a path to the Terraform project (tfdir) if providing a Terraform Plan file (tfplan)")
	}

	return nil
}

func (p *terraformProvider) LoadResources() ([]*schema.Resource, error) {
	err := p.preChecks()
	if err != nil {
		return []*schema.Resource{}, err
	}

	plan, err := p.loadPlanJSON()
	if err != nil {
		return []*schema.Resource{}, err
	}

	resources, err := parsePlanJSON(plan)
	if err != nil {
		return resources, errors.Wrap(err, "Error parsing plan JSON")
	}

	return resources, nil
}

func (p *terraformProvider) preChecks() error {
	if p.jsonFile == "" {
		_, err := exec.LookPath(terraformBinary())
		if err != nil {
			return errors.Errorf("Could not find terraform binary \"%s\" in path.\nYou can set a custom terraform binary using the environment variable TERRAFORM_BINARY.", terraformBinary())
		}

		if !p.inTerraformDir() {
			return errors.Errorf("Directory \"%s\" does not have any .tf files.\nYou can pass a path to a Terraform directory using the --tfdir option.", p.dir)
		}
	}
	return nil
}

func (p *terraformProvider) loadPlanJSON() ([]byte, error) {
	if p.jsonFile == "" {
		return p.generatePlanJSON()
	}

	f, err := os.Open(p.jsonFile)
	if err != nil {
		return []byte{}, errors.Wrapf(err, "Error reading plan file")
	}
	defer f.Close()

	out, err := ioutil.ReadAll(f)
	if err != nil {
		return []byte{}, errors.Wrapf(err, "Error reading plan file")
	}

	return out, nil
}

func (p *terraformProvider) generatePlanJSON() ([]byte, error) {
	opts := &CmdOptions{
		TerraformDir: p.dir,
	}

	if p.planFile == "" {
		spinner := spin.NewSpinner("Running terraform init")
		_, err := TerraformCmd(opts, "init", "-no-color")
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

		_, err = TerraformCmd(opts, "plan", "-input=false", "-lock=false", "-no-color", fmt.Sprintf("-out=%s", f.Name()))
		if err != nil {
			spinner.Fail()
			terraformError(err)
			return []byte{}, errors.Wrap(err, "Error running terraform plan")
		}
		spinner.Success()

		p.planFile = f.Name()
	}

	spinner := spin.NewSpinner("Running terraform show")
	out, err := TerraformCmd(opts, "show", "-no-color", "-json", p.planFile)
	if err != nil {
		spinner.Fail()
		terraformError(err)
		return []byte{}, errors.Wrap(err, "Error running terraform show")
	}
	spinner.Success()

	return out, nil
}

func (p *terraformProvider) inTerraformDir() bool {
	matches, err := filepath.Glob(filepath.Join(p.dir, "*.tf"))
	return matches != nil && err == nil
}

func terraformError(err error) {
	if terr, ok := err.(*TerraformCmdError); ok {
		fmt.Fprintln(os.Stderr, indent(color.HiRedString("Terraform command failed with:"), "  "))
		fmt.Fprintln(os.Stderr, indent(color.HiRedString(stripBlankLines(string(terr.Stderr))), "    "))
	}
}

func indent(s, indent string) string {
	result := ""
	for _, j := range strings.Split(strings.TrimRight(s, "\n"), "\n") {
		result += indent + j + "\n"
	}
	return result
}

func stripBlankLines(s string) string {
	return regexp.MustCompile(`[\t\r\n]+`).ReplaceAllString(strings.TrimSpace(s), "\n")
}
