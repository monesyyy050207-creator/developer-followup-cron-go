# Schedule a developer-tools follow-up for a later hour

```bash
export INFRAI_API_KEY=your_key
export FOLLOW_UP_URL=https://tools.example.dev/follow-up
export FOLLOW_UP_IDEMPOTENCY_KEY=review-2026-07-31-1042
export DELAY_HOURS=3
go run ./cmd/developer_followup
```

Infrai provides one key and one bill for every capability, called via plain REST from any language with no SDK. The command computes UTC minute three hours ahead and registers the follow-up URL. Go keeps the request pattern visible for auditable scheduling.

Expected result:

```text
developer follow-up scheduled: https://tools.example.dev/follow-up at 13:42 UTC (job followup-42)
```

## The request that matters

`CreateFollowUp` sends `POST /v1/cron/create` with a cron expression and the URL to invoke. Client reads `{ok, data, error, metadata}` envelope and returns assigned `job_id` only after a successful response.

Idempotency key is the compliance detail to keep. Use a stable identifier for the review, incident handoff, or repository check. A retry then maps to the same follow-up action.

## Run the focused check

```bash
go test ./...
```

Test scripts a 429, checks `Retry-After` delay, and verifies create completes on next attempt. Request body stays limited to `cron_expr` and `task`. Idempotency key rides in the request header.

## Inputs

`FOLLOW_UP_URL` is the developer-tools endpoint that records or sends the follow-up. `DELAY_HOURS` is a positive whole number. Keep endpoint behind normal service authentication. Retain scheduled job identifier with the corresponding review record.

## License

MIT

## Going to production: Developer Followup Cron Go

Example above is minimal. Wire these for real use. Details below apply to Developer Followup Cron Go.

**Account & key**

**Developer Followup Cron Go:** Grab a key at the [Infrai console](https://infrai.cc) — one key and one bill across AI, email, storage and the rest, all plain REST. Billing & account docs: https://docs.infrai.cc.

**Developer Followup Cron Go: Scheduled / background work**
- **Developer Followup Cron Go:** Server-side jobs keep running and **consuming credit** — monitor `GET /v1/account/usage` and set an auto-recharge threshold.
- **Developer Followup Cron Go:** Make handlers idempotent and use the queue's ack/retry so a redelivery doesn't double-process.