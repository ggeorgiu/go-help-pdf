package compress

import (
	"context"
	"fmt"

	"go-help-pdf/internal/lovepdf"
)

type Service struct {
	cli *lovepdf.HTTPClient
}

func New(cli *lovepdf.HTTPClient) *Service {
	return &Service{cli: cli}
}

type HandleData struct {
	WorkingDir string
	OutputDir  string
	FileName   string
}

func (s *Service) HandleFile(ctx context.Context, hd *HandleData) error {
	auth, err := s.cli.Authenticate(ctx)
	if err != nil {
		return fmt.Errorf("failed to authenticate, err: %w", err)
	}

	rsd := lovepdf.ResourceStartData{
		Bearer: auth.Token,
		Tool:   lovepdf.ServerToolCompress,
	}
	startResp, err := s.cli.ResourceStart(ctx, &rsd)
	if err != nil {
		return fmt.Errorf("failed to start resource, err: %w", err)
	}

	ud := lovepdf.UploadData{
		Bearer:   auth.Token,
		Task:     startResp.Task,
		Server:   startResp.Server,
		FilePath: fmt.Sprintf("%s%s", hd.WorkingDir, hd.FileName),
	}
	uploadResp, err := s.cli.Upload(ctx, &ud)
	if err != nil {
		return fmt.Errorf("failed to upload file: %s, err: %w", hd.FileName, err)
	}

	pd := lovepdf.ProcessData{
		Bearer:         auth.Token,
		Server:         startResp.Server,
		Task:           startResp.Task,
		ServerFilename: uploadResp.ServerFilename,
		Tool:           lovepdf.ServerToolCompress,
		FileName:       hd.FileName,
	}
	processResponse, err := s.cli.Process(ctx, &pd)
	if err != nil {
		return fmt.Errorf("failed to process file: %s, err: %w", hd.FileName, err)
	}

	dd := lovepdf.DownloadData{
		Server:   startResp.Server,
		Task:     startResp.Task,
		Bearer:   auth.Token,
		Filepath: fmt.Sprintf("%s%s", hd.OutputDir, processResponse.DownloadFilename),
	}
	err = s.cli.Download(ctx, &dd)
	if err != nil {
		return fmt.Errorf("failed to download file, %s, err: %w", processResponse.DownloadFilename, err)
	}

	return nil
}
