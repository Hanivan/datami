# Otakotaku

## What

ID cross-reference lookups from `https://otakotaku.com` (Indonesian site). No key. Accessed as `datami.Client.Otakotaku`.

| Method                 | Result                                                              |
| ---------------------- | ------------------------------------------------------------------- |
| `Get(ctx, id)`         | `GET /api/anime/view/{id}/placeholder` (Otakotaku ID; empty gives `ErrInvalidInput`) |
| `ListAll(ctx, opts)`   | same endpoint with `opts.Page` as the ID (min 1); returns one item  |
| `GetByMalID(ctx, malID)` | `ListAll` with ID = `malID`, then checks the MAL external ID matches |
| `Search`               | `ErrNotImplemented`                                                 |

## How it works

JSON via `helpers.DecodeJSON`. Non-200 via `helpers.StatusError`: 403 `ErrRateLimited` (bot/WAF block, keyless source), 401 `ErrAuthFailed`, 404 `ErrNotFound`, 429 `ErrRateLimited`. Mapping: title from `data.title`; `ExternalIDs` for MAL, AniDB, AnimeNewsNetwork, AnimePlanet when present; `Raw` keeps `otakotaku` ID and `title`. httpx rate 1 req/s, 3 attempts on network or 5xx errors.

## Config

`otakotaku.Config{Timeout, Metrics}`. Timeout default 10s. No custom transport or headers. `NewWithTestServer(baseURL, cfg)` for tests.

## Code

- `provider/otakotaku/client.go`, `config.go`, `mapping.go`, `models.go`

## Test

- `provider/otakotaku/client_test.go`: `TestGet`, `TestSearchNotImplemented`

## Gotchas

- `GetByMalID` passes the MAL ID as the Otakotaku ID; it only succeeds when both IDs coincide, else `ErrNotFound`. Non-numeric `malID` gives `ErrInvalidInput`.
- Value of `ListAll` is crawling by sequential ID, not a catalog listing.
