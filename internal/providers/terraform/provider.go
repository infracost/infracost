package terraform

import (
	"fmt"
	"io/ioutil"
	"os"

	"infracost/pkg/schema"

	"github.com/urfave/cli/v2"
)

type terraformProvider struct {
	jsonFile string
	planFile string
	dir string
}

func Provider() schema.Provider {
	return &terraformProvider{}
}

func (p *terraformProvider) ProcessArgs(c *cli.Context) error {
	p.jsonFile = c.String("tfjson")
	p.planFile = c.String("tfplan")
	p.dir = c.String("tfdir")

	if p.jsonFile != "" && p.planFile != "" {
		return fmt.Errorf("Please only provide one of either a Terraform Plan JSON file (tfjson) or a Terraform Plan file (tfplan)")
	}

	if p.planFile != "" && p.dir == "" {
		return fmt.Errorf("Please provide a path to the Terrafrom project (tfdir) if providing a Terraform Plan file (tfplan)\n\n")
	}

	if p.jsonFile == "" && p.dir == "" {
		return fmt.Errorf("Please provide either the path to the Terrafrom project (tfdir) or a Terraform Plan JSON file (tfjson)")
	}

	return nil
}

func (p *terraformProvider) LoadResources() ([]*schema.Resource, error) {
	var planJSON []byte
	var err error

	if p.jsonFile != "" {
		planJSON, err = loadPlanJSON(p.jsonFile)
		if err != nil {
			return []*schema.Resource{}, err
		}
	} else {
		planJSON, err = generatePlanJSON(p.dir, p.planFile)
		if err != nil {
			return []*schema.Resource{}, err
		}
	}

	schemaResources := parsePlanJSON(planJSON)
	return schemaResources, nil
}


func loadPlanJSON(path string) ([]byte, error) {
	planFile, err := os.Open(path)
	if err != nil {
		return []byte{}, err
	}
	defer planFile.Close()
	out, err := ioutil.ReadAll(planFile)
	if err != nil {
		return []byte{}, err
	}
	return out, nil
}

func generatePlanJSON(tfdir string, planPath string) ([]byte, error) {
	var err error

	opts := &cmdOptions{
		TerraformDir: tfdir,
	}

	if planPath == "" {
		_, err = terraformCmd(opts, "init")
		if err != nil {
			return []byte{}, err
		}

		planfile, err := ioutil.TempFile(os.TempDir(), "tfplan")
		if err != nil {
			return []byte{}, err
		}
		defer os.Remove(planfile.Name())

		_, err = terraformCmd(opts, "plan", "-input=false", "-lock=false", fmt.Sprintf("-out=%s", planfile.Name()))
		if err != nil {
			return []byte{}, err
		}

		planPath = planfile.Name()
	}

	out, err := terraformCmd(opts, "show", "-json", planPath)
	if err != nil {
		return []byte{}, err
	}

	return out, nil
}
