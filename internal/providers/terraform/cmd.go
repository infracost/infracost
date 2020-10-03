package terraform

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

type CmdOptions struct {
	TerraformDir string
}

func TerraformCmd(options *CmdOptions, args ...string) ([]byte, error) {
	terraformBinary := os.Getenv("TERRAFORM_BINARY")
	if terraformBinary == "" {
		terraformBinary = "terraform"
	}

	cmd := exec.Command(terraformBinary, args...)
	log.Infof("Running command: %s", cmd.String())
	cmd.Dir = options.TerraformDir

	logWriter := &cmdLogWriter{
		logger: log.StandardLogger(),
		level:  log.ErrorLevel,
	}

	terraformLogWriter := &cmdLogWriter{
		logger: log.StandardLogger().WithField("binary", "terraform"),
		level:  log.DebugLevel,
	}

	var outbuf bytes.Buffer
	b := bufio.NewWriter(&outbuf)
	cmd.Stdout = io.MultiWriter(b, terraformLogWriter)
	cmd.Stderr = logWriter
	err := cmd.Run()

	b.Flush()
	terraformLogWriter.Flush()
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

type cmdLogger interface {
	Log(level logrus.Level, args ...interface{})
}

// Adapted from https://github.com/sirupsen/logrus/issues/564#issuecomment-345471558
// Needed to ensure we can log large Terraform output lines
type cmdLogWriter struct {
	logger cmdLogger
	level  logrus.Level
	buf    bytes.Buffer
	mu     sync.Mutex
}

func (w *cmdLogWriter) Write(b []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	origLen := len(b)
	for {
		if len(b) == 0 {
			return origLen, nil
		}
		i := bytes.IndexByte(b, '\n')
		if i < 0 {
			w.buf.Write(b)
			return origLen, nil
		}

		w.buf.Write(b[:i])
		w.alwaysFlush()
		b = b[i+1:]
	}
}

func (w *cmdLogWriter) alwaysFlush() {
	w.logger.Log(w.level, w.buf.String())
	w.buf.Reset()
}

func (w *cmdLogWriter) Flush() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.buf.Len() != 0 {
		w.alwaysFlush()
	}
}
