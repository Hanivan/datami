# AniList GraphQL API — AI Implementation Reference

> Derived from the supplied AniList API Docs export.
> Use this as implementation context for an AI coding agent.
> **Do not invent undocumented fields, arguments, enum values, limits, or behavior.**
> The original supplied documentation remains the authority for details not captured here.

## API Identity

AniList provides a public GraphQL API for anime, manga, tracking, and social data.

**GraphQL endpoint**

`https://graphql.anilist.co`

All API requests are **HTTP POST** requests.

Request body:

```json
{
  "query": "...",
  "variables": {}
}
```

`variables` is optional, but the docs recommend using variables wherever appropriate.

Required headers:

```http
Content-Type: application/json
Accept: application/json
```

Authenticated requests additionally use:

```http
Authorization: Bearer <ACCESS_TOKEN>
```

## Core Mental Model

AniList is **GraphQL, not REST**.

A client should:

1. Choose a top-level GraphQL query (`Media`, `Page`, `Viewer`, `MediaList`, etc.).
2. Pass typed arguments.
3. Explicitly select the fields it needs.
4. Use `Page` for paginated results.
5. Use `edges`, `nodes`, and `pageInfo` for connections.
6. Use mutations for writes.
7. Always inspect the GraphQL `errors` array.

## Media

`Media` represents both **anime and manga**.

Use:

```graphql
type: ANIME
```

or:

```graphql
type: MANGA
```

The docs state that IDs are not shared between the two media types: an ID used by an anime entry will not also identify a manga entry.

### Critical identity rule

Do **not** identify media by title alone.

AniList explicitly states that titles are not unique. There is no guaranteed "get anime by name" operation; search is the correct approach.

## Basic Media Query

```graphql
query ($id: Int) {
  Media(id: $id) {
    id
    title {
      romaji
      english
      native
    }
  }
}
```

If the ID does not match a media entry, the docs state that the API returns HTTP 404.

If a `type` is supplied and the ID belongs to the other media type, the docs also state that it returns 404.

## Search Anime

Use `Page` because search returns multiple results:

```graphql
query ($search: String!) {
  Page {
    media(search: $search, type: ANIME) {
      id
      title {
        romaji
        english
        native
      }
    }
  }
}
```

The `search` variable is required in this example.

AI implementation rule:

> Search by title, then resolve/match the correct entity using additional fields when available. Never assume the first title match is guaranteed to be correct.

## Pagination

### Page

Top-level queries such as `Media`, `Character`, and `Staff` normally return one object.

To retrieve multiple objects, wrap the query in `Page`.

Example:

```graphql
query ($page: Int, $perPage: Int, $search: String) {
  Page(page: $page, perPage: $perPage) {
    pageInfo {
      currentPage
      hasNextPage
      perPage
    }
    media(search: $search) {
      id
      title {
        romaji
      }
    }
  }
}
```

### One paginated data field per Page

A `Page` query may contain only **one** paginated data field.

Valid:

```graphql
Page {
  pageInfo {
    hasNextPage
  }
  media {
    id
  }
}
```

Invalid:

```graphql
Page {
  media {
    id
  }
  characters {
    id
  }
}
```

`pageInfo` is exempt from this restriction.

### PageInfo warning

The supplied docs explicitly state:

- `total` is currently inaccurate.
- `lastPage` is currently inaccurate.
- **Only rely on `hasNextPage` for pagination logic.**

Recommended:

```text
page = 1

loop:
    request Page(page)
    process data

    if pageInfo.hasNextPage == false:
        stop

    page += 1
```

Do not terminate based on `total` or `lastPage`.

### Optional variables

A variable can be declared but omitted from the variables object; the server can ignore it.

For stronger type safety:

```graphql
query ($id: Int!) { ... }
```

The `!` means the variable is required.

## Connections

Connections model one-to-many and many-to-many relationships.

They expose:

- `edges`
- `nodes`
- `pageInfo`

### edges

Use `edges` when intermediate relationship data matters.

Example:

```graphql
Media {
  characters(page: 1) {
    edges {
      role
      voiceActors {
        id
        name {
          first
          last
        }
      }
    }
  }
}
```

### nodes

Use `nodes` when only the related entities are needed.

Rule:

```text
Need relationship metadata -> edges
Need only related objects -> nodes
```

