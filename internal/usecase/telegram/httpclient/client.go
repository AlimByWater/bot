package httpclient

import (
	"bytes"
	"context"
	"elysium/internal/entity"
	"encoding/json"
	"fmt"
	"golang.org/x/time/rate"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// RLHTTPClient Rate Limited HTTP Client
type RLHTTPClient struct {
	client      *http.Client
	Ratelimiter *rate.Limiter
	logger      *slog.Logger
}

// Do dispatches the HTTP request to the network
func (c *RLHTTPClient) Do(req *http.Request) (*http.Response, error) {
	ctx := context.Background()
	err := c.Ratelimiter.Wait(ctx)
	if err != nil {
		c.logger.Debug("CLIENT: wait", slog.String("err", err.Error()), slog.String("url", req.URL.String()))
		return nil, err
	}

	var resp *http.Response
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	for {
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		resp, err = c.client.Do(req)
		if err != nil {
			c.logger.Debug("CLIENT: request error", slog.String("err", err.Error()), slog.String("url", req.URL.String()))
			return nil, err
		}

		if resp.StatusCode == http.StatusOK {
			break
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			var retryAfter int

			retryAfterHeader := resp.Header.Get("Retry-After")
			if retryAfterHeader != "" {
				waitTimeInt, err := strconv.Atoi(retryAfterHeader)
				if err == nil {
					retryAfter = waitTimeInt
				}
			}

			if retryAfter == 0 {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					c.logger.Error("CLIENT: read body", slog.String("err", err.Error()))
					return nil, err
				}

				type apiResponse struct {
					Description string `json:"description,omitempty"`
					ErrorCode   int    `json:"error_code,omitempty"`
					Parameters  struct {
						RetryAfter int `json:"retry_after,omitempty"`
					} `json:"parameters,omitempty"`
				}

				var r apiResponse
				err = json.Unmarshal(body, &r)
				if err != nil {
					c.logger.Error("CLIENT: unmarshal body", slog.String("err", err.Error()))
					return nil, err
				}

				retryAfter = r.Parameters.RetryAfter
			}

			if retryAfter > 0 {
				if retryAfter > 100 {
					if strings.Contains(req.URL.String(), "uploadStickerFile") || strings.Contains(req.URL.String(), "createNewStickerSet") || strings.Contains(req.URL.String(), "addStickerToSet") {
						return nil, fmt.Errorf("%w: попробуйте через %.f минуты", entity.ErrEmojiPacksLimitExceeded, float64(retryAfter)/60)
					} else {
						return nil, fmt.Errorf("лимит запросов в секунду превышен")
					}
				}

				c.logger.Debug("CLIENT: rate limited", slog.Int("retry_after", retryAfter))
				resp.Body.Close()
				time.Sleep(time.Duration(retryAfter) * time.Second)
				continue
			} else {
				return nil, fmt.Errorf("получен статус 429, но не удалось определить время ожидания")
			}
		}

		break
	}

	return resp, nil
}

// NewClient return http client with a ratelimiter
func NewClient(rl *rate.Limiter, logger *slog.Logger) *RLHTTPClient {
	c := &RLHTTPClient{
		client:      http.DefaultClient,
		Ratelimiter: rl,
		logger:      logger,
	}
	return c
}
