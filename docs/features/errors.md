# Errors

## What

Every provider failure is a `*types.Error` (`types/errors.go`) wrapping a sentinel, so `errors.Is` and `types.StatusOf` work across all sources. Root package re-exports `Error` and the five main sentinels (not `ErrNotImplemented`).

## How it works

`types.Error` fields:

| Field | Meaning |
| --- | --- |
| `Source` | `types.Source*` string, e.g. "anilist" |
| `Op` | provider `Op*` const, e.g. `OpGet` |
| `Err` | wrapped cause; `Unwrap()` returns it |
| `Status` | HTTP status answered, 0 if not an HTTP response |
| `RetryAfter` | seconds from `Retry-After` header, 0 if none |

`Error()` renders `source.op: cause`. Helpers: `types.StatusOf(err)` and `types.RetryAfterOf(err)` (both 0 when no `*Error` in chain). Branch on `Status`, not message text.

Sentinels: `ErrNotFound`, `ErrRateLimited`, `ErrAuthFailed`, `ErrUnavailable`, `ErrInvalidInput`, `ErrNotImplemented`.

Shared mapper `helpers.StatusError` / `helpers.DecodeJSON` (`internal/helpers/errors.go`, `response.go`):

| HTTP | Sentinel |
| --- | --- |
| 404 | `ErrNotFound` |
| 429 | `ErrRateLimited` |
| 403 | `ErrAuthFailed`; `ErrRateLimited` for hikka, kaize, nautiljon, otakotaku (keyless, 403 = bot/WAF block) |
| 401 | `ErrAuthFailed` |
| other non-200 | plain `status N` error (no sentinel) |

`DecodeJSON` JSON parse failure: `Error{Err: "parse: ..."}` with `Status` 0. Used by mal, kitsu, hikka, kaize, nautiljon, otakotaku.

AniList own mapper (`statusErr`, `provider/anilist/client.go`), sets `Status` and `RetryAfter`: 429 and 403 = `ErrRateLimited` (403 means suspension or IP block), 404 = `ErrNotFound`, 401 = `ErrAuthFailed`, 400 = `ErrInvalidInput`, else `status N`. Also `anilist.ErrMediaNotFound` (`errors.New("anilist: media not found")`) from tags/links fetch when Media is missing, wrapped in `types.Error` (`OpGetTagsLinks`).

AniDB own mappers (`provider/anidb/client.go`): `statusError` 404 = `ErrNotFound`; 401/403 = `ErrAuthFailed`; 429/503 = `ErrRateLimited`; other `>= 500` = `ErrUnavailable`; rest = `ErrUnavailable` with "status N". `bodyError` maps HTTP 200 `<error>` text: prefix "banned" = `ErrRateLimited`, contains "not found" = `ErrNotFound`, contains "client" = `ErrAuthFailed`, else `ErrUnavailable`. Bad aid (non-positive-integer) or unconfigured client = `ErrInvalidInput`.

httpx errors (see [http-client.md](http-client.md)): `httpx.ErrLimiterWait` (limiter or ctx failed, no request sent), `httpx.ErrBodyTooLarge` (body over cap, not retried). Exhausted retries: `after N attempts: ...`.

`ErrNotImplemented` stubs: Hikka `Get`/`GetByMalID`; Kaize `Get`/`GetByMalID`; Nautiljon `Get`/`GetByMalID`; Otakotaku `Search`; AniDB `GetByMalID`, `GetManga`, `SearchManga`, `GetMangaByMalID`; MAL forum, user list/info, list mutation methods; AniList viewer, media list, save/delete list entry.

Empty input: `ErrInvalidInput` before any network call.

## Code

```go
_, err := c.MAL.Get(ctx, "1")
switch {
case errors.Is(err, datami.ErrNotFound):
case errors.Is(err, datami.ErrRateLimited):
    wait := types.RetryAfterOf(err) // 0 if header absent
case errors.Is(err, types.ErrNotImplemented):
}
if types.StatusOf(err) == 403 { /* ... */ }
var e *types.Error
if errors.As(err, &e) { log.Println(e.Source, e.Op, e.Status) }
```

## Test

```bash
go test ./types/ ./internal/helpers/ ./provider/anilist/ ./provider/anidb/
```

`types/errors_status_test.go` covers `Status`/`StatusOf`.

## Gotchas

- `Status` is 0 for validation, parse and transport errors.
- `RetryAfter` is set by every provider on a non-200 response when the server sends a delta-seconds `Retry-After`; 0 otherwise.
- After 5xx retries are exhausted the error is a plain wrapped error from httpx; sentinel mapping applies only when a non-5xx status is returned (AniDB also maps the leftover 5xx status).
- 403 differs per source: keyless and AniList = rate limited, MAL/Kitsu = auth failed, AniDB = auth failed.
