package httpcaller

import (
	"context"
	"elysium/internal/entity"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/mymmrac/telego/telegoapi"
	"github.com/valyala/fasthttp"
	"golang.org/x/time/rate"
)

type FastHttpCallerWithLimiter struct {
	Client      *fasthttp.Client
	RateLimiter *rate.Limiter
	Logger      *slog.Logger
}

func NewFastHttpCallerWithLimiter(rateLimiter *rate.Limiter, logger *slog.Logger) *FastHttpCallerWithLimiter {

	return &FastHttpCallerWithLimiter{
		Client:      &fasthttp.Client{},
		RateLimiter: rateLimiter,
		Logger:      logger,
	}
}

func (c *FastHttpCallerWithLimiter) Call(url string, data *telegoapi.RequestData) (*telegoapi.Response, error) {
	// Рейтлимитинг
	ctx := context.Background()
	err := c.RateLimiter.Wait(ctx)
	if err != nil {
		c.Logger.Debug("CLIENT: wait", slog.String("err", err.Error()), slog.String("url", url))
		return nil, err
	}

	for {
		req := fasthttp.AcquireRequest()
		req.SetRequestURI(url)
		req.Header.SetMethod(fasthttp.MethodPost)

		req.SetBodyRaw(data.Buffer.Bytes())
		req.Header.SetContentType(data.ContentType)

		resp := fasthttp.AcquireResponse()
		err = c.Client.Do(req, resp)
		if err != nil {
			c.Logger.Debug("CLIENT: request error", slog.String("err", err.Error()), slog.String("url", url))
			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
			return nil, err
		}

		statusCode := resp.StatusCode()
		respBody := resp.Body()

		if statusCode == fasthttp.StatusOK {
			apiResp := &telegoapi.Response{}
			err = json.Unmarshal(respBody, apiResp)
			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
			if err != nil {
				return nil, fmt.Errorf("decode json: %w", err)
			}
			return apiResp, nil
		}

		if statusCode == fasthttp.StatusTooManyRequests {
			retryAfter := 0

			retryAfterHeader := string(resp.Header.Peek("Retry-After"))
			if retryAfterHeader != "" {
				waitTime, err := strconv.Atoi(retryAfterHeader)
				if err == nil {
					retryAfter = waitTime
				}
			}

			if retryAfter == 0 {
				apiResp := &telegoapi.Response{}
				err = json.Unmarshal(respBody, apiResp)
				if err != nil {
					c.Logger.Error("CLIENT: unmarshal body", slog.String("err", err.Error()))
					fasthttp.ReleaseRequest(req)
					fasthttp.ReleaseResponse(resp)
					return nil, err
				}

				if apiResp.Parameters != nil && apiResp.Parameters.RetryAfter > 0 {
					retryAfter = apiResp.Parameters.RetryAfter
				}
			}

			if retryAfter > 100 {
				if retryAfter > 100 {
					return nil, fmt.Errorf("%w: попробуйте через %.f минуты", entity.ErrEmojiPacksLimitExceeded, float64(retryAfter)/60)
				}
			}

			if retryAfter > 0 {
				c.Logger.Debug("CLIENT: rate limited", slog.Int("retry_after", retryAfter))
				fasthttp.ReleaseRequest(req)
				fasthttp.ReleaseResponse(resp)
				time.Sleep(time.Duration(retryAfter+1) * time.Second)
				continue
			} else {
				fasthttp.ReleaseRequest(req)
				fasthttp.ReleaseResponse(resp)
				return nil, fmt.Errorf("получен статус 429, но не удалось определить время ожидания")
			}
		}

		if statusCode >= fasthttp.StatusInternalServerError {
			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
			return nil, fmt.Errorf("internal server error: %d", statusCode)
		}

		apiResp := &telegoapi.Response{}
		err = json.Unmarshal(respBody, apiResp)
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
		if err != nil {
			return nil, fmt.Errorf("decode json: %w", err)
		}

		return apiResp, nil
	}
}
