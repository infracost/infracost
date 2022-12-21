package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strings"
	"sync"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/infracost/infracost/internal/config"
)

var (
	pass = color.FgGreen
	skip = color.FgYellow
	fail = color.FgHiRed
)

type test struct {
	path           string
	showCoverage   bool
	uploadCoverage bool
	pretty         bool

	coverage   *bytes.Buffer
	coverTable bool

	cmd *cobra.Command
}

func (t *test) run(ctx *config.RunContext) error {
	var wg sync.WaitGroup
	wg.Add(1)
	defer wg.Wait()

	r, w := io.Pipe()
	defer w.Close()
	p := t.path
	if !path.IsAbs(p) && !strings.HasPrefix(p, ".") {
		p = fmt.Sprintf("./%s", p)
	}

	color.Output = t.cmd.OutOrStderr()

	if !t.pretty {
		color.NoColor = true
	}

	args := []string{"test", p, "-test.v", "-costtest.porcelain"}
	if t.showCoverage {
		t.coverage = bytes.NewBuffer([]byte{})
		args = append(args, "-costtest.cover")
	}

	cmd := exec.Command("go", args...)
	cmd.Stderr = w
	cmd.Stdout = w
	cmd.Env = os.Environ()

	if err := cmd.Start(); err != nil {
		wg.Done()
		return err
	}

	go t.consume(&wg, r)

	sigc := make(chan os.Signal)
	done := make(chan struct{})
	defer func() {
		done <- struct{}{}
	}()
	signal.Notify(sigc)

	go func() {
		for {
			select {
			case sig := <-sigc:
				cmd.Process.Signal(sig)
			case <-done:
				return
			}
		}
	}()

	return cmd.Wait()
}

func (t test) consume(wg *sync.WaitGroup, r io.Reader) {
	defer wg.Done()
	reader := bufio.NewReader(r)
	for {
		l, _, err := reader.ReadLine()
		if err == io.EOF {
			if t.coverage != nil {
				t.cmd.Printf("\n%s", t.coverage.String())
			}

			return
		}

		if err != nil {
			log.Println(err)
			return
		}

		t.parse(string(l))
	}
}

func (t *test) parse(line string) {
	trimmed := strings.TrimSpace(line)
	defer color.Unset()

	var c color.Attribute
	switch {
	case strings.HasPrefix(trimmed, "--- COSTTEST"):
		t.coverTable = true
		return
	case strings.HasPrefix(trimmed, "COSTTEST"):
		t.coverTable = false
		return
	case strings.Contains(trimmed, "[no test files]"):
		return
	case strings.HasPrefix(trimmed, "--- PASS"): // passed
		fallthrough
	case strings.HasPrefix(trimmed, "ok"):
		fallthrough
	case strings.HasPrefix(trimmed, "PASS"):
		c = pass
	case strings.HasPrefix(trimmed, "--- SKIP"):
		c = skip
	case strings.HasPrefix(trimmed, "--- FAIL"):
		fallthrough
	case strings.HasPrefix(trimmed, "FAIL"):
		c = fail
	}

	color.Set(c)
	if t.coverTable {
		t.coverage.WriteString(fmt.Sprintf("%s\n", line))
		return
	}

	t.cmd.Printf("%s\n", line)
}

func testCmd(ctx *config.RunContext) *cobra.Command {
	t := new(test)

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Run a suite of defined cost tests.",
		Long:  "Run a cost of defined cost tests.",
		Example: `
      infracost test --path tests
      `,
		ValidArgs: []string{"--", "-"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkAPIKey(ctx.Config.APIKey, ctx.Config.PricingAPIEndpoint, ctx.Config.DefaultPricingAPIEndpoint); err != nil {
				return err
			}

			return t.run(ctx)
		},
	}

	t.cmd = cmd

	cmd.Flags().StringVarP(&t.path, "path", "p", "", "The costtest package (normally /tests)")
	cmd.Flags().BoolVar(&t.showCoverage, "coverage", false, "Show coverage of your project resources by file, resource and cost")
	cmd.Flags().BoolVar(&t.uploadCoverage, "upload", false, "Upload a coverage report to your Infracost Cloud organization")
	cmd.Flags().BoolVar(&t.pretty, "pretty", true, "Run the tests with colored output")

	return cmd
}