## Media Relations

Relations are a connection, so use `edges` when the relation type is needed:

```graphql
query ($id: Int!) {
  Media(id: $id) {
    id
    title {
      romaji
    }
    relations {
      edges {
        relationType
        node {
          id
          title {
            romaji
          }
        }
      }
    }
  }
}
```

## FuzzyDateInt

AniList's `FuzzyDateInt` format is:

`YYYYMMDD`

Example:

```text
20200000 = year 2020 lower bound
20210000 = year 2021 upper bound
```

Example from the docs: query light novels published in 2020:

```graphql
query ($page: Int = 1) {
  Page(page: $page) {
    pageInfo {
      hasNextPage
    }
    media(
      type: MANGA
      format: NOVEL
      startDate_greater: 20200000
      startDate_lesser: 20210000
    ) {
      id
      title {
        romaji
      }
    }
  }
}
```

## Advanced Media Browse

The supplied docs demonstrate these browse variables:

```text
page: Int = 1
id: Int
type: MediaType
isAdult: Boolean = false
search: String
format: [MediaFormat]
status: MediaStatus
countryOfOrigin: CountryCode
source: MediaSource
season: MediaSeason
seasonYear: Int
year: String
onList: Boolean
yearLesser: FuzzyDateInt
yearGreater: FuzzyDateInt
episodeLesser: Int
episodeGreater: Int
durationLesser: Int
durationGreater: Int
chapterLesser: Int
chapterGreater: Int
volumeLesser: Int
volumeGreater: Int
licensedBy: [Int]
isLicensed: Boolean
genres: [String]
excludedGenres: [String]
tags: [String]
excludedTags: [String]
minimumTagRank: Int
sort: [MediaSort] = [POPULARITY_DESC, SCORE_DESC]
```

Their documented argument mapping is:

| Variable | `media` argument |
|---|---|
| `id` | `id` |
| `type` | `type` |
| `season` | `season` |
| `format` | `format_in` |
| `status` | `status` |
| `countryOfOrigin` | `countryOfOrigin` |
| `source` | `source` |
| `search` | `search` |
| `onList` | `onList` |
| `seasonYear` | `seasonYear` |
| `year` | `startDate_like` |
| `yearLesser` | `startDate_lesser` |
| `yearGreater` | `startDate_greater` |
| `episodeLesser` | `episodes_lesser` |
| `episodeGreater` | `episodes_greater` |
| `durationLesser` | `duration_lesser` |
| `durationGreater` | `duration_greater` |
| `chapterLesser` | `chapters_lesser` |
| `chapterGreater` | `chapters_greater` |
| `volumeLesser` | `volumes_lesser` |
| `volumeGreater` | `volumes_greater` |
| `licensedBy` | `licensedById_in` |
| `isLicensed` | `isLicensed` |
| `genres` | `genre_in` |
| `excludedGenres` | `genre_not_in` |
| `tags` | `tag_in` |
| `excludedTags` | `tag_not_in` |
| `minimumTagRank` | `minimumTagRank` |
| `sort` | `sort` |
| `isAdult` | `isAdult` |

The docs use:

```graphql
Page(page: $page, perPage: 20)
```

and rely on `pageInfo.hasNextPage`.

## Media Lists

There are two primary ways to retrieve a user's list:

1. `MediaList` — paginated/specific list entries.
2. `MediaListCollection` — full list grouped by status/custom lists.

### User is NOT inferred

Even with an authenticated request, AniList does not infer the target user for media list queries.

You must specify the user.

### Single MediaList entry

A list entry can be retrieved by:

- list entry ID
- media ID + user ID/user name

Example:

```graphql
query ($id: Int, $mediaId: Int, $userId: Int) {
  MediaList(id: $id, mediaId: $mediaId, userId: $userId) {
    id
    media {
      id
      title {
        romaji
      }
    }
  }
}
```

Important:

> `MediaList.id` is the **list-entry ID**, not the media ID.

If looking up by media ID, pair it with `userId` or `userName`.

### Media.mediaListEntry

For an authenticated user:

```graphql
query ($id: Int!) {
  Media(id: $id) {
    id
    title {
      romaji
    }
    mediaListEntry {
      id
    }
  }
}
```

`mediaListEntry` represents the authenticated user's list entry.

