package main

import (
	"fmt"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/hcl"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/ui"
)

type InitCommand struct {
	Path         string
	Overwrite    bool
	ExcludedDirs []string

	cmd *cobra.Command
}

func (ic InitCommand) run(ctx *config.RunContext) error {
	pl := hcl.NewProjectLocator(logging.Logger.WithFields(logrus.Fields{
		"path":      ic.Path,
		"overwrite": ic.Overwrite,
	}), ic.ExcludedDirs)

	paths := pl.FindRootModules(ic.Path)
	if len(paths) == 0 {
		return fmt.Errorf("No valid Terraform directories detected at %s.", ic.Path)
	}

	for i, path := range paths {
		paths[i], _ = filepath.Rel(ic.Path, path)
	}

	config.CreateConfigFile(ic.Path, paths, ic.Overwrite)
	ui.PrintSuccessf(ic.cmd.ErrOrStderr(), "initialized Infracost config file: %s", filepath.Join(ic.Path, "infracost.yml"))

	return nil
}

func initCommand(ctx *config.RunContext) *cobra.Command {
	var initCmd InitCommand

	cmd := &cobra.Command{
		Use:    "init",
		Short:  "Initialize a new Infracost working directory by creating initial files.",
		Long:   "Initialize a new Infracost working directory by creating initial files.",
		Hidden: true,
		Example: `
      infracost init --path /code --overwrite --exclude-path k8s-conf
      `,
		ValidArgs: []string{"--", "-"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return initCmd.run(ctx)
		},
	}

	initCmd.cmd = cmd

	cmd.Flags().StringVarP(&initCmd.Path, "path", "p", "", "Path to your project.")
	cmd.Flags().BoolVar(&initCmd.Overwrite, "overwrite", false, "Overwrite existing infracost.yml config files with newly detected paths.")
	cmd.Flags().StringSliceVar(&initCmd.ExcludedDirs, "exclude-path", nil, "Paths of directories to exclude, glob patterns need quotes")

	return cmd
}
