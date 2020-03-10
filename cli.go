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
	nThreads   int
}

var client *sypht.Client
var cliFlags *flags

func initFunc() {
	cliFlags = &flags{}
	currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
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
				Name:        "scan",
				Usage:       "sypht-cli scan [OPTIONS] [directory]",
				UsageText:   "sypht-cli scan [OPTIONS] [directory]",
				Description: "Scan and upload all documents in a directory to Sypht's API.",
				ArgsUsage:   "Full path of the directory to scan",
				Action: func(ctx *cli.Context) error {
					var dir string
					if ctx.NArg() > 0 {
						dir = ctx.Args().Get(0)
						return scan(dir, ctx)
					}
					currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
					if err != nil {
						log.Fatalf("Unknown error : %v", err)
						return nil
					}
					return scan(currentDir, ctx)
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "workflow",
						Aliases:     []string{"w"},
						Value:       "process",
						Usage:       "Sypht workflow",
						Destination: &cliFlags.workflowID,
						Hidden:      true,
					},
					&cli.IntFlag{
						Name:        "rate-limit",
						Value:       1,
						Usage:       "Number of files to upload per second",
						Destination: &cliFlags.uploadRate,
					},
					&cli.BoolFlag{
						Name:        "recursive",
						Aliases:     []string{"R"},
						Value:       false,
						Usage:       "Recursively scan files in subdirectories",
						Destination: &cliFlags.recursive,
					},
					&cli.IntFlag{
						Name:        "nThreads",
						Value:       1,
						Usage:       "Number of go routines running at same time",
						Destination: &cliFlags.nThreads,
						Hidden:      true,
					},
				},
			},
			{
				Name:        "watch",
				Usage:       "sypht-cli watch [OPTIONS] [directory]",
				UsageText:   "sypht-cli watch [OPTIONS] [directory]",
				Description: "Watch and upload all newly added documents in a directory to Sypht's API.",
				ArgsUsage:   "Full path of the directory to watch",
				Action: func(ctx *cli.Context) error {
					var dir string
					if ctx.NArg() > 0 {
						dir = ctx.Args().Get(0)
						return watch(dir, ctx)
					}
					currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
					if err != nil {
						log.Fatalf("Unknown error : %v", err)
						return nil
					}
					return watch(currentDir, ctx)
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "workflow",
						Aliases:     []string{"w"},
						Value:       "process",
						Usage:       "Sypht workflow",
						Destination: &cliFlags.workflowID,
						Hidden:      true,
					},
					&cli.IntFlag{
						Name:        "rate-limit",
						Value:       1,
						Usage:       "Number of files to upload per second",
						Destination: &cliFlags.uploadRate,
					},
					&cli.BoolFlag{
						Name:        "recursive",
						Aliases:     []string{"R"},
						Value:       true,
						Usage:       "Recursively watch files in subdirectories",
						Destination: &cliFlags.recursive,
					},
					&cli.IntFlag{
						Name:        "nThreads",
						Value:       1,
						Usage:       "Number of go routines running at same time",
						Destination: &cliFlags.nThreads,
						Hidden:      true,
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