When unauthenticated, the docs state this field returns `null`.

### Full MediaListCollection

```graphql
query ($type: MediaType!, $userId: Int!) {
  MediaListCollection(type: $type, userId: $userId) {
    lists {
      name
      entries {
        id
        media {
          id
          title {
            romaji
          }
        }
      }
    }
  }
}
```

`type` is required.

Provide either `userId` or `userName`.

### Custom lists matter

Do not only process the default status lists.

The docs explicitly warn that users can hide entries from default status lists while retaining them in custom lists. Skipping custom lists can therefore cause missing entries.

### 11,000-entry limitation

The supplied docs state that `MediaListCollection` is currently limited to the **11,000 most recently updated unique entries**.

Therefore:

> Do not claim that `MediaListCollection` is unconditionally complete for every user.

## Viewer / Current User

`Viewer` infers the current authenticated user from the access token:

```graphql
query {
  Viewer {
    id
    name
  }
}
```

If only the user ID is needed, the docs state that it can also be obtained from the JWT `sub` field.

## Authentication

AniList uses OAuth2.

Supported flows:

- Authorization Code Grant
- Implicit Grant

### Important OAuth behavior

The supplied docs state:

- OAuth scopes are **not supported**.
- Access tokens provide almost full access to a user's data.
- Tokens remain valid for **1 year** from issuance.
- Refresh tokens are **not currently supported**.
- After expiration, the user must authenticate again.

### Public/no-auth operations

Authentication is not required for:

- anime and manga data
- character searches
- public and unlisted user data

### Auth-required operations

Authentication is required for:

- modifying user lists
- private user data (only the current authenticated user)
- user-specific fields on other objects, such as `Media.mediaListEntry`

## Authorization Code Grant

Use when credentials/secret can be securely stored, typically server-side.

Authorization:

`https://anilist.co/api/v2/oauth/authorize`

Parameters:

```text
client_id
redirect_uri
response_type=code
```

The redirect URI must exactly match the configured application redirect URI.

Exchange the authorization code:

`POST https://anilist.co/api/v2/oauth/token`

Body:

```json
{
  "grant_type": "authorization_code",
  "client_id": "{client_id}",
  "client_secret": "{client_secret}",
  "redirect_uri": "{redirect_uri}",
  "code": "{code}"
}
```

The response contains an `access_token` JWT.

## Implicit Grant

Use when client credentials/secrets cannot safely be stored, typically:

- websites
- mobile apps

Authorization:

`https://anilist.co/api/v2/oauth/authorize`

Parameters:

```text
client_id
response_type=token
```

The access token is returned in the redirect URL fragment.

## Rate Limiting

The supplied docs state:

- Normal limit: **90 requests/minute**.
- The documentation currently warns that the API is degraded and temporarily limited to **30 requests/minute**.
- Response headers:
  - `X-RateLimit-Limit`
  - `X-RateLimit-Remaining`
- Exceeding the limit causes a **1-minute timeout**.
- During timeout, responses may include:
  - `Retry-After`
  - `X-RateLimit-Reset`
- There is also a burst limiter.

429 example:

```json
{
  "data": null,
  "errors": [
    {
      "message": "Too Many Requests.",
      "status": 429
    }
  ]
}
```

### Client behavior

Implement:

```text
read rate-limit headers
        |
        v
avoid aggressive bursts
        |
        v
on 429 -> respect Retry-After
        |
        v
retry with backoff
```

Do not blindly assume 90 RPM is always available because the supplied docs explicitly describe a temporary 30 RPM degraded state.

## GraphQL Errors

Errors are returned inside the GraphQL response's `errors` field.

Errors may occur because:

- query is invalid
- argument is missing
- field is unavailable
- server cannot fulfill the request
- mutation validation fails

**HTTP 200 can still contain GraphQL errors.**

Example:

```json
{
  "data": null,
  "errors": [
    {
      "message": "Cannot query field \"nonexistentField\" on type \"MediaList\".",
      "status": 400,
      "locations": [
        {
          "line": 4,
          "column": 5
        }
      ]
    }
  ]
}
```

Mutation validation errors may include:

```json
{
  "validation": {
    "id": ["The selected id is invalid."],
    "score": ["The score may not be greater than 100."]
  }
}
```

### Error handling rule

Always inspect:

