package main

import (
	"context"
	"fmt"
	"go-help-pdf/internal/lovepdf"
	"log/slog"
	"os"

	"github.com/urfave/cli/v2"
)

type appOpts struct {
	projectKey string
	secretKey  string
	workingDir string
	outputDir  string
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

func compressFileCmd(ctx context.Context, fileName string) error {
	pdfClient := lovepdf.New("https://api.ilovepdf.com/v1/", opts.projectKey)

	auth, err := pdfClient.Authenticate(ctx)
	if err != nil {
		return fmt.Errorf("failed to authenticate, err: %w", err)
	}

	rsd := lovepdf.ResourceStartData{
		Bearer: auth.Token,
		Tool:   lovepdf.ServerToolCompress,
	}
	startResp, err := pdfClient.ResourceStart(ctx, &rsd)
	if err != nil {
		return fmt.Errorf("failed to start resource, err: %w", err)
	}

	ud := lovepdf.UploadData{
		Bearer:   auth.Token,
		Task:     startResp.Task,
		Server:   startResp.Server,
		FilePath: fmt.Sprintf("%s%s", opts.workingDir, fileName),
	}
	uploadResp, err := pdfClient.Upload(ctx, &ud)
	if err != nil {
		fmt.Println(uploadResp)
		return fmt.Errorf("failed to upload file: %s, err: %w", fileName, err)
	}

	pd := lovepdf.ProcessData{
		Bearer:         auth.Token,
		Server:         startResp.Server,
		Task:           startResp.Task,
		ServerFilename: uploadResp.ServerFilename,
		Tool:           lovepdf.ServerToolCompress,
		FileName:       fileName,
	}
	processResponse, err := pdfClient.Process(ctx, &pd)
	if err != nil {
		return fmt.Errorf("failed to process file: %s, err: %w", fileName, err)
	}

	dd := lovepdf.DownloadData{
		Server:   startResp.Server,
		Task:     startResp.Task,
		Bearer:   auth.Token,
		Filepath: fmt.Sprintf("%s%s", opts.outputDir, processResponse.DownloadFilename),
	}
	err = pdfClient.Download(ctx, &dd)
	if err != nil {
		return fmt.Errorf("failed to download file, %s, err: %w", processResponse.DownloadFilename, err)
	}

	return nil
}
