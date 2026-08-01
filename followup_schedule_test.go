package delayedfollowup

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type scriptedDoer struct {
	requests int
}

func (d *scriptedDoer) Do(req *http.Request) (*http.Response, error) {
	d.requests++
	if req.Method != http.MethodPost {
		return nil, nil
	}
	if d.requests == 1 {
		return response(http.StatusTooManyRequests, `{"ok":false}`, "1"), nil
	}
	return response(http.StatusOK, `{"ok":true,"data":{"job_id":"followup-42"}}`, ""), nil
}

func response(status int, body, retryAfter string) *http.Response {
	headers := make(http.Header)
	if retryAfter != "" {
		headers.Set("Retry-After", retryAfter)
	}
	return &http.Response{StatusCode: status, Header: headers, Body: io.NopCloser(strings.NewReader(body))}
}

func TestCreateFollowUpRetriesWithSameIdempotencyKey(t *testing.T) {
	doer := &scriptedDoer{}
	var slept time.Duration
	client := &CronClient{apiKey: "test-key", http: doer, sleep: func(d time.Duration) { slept = d }}

	jobID, err := client.CreateFollowUp(context.Background(), "0 14 * * *", "https://tools.example.dev/follow-up", "review-123")
	if err != nil {
		t.Fatal(err)
	}
	if jobID != "followup-42" || doer.requests != 2 || slept != time.Second {
		t.Fatalf("job=%q requests=%d sleep=%s", jobID, doer.requests, slept)
	}
}
