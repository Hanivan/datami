# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

`datami` - Go 1.27 library providing a unified client for anime/manga metadata across multiple sources: MyAnimeList (MAL), AniList, AniDB, Kitsu, Hikka, Kaize, Nautiljon, Otakotaku. Module path: `github.com/Hanivan/datami`.

## Commands

```bash
go test ./...                       # all tests
go test ./anilist/                  # single package
go test -run TestClientGetSuccess ./anilist/  # single test
go build ./...
go vet ./...
```

## Architecture

**Root package (`datami`)** is a facade: it aggregates all source clients into one `Client` struct and re-exports `types` symbols as type aliases so consumers write `datami.Media` instead of `types.Media`.

**`types/`** - shared domain model. `Media` is the canonical response shape every source maps to. Sentinel errors (`ErrNotFound`, `ErrRateLimited`, `ErrAuthFailed`, `ErrUnavailable`, `ErrInvalidInput`) are defined here. `Error` wraps them with `Source` + `Op` for provenance and supports `Unwrap()` for `errors.Is`.

**Source packages (`provider/anilist/`, `provider/mal/`, `provider/anidb/`, `provider/kitsu/`, `provider/hikka/`, `provider/kaize/`, `provider/nautiljon/`, `provider/otakotaku/`)** - each follows an identical structure:
- `constants.go` - status/format normalization maps, when the provider has any (use `types.Source*` directly, no local `sourceName` const)
- `config.go` - source-specific `Config` + `New()` constructor
- `client.go` - `Get`, `Search`, `GetByMalID` methods
- `models.go` - raw API response structs
- `mapping.go` - `toMedia()` converts raw → `types.Media`

**`internal/httpx/`** - shared HTTP client with timeout, rate limiting, and a metrics callback. Sources inject the source name into the callback; `op` is the provider `Op*` const passed to `Get(ctx, op, url)`/`Post`/`Do(op, req)`, which also return Retry-After seconds. Accepts a `Transport http.RoundTripper` in `Config` for providers that need custom bot-handling (cloudscraper, CSRF, custom headers). Caps response bodies at `Config.MaxBody` (default `DefaultMaxBody` 32 MiB; over returns `ErrBodyTooLarge`, not retried). Exports shared consts: `DefaultTimeout`, `MaxRetries`, `RetryBackoff`, `DefaultSearchLimit`, `ServerErrThreshold`, `ContentTypeJSON`, `HeaderContentType`, `HeaderAccept`, `HeaderAuthorization`, `BearerPrefix`.

**`internal/helpers/`** - shared mapping helpers. Use these instead of redefining per-provider: `ParseDate`, `RawPayload`, `NormalizeRelation`, `DerefString`, `DerefInt`, `TextContent`, `HasClass`, `TdNodes`.

**`types/enums.go`** - typed enums. Use these instead of raw strings: `MediaStatus` (`StatusFinished`, `StatusOngoing`, etc.), `MediaType` (`TypeTV`, `TypeManga`, etc.), `TitleKind` (`KindCanonical`, `KindRomaji`, etc.), `Season` (`SeasonWinter`, etc.), `RelationKind` (`RelationSequel`, etc.), `ImageSize` (`SizeMedium`, `SizeLarge`, `SizeOriginal`), `ResourceType` (`ResourceAnime`, `ResourceManga`), `CharacterRoleType` (`RoleMain`, `RoleSupporting`).

**`types.Media` typed fields** - `Status MediaStatus`, `MediaType MediaType`, `Season Season`, `Images map[ImageSize]Image`, `Title.Kind TitleKind`, `Relation.Kind RelationKind`, `ExternalID.ResourceType ResourceType`, `CharacterRole.Role CharacterRoleType`. Assign with explicit casts when needed.
**`types.SearchOpts`** includes a `Page` field for paginated `ListAll` calls.

## Conventions

Do not use emojis or em-dashes in any output, comments, or documentation.

