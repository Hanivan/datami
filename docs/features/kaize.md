# Kaize

## What

Top-anime listing scraped from `https://kaize.io`. No key. Accessed as `datami.Client.Kaize`.

| Method                | Result                                      |
| --------------------- | ------------------------------------------- |
| `Search(ctx, query, opts)` | `GET /anime/top?page=P` (`opts.Page`, min 1) |
| `ListAll(ctx, opts)`  | `Search` with empty query                   |
| `Get`, `GetByMalID`   | `ErrNotImplemented`                         |

## How it works

HTML parsed with `golang.org/x/net/html`: elements with class `anime-list-element`, then `<a>` whose class contains `name` gives title (text) and slug (last `href` segment). Items without title are skipped. Result has `Titles` (text only) and `Raw` (`slug`, `kaize`). Non-200 via `helpers.StatusError`: 403 `ErrRateLimited` (bot/WAF block, keyless source), 401 `ErrAuthFailed`, 404 `ErrNotFound`, 429 `ErrRateLimited`. Parse failure wraps as `parse: ...`. httpx rate 1 req/s, 3 attempts on network or 5xx errors.

## Config

`kaize.Config{Timeout, Metrics}`. Timeout default 10s. No custom transport or headers. `NewWithTestServer(baseURL, cfg)` for tests.

## Code

- `provider/kaize/client.go`, `config.go`, `mapping.go`, `models.go`

## Test

- `provider/kaize/client_test.go`: `TestSearch`, `TestGetNotImplemented`, `TestSearchForbidden`

## Gotchas

- `query` is ignored: `Search` always returns the top list for the page.
- Selector-based scraping: a site markup change yields empty results, not an error.
