package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/docs"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func usageError(c *cli.Context, msg string) {
	color.HiRed(fmt.Sprintf("%v\n", msg))
	cli.ShowAppHelpAndExit(c, 1)
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
			usageError(c, err.Error())
			return nil
		},
		Action: func(c *cli.Context) error {
			err := config.Config.SetLogLevel(c.String("log-level"))
			if err != nil {
				usageError(c, err.Error())
			}

			templatesPath := c.String("input")
			if templatesPath == "" {
				templatesPath = getcwd() + "/docs/templates"
			}

			outputPath := c.String("output")
			if outputPath == "" {
				outputPath = getcwd() + "/docs/generated"
			}

			return docs.GenerateDocs(templatesPath, outputPath)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