Do NOT commit without telling the user first. Exception: if the user invokes `/commit-commands:commit`, proceed without asking.

- Every error from a source is wrapped as `types.Error{Source: types.Source*, Op: Op*, Err: ...}`. Use `types.SourceMAL`, `types.SourceAniList`, etc. from `types/source.go` — never hardcode provider name strings. Each provider's `client.go` has an `Op*` const block (`OpGet`, `OpSearch`, etc.) — use these instead of raw op strings.
- Empty-input validation returns `types.ErrInvalidInput` before any network call.
- Tests use `httptest.NewServer` + a `NewWithTestServer(baseURL, cfg)` constructor that bypasses real endpoints.
- Source status/format strings (e.g. `"RELEASING"` → `"ongoing"`) are normalized in `mapping.go` maps, not in the client.
- Some providers don't support all methods - they return `types.ErrNotImplemented` (e.g., `Hikka.Get`, `Kaize.Get`, `Nautiljon.Get`, `Otakotaku.Search`, `Otakotaku.GetByMalID`, `AniDB.GetManga`, `AniDB.GetByMalID`). Don't treat as a hard error.
- Non-goal endpoints (not in mono-crawler catalog path) are stubbed with `ErrNotImplemented`: MAL forum/user-list/mutation methods, AniList viewer/list/mutation methods. `types/forum.go` (`ForumBoard`, `ForumPost`, `ForumTopic`) and `types/user.go` (`UserInfo`, `MediaListEntry`, `MediaListCollection`) hold the stub types for these.
- AniDB: needs registered client (`Config.ClientName`/`ClientVersion`, root `AniDBClientName`/`AniDBClientVersion`/`AniDBCacheDir`); empty name returns `ErrInvalidInput`. `Search` reads the anime-titles dump (cached in CacheDir, 24h refresh, 1h retry after failure, `LoadTitles(io.Reader)` preloads). `GetCharacters(ctx, aid)` exists; `Character.Tags` reserved for a future scraper. Rate limit 1 req/2s, no retries. XML structures adapted from go-anidb (see NOTICE).
- AniList `Get`/`GetManga`/`GetByMalID`/`GetMangaByMalID` take variadic `types.MediaOpts{CharacterPages, StaffPages}` to embed cast/staff pages (25 edges each, aliases `c1..cN`/`s1..sN`) into the media query, with a standalone follow-up when more remain; `GetCharacters*`/`GetStaff*` take variadic `types.PageOpts{StartPage, MaxPages}`. Total pages per connection never exceed 10.
- AniList `GetByMalIDs`/`GetMangaByMalIDs` (`batch.go`) fetch up to 50 works by MAL id in one `Page(media(idMal_in))` request, same `MediaOpts` embedding per media and follow-ups only where more remain; result is keyed by MAL id, unknown ids absent, empty input makes no request, more than 50 is `ErrInvalidInput`.
- HTML-based providers (`kaize`, `nautiljon`) parse responses with `golang.org/x/net/html`.
- HTTP status mapping: `helpers.StatusError` / `helpers.DecodeJSON[T]`. 401 is `ErrAuthFailed`; 403 is `ErrAuthFailed` for keyed sources but `ErrRateLimited` for keyless ones (hikka, kaize, nautiljon, otakotaku), where 403 means a bot/WAF block. AniList (403 = rate limit) and AniDB (body errors, 503) keep their own mappers.

## Quality

```bash
gofmt -l .        # list unformatted files
gofmt -w <files>  # format specific files
```

This repo does not auto-format. Always run `gofmt -l .` before finishing — CI will reject unformatted files.

## Docs

- `docs/features/` - per-feature docs (one file per provider + configuration, errors, http-client, media-model). Keep README.md short; put details here and update the matching file when behaviour changes.
- `docs/api-docs/` - provider API reference docs (MAL, AniList, AniDB)
- `docs/superpowers/specs/` - design specs (gitignored, local only)
