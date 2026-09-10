# AniList GraphQL API — alternate source (verified)

Verified live 2026-08-02. AniList fills the gaps MAL client-id leaves: **voice actors, anime staff, streaming links**.

- Endpoint: `POST https://graphql.anilist.co` (note `.co`, not `.io` — `.io` does not resolve)
- Auth: **none** (free). **Rate limit corrected 2026-09-05:** response headers report `x-ratelimit-limit: 30` (not 90) — verify against `x-ratelimit-remaining` / `x-ratelimit-reset`, which AniList exposes via CORS. `ANILIST_REQUEST_DELAY_MS` should respect 30/min, not 90.
- **Introspection is enabled** — `__schema` / `__type` answer authoritatively what exists. Never brute-force field names here (unlike MAL, see `mal-api-endpoints.md`); one introspection query returns the whole type.
- Join key: **only AniList `Media` carries `idMal`** (the MAL id) → reconcile Anime by `Media.idMal`. AniList `Character` (14 fields) and `Staff` (25 fields) do **NOT** carry `idMal` (verified via `__type` introspection 2026-08-02) → reconcile characters by **name** (case-insensitive `firstName`+`lastName`, then `alternative` spelling), not by idMal.
- Response: GraphQL `{ data, errors }`. Query exactly the fields you need; no `fields=` param.

Ref anime: Youjo Senki II (`idMal: 49233` → AniList `id: 135865`). Naruto (`idMal: 20`) for the VA check.

## Verified query shape

```graphql
query ($idMal: Int) {
  Media (idMal: $idMal, type: ANIME) {
    id idMal title { romaji english native }
    episodes duration
    staff(perPage: 20, sort: RELEVANCE) {
      edges { role }                            # "Director", "Original Story", "Character Design", ...
      nodes { id name { full native } primaryOccupations }
    }
    characters(perPage: 50, sort: ROLE) {
      edges {
        role                                    # MAIN | SUPPORTING
        node { id name { full native alternative } }
        voiceActors { id name { full native } languageV2 }   # multi-language VAs
      }
    }
    externalLinks { id url site siteId type language color icon notes isDisabled }  # STREAMING / INFO / SOCIAL
    streamingEpisodes { title url site thumbnail }           # per-episode stream URLs (sparse)
    relations {
      edges { relationType(version: 2) }                     # SOURCE, PREQUEL, SEQUEL, SIDE_STORY, ALTERNATIVE, ...
      nodes { id idMal type title { english } }              # type: ANIME | MANGA — cross-medium
    }
  }
}
```

## What it unblocks (vs MAL client-id)

| Data | MAL client-id | AniList | Sample value |
|---|---|---|---|
| **Voice actors** (per character, per language) | ❌ edge removed | ✅ `characters.edges[].voiceActors[]` | Naruto→Sasuke→Noriaki Sugiyama (Japanese), Yuri Lowenthal (English), +10 langs |
| **Anime staff** (director/composer/designer) | ❌ 404 | ✅ `staff.edges[].role` + `nodes[]` | "Director", "Original Story", "Character Design" + `primaryOccupations` |
| **Streaming links** (where to watch) | ❌ none | ✅ `externalLinks` where `type=STREAMING` | Crunchyroll, YouTube playlist URLs (series-level) |
| Per-episode stream URLs | ❌ none | 🟡 `streamingEpisodes` (sparse, often empty) | — |
| Franchise relations (cross-medium) | ✅ MAL has it | ✅ `relations` (ANIME+MANGA mixed, `relationType`) | SOURCE (manga), PREQUEL, SIDE_STORY |
| Characters (role) | ✅ MAL has it | ✅ `characters.edges[].role` (MAIN/SUPPORTING) | same semantics |

## Data caveats observed

