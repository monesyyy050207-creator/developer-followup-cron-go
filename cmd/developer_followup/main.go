package main

import (
	"context"
	"fmt"
	"os"
	"time"

	delayedfollowup "example.com/developer-followup-cron"
)

func main() {
	client, err := delayedfollowup.NewCronClientFromEnv()
	if err != nil {
		fatal(err)
	}

	hours, err := parseHours(os.Getenv("DELAY_HOURS"))
	if err != nil {
		fatal(err)
	}
	taskURL := os.Getenv("FOLLOW_UP_URL")
	key := os.Getenv("FOLLOW_UP_IDEMPOTENCY_KEY")
	if taskURL == "" || key == "" {
		fatal(fmt.Errorf("FOLLOW_UP_URL and FOLLOW_UP_IDEMPOTENCY_KEY are required"))
	}

	target := time.Now().UTC().Add(time.Duration(hours) * time.Hour)
	cronExpr := fmt.Sprintf("%d %d * * *", target.Minute(), target.Hour())
	jobID, err := client.CreateFollowUp(context.Background(), cronExpr, taskURL, key)
	if err != nil {
		fatal(err)
	}
	fmt.Printf("developer follow-up scheduled: %s at %s UTC (job %s)\n", taskURL, target.Format("15:04"), jobID)
}

func parseHours(value string) (int, error) {
	var hours int
	if _, err := fmt.Sscanf(value, "%d", &hours); err != nil || hours < 1 {
		return 0, fmt.Errorf("DELAY_HOURS must be a positive whole number")
	}
	return hours, nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
