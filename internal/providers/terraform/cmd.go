package terraform

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

type CmdOptions struct {
	TerraformDir string
}

type CmdError struct {
	err    error
	Stderr []byte
}

func (e *CmdError) Error() string {
	return e.err.Error()
}

func terraformBinary() string {
	terraformBinary := os.Getenv("TERRAFORM_BINARY")
	if terraformBinary == "" {
		terraformBinary = "terraform"
	}
	return terraformBinary
}

func Cmd(opts *CmdOptions, args ...string) ([]byte, error) {
	os.Setenv("TF_IN_AUTOMATION", "true")

	exe := terraformBinary()
	cmd := exec.Command(exe, args...)
	log.Infof("Running command: %s", cmd.String())
	cmd.Dir = opts.TerraformDir

	logWriter := &cmdLogWriter{
		logger: log.StandardLogger(),
		level:  log.ErrorLevel,
	}

	terraformLogWriter := &cmdLogWriter{
		logger: log.StandardLogger().WithField("binary", "terraform"),
		level:  log.DebugLevel,
	}

	var outbuf bytes.Buffer
	outw := bufio.NewWriter(&outbuf)
	var errbuf bytes.Buffer
	errw := bufio.NewWriter(&errbuf)

	cmd.Stdout = io.MultiWriter(outw, terraformLogWriter)
	cmd.Stderr = io.MultiWriter(errw, logWriter)
	err := cmd.Run()

	outw.Flush()
	errw.Flush()
	terraformLogWriter.Flush()
	logWriter.Flush()

	if err != nil {
		return outbuf.Bytes(), &CmdError{err, errbuf.Bytes()}
	}

	return outbuf.Bytes(), nil
}

func Version() (string, error) {
	exe := terraformBinary()
	out, err := exec.Command(exe, "-version").Output()
	return strings.SplitN(string(out), "\n", 2)[0], err
}

type cmdLogger interface {
	Log(level log.Level, args ...interface{})
}

// Adapted from https://github.com/sirupsen/logrus/issues/564#issuecomment-345471558
// Needed to ensure we can log large Terraform output lines.
type cmdLogWriter struct {
	logger cmdLogger
	level  log.Level
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
