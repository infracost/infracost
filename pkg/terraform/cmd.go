package terraform

import (
	"bufio"
	"bytes"
	"fmt"
	"infracost/pkg/config"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
)

type TerraformOptions struct {
	TerraformDir string
}

func terraformCommand(options *TerraformOptions, args ...string) ([]byte, error) {
	terraformBinary := os.Getenv("TERRAFORM_BINARY")
	if terraformBinary == "" {
		terraformBinary = "terraform"
	}

	cmd := exec.Command(terraformBinary, args...)
	if config.Config.NoColor {
		log.Infof("Running command: %s", cmd.String())
	} else {
		log.Info(color.HiBlackString("Running command: %s", cmd.String()))
	}
	cmd.Dir = options.TerraformDir

	var outbuf bytes.Buffer
	cmd.Stdout = bufio.NewWriter(&outbuf)
	cmd.Stderr = log.StandardLogger().WriterLevel(log.ErrorLevel)
	err := cmd.Run()
	return outbuf.Bytes(), err
}

func LoadPlanJSON(path string) ([]byte, error) {
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

func GeneratePlanJSON(tfdir string, planPath string) ([]byte, error) {
	var err error

	opts := &TerraformOptions{
		TerraformDir: tfdir,
	}

	if planPath == "" {
		_, err = terraformCommand(opts, "init")
		if err != nil {
			return []byte{}, err
		}

		planfile, err := ioutil.TempFile(os.TempDir(), "tfplan")
		if err != nil {
			return []byte{}, err
		}
		defer os.Remove(planfile.Name())

		_, err = terraformCommand(opts, "plan", "-input=false", "-lock=false", fmt.Sprintf("-out=%s", planfile.Name()))
		if err != nil {
			return []byte{}, err
		}

		planPath = planfile.Name()
	}

	out, err := terraformCommand(opts, "show", "-json", planPath)
	if err != nil {
		return []byte{}, err
	}

	return out, nil
}
