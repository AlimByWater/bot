package httpclient

import (
	"context"
	"elysium/internal/entity"
	"fmt"
	"golang.org/x/time/rate"
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
}

// Do dispatches the HTTP request to the network
func (c *RLHTTPClient) Do(req *http.Request) (*http.Response, error) {
	// Comment out the below 5 lines to turn off ratelimiting
	ctx := context.Background()
	err := c.Ratelimiter.Wait(ctx) // This is a blocking call. Honors the rate limit
	if err != nil {
		slog.Debug("CLIENT: wait", slog.String("err", err.Error()), slog.String("url", req.URL.String()))
		return nil, err
	}
	slog.Debug("CLIENT: wait no error", slog.String("url", req.URL.String()))

	var resp *http.Response

	for {
		resp, err = c.client.Do(req)
		slog.Debug("CLIENT: do", slog.String("url", req.URL.String()))
		if err != nil {
			if strings.Contains(err.Error(), "retry_after") {
				// Извлекаем время ожидания из ошибки
				parts := strings.Split(err.Error(), "retry_after ")
				if len(parts) < 2 {
					slog.Error("CLIENT: split", "err", err.Error())
					return nil, err
				}
				waitTimeInt, parseErr := strconv.Atoi(strings.TrimSpace(parts[len(parts)-1]))
				if parseErr != nil {
					slog.Error("CLIENT: parse", "err", parseErr.Error())
					return nil, err
				}

				waitTime := time.Duration(waitTimeInt) * time.Second

				if waitTime.Seconds() > 100 {
					return nil, fmt.Errorf("%w: попробуйте через %.0f секунд", entity.ErrEmojiPacksLimitExceeded, waitTime.Minutes())
				}

				time.Sleep(waitTime)
				continue
			} else {
				return nil, err
			}
		}
		break
	}

	return resp, nil
}

// NewClient return http client with a ratelimiter
func NewClient(rl *rate.Limiter) *RLHTTPClient {
	c := &RLHTTPClient{
		client:      http.DefaultClient,
		Ratelimiter: rl,
	}
	return c
}