```text
HTTP status
GraphQL errors[]
errors[].status
errors[].locations
errors[].validation
```

Do not use HTTP status alone to decide whether a GraphQL operation succeeded.

## API Stability / Blocking

The docs do not guarantee continuous API availability.

During severe outages AniList may:

- lower rate limits
- temporarily suspend API access

API-wide suspension may return HTTP 403 with a GraphQL error.

The docs also describe rare manual IP blocking caused by excessive requests from a single IP.

Implementation should therefore:

- distinguish service suspension from ordinary errors
- use backoff
- avoid request bursts
- avoid hammering a failing API

## Adult Content

AniList entries may contain adult content.

The supplied docs state:

- filtering is available where possible
- filtering is not guaranteed to be accurate
- AniList's definition may differ from platform/app-store standards
- user activity/forum data may also contain adult content
- clients should perform additional filtering where required

Important:

> "Ecchi" shows are not considered adult content by AniList.

Do not assume AniList filtering alone satisfies a third-party platform's content policy.

## Recommended Crawler / Ingestion Architecture

For a multi-source anime/manga aggregation system:

```text
AniList GraphQL
      |
      v
Typed GraphQL Client / Query Builder
      |
      v
Rate Limiter + Retry / Backoff
      |
      v
Raw Payload Storage
      |
      v
Normalizer / Entity Resolver
      |
      v
Canonical Anime / Manga Entities
```

### Source identity

Preserve:

```text
source = "anilist"
source_id = Media.id
media_type = ANIME | MANGA
```

Never use title as the canonical source identity.

### Raw payloads

For a multi-source ingestion pipeline, retain the raw AniList response when practical.

Reason:

- GraphQL lets the client choose fields.
- Different ingestion jobs may request different fields.
- Future normalization may need data that was not originally modeled.

### Query efficiency

Request only required fields.

Good:

```graphql
Media(id: $id, type: ANIME) {
  id
  title {
    romaji
    english
    native
  }
}
```

Avoid requesting an enormous nested graph unless the job actually needs it.

## AI Coding Rules

When generating AniList integration code:

1. Use `POST https://graphql.anilist.co`.
2. Do not invent REST endpoints.
3. Prefer GraphQL variables over string interpolation.
4. Explicitly select required fields.
5. Use `Page` for paginated lists.
6. Use only one paginated data field inside each `Page`.
7. Stop pagination using `pageInfo.hasNextPage`.
8. Never rely on `PageInfo.total` or `lastPage`.
9. Use `edges` when relationship metadata is needed.
10. Use `nodes` when only related objects are needed.
11. Always inspect `errors`, even for HTTP 200.
12. Respect `Retry-After` on 429.
13. Implement rate limiting and burst protection.
14. Do not assume 90 RPM is always currently available.
15. Do not assume titles are unique.
16. Use `MediaType` to distinguish anime and manga.
17. Treat `MediaList.id` as a list-entry ID.
18. Explicitly specify the target user for user-list queries.
19. Include custom lists when processing `MediaListCollection`.
20. Account for the documented 11,000-entry `MediaListCollection` limitation.
21. Public media data does not require authentication.
22. Mutations/private/current-user-specific data require authentication.
23. AniList OAuth does not support scopes.
24. Access tokens currently last one year and there is no refresh-token flow according to the supplied docs.
25. Never expose a client secret in an environment that cannot securely store it.
26. Preserve AniList source IDs and raw payloads for multi-source ingestion.

## Minimal Request

```http
POST https://graphql.anilist.co
Content-Type: application/json
Accept: application/json
```

```json
{
  "query": "query ($id: Int!) { Media(id: $id, type: ANIME) { id title { romaji english native } } }",
  "variables": {
    "id": 1
  }
}
```

## Minimal Paginated Request

```graphql
query ($page: Int = 1, $search: String) {
  Page(page: $page, perPage: 20) {
    pageInfo {
      currentPage
      hasNextPage
      perPage
    }
    media(
      type: ANIME
      search: $search
    ) {
      id
      title {
        romaji
        english
        native
      }
    }
  }
}
```

## Source Boundary

This reference is based on the supplied AniList API Docs export.

It intentionally does not silently add schema information from third-party clients or model memory. For complete schema/type/enum/field coverage, use the actual AniList GraphQL schema/reference as the authoritative source.
