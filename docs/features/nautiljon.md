# Nautiljon

## What

Anime listing scraped from `https://www.nautiljon.com` (French). No key. Accessed as `datami.Client.Nautiljon`.

| Method                | Result                                              |
| --------------------- | --------------------------------------------------- |
| `Search(ctx, query, opts)` | `GET /animes/?dbt={(page-1)*15}` (`opts.Page`, min 1) |
| `ListAll(ctx, opts)`  | `Search` with empty query                           |
| `Get`, `GetByMalID`   | `ErrNotImplemented`                                 |

## How it works

HTML parsed with `golang.org/x/net/html`: finds a `<table>` whose class contains `search`, then each `<tr>` with at least 4 `<td>`; columns 1, 2, 3 are title, format, status (text). Status is lowercased/trimmed and mapped via `nautiljonStatusMap` (`en cours` ongoing, `terminé` finished); unmatched leaves `Status` empty. `Raw` keeps `slug`, `format`, `status`. Non-200 via `helpers.StatusError`: 403 `ErrRateLimited` (bot/WAF block, keyless source), 401 `ErrAuthFailed`, 404 `ErrNotFound`, 429 `ErrRateLimited`. httpx rate 1 req/s, 3 attempts on network or 5xx errors.

## Config

`nautiljon.Config{Timeout, Metrics}`. Timeout default 10s. Uses a custom `rewriteTransport` that currently just delegates to `http.DefaultTransport` (placeholder for a cloudscraper-style transport). `NewWithTestServer(baseURL, cfg)` for tests.

## Code

- `provider/nautiljon/client.go`, `config.go`, `constants.go`, `mapping.go`, `models.go`

## Test

- `provider/nautiljon/client_test.go`: `TestSearch`, `TestGetNotImplemented`

## Gotchas

- `query` is ignored: always returns the page of the full list (15 rows per page assumed).
- `Slug` is never filled by the parser, so `Raw["slug"]` is empty.
- No bot-bypass in the transport: Cloudflare blocks surface as `ErrRateLimited` (403).