- **Some titles return `voiceActors: null`** even when populated elsewhere (saw null on Death Note `idMal:1535` and Youjo Senki II; full data on Naruto). Cause unclear (stale cache / unlinked on AniList). **Not every title will have VA edges** — sync must treat VA as best-effort, nullable. Does not block the table; rows just absent where AniList lacks them.
- `externalLinks` `type` ∈ `STREAMING` | `INFO` | `SOCIAL`. Only `STREAMING` feeds the "where to watch" feature; `INFO`/`SOCIAL` are official-site/Twitter (could store separately or skip).
- **Undocumented `externalLinks` subfields (verified 2026-09-05, `idMal:49233`):** `siteId` (stable int per provider — 1=Official Site, 5=Crunchyroll, 13=YouTube, 17=Twitter; better join key than the free-text `site`), `icon` (AniList CDN png, `https://s4.anilist.co/file/anilistcdn/link/icon/<siteId>-<hash>.png`, null for Official Site), `notes` (**the real brand when `site` is generic** — the Ani-One Asia row comes back as `site:"YouTube", notes:"Ani-One Asia"`, so display name should be `notes ?? site`), `isDisabled` (bool — filter these out).
- **vs MAL:** MAL has no API path for its own links block at all (see `mal-api-endpoints.md` → "hidden fields"). AniList covers Official Site / Twitter / Crunchyroll / YouTube but **not** the Resources row (AniDB, ANN, Wikipedia, Syoboi) nor region-locked streamers (Bahamut Anime Crazy). ~60% coverage of the MAL UI block, but it is the only *official-API* source for any of it.
- `streamingEpisodes` was empty for the tested title — don't rely on it as the primary episode source. Use it as a bonus when present.
- `relationType(version: 2)` is the modern enum (`SOURCE`, `PREQUEL`, `SIDE_STORY`, `ALTERNATIVE`, `SEQUEL`, `SPIN_OFF`, `SUMMARY`, `OTHER`, `CHARACTER`, `CONTAINS`, `COMPILATION`, `PARENT`). Omit `version` for the legacy strings — always pass `2`.

## Schema impact (ties to `docs/todos/schema-split-decisions.md`)

AniList changes three "blocked" verdicts to "shippable via second source":

1. **Voice actors → shippable.** New `Person` table (nullable `malId` + `anilistId`) + `AnimeCharacterVa{animeId,characterId,personId,language}` join. `language` from `voiceActors[].languageV2`.
2. **Anime staff → shippable.** Reuse `Person` + `AnimeStaff{animeId,personId,role}` join (`role` from `staff.edges[].role`).
3. **Streaming sources → seedable.** `StreamingSource` rows directly from `externalLinks` where `type=STREAMING` (`provider`=site, `url`, `language`).

**`malId` nullability confirmed necessary:** AniList `Staff`/`Character` ids differ from MAL ids. Reconcile via `idMal` where AniList has it, but allow rows sourced purely from AniList (nullable `malId`, set `anilistId`). Matches the nullable-`malId` + `source` discriminator rule in `schema-split-decisions.md`.

Pillar 03 ("cast and crew, connected") becomes **shippable** on AniList — no longer hard-blocked.

## Sync sketch (if adopted)

- New `AniListClient` (libs/anilist-client), mirrors `MalClient` shape: one `POST graphql.anilist.co` method, GraphQL query const per operation, `getWithRetry`-style 429 handling (AniList sends HTTP 429; also `X-RateLimit-*` headers).
- Sync keyed off existing MAL-sourced rows: for each `Anime.malId`, `Media(idMal: $malId)` → fetch VA/staff/externalLinks/relations. One request per title covers everything (batchable field-wise).
- Page characters/staff with `perPage` (max 25 default; bump with pagination via `Page` for large casts).
- Best-effort VA: log null/missing edges, continue.
- Reconciliation: upsert `Person` by `anilistId` (or `idMal` when present); link via join tables.

## Manga media (`type: MANGA`) — probed live 2026-08-09

Same `Media(idMal:, type: MANGA)` shape. What manga can/can't pull beyond MAL list+detail:

| Data | Availability | Sample |
|---|---|---|
| **Characters** | ✅ `characters(perPage: 50, sort: ROLE)` — role MAIN/SUPPORTING, name+alt+native+image | Monster `idMal:1` → 25 chars; Super no Ura `idMal:148467` → 8 |
| **Staff** | ✅ `staff` — role ("Story & Art", "Assistant", "Supervisor", translators) + `primaryOccupations` | Monster → Naoki Urasawa/Mangaka + 9 more |
| **Relations** | ✅ cross-medium from manga root | Monster → ADAPTATION anime `idMal:19`, SIDE_STORY manga `idMal:10968` |
| **Voice actors** | ❌ **always `[]` on MANGA media** — VAs exist on ANIME media only. Manga VA is structurally impossible, not a data gap | tested Monster, One Piece, AoT, Berserk, 148467 — all empty |
| **Publisher** | ❌ `Media.publisher` doesn't exist (400, full `__type` introspection confirms no such field) — see `mal-api-endpoints.md` for the MAL side. Manga's publication venue is `serialization`/magazine, which is MAL-sourced | — |
| **Studios** | ❌ `studios.edges` always `[]` on MANGA media (Monster, One Piece tested) — animation studios are anime-only | — |

**Manga characters dedup cleanly into the shared `Character` table** — AniList character ids are global across media, not per-media. One Piece manga (`idMal:13`) and anime (`idMal:21`) both return character `id:724` (Usopp), `id:309` (Chopper). `upsertAniListCharacters` keyed on `anilistId` therefore reuses anime-sourced rows instead of duplicating.

**What a manga character sync looks like** (mirrors `fetchAniListCharacters`):
- `AniListClient.getMangaCharacters(malId)` — same query, `type: MANGA`, reuse `AniListCharacterEdge`.
- Reuse `upsertAniListCharacters` + `ensureAnimeCharacterRoles` — but the role join is `AnimeCharacter` (anime-scoped PK `(anime_id, character_id)`), so manga needs its own `MangaCharacter` junction (same `status` column convention). No VA leg — skip `Person`/VA entirely for manga.
- Optional staff leg: `MangaAuthor` already exists (MAL-sourced). AniList staff is richer (role names + occupations) but reconciles to `Person`/`MangaAuthor` differently — deferred unless manga-staff UI is wanted.

**Manga → anime adaptation edges: AniList is the source, MAL is unreliable** (probed 2026-08-09, ref manga `idMal:148467`):

The manga's TV adaptation ("Super no Ura de Yani Suu Futari", announced on MAL website under Related Entries → Adaptation (TV)) is **absent from the MAL API** — `/manga/148467?fields=related_anime` and `related_manga` both return `[]` (API lags/drops the website's relation). AniList has it:
```
Media(idMal: 148467, type: MANGA).relations.edges →
  { relationType: "ADAPTATION",
    node: { idMal: 62076, type: "ANIME",
            title: "Smoking Behind the Supermarket with You",
            format: "TV", status: "RELEASING" } }
```
Implication: manga↔anime franchise edges (`anime_manga_relation`) sourced only from MAL `related_anime` will silently miss adaptations when MAL's API is stale. Fix path (built): `AniListClient.getMangaRelations(malId)` — same `ANIME_RELATIONS_QUERY` const, `type: MANGA` — called from `MangaSyncService.upsertNode` → feed `upsertAnimeMangaRelations`. The anime worker already has this logic (`fetchAniListRelations`); the manga-side caller was added.

**LN routing in both relation fetchers (built):** AniList marks light novels/novels as `type: MANGA, format: NOVEL`. Both `MangaSyncService.fetchAniListRelations` and `AnimeSyncService.fetchAniListRelations` route `format === ANILIST_NOVEL_FORMAT` (`'NOVEL'`, shared const in `@libs/anilist-client`) targets to the light-novel table (`upsertMangaLightNovelRelations` / `upsertAnimeLightNovelRelations`), not the manga table. Still gated by `idMal` — AniList's sparse LN→MAL mapping means novel targets frequently carry `idMal: null` and get skipped (documented in `docs/notes/light-novel-sync.md`).

See:
- `docs/notes/mal-api-endpoints.md` — MAL client-id surface (what AniList complements)
- `docs/todos/roadmap-api-mapping.md` — pillar 03 was ❌, now 🟡/✅ via AniList
- `docs/todos/schema-split-decisions.md` — nullable `malId`, Person/VA/staff/streaming tables


