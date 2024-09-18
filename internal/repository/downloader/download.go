package downloader

import (
	"context"
	"elysium/pkg/proto"
	"fmt"
)

func (m *Module) DownloadByLink(ctx context.Context, link string, format string) (string, []byte, error) {
	resp, err := m.client.DownloadByLink(ctx, &proto.DownloadRequest{
		Url:    link,
		Format: format,
	})

	if err != nil {
		return "", nil, fmt.Errorf("download by link: %w", err)
	}

	return resp.FileName, resp.FileData, nil
}
