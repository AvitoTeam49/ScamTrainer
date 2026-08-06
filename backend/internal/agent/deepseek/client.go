package deepseek

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/agent"
)

var _ agent.Agent = (*Client)(nil)

const (
	defaultBaseURL  = "https://api.deepseek.com"
	defaultModel    = "deepseek-v4-flash"
	defaultTimeout  = 60 * time.Second
	completionsPath = "/chat/completions"
	retryBaseDelay  = time.Second
	maxErrorBody    = 512
)

type Config struct {
	BaseURL    string
	APIKey     string
	Model      string
	Timeout    time.Duration
	MaxRetries int
}

type Client struct {
	baseURL    string
	apiKey     string
	model      string
	timeout    time.Duration
	maxRetries int
	retryDelay time.Duration
	httpClient *http.Client
}

func New(config Config) *Client {
	if config.BaseURL == "" {
		config.BaseURL = defaultBaseURL
	}
	if config.Model == "" {
		config.Model = defaultModel
	}
	if config.Timeout <= 0 {
		config.Timeout = defaultTimeout
	}
	if config.MaxRetries < 0 {
		config.MaxRetries = 0
	}

	return &Client{
		baseURL:    strings.TrimRight(config.BaseURL, "/"),
		apiKey:     config.APIKey,
		model:      config.Model,
		timeout:    config.Timeout,
		maxRetries: config.MaxRetries,
		retryDelay: retryBaseDelay,
		httpClient: &http.Client{Timeout: config.Timeout},
	}
}

func (c *Client) Complete(ctx context.Context, chatRequest agent.Request) (*agent.Reply, error) {
	body, err := json.Marshal(request{
		Model:    c.model,
		Messages: toAPIMessages(chatRequest.Messages),
		Tools:    toAPITools(chatRequest.Tools),
		Stream:   false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to encode deepseek request: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			if err := wait(ctx, c.retryDelay<<(attempt-1)); err != nil {
				return nil, err
			}
		}

		reply, retryable, err := c.do(ctx, body)
		if err == nil {
			return reply, nil
		}

		lastErr = err
		if !retryable {
			return nil, err
		}
	}

	return nil, lastErr
}

func (c *Client) do(ctx context.Context, body []byte) (*agent.Reply, bool, error) {
	requestCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	httpRequest, err := http.NewRequestWithContext(requestCtx, http.MethodPost, c.baseURL+completionsPath, bytes.NewReader(body))
	if err != nil {
		return nil, false, fmt.Errorf("failed to build deepseek request: %w", err)
	}

	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Authorization", "Bearer "+c.apiKey)

	httpResponse, err := c.httpClient.Do(httpRequest)
	if err != nil {
		if ctx.Err() != nil {
			return nil, false, fmt.Errorf("deepseek request cancelled: %w", ctx.Err())
		}
		return nil, true, fmt.Errorf("%w: %v", agent.ErrUnavailable, err)
	}
	defer func() { _ = httpResponse.Body.Close() }()

	payload, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, true, fmt.Errorf("%w: failed to read response body: %v", agent.ErrUnavailable, err)
	}

	switch {
	case httpResponse.StatusCode == http.StatusOK:
	case httpResponse.StatusCode == http.StatusTooManyRequests:
		return nil, true, fmt.Errorf("%w: %s", agent.ErrRateLimited, errorMessage(payload))
	case httpResponse.StatusCode >= http.StatusInternalServerError:
		return nil, true, fmt.Errorf("%w: status %d: %s", agent.ErrUnavailable, httpResponse.StatusCode, errorMessage(payload))
	default:
		return nil, false, fmt.Errorf("%w: status %d: %s", agent.ErrBadResponse, httpResponse.StatusCode, errorMessage(payload))
	}

	var decoded response
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return nil, false, fmt.Errorf("%w: failed to decode response: %v", agent.ErrBadResponse, err)
	}

	if len(decoded.Choices) == 0 {
		return nil, false, fmt.Errorf("%w: response contains no choices", agent.ErrBadResponse)
	}

	return replyFromMessage(decoded.Choices[0].Message), false, nil
}

func errorMessage(payload []byte) string {
	var decoded errorResponse
	if err := json.Unmarshal(payload, &decoded); err == nil && decoded.Error.Message != "" {
		return decoded.Error.Message
	}

	if len(payload) > maxErrorBody {
		payload = payload[:maxErrorBody]
	}

	return string(payload)
}

func wait(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return fmt.Errorf("deepseek retry cancelled: %w", ctx.Err())
	case <-timer.C:
		return nil
	}
}
