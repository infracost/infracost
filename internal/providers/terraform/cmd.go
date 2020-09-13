package terraform

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"strings"

	"github.com/infracost/infracost/pkg/config"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
)

type cmdOptions struct {
	TerraformDir string
}

func terraformCmd(options *cmdOptions, args ...string) ([]byte, error) {
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

func TerraformVersion() (string, error) {
	terraformBinary := os.Getenv("TERRAFORM_BINARY")
	if terraformBinary == "" {
		terraformBinary = "terraform"
	}
	out, err := exec.Command(terraformBinary, "-version").Output()
	return strings.SplitN(string(out), "\n", 2)[0], err
}
