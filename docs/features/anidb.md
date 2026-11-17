# AniDB

## What

`provider/anidb`: client for the AniDB HTTP API (`http://api.anidb.net:9001/httpapi`, plain HTTP, no TLS on 9001) plus a local search over the official anime-titles dump (`https://anidb.net/api/anime-titles.xml.gz`). Reached as `datami.Client.AniDB`. Anime only. XML structures adapted from go-anidb (see NOTICE).

## How it works

| Method | Returns | Notes |
|--------|---------|-------|
| `Get(ctx, aid)` | `*types.Media` | `request=anime`, `protover=1`; `aid` must be a canonical positive integer string, else `ErrInvalidInput` with no request |
| `GetCharacters(ctx, aid)` | `[]types.Character` | same fetch as `Get`; keeps only charactertype id 1; "main character in" = `RoleMain`, all else `RoleSupporting`; `Character.Tags` reserved for a future scraper |
| `Search(ctx, q, opts)` | `[]types.Media` | local title index, no API call; titles + AniDB `ExternalID` only, call `Get` for detail; exact matches first then substring, aid order, deduped, default limit 10; empty or all-punctuation query is `ErrInvalidInput` |
| `LoadTitles(r io.Reader)` | `error` | preload dump (gzip or plain), marks it loaded now |
| `GetByMalID`, `GetManga`, `SearchManga`, `GetMangaByMalID` | none | `ErrNotImplemented` |

Responses are gunzipped by magic bytes. Titles are normalized (lowercase, letters/digits only, other runs collapse to one space).

Errors (all `types.Error{Source: SourceAniDB}`):

| Source | Error |
|--------|-------|
| HTTP 404 | `ErrNotFound` |
| 401, 403 | `ErrAuthFailed` |
| 429, 503 | `ErrRateLimited` |
| other >= 500 | `ErrUnavailable` |
| other non-200 | `ErrUnavailable` wrapping `status <code>` |
| `<error>` body with HTTP 200: starts "banned" | `ErrRateLimited` |
| `<error>` contains "not found" | `ErrNotFound` |
| `<error>` contains "client" | `ErrAuthFailed` |
| other `<error>` text | `ErrUnavailable` |

Titles dump cache (`titles.go`), file `<CacheDir>/anime-titles.xml.gz`:

- In-memory index fresh for 24h (`titlesMaxAge`); a fresh on-disk file (mtime within 24h, not in the future) is read instead of downloading. Else download, parse, write atomically (temp + rename, mtime stamped).
- Any failure after a request was sent starts a 1h cooldown (`titlesRetryCooldown`); during it Search serves the existing index or returns the last error. A failed refresh serves the stale file or index and retries after the cooldown.
- No cooldown when nothing was sent (limiter wait refused, ctx already done).
- One download at a time (1-slot lock); concurrent Searches share it; waiters honor ctx.
- Caps: 256 MiB decompressed dump (`maxTitlesDump`), 16 MiB decompressed anime XML (`maxAnimeXML`), 64 MiB raw HTTP body. An empty dump (no `<anime>`) is rejected as `ErrUnavailable`. Cache write failure is reported to `Metrics` only, not returned.

## Config

| Field | Root field | Meaning |
|-------|-----------|---------|
| `ClientName` | `AniDBClientName` | registered AniDB client name, required |
| `ClientVersion` | `AniDBClientVersion` | registered version, required, > 0 |
| `CacheDir` | `AniDBCacheDir` | default `os.UserCacheDir()/datami/anidb` |
| `Timeout` | `Timeout` | default 10s |
| `Metrics` | `Metrics` | `func(source, op, dur, err)` |

`anidb.New` returns `ErrInvalidInput` if name is empty or version <= 0. `datami.New` with empty `AniDBClientName` installs `anidb.Unconfigured()`: every method returns `ErrInvalidInput` with no network.

```go
c, _ := datami.New(datami.Config{AniDBClientName: "myclient", AniDBClientVersion: 1, AniDBCacheDir: "/var/cache/datami"})
m, _ := c.AniDB.Get(ctx, "1")
chars, _ := c.AniDB.GetCharacters(ctx, "1")
res, _ := c.AniDB.Search(ctx, "bebop", datami.SearchOpts{})
```

Fixed: `Rate: 0.5` (1 req per 2s, shared by API and dump download), burst 1, `MaxRetries: 1` (no retries), body cap 64 MiB.

## Code

`provider/anidb/`: `config.go`, `client.go` (API, status/body errors), `titles.go` (dump, cache, search), `mapping.go`, `models.go`, `constants.go`, `testdata/`.

## Test

`client_test.go`, `titles_test.go`, `characters_test.go`, `mapping_test.go`; fixtures in `testdata/` (`anime-titles.xml`, `anime_full.xml`, `anime_partial.xml`, `error_banned.xml`, `error_client.xml`, `error_notfound.xml`). `NewWithTestServer(apiURL, titlesURL, cfg)` defaults the client to `test`/1. Run `go test ./provider/anidb/`.

## Gotchas

- API is HTTP, not HTTPS; the dump is HTTPS.
- Rate limit is shared: a dump download delays the next `Get` by 2s.
- AniDB ids are their own namespace; no MAL lookup (`GetByMalID` is `ErrNotImplemented`, not a hard failure).
- Search results are thin (titles + id only).
- AniDB bans clients that exceed limits; a ban shows as `ErrRateLimited` via a 200 `<error>` body.