---

## Schema introspection — full field inventory (2026-09-05)

Obtained via `__schema`/`__type`, not guesswork. **Zero deprecated fields on `Media`** — nothing we use is at risk.

### Root `Query` — 27 entry points

Used by us: `Media`. Unused and potentially useful:

| entry point | what it gives |
|---|---|
| `Page(page, perPage)` | The generic paginated wrapper — `Page { media, characters, staff, studios, ... }`. **Batch many works in one request** instead of one `Media` call per `idMal`. The single biggest available call-count saving. |
| `ExternalLinkSourceCollection(mediaType, type)` | Catalogue of every link provider AniList knows, independent of any title — `{id, site, siteId, type, language, icon, color}`. 37 STREAMING sources for ANIME, 1 INFO (`Official Site`). A ready-made seed table for a `streaming_source` entity. |
| `GenreCollection` | All 19 genre strings in one call — the canonical list for `genre_alias` seeding. |
| `MediaTagCollection` | **428** tags, each `{name, category, isAdult}`, e.g. `4-koma`/Technical, `Achronological Order`/Setting-Time. Far richer than the 19 genres. |
| `Studio(id, search)` / `Character` / `Staff` | Direct lookup + `search` by name — AniList has entity search where MAL 404s on `/people?q=`. |
| `Recommendation`, `Review`, `AiringSchedule`, `MediaTrend`, `SiteStatistics` | Unused; available if ever wanted. |
| `Markdown(markdown)` | Renders **AniList-flavoured** markdown to HTML. WARNING: does **not** parse MAL's `[i]`/`[url=]`/`[spoiler]` bbcode — tested, passes it through unchanged. Not a fix for MAL's `biography`/`more` blobs. |
| `MediaList`, `MediaListCollection`, `User`, `Viewer`, `Following`, `Activity`, `Thread`, `Notification`, `Like` | User/social surface — needs auth, out of scope. |

### `Media` — 55 fields; we currently request ~12

Full set: `airingSchedule autoCreateForumThread averageScore bannerImage chapters characters countryOfOrigin coverImage description duration endDate episodes externalLinks favourites format genres hashtag id idMal isAdult isFavourite isFavouriteBlocked isLicensed isLocked isRecommendationBlocked isReviewBlocked meanScore mediaListEntry modNotes nextAiringEpisode popularity rankings recommendations relations reviews season seasonInt seasonYear siteUrl source staff startDate stats status streamingEpisodes studios synonyms tags title trailer trending trends type updatedAt volumes`

Notable unused, verified live on `idMal:49233`:

| field | value / shape |
|---|---|
| `tags` | 23 entries `{name, category, rank 0-100, isGeneralSpoiler, isMediaSpoiler}` — e.g. Military/Theme-Other-Organisations/100, War/97. **Ranked and spoiler-flagged**, unlike genres. |
| `rankings` | `[{rank, type: RATED\|POPULAR, format, year, season, allTime, context}]` — 5 entries: "highest rated all time" #289, "most popular" 2026 #15, seasonal #7. Pre-computed rank slices MAL only gives as one flat number. |
| `stats` | `{scoreDistribution:[{score,amount}], statusDistribution:[{status,amount}]}` — the histogram behind the mean. |
| `trailer` | `{id, site:"youtube", thumbnail}` — pairs with MAL's undocumented `videos`. |
| `synonyms` | `["Saga of Tanya the Evil II","Youjo Senki 2"]` — extra name keys for `NameKeyUtil`. |
| `countryOfOrigin` | `"JP"` — distinguishes manhwa/manhua/donghua at the source. |
| `source(version: 3)` | `"LIGHT_NOVEL"` — **note the `version` arg**; the unversioned enum is coarser. |
| `hashtag` | `"#youjosenki"` |
| `isLicensed`, `isAdult`, `siteUrl`, `favourites`, `bannerImage`, `nextAiringEpisode{episode,airingAt,timeUntilAiring}` | self-explanatory; `bannerImage` is a wide art asset MAL has no equivalent for. |

