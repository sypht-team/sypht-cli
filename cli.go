package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/sypht-team/sypht-golang-client"
	"github.com/urfave/cli/v2"
)

type flags struct {
	recursive  bool
	workflowID string
	uploadRate int
	nRoutines int
}

var client *sypht.Client
var currentDir string
var cliFlags = flags{
	recursive:  true,
	workflowID: "process",
	uploadRate: 1,
	nRoutines : 5,
}

const nRoutines = 2

func initFunc() {
	var err error
	currentDir, err = filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	config := getConfig(filepath.Join(currentDir, "config.json"))
	client, err = sypht.NewSyphtClient(fmt.Sprintf("%s:%s", config.ClientID, config.ClientSecret), nil)
	if err != nil {
		log.Fatalf("Unable to start Sypht client , %v", err)
	}
}

func main() {
	initFunc()
	app := &cli.App{
		Name:    "Sypht-cli",
		Usage:   "Upload docs to Sypht's API",
		Version: "v0.1.0",
		Commands: []*cli.Command{
			{
				Name:        "watch",
				Usage:       "sypht-cli watch [directory] [OPTIONS]",
				UsageText:   "sypht-cli watch [directory] [OPTIONS]",
				Description: "Scan and upload all documents in a directory to Sypht's API.",
				ArgsUsage:   "Full path of the directory to scan",
				Action: func(ctx *cli.Context) error {
					var dir string
					if ctx.NArg() > 0 {
						dir = ctx.Args().Get(0)
						return watch(dir, ctx)
					}
					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "workflow",
						Aliases:     []string{"w"},
						Value:       "process",
						Usage:       "Sypht workflow",
						Destination: &cliFlags.workflowID,
					},
					&cli.IntFlag{
						Name:        "rate-limit",
						Aliases:     []string{"r"},
						Value:       1,
						Usage:       "Upload number of files per second",
						Destination: &cliFlags.uploadRate,
					},
					&cli.BoolFlag{
						Name:        "recursive",
						Aliases:     []string{"R"},
						Value:       true,
						Usage:       "Recursively scan subdirectory",
						Destination: &cliFlags.recursive,
					},
					&cli.IntFlag{
						Name : "nRoutines",
						Value : 5,
						Usage : "Number of go routines running at same time",
						Destination: &cliFlags.recursive,
					}
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
