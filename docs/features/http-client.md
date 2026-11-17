# HTTP Client

## What

`internal/httpx` is the shared HTTP client every provider uses: timeout, token-bucket rate limiter, retries, body cap, Retry-After parsing and a metrics callback. Internal package, not importable by consumers.

## How it works

`httpx.New(Config)` then `Get(ctx, op, url)`, `Post(ctx, op, url, body, headers)`, `Do(op, req)`. All return status, body, retry-after seconds, error. `op` is the provider `Op*` constant, reported to the metrics callback.

Per request, up to `maxRetries` attempts (total attempts, not extra):

1. attempt > 0: sleep `attempt * RetryBackoff` (100ms, 200ms, ...).
2. If limiter set, `limiter.Wait(ctx)` on every attempt. Failure returns `ErrLimiterWait` (wrapping the cause), no request sent.
3. Send. Transport error: record, retry.
4. Read body through `io.LimitReader(maxBody+1)`. Over cap: `ErrBodyTooLarge`, returned immediately (status kept, body nil), not retried.
5. Read error: retry. Status `>= 500` (`ServerErrThreshold`): retry. Else return status, body, Retry-After.

After exhaustion: last status and body with error `after N attempts: <cause>`. 4xx (including 429) is returned as a normal response, not an error, not retried.

Retry-After: delta-seconds form only; missing, negative or HTTP-date value gives 0.

Metrics callback `func(op string, dur time.Duration, err error)` fires after every attempt with `op` = the provider operation name passed to `Get`/`Post`/`Do` (the provider `Op*` const, e.g. `mal.OpGet`); `err` is the transport or read/size error (nil on any HTTP response). Providers wrap it to prepend the source name.

## Config

| Field | Default |
| --- | --- |
| `Timeout` | `DefaultTimeout` 10s |
| `Rate` | 0 = no limiter; else req/s |
| `Burst` | 1 |
| `MaxRetries` | 3 (`MaxRetries` const) |
| `MaxBody` | `DefaultMaxBody` = 32 MiB |
| `Transport` | http default |
| `Headers` | none; set on every request |
| `Metrics` | nil |

Per provider:

| Provider | Rate (req/s) | Burst | MaxRetries | MaxBody | Transport / Headers |
| --- | --- | --- | --- | --- | --- |
| mal | 2 | 1 | 3 | 32 MiB | Header `X-MAL-CLIENT-ID` |
| anilist | 1.5 | 1 | 3 | 32 MiB | per-request `Authorization: Bearer` if token |
| anidb | 0.5 | 1 | 1 | 64 MiB | none |
| kitsu | 4 | 1 | 3 | 32 MiB | none |
| hikka | 2 | 1 | 3 | 32 MiB | per-request headers on POST (Accept, Accept-Language, Content-Type, Origin) |
| kaize | 1 | 1 | 3 | 32 MiB | none |
| nautiljon | 1 | 1 | 3 | 32 MiB | `rewriteTransport` (currently delegates to `http.DefaultTransport`) |
| otakotaku | 1 | 1 | 3 | 32 MiB | none |

Burst is unset (defaults to 1) in all providers except AniDB, which sets it explicitly to 1.

## Code

```go
c := httpx.New(httpx.Config{Rate: 2, MaxRetries: 3, MaxBody: 1 << 20})
code, body, retryAfter, err := c.Get(ctx, "get", url)
if errors.Is(err, httpx.ErrBodyTooLarge) { /* ... */ }
```

## Test

```bash
go test ./internal/httpx/
```

Covers success, `MaxBody`, `MaxRetries`, per-attempt limiter, `Retry-After`, metrics `op`, POST body resend on retry.

## Gotchas

- `MaxRetries: 1` means one request, no retry (AniDB, to respect its ban rules).
- Sleep between retries ignores ctx cancellation (`time.Sleep`); the limiter wait does honor ctx.
- 4xx statuses mapped to errors by providers, see [errors.md](errors.md); config entry point in [configuration.md](configuration.md).
