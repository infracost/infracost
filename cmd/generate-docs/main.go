package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/infracost/infracost/internal/docs"
	"github.com/infracost/infracost/pkg/config"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func customError(c *cli.Context, msg string, showHelp bool) error {
	color.HiRed(fmt.Sprintf("%v\n", msg))
	if showHelp {
		_ = cli.ShowAppHelp(c)
	}

	return fmt.Errorf("")
}

func getcwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Warn(err)
		cwd = ""
	}

	return cwd
}

func main() {

	app := &cli.App{
		Name:                 "infracost-geneate-docs",
		Usage:                "Generate infracost documentations",
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:      "input",
				Usage:     "Path to docs templates",
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:      "output",
				Usage:     "Path to output of docs",
				TakesFile: true,
			},
		},
		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
			return customError(c, err.Error(), true)
		},
		Action: func(c *cli.Context) error {
			err := config.Config.SetLogLevel(c.String("log-level"))
			if err != nil {
				return customError(c, err.Error(), true)
			}

			templatesPath := c.String("input")
			if templatesPath == "" {
				templatesPath = getcwd() + "/docs/templates"
			}

			outputPath := c.String("output")
			if outputPath == "" {
				outputPath = getcwd() + "/docs/generated"
			}

			err = docs.GenerateDocs(templatesPath, outputPath)
			if err != nil {
				return errors.Wrap(err, "")
			}

			return nil

		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}
