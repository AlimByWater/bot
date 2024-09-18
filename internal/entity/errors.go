package entity

import "fmt"

var ErrUserNotFound = fmt.Errorf("user not found")

var (
	ErrCacheLayoutIDRequired          = fmt.Errorf("layout id is required")
	ErrCacheIncrementReachedMaxNumber = fmt.Errorf("increment reached max number of retries")

	ErrCacheTokenNotFound   = fmt.Errorf("token not found")
	ErrCacheTokenIDRequired = fmt.Errorf("token id is required")
)

var (
	ErrInvalidDownloadLink = fmt.Errorf("invalid download link")
)
