package httpcaller

import (
	"elysium/internal/application/logger"
	"github.com/mymmrac/telego/telegoapi"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
	"log/slog"
	"os"
	"testing"
	"time"
)

func constructRequestData(parameters any) (*telegoapi.RequestData, error) {
	data, err := telegoapi.DefaultConstructor{}.JSONRequest(parameters)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func TestFastHttpCallerWithLimiter(t *testing.T) {
	loggerModule := logger.New(
		logger.Options{
			AppName: "httpclient",
			Writer:  os.Stdout,
			HandlerOptions: &slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		},
	)
	rl := rate.NewLimiter(rate.Every(1*time.Second), 100)
	client := NewFastHttpCallerWithLimiter(rl, loggerModule)

	parameters := map[string]interface{}{
		"chat_id": "251636949",
		"text":    "Test message",
	}

	data, err := constructRequestData(parameters)
	if err != nil {
		t.Fatalf("Error constructing request data: %v", err)
	}

	url := "https://api.telegram.org/bot7894673045:AAHgosEAHjdW78q44bPTSuwVqSZl8SEN0-w/sendMessage"

	for i := 0; i < 300; i++ {
		resp, err := client.Call(url, data)
		assert.NoError(t, err)
		assert.True(t, resp.Ok)
		t.Logf(resp.String())
	}
}

func TestFastHttpCallerWithLimiter_RateLimitExceeded(t *testing.T) {
	loggerModule := logger.New(
		logger.Options{
			AppName: "httpclient",
			Writer:  nil,
			HandlerOptions: &slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		},
	)
	rl := rate.NewLimiter(rate.Every(1*time.Second), 1) // Strict limit: 1 request per second
	client := NewFastHttpCallerWithLimiter(rl, loggerModule)

	parameters := map[string]interface{}{
		"chat_id": "123456789",
		"text":    "Test message",
	}

	data, err := constructRequestData(parameters)
	assert.NoError(t, err)

	url := "https://api.telegram.org/bot7486051673:AAEg2bzMqec1NkFK8tHycLn8gvGxK6xQ6ww/sendMessage"

	// First request should pass
	resp, err := client.Call(url, data)
	assert.NoError(t, err)
	assert.True(t, resp.Ok)

	// Second request should wait due to rate limit
	start := time.Now()
	resp, err = client.Call(url, data)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.True(t, resp.Ok)
	assert.True(t, duration >= 1*time.Second)
}
