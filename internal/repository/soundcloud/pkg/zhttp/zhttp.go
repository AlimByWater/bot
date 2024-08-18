package zhttp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Zhttp client
type Zhttp struct {
	client *http.Client
}

// New ...
func New(timeout time.Duration, proxy string) (*Zhttp, error) {
	zhttp := &Zhttp{
		client: http.DefaultClient,
	}

	if timeout > 0 {
		zhttp.client.Timeout = timeout
	}

	if proxy != "" {
		p, err := url.Parse(proxy)
		if err != nil {
			return nil, fmt.Errorf("parse proxy: %w", err)
		}

		t := http.DefaultTransport.(*http.Transport)
		t.Proxy = func(*http.Request) (*url.URL, error) {
			return p, nil
		}
		zhttp.client.Transport = t
	}

	return zhttp, nil
}

// Get ...
func (zhttp *Zhttp) Get(url string) (int, []byte, error) {
	var code int
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, nil, fmt.Errorf("new request: %w", err)
	}

	resp, err := zhttp.client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	// response body
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return 0, nil, fmt.Errorf("reaad body: %w", err)
	}

	code = resp.StatusCode

	return code, body, nil
}
