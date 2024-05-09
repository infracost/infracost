package terraform

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/rs/zerolog"

	"github.com/infracost/infracost/internal/logging"
)

var defaultTerraformBinary = "terraform"

type CmdOptions struct {
	TerraformBinary     string
	Dir                 string
	TerraformWorkspace  string
	TerraformConfigFile string
	Env                 map[string]string
	Flags               []string
}

type CmdError struct {
	err    error
	Stderr []byte
}

func (e *CmdError) Error() string {
	return e.err.Error()
}

func Cmd(opts *CmdOptions, args ...string) ([]byte, error) {
	exe := opts.TerraformBinary
	if exe == "" {
		exe = defaultTerraformBinary
	}

	cmd := exec.Command(exe, append(args, opts.Flags...)...)
	logging.Logger.Debug().Msgf("Running command: %s", cmd.String())
	cmd.Dir = opts.Dir
	cmd.Env = os.Environ()

	for k, v := range opts.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	cmd.Env = append(cmd.Env, "TF_IN_AUTOMATION=true")

	if opts.TerraformWorkspace != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("TF_WORKSPACE=%s", opts.TerraformWorkspace))
	}

	if opts.TerraformConfigFile != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("TF_CLI_CONFIG_FILE=%s", opts.TerraformConfigFile))
	}

	subLogger := logging.Logger.With().Str("binary", "terraform").Logger().Level(zerolog.DebugLevel)
	logWriter := &cmdLogWriter{
		logger: subLogger,
		level:  zerolog.DebugLevel,
	}

	terraformLogWriter := &cmdLogWriter{
		logger: subLogger,
		level:  zerolog.DebugLevel,
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

// Adapted from https://github.com/sirupsen/logrus/issues/564#issuecomment-345471558
// Needed to ensure we can log large Terraform output lines.
type cmdLogWriter struct {
	logger zerolog.Logger
	buf    bytes.Buffer
	mu     sync.Mutex
	level  zerolog.Level
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
	w.logger.WithLevel(w.level).Msg(w.buf.String())
	w.buf.Reset()
}

func (w *cmdLogWriter) Flush() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.buf.Len() != 0 {
		w.alwaysFlush()
	}
}

func CreateConfigFile(dir string, terraformCloudHost string, terraformCloudToken string) (string, error) {
	if terraformCloudToken == "" {
		return "", nil
	}

	logging.Logger.Debug().Msg("Creating temporary config file for Terraform credentials")
	tmpFile, err := os.CreateTemp("", "")
	if err != nil {
		return "", err
	}

	if os.Getenv("TF_CLI_CONFIG_FILE") != "" {
		logging.Logger.Debug().Msgf("TF_CLI_CONFIG_FILE is set, copying existing config from %s to config to temporary config file %s", os.Getenv("TF_CLI_CONFIG_FILE"), tmpFile.Name())
		path := os.Getenv("TF_CLI_CONFIG_FILE")

		if !filepath.IsAbs(path) {
			path, err = filepath.Abs(filepath.Join(dir, path))
			if err != nil {
				logging.Logger.Warn().Msgf("Unable to copy existing config from %s: %v", path, err)
			}
		}

		if err == nil {
			err = copyFile(path, tmpFile.Name())
			if err != nil {
				logging.Logger.Warn().Msgf("Unable to copy existing config from %s: %v", path, err)
			}
		}
	}

	f, err := os.OpenFile(tmpFile.Name(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return tmpFile.Name(), err
	}
	defer f.Close()

	host := terraformCloudHost
	if host == "" {
		host = "app.terraform.io"
	}

	contents := fmt.Sprintf(`credentials "%s" {
	token = "%s"
}
`, host, terraformCloudToken)

	logging.Logger.Debug().Msgf("Writing Terraform credentials to temporary config file %s", tmpFile.Name())
	if _, err := f.WriteString(contents); err != nil {
		return tmpFile.Name(), err
	}

	return tmpFile.Name(), nil
}

func copyFile(srcPath string, dstPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}

	return nil
}
