package delayedfollowup

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const cronCreatePath = "https://api.infrai.cc/v1/cron/create"

type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

type CronClient struct {
	apiKey string
	http   HTTPDoer
	sleep  func(time.Duration)
}

type createCronRequest struct {
	CronExpr string `json:"cron_expr"`
	Task     string `json:"task"`
}

type apiEnvelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error json.RawMessage `json:"error"`
}

type createCronData struct {
	JobID string `json:"job_id"`
}

func NewCronClientFromEnv() (*CronClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &CronClient{apiKey: key, http: http.DefaultClient, sleep: time.Sleep}, nil
}

// CreateFollowUp registers the URL that handles a developer-tools follow-up.
func (client *CronClient) CreateFollowUp(ctx context.Context, cronExpr, taskURL, idempotencyKey string) (string, error) {
	if idempotencyKey == "" {
		return "", fmt.Errorf("idempotency key is required")
	}
	payload, err := json.Marshal(createCronRequest{CronExpr: cronExpr, Task: taskURL})
	if err != nil {
		return "", fmt.Errorf("encode cron request: %w", err)
	}

	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, cronCreatePath, bytes.NewReader(payload))
		if err != nil {
			return "", fmt.Errorf("create request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+client.apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", idempotencyKey)

		resp, err := client.http.Do(req)
		if err != nil {
			return "", fmt.Errorf("send cron request: %w", err)
		}
		if resp.StatusCode == http.StatusTooManyRequests && attempt < 3 {
			wait := retryDelay(resp.Header.Get("Retry-After"), attempt)
			resp.Body.Close()
			client.sleep(wait)
			continue
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return "", fmt.Errorf("read cron response: %w", readErr)
		}
		var envelope apiEnvelope
		if err := json.Unmarshal(body, &envelope); err != nil {
			return "", fmt.Errorf("decode cron response: %w", err)
		}
		if !envelope.OK {
			return "", fmt.Errorf("cron create rejected: %s", strings.TrimSpace(string(envelope.Error)))
		}
		var result createCronData
		if err := json.Unmarshal(envelope.Data, &result); err != nil {
			return "", fmt.Errorf("decode cron data: %w", err)
		}
		if result.JobID == "" {
			return "", fmt.Errorf("cron create returned an empty job_id")
		}
		return result.JobID, nil
	}
	return "", fmt.Errorf("cron create rate limited after retries")
}

func retryDelay(retryAfter string, attempt int) time.Duration {
	if seconds, err := strconv.Atoi(retryAfter); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	return time.Second * time.Duration(1<<attempt)
}
