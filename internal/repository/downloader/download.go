package downloader

import (
	"context"
	"elysium/pkg/proto"
	"fmt"
)

const (
	tryCount = 3
)

func (m *Module) DownloadByLink(ctx context.Context, link string, format string) (string, []byte, error) {
	for i := 0; i < tryCount; i++ {
		resp, err := m.client.DownloadByLink(ctx, &proto.DownloadRequest{
			Url:    link,
			Format: format,
		})

		if err != nil {
			if i == tryCount-1 {
				return "", nil, fmt.Errorf("download by link: %w", err)
			}
			continue
		}

		return resp.FileName, resp.FileData, nil
	}

	return "", nil, fmt.Errorf("download by link: try count exceeded")
}
