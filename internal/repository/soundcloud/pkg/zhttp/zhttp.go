package zhttp

import (
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
	req, err := http.Get(url)
	if err != nil {
		return 0, nil, err
	}
	// response body
	body, err := io.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		return 0, nil, fmt.Errorf("reaad body: %w", err)
	}

	code = req.StatusCode

	return code, body, nil
}
