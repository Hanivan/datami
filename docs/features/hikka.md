# Hikka

## What

Anime listing from the Hikka API (`https://api.hikka.io`, Ukrainian community DB). No key. Accessed as `datami.Client.Hikka`.

| Method                     | Result                                                 |
| -------------------------- | ------------------------------------------------------ |
| `Search(ctx, query, opts)` | `POST /anime?page=P&size=N` with JSON filter body      |
| `ListAll(ctx, opts)`       | `Search` with empty query                              |
| `Get`, `GetByMalID`        | `ErrNotImplemented`                                    |

## How it works

Body has empty filter arrays, `only_translated: false`, `sort: ["score:desc"]`, and `search` only when query is non-empty. `opts.Page` (min 1) and `opts.Limit` (default 100) set `page` and `size`. Sent with headers `Accept: */*`, `Accept-Language: en`, `Content-Type: application/json`, `Origin: https://hikka.io`. Response is JSON (`list` plus `pagination`), decoded via `helpers.DecodeJSON`. Non-200 via `helpers.StatusError`: 403 `ErrRateLimited` (bot/WAF block, keyless source), 401 `ErrAuthFailed`, 404 `ErrNotFound`, 429 `ErrRateLimited`. httpx rate 2 req/s, 3 attempts on network or 5xx errors.

Mapping: titles `en`, `ja`, `ua` (no `Kind`); if none present, slug is the title (only when `title_ja` is also empty). MAL ID goes to `ExternalIDs` (`types.SourceMAL`). `Raw` keeps `slug` and the three titles. No status, type or images.

## Config

`hikka.Config{Timeout, Metrics}`. Timeout default 10s, fed from root `datami.Config`. `NewWithTestServer(baseURL, cfg)` for tests.

## Code

- `provider/hikka/client.go`, `config.go`, `mapping.go`, `models.go`

## Test

- `provider/hikka/client_test.go`: `TestSearch`, `TestGetNotImplemented`, `TestSearchSkipsZeroMalID`

## Gotchas

- Items with `mal_id` 0 are dropped from results (`handleListResponse`), so page sizes can be smaller than `size`.
- Empty query is valid (lists by score).
