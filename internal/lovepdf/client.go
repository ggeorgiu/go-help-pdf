package lovepdf

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type HTTPClient struct {
	baseURL   string
	publicKey string
	client    http.Client
}

func New(baseURL string, publicKey string) HTTPClient {
	c := HTTPClient{
		baseURL:   baseURL,
		publicKey: publicKey,
		client:    http.Client{},
	}

	return c
}

type AuthRequest struct {
	PublicKey string `json:"public_key"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

func (cli *HTTPClient) Authenticate(ctx context.Context) (*AuthResponse, error) {
	body := AuthRequest{PublicKey: cli.publicKey}

	encoded, _ := json.Marshal(body)
	rb := bytes.NewBuffer(encoded)

	req := internalRequest{
		method: http.MethodPost,
		path:   fmt.Sprintf("%s%s", cli.baseURL, "auth"),
		headers: map[string]string{
			"Content-Type": "application/json",
		},
		body: rb,
	}
	res, err := cli.doRequest(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to do request, %w", err)
	}

	resp := AuthResponse{}
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response, err: %w", err)
	}

	return &resp, nil
}

type ServerTool string

var ServerToolCompress ServerTool = "compress"

type ResourceStartData struct {
	Bearer string
	Tool   ServerTool
}

type ResourceStartResponse struct {
	Server string `json:"server"`
	Task   string `json:"task"`
}

func (cli *HTTPClient) ResourceStart(ctx context.Context, data *ResourceStartData) (*ResourceStartResponse, error) {
	req := internalRequest{
		method: http.MethodGet,
		path:   fmt.Sprintf("%s%s%s", cli.baseURL, "start/", string(data.Tool)),
		headers: map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", data.Bearer),
			"Content-Type":  "application/json",
		},
		body: nil,
	}
	res, err := cli.doRequest(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to start resource, err: %w", err)
	}

	resp := ResourceStartResponse{}
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response, err: %w", err)
	}

	return &resp, nil
}

type UploadData struct {
	Bearer   string
	Task     string
	Server   string
	FilePath string
}

type UploadResponse struct {
	ServerFilename string `json:"server_filename"`
}

func (cli *HTTPClient) Upload(ctx context.Context, data *UploadData) (*UploadResponse, error) {
	f, err := os.Open(data.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file, %w", err)
	}

	body := bytes.Buffer{}
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", filepath.Base(data.FilePath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file, %w", err)
	}
	if _, err = io.Copy(part, f); err != nil {
		return nil, fmt.Errorf("failed to copy file content, err:%w", err)
	}

	fmt.Println(data.Task)
	field, err := writer.CreateFormField("task")
	if _, err := io.Copy(field, strings.NewReader(data.Task)); err != nil {
		return nil, fmt.Errorf("failed to create form field, err: %w", err)
	}

	writer.Close()

	req := internalRequest{
		method: http.MethodPost,
		path:   fmt.Sprintf("https://%s/v1/upload", data.Server),
		headers: map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", data.Bearer),
			"Content-Type":  writer.FormDataContentType(),
		},
		body: &body,
	}

	res, err := cli.doRequest(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to upload, err:%w", err)
	}

	resp := UploadResponse{}
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode upload response, err:%w", err)
	}

	return &resp, nil
}

type ProcessData struct {
	Bearer         string
	Server         string
	Task           string
	Tool           ServerTool
	FileName       string
	ServerFilename string
	WebhookURL     string
}

type ProcessRequest struct {
	Task  string        `json:"task"`
	Tool  string        `json:"tool"`
	Files []FileRequest `json:"files"`
}

type FileRequest struct {
	ServerFilename string `json:"server_filename"`
	Filename       string `json:"filename"`
}

type ProcessResponse struct {
	DownloadFilename string `json:"download_filename"`
	Filesize         int    `json:"filesize"`
	OutputFilesize   int    `json:"output_filesize"`
	OutputFilenumber int    `json:"output_filenumber"`
	OutputExtensions string `json:"output_extensions"`
	Timer            string `json:"timer"`
	Status           string `json:"status"`
}

func (cli *HTTPClient) Process(ctx context.Context, data *ProcessData) (*ProcessResponse, error) {
	body := ProcessRequest{
		Task: data.Task,
		Tool: string(data.Tool),
		Files: []FileRequest{
			{
				ServerFilename: data.ServerFilename,
				Filename:       data.FileName,
			},
		},
	}

	encoded, _ := json.Marshal(body)
	rb := bytes.NewBuffer(encoded)

	req := internalRequest{
		method: http.MethodPost,
		path:   fmt.Sprintf("https://%s/v1/process", data.Server),
		headers: map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", data.Bearer),
			"Content-Type":  "application/json",
		},
		body: rb,
	}

	res, err := cli.doRequest(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("faild to do process request, err: %w", err)
	}

	resp := ProcessResponse{}
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response, err: %w", err)
	}

	return &resp, nil

}

type DownloadData struct {
	Server   string
	Task     string
	Bearer   string
	Filepath string
}

func (cli *HTTPClient) Download(ctx context.Context, data *DownloadData) error {
	f, err := os.Create(data.Filepath)
	if err != nil {
		return fmt.Errorf("failed to create file, err: %w", err)
	}
	defer f.Close()

	req := internalRequest{
		method: http.MethodGet,
		path:   fmt.Sprintf("https://%s/v1/download/%s", data.Server, data.Task),
		headers: map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", data.Bearer),
		},
	}

	resp, err := cli.doRequest(ctx, &req)
	if err != nil {
		return fmt.Errorf("faild to do download request, err: %w", err)
	}

	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("failed to copy content, err: %w", err)
	}

	return nil
}

type internalRequest struct {
	method  string
	path    string
	headers map[string]string
	body    io.Reader
}

func (cli *HTTPClient) doRequest(ctx context.Context, intReq *internalRequest) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, intReq.method, intReq.path, intReq.body)
	if err != nil {
		return nil, fmt.Errorf("failed building request, err: %w", err)
	}

	for k, v := range intReq.headers {
		req.Header.Set(k, v)
	}

	res, err := cli.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed doing request, err: %w", err)
	}

	if res.StatusCode != 200 {
		return res, fmt.Errorf("got response status: %d", res.StatusCode)
	}

	return res, nil
}