### `MediaExternalLink` — complete (10 fields)
`id url site siteId type language color icon notes isDisabled`. Nothing beyond what is documented above — this type is fully mined.

### `Character` — 14 fields
`id name image description gender dateOfBirth age bloodType isFavourite isFavouriteBlocked siteUrl media favourites modNotes`

Unused: `gender`, `dateOfBirth` (`{year,month,day}` fuzzy date), `age`, `bloodType`, `favourites`, `siteUrl`, and **`media`** — the reverse character→works connection (AniList's answer to MAL's `animeography`). Still **no `idMal`**, so the name-key reconciliation in `identity-resolver` stays necessary.

### `Staff` — 25 fields, and it breaks MAL's hardest block
`id name languageV2 image description primaryOccupations gender dateOfBirth dateOfDeath age yearsActive homeTown bloodType isFavourite isFavouriteBlocked siteUrl staffMedia characters characterMedia staff submitter submissionStatus submissionNotes favourites modNotes`

**`staffMedia`, `characters` and `characterMedia` are a full reverse person->works index** — exactly what `mal-api-endpoints.md` documents as permanently blocked on MAL (`/people/{id}` has no `animeography`). Verified on `Staff(id:101686)` (Aoi Yuuki): `characters` -> Maomao, Mami Nanami...; `staffMedia` -> `[{staffRole, node{idMal,title}}]`, both paginated. Also live: `dateOfBirth {1992,3,27}`, `age 34`, `bloodType "A"`, `homeTown "Sanmu, Chiba Prefecture, Japan"`, `yearsActive [2003]`, `favourites 12541`, `dateOfDeath` (null when living).

### `Studio` — 7 fields
`id name isAnimationStudio media siteUrl isFavourite favourites`. **`isAnimationStudio`** separates real animation studios from producers/licensors — MAL's `studios` field conflates them. `media` is the reverse studio->works connection.

---

## WARNING: Pagination — our queries silently truncate, and `pageInfo.total` lies

`ANIME_CHARACTERS_QUERY` / `ANIME_STAFF_QUERY` (`libs/anilist-client/src/anilist-client.const.ts`) request `perPage: 50` with **no `page` arg and no `pageInfo`**, so any roster past 50 is dropped with no signal. The in-file comment acknowledges this ("large casts (>50) truncate VA coverage").

Every connection (`CharacterConnection`, `StaffConnection`, ...) exposes `pageInfo { total perPage currentPage lastPage hasNextPage }`. **But `total` and `lastPage` are not to be trusted.** Measured on `idMal:49233`:

| request | reported | actual |
|---|---|---|
| `characters(perPage:2)` | `total: 500`, `lastPage: 250` | — |
| `characters(perPage:25, page:1)` | `total: 500`, `lastPage: 20`, `hasNextPage: true` | 25 edges |
| `characters(perPage:25, page:2)` | `hasNextPage: false` | **17 edges** |
| `characters(perPage:25, page:3)` | `hasNextPage: false` | 0 edges |

Real roster is **42**. `total` reports a constant 500 and `lastPage` recomputes from that fiction (`500/25=20`, `500/2=250`). **Only `hasNextPage` is accurate.** Any pagination loop must terminate on `hasNextPage === false` and must never size itself from `total`/`lastPage`. Same caveat applies to `staff` (also claimed `total: 500`).

## `CharacterEdge.voiceActorRoles` — strictly better than `voiceActors`

`CharacterEdge` fields: `node id role name voiceActors voiceActorRoles media favouriteOrder`.

We use `voiceActors { ... }`. The sibling `voiceActorRoles` wraps the same person plus context:

```graphql
voiceActorRoles { voiceActor { id name{full} languageV2 } roleNotes dubGroup }
```

`dubGroup` names the dub production (e.g. distinguishing two different English dubs of one title) and `roleNotes` carries per-credit annotations. Both were `null` on the 49233 sample, but this is the field that can disambiguate the duplicate `(personId, language)` rows recorded in [Known Issues](../claude/known-issues.md) — worth checking on a title with multiple dubs before assuming dedup is the only fix.
