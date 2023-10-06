package main

import (
	"context"
	"fmt"
	"go-help-pdf/internal/compress"
	"go-help-pdf/internal/lovepdf"
	"log/slog"
	"os"

	"github.com/urfave/cli/v2"
)

type appOpts struct {
	projectKey string
	workingDir string
	outputDir  string
	baseURL    string
}

var opts = appOpts{}

func main() {
	app := cli.NewApp()

	app.Commands = []*cli.Command{
		{
			Name: "compress-file",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "project-key",
					Destination: &opts.projectKey,
					EnvVars:     []string{"PDF_API_PROJECT_KEY"},
					Required:    true,
					Hidden:      true,
				},
				&cli.StringFlag{
					Name:        "working-dir",
					Destination: &opts.workingDir,
					EnvVars:     []string{"PDF_API_WORKING_DIR"},
					Value:       "/Users/gabrielgeorgiu/Projects/learning/go-help-pdf/pdfs/",
				},
				&cli.StringFlag{
					Name:        "output-dir",
					Destination: &opts.outputDir,
					EnvVars:     []string{"PDF_API_OUTPUT_DIR"},
					Value:       "/Users/gabrielgeorgiu/Projects/learning/go-help-pdf/pdfc/",
				},
				&cli.StringFlag{
					Name:        "base-url",
					Destination: &opts.baseURL,
					EnvVars:     []string{"PDF_API_BASE_URL"},
					Value:       "https://api.ilovepdf.com/v1/",
				},
			},
			Action: func(c *cli.Context) error {
				return compressFileCmd(c.Context, c.Args().Get(0))
			},
		},
	}

	if err := app.RunContext(context.Background(), os.Args); err != nil {
		slog.Error("got error while running app, %v", err)
	}
}

func compressFileCmd(ctx context.Context, filename string) error {
	if filename == "" {
		return fmt.Errorf("invalid filename")
	}

	pdfClient := lovepdf.New(opts.baseURL, opts.projectKey)
	svc := compress.New(pdfClient)

	hd := compress.HandleData{
		FileName:   filename,
		WorkingDir: opts.workingDir,
		OutputDir:  opts.outputDir,
	}

	err := svc.HandleFile(ctx, &hd)
	if err != nil {
		return fmt.Errorf("failed to handle file: %s, err: %w", filename, err)
	}

	return nil
}
