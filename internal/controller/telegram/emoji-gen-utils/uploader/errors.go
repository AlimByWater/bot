package uploader

import "fmt"

const (
	ErrCodeNoFiles                  = "NoFiles"
	ErrCodeExceedLimit              = "ExceedLimit"
	ErrCodeUploadSticker            = "UploadSticker"
	ErrCodeCreateStickerSet         = "CreateStickerSet"
	ErrCodeAddStickers              = "AddStickers"
	ErrCodeGetStickerSet            = "GetStickerSet"
	ErrCodeOpenFile                 = "OpenFile"
	ErrCodeUploadTransparentSticker = "UploadTransparentSticker"
)

// UploaderError представляет ошибку с кодом.
type UploaderError struct {
	Code   string
	Params map[string]any
	Err    error
}

func (e *UploaderError) Error() string {
	if len(e.Params) > 0 {
		return fmt.Sprintf("[%s] %v, Params: %v", e.Code, e.Err, e.Params)
	}
	return fmt.Sprintf("[%s] %v", e.Code, e.Err)
}

func (e *UploaderError) Unwrap() error {
	return e.Err
}
