# Kitsu

## What

Anime and manga metadata from the Kitsu JSON:API (`https://kitsu.io/api/edge`). No key needed. Accessed as `datami.Client.Kitsu`.

| Method                         | Result                                                                    |
| ------------------------------ | ------------------------------------------------------------------------- |
| `Get(ctx, id)`                 | `GET /anime/{id}`                                                         |
| `Search(ctx, query, opts)`     | `GET /anime?filter[text]=...&page[limit]=N` (`opts.Limit`, default `httpx.DefaultSearchLimit`) |
| `GetManga(ctx, id)`            | `GET /manga/{id}`                                                         |
| `SearchManga(ctx, query, opts)` | `GET /manga?filter[text]=...&page[limit]=N`                              |
| `GetByMalID(ctx, malID)`       | always `ErrNotFound` (no lookup implemented)                              |

Empty `id` or `query` returns `ErrInvalidInput` before any network call. No `ListAll`.

## How it works

JSON via `helpers.DecodeJSON`; non-200 goes through `helpers.StatusError`: 404 `ErrNotFound`, 429 `ErrRateLimited`, 401 and 403 `ErrAuthFailed` (Kitsu is not a keyless source, so 403 is not rate limit). `mapping.go` maps to `types.Media`: status via `kitsuStatusMap` (`current` ongoing, `finished`, `tba`/`upcoming`/`unreleased` planned; unknown lowercased), show type via `kitsuShowTypeMap` (`TV`, `OVA`, `ONA`, `movie`, ...), titles romaji (`en_jp`), english (`en`, skipped if equal to `en_jp`), native (`ja_jp`), score mean from `averageRating`, popularity from `userCount`, poster as `SizeMedium` and `SizeLarge` (original). httpx rate 4 req/s, 3 attempts on network or 5xx errors with 100ms * attempt backoff.

## Config

`kitsu.Config{Timeout, Metrics}`. Timeout default 10s. Root `datami.Config.Timeout` and `Metrics` are passed through. `Metrics(source, op, dur, err)` gets source `types.SourceKitsu`; `op` is the HTTP method from httpx. `NewWithTestServer(baseURL, cfg)` overrides the base URL.

## Code

- `provider/kitsu/client.go`, `config.go`, `constants.go`, `mapping.go`, `models.go`

## Test

- `provider/kitsu/client_test.go`: `TestGetAnime`, `TestGetNotFound`, `TestGetManga`, `TestSearchManga`

## Gotchas

- `GetManga`/`SearchManga` reuse the anime mapper: `ExternalID.ResourceType` is always `ResourceAnime`, and `SearchManga` decode errors carry `Op` `search`, not `searchmanga`.
- `GetByMalID` returns `ErrNotFound`, not `ErrNotImplemented`.
