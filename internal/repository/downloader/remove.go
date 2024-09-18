package downloader

import (
	"context"
	"elysium/pkg/proto"
	"fmt"
)

func (m *Module) RemoveFile(ctx context.Context, fileName string) error {
	resp, err := m.client.DeleteFile(ctx, &proto.DeleteRequest{
		FileName: fileName,
	})
	if err != nil {
		return fmt.Errorf("remove file: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("remove file: failed")
	}

	return nil
}
