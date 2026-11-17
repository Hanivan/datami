# AniList

## What

`provider/anilist`: client for the AniList GraphQL API (`https://graphql.anilist.co`, POST). Reached as `datami.Client.AniList`. Media, search, characters with voice actors, staff, tags and external links, with by-AniList-id and by-MAL-id variants.

## How it works

| Method | Returns | Notes |
|--------|---------|-------|
| `Get(ctx, id)` | `*types.Media` | AniList id, no type filter |
| `Search(ctx, q, opts)` | `[]types.Media` | `Page(perPage: limit)` over all types; default limit 10; no page offset |
| `GetByMalID(ctx, malID)` | `*types.Media` | `idMal` with `type: ANIME` forced |
| `GetManga` / `SearchManga` | `*types.Media` / `[]types.Media` | `type: MANGA` |
| `GetMangaByMalID(ctx, malID)` | `*types.Media` | `idMal` with `type: MANGA` |
| `GetCharacters(ctx, id)` / `GetCharactersByMalID(ctx, malID, mediaType)` | `[]types.Character` | role + voice actors; 25 per page, up to 10 pages (250 max) |
| `GetStaff(ctx, id)` / `GetStaffByMalID(ctx, malID, mediaType)` | `[]types.Person` | role + primary occupations; same paging |
| `GetTagsAndLinks(ctx, id)` / `GetTagsAndLinksByMalID(ctx, malID, mediaType)` | `*TagsLinks` | tags (spoiler/adult flags) + external links |

`mediaType` is `"ANIME"`, `"MANGA"` or `""` (no type filter). Empty id/query returns `ErrInvalidInput`. A null `Media` returns `ErrNotFound`; for tags/links it returns `anilist.ErrMediaNotFound` instead, distinct from nil error with empty slices (media resolved, has none).

Stubs returning `ErrNotImplemented`: `GetViewer`, `GetMediaListEntry`, `GetMediaListCollection`, `SaveMediaListEntry`, `DeleteMediaListEntry`.

Status mapping (`statusErr` in `client.go`, not `helpers.StatusError`):

| HTTP | Error |
|------|-------|
| 429, 403 | `ErrRateLimited` (403 = global suspension or per-IP block, not authorization) |
| 404 | `ErrNotFound` |
| 401 | `ErrAuthFailed` |
| 400 | `ErrInvalidInput` |
| other | `status <code>` |

`types.Error` carries `Status` and `RetryAfter` (seconds, from `Retry-After`, delta-seconds form only; 0 if absent) via `DoWithRetryAfter`.

## Config

| Field | Default | Meaning |
|-------|---------|---------|
| `Token` (root: `AniListToken`) | empty | optional; sent as `Authorization: Bearer <token>` when non-empty |
| `DescriptionAsHTML` (root: `AniListDescriptionAsHTML`) | nil = true | `*bool`; sets `asHtml` on Media and Character descriptions; `false` returns AniList markdown |
| `Timeout` | 10s | per-request |
| `Metrics` | nil | `func(source, op, dur, err)` |

Fixed: `Rate: 1.5` req/s, burst 1, 3 attempts, backoff `attempt * 100ms`, retries on transport error or 5xx.

## Code

- `provider/anilist/`: `config.go`, `client.go` (media, `post`, `statusErr`), `characters.go`, `staff.go`, `tags.go`, `stubs.go`, `mapping.go`, `models.go`, `constants.go`

## Test

`provider/anilist/`: `client_test.go`, `description_test.go`, `query_test.go`, `serialization_test.go` (`httptest` + `NewWithTestServer`). Run `go test ./provider/anilist/`.

## Gotchas

- AniList ids and MAL ids are different namespaces. Passing a MAL id to `Get`/`GetCharacters`/`GetStaff`/`GetTagsAndLinks` resolves to an unrelated media. Use the `*ByMalID` variant.
- MAL anime and manga ids also collide. `GetByMalID` is anime only, `GetMangaByMalID` manga only; for characters/staff/tags pass `mediaType` or the `idMal` lookup is unfiltered.
- `parseID` ignores Atoi errors: a non-numeric id becomes 0.
- 403 is `ErrRateLimited`, so a bad token (401) and a block (403) are told apart by sentinel; back off on 403.
- Descriptions default to HTML, not markdown.
- GraphQL-level `errors` in a 200 body are not inspected; a null `Media` surfaces as `ErrNotFound`.
