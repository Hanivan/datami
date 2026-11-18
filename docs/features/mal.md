# MAL

## What

`provider/mal`: client for the official MyAnimeList API v2 (`https://api.myanimelist.net/v2`). Reached as `datami.Client.MAL`. Anime and manga detail, search, seasonal, ranking, characters, people. Forum, user-list and mutation methods are stubs.

## How it works

| Method | Returns | Notes |
|--------|---------|-------|
| `Get(ctx, id)` | `*types.Media` | anime by MAL id, one wide `fields` list |
| `Search(ctx, q, opts)` | `[]types.Media` | anime; limit defaults to `httpx.DefaultSearchLimit` (10) |
| `GetByMalID(ctx, malID)` | `*types.Media` | alias of `Get` |
| `GetManga` / `SearchManga` / `GetMangaByMalID` | manga equivalents | `GetMangaByMalID` aliases `GetManga` |
| `GetSeasonal(ctx, year, season, opts)` | `[]types.Media` | `opts.Page` 1-based, mapped to `offset=(Page-1)*limit`; no total, last page = fewer than `Limit` items |
| `GetRanking` / `GetMangaRanking(ctx, rankingType, opts)` | `[]types.Media` | `ranking_type` passed through |
| `GetAnimeCharacters(ctx, animeID)` | `[]types.Character` | `limit=100`, single request, no paging |
| `GetCharacter(ctx, charID)` | `*types.Character` | includes animeography |
| `GetPerson(ctx, personID)` | `*types.Person` | |

Empty id or query returns `types.ErrInvalidInput` before any network call. List endpoints (search, seasonal, ranking) return only id/title/picture unless `fields` asks for more, and unsupported fields are silently dropped, so anime and manga use separate field sets (`animeListFields`, `mangaListFields` in `client.go`).

Stubs returning `types.ErrNotImplemented` (non-goal, not in mono-crawler catalog path): `GetForumBoards`, `GetForumTopic`, `SearchForumTopics`, `GetUserAnimeList`, `GetUserMangaList`, `GetUserInfo`, `UpdateAnimeListStatus`, `DeleteAnimeListItem`.

Errors: all wrapped `types.Error{Source: types.SourceMAL, Op: Op*}`. Non-200 goes through `helpers.StatusError` via `helpers.DecodeJSON`:

| HTTP | Error |
|------|-------|
| 404 | `ErrNotFound` |
| 429 | `ErrRateLimited` |
| 401, 403 | `ErrAuthFailed` (403 is auth here; only keyless sources map 403 to rate limited) |
| other | `status <code>` text; `Error.Status` carries the code |

Transport failures after retries return the wrapped httpx error. `Retry-After` is not surfaced for MAL (`Do` discards it).

## Config

| Field | Default | Meaning |
|-------|---------|---------|
| `UserAgent` | `datami (+https://github.com/Hanivan/datami)` | `User-Agent` on every request; set your app name so MAL can identify and throttle you |
| `ClientID` (root: `MALClientID`) | none | sent as `X-MAL-CLIENT-ID` header on every request; not validated, empty is sent as-is |
| `Timeout` | 10s | per-request |
| `Metrics` | nil | `func(source, op, dur, err)` |

Fixed `httpx.Config`: `Rate: 2` req/s, `Burst` 1, 3 total attempts (`httpx.MaxRetries`), backoff `attempt * 100ms`, retries on transport error or status >= 500, body cap 32 MiB.

## Code

- `provider/mal/config.go`: `Config`, `New`, `NewWithTestServer`
- `client.go`: media methods, stubs, response handlers
- `characters.go`: characters and people
- `mapping.go`, `models.go`, `constants.go`: raw structs, `toMedia()`, status/format maps

## Test

`provider/mal/client_test.go` (`httptest.NewServer` + `mal.NewWithTestServer(baseURL, cfg)`): `go test ./provider/mal/`.

## Gotchas

- MAL anime and manga ids are separate namespaces. `Get` is anime only; use `GetManga` for manga.
- The "MAL id" in `*ByMalID` is the same id as `Get`; those exist for interface symmetry with AniList/AniDB.
- Manga list endpoints do not return `num_episodes`; asking for it yields no error and no value.
- Characters capped at 100 per anime (no pagination).
