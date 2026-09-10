# MAL API v2 — usable endpoints (client-id auth)

Tested live 2026-08-02, `X-MAL-CLIENT-ID` header only (no OAuth).
Ref anime: **49233** (Youjo Senki II), ref manga: **1** (Monster), ref character: **137825**, ref person: **1870** (Miyazaki).

All paths under `https://api.myanimelist.net/v2`. `limit` capped 100. Paged via `offset` / `paging.next`.

## Anime

### `GET /anime/{id}` — detail
**Field is returned only if requested in `fields=`.** `id`+`title`+`main_picture` always present.

| field | detail? | list/seasonal/ranking? | type |
|---|---|---|---|
| `id`,`title`,`main_picture` | ✅ | ✅ | — |
| `alternative_titles`,`synopsis`,`mean`,`rank`,`popularity` | ✅ | ✅ | — |
| `num_list_users`,`num_scoring_users`,`nsfw`,`status`,`genres`,`num_episodes` | ✅ | ✅ | — |
| `start_season`,`broadcast`,`source`,`average_episode_duration`,`rating` | ✅ | ✅ | — |
| `start_date`,`end_date` | ✅ | ✅ | date (end null if unfinished) |
| `pictures`,`background`,`studios` | ✅ | ❌ dropped | — |
| `created_at`,`updated_at`,`media_type` | ✅ | ❌ dropped | — |
| `opening_themes`,`ending_themes` | ✅ | ❌ dropped | `[{id,anime_id,text}]` |
| `videos` **(undocumented)** | ✅ | ❌ dropped | `[{id,title,url,thumbnail,created_at,updated_at}]` — PVs/trailers, YouTube URLs, epoch-**ms** timestamps |
| `num_favorites` **(undocumented)** | ✅ | ✅ **survives lists** | int (49233 → `3825`) |
| `recommendations` | ✅ | ❌ dropped | `[{node:{id,title,main_picture}}]` |
| `related_anime`,`related_manga` | ✅ | ❌ dropped | `[{node:{...},relation_type,relation_type_formatted}]` |
| `airing`,`external`,`relations`,`my_list_status` | ❌ dropped | ❌ | (use `status`; my_list_status needs OAuth) |

**List-endpoint gotcha** (per CLAUDE.md): `/anime/season/...`, `/anime/ranking`, `/anime?q=` silently drop `pictures`,`background`,`opening_themes`,`ending_themes`,`recommendations`,`related_*`,`created_at`,`updated_at`,`media_type`. Detail-only → per-item `GET /anime/{id}` backfill.

### `GET /anime?q={q}&limit=&offset=&fields=` — search
Works. `q` required (bare `/anime` → 400 bad_request).

### `GET /anime/season/{year}/{season}?limit=&offset=&fields=` — seasonal
`season` ∈ `winter`|`spring`|`summer`|`fall`. Exhausts `paging.next`. (Used by `AnimeSyncService`.)

### `GET /anime/ranking?ranking_type={type}&limit=&offset=&fields=` — ranking
`ranking_type` is **query param, not path** (`/anime/ranking/all` → 404). Types:
`all` | `bypopularity` | `airing` | `upcoming` | `tv` | `ova` | `movie` | `special` | `favorite`.
Returns `[{node, ranking:{rank}}]`. Bare `/anime/ranking` (no type) = default ranking.

### `GET /anime/{id}/characters` — character roster
`[{node:{...}, role}]`, `role` ∈ `Main`|`Supporting`. Paged.

**⚠️ CORRECTED 2026-09-05 — `fields` IS honored here.** The earlier note ("`fields` ignored, node stays `{id}`") is wrong as of this re-test. Accepted on the node: `first_name`, `last_name`, `alternative_name`, `main_picture`, `num_favorites`, `biography`. This **removes the N+1** — the whole roster with names, pictures and biographies comes back in one paged call, no per-character `GET /characters/{id}` needed.

`animeography` and every VA-shaped key (`voice_actors`, `va`, `cast`, `seiyuu`, `actors`, `people`) are **silently dropped here** — VA linkage stays blocked (see below).

### `GET /anime/{id}/episodes` — episode list
`403 forbidden` on client-id. Exists but blocked.

### Dead anime sub-paths (404)
`staff`, `characters_staff`, `pictures`, `recommendations` (as sub-path), `related_anime`, `related_manga`, `reviews`, `news`, `forum`, `external`, `statistics`, `streaming`, `ranking`, `themes`, `producers`, `videos`, `clubs`, `userrecs`, `my_list_status` (405). Use the detail `fields=` form instead. Re-swept 2026-09-05 — all still 404, `episodes` still 403.

### Dead top-level endpoints (404, re-swept 2026-09-05)
`/genres`, `/studios`, `/producers`, `/seasons`, `/magazines`, `/companies`, `/characters?q=`, `/people?q=`. No browse-by or search for any non-anime/manga entity.

## Manga

### `GET /manga/{id}` — detail
Same field-on-request model. Audited on Monster (id 1):

| field | detail? |
|---|---|
| `id`,`title`,`main_picture`,`alternative_titles`,`synopsis`,`mean`,`rank`,`popularity` | ✅ |
| `num_list_users`,`num_scoring_users`,`nsfw`,`status`,`genres` | ✅ |
| `num_chapters`,`num_volumes`,`start_date`,`end_date`,`media_type` | ✅ |
| `authors`,`pictures`,`background`,`created_at`,`updated_at` | ✅ |
| `recommendations`,`related_manga`,`related_anime` | ✅ |
| `num_favorites` **(undocumented)** | ✅ (also survives `/manga/ranking`) |
| `serializations` (plural) | ❌ dropped |
| `videos`,`opening_themes`,`ending_themes` | ❌ dropped (anime-only) |
| `my_list_status` | ❌ dropped (OAuth) |

`authors` shape: `[{node:{id}, role:"Story & Art"}]` — id only *by default*, but **subfields work** (see "Manga `authors` inline names" below) and now reach further than documented: `authors{first_name,last_name,main_picture,birthday,num_favorites,more}` returns the full person inline, killing the per-author `/people/{id}` round-trip.

### `GET /manga?q={q}` — search ✅
### `GET /manga/ranking?ranking_type={type}` — ranking ✅
Types: `manga` | `novels` | `oneshots` | `doujin` | `manhwa` | `manhua` | `bypopularity` | `favorite`. (`/manga/ranking/{type}` path also 404 — use query param.)

## Character

### `GET /characters/{id}` (plural)
`{id}` by default. Fields on request:
- ✅ `first_name`,`last_name`,`alternative_name`,`main_picture`
- ✅ **`num_favorites`** (undocumented) — int
- ✅ **`biography`** (undocumented) — HTML-ish string, `<br />`, `&#039;` entities, and `[spoiler]...[/spoiler]` bbcode. Needs sanitizing before display.
- ✅ **`animeography`** (undocumented) — `[{node:{id,title,main_picture}, role}]`. **Reverse character→anime link**, the only reverse-index MAL exposes on client-id.
- ❌ `name` (does not exist — use first+last)
- ❌ dropped: `voice_actors`,`va`,`cast`,`seiyuu`,`actors`,`people`,`anime`,`manga`,`mangaography`

### `GET /character/{id}` (singular) → 404

## Person (authors, staff, VAs)

### `GET /people/{id}`
`{id}` by default. Fields on request:
- ✅ `first_name`,`last_name`,`main_picture`
- ✅ **`birthday`** (undocumented) — `"1960-01-02"`
- ✅ **`num_favorites`** (undocumented) — int
- ✅ **`more`** (undocumented) — biography blob. Carries `[url=...]...[/url]` / `[i]...[/i]` bbcode plus `<br />` and HTML entities; often holds Twitter/YouTube links and "Date of Death:" for the deceased.
- ❌ dropped: `voice_acting_roles`,`animeography`,`mangaography`,`anime`,`manga`,`roles`,`credits`,`works`,`alternative_name`,`website`
- ⚠️ **no reverse index** — unlike `/characters/{id}`, people have no `animeography`. Person→works is one-way only.

**Unreachable from anime/manga for VAs** — no anime→person link exposed (no `/anime/{id}/staff`). Authors reachable only via `/manga/{id}?fields=authors` → `{node:{id}, role}` → `/people/{id}`.

## Manga characters / VAs — probed live 2026-08-09, still blocked

Confirmed **no manga character/VA path exists** on client-id:
- `/manga/{id}/characters` → **404** (`{"error":"not_found"}`).
- `characters` / `voice_actors` fields on `/manga/{id}?fields=` → **200 but silently dropped** (like the ranking/list drops — MAL just has no manga→character model exposed).

Only escape hatch: **AniList `type: MANGA`** (`characters` + `staff` available, VAs still `[]` — VAs are anime-media-only). See `docs/notes/anilist-api.md` → "Manga media" for the full probe.

## Manga publisher / studios — don't exist (probed live 2026-08-09)

Manga has **no publisher or studio** on either source:

- **MAL** `/manga/{id}?fields=publisher,publishers` → **200 but silently dropped** (same drop behavior as ranking/list). No `studios` field on manga detail at all (anime-only).
- **AniList** `Media(type: MANGA).publisher` → **400** `Cannot query field "publisher" on type "Media"` — field doesn't exist (full `__type` field introspection of `Media` shows no publisher anywhere). `studios` field exists but returns `edges: []` always on MANGA media (Monster, One Piece tested — animation studios are anime-only).

**Manga's publication-venue field is `serialization` (magazine)** — the closest thing to "publisher", but it names the magazine (e.g. Monster → "Big Comic Original"), not the publishing company that owns it (Shogakukan). Neither API exposes the magazine→publisher link, so a publisher entity is **not implementable** from either source.

**`serialization` IS implemented in the codebase** (verified):
- Interface: `MalMangaNode.serialization?: MalMagazine[]` (`libs/mal-client/src/interfaces/mal-manga-node.interface.ts`)
- Worker: `serialization` in `DETAIL_FIELDS` (`manga-sync.service.ts`), pulled only via `/manga/{id}` detail backfill (ranking endpoint silently drops it), mapped in `upsertNode` → `upsertMagazines`
- Repository: `upsertMagazines` (`MangaRepository`), `Magazine` entity + `manga_magazine` join table, eager-loaded via `findByMalId`/`findOne`
- Modules: `Magazine` registered in `worker manga-sync.module.ts` + `mono-api manga.module.ts`

## Manga `related_anime` / `related_manga` — can be stale (probed live 2026-08-09)

`/manga/{id}?fields=related_anime` works, but **isn't always populated even when the MAL website shows the relation**. Ref manga `idMal:148467` (Super no Ura de Yani Suu Futari): MAL website lists "Adaptation (TV)" under Related Entries, but both `related_anime` and `related_manga` return `[]` from the API. The same anime is present on **AniList** as `ADAPTATION` (`idMal:62076`, format TV). So:

- MAL relation edges are best-effort — missing edges are a MAL data gap, not a code bug.
- For complete manga↔anime adaptation edges, fall back to **AniList `relations`** (`type: MANGA`). See `docs/notes/anilist-api.md` → "Manga → anime adaptation edges".

## Manga `authors` inline names — works with subfields

`/manga/{id}?fields=authors` → `[{node:{id}, role}]` (id-only, per note above). But adding subfields works:
`?fields=authors{first_name,last_name}` → `[{"node":{"id":1867,"first_name":"Naoki","last_name":"Urasawa"},"role":"Story & Art"}]`.
That's why `MangaSyncService` syncs author names in one call — no per-author `/people/{id}` roundtrip needed.

## Studio / genre / producer

No standalone list/by-id endpoint — all probed (`/genres/anime`, `/producer/{id}`, `/producers/{id}`, `/studio/{id}`, `/anime/genres/{id}`) → 404.
**Studios/genres come inline only** via `/anime/{id}?fields=studios,genres` → `[{id,name}]`. No way to list all studios or browse-by-studio.

## User (OAuth-only)

`/users/@me/animelist` → `401 invalid_token`. All user/list-status endpoints need OAuth2 token, not client-id.

---

## Summary — what we can build

| Want | Source | Calls |
|---|---|---|
| Anime list + slim fields | `/anime/season/{y}/{s}` | 1/page |
| Anime full entity | `/anime/{id}?fields=...` | 1 |
| Anime character roster **+ names/pics/bios** | `/anime/{id}/characters?fields=first_name,last_name,alternative_name,main_picture,num_favorites,biography` | 1/page (**no longer N+1**) |
| Character detail + reverse anime list | `/characters/{id}?fields=...,biography,animeography` | 1 per char |
| PVs / trailers | `/anime/{id}?fields=videos` | 1 |
| Favorites count | `?fields=num_favorites` (anime, manga, character, person) | free, rides any request |
| Franchise graph | `/anime/{id}?fields=related_anime,related_manga` | 1 per anime |
| Themes (OP/ED text) | `/anime/{id}?fields=opening_themes,ending_themes` | 1 |
| Recommendations | `/anime/{id}?fields=recommendations` | 1 |
| Manga authors **inline w/ person detail** | `/manga/{id}?fields=authors{first_name,last_name,main_picture,birthday,num_favorites,more}` | 1 (**no longer N+1**) |

**Blocked (no client-id path):** voice actors (anime↔VA link removed), anime staff, episode list (403), browse-by-genre/studio, user lists.

---

## Character / VA / Staff — deep-dive

### Characters — viable
**Now a one-call flow per anime** (re-tested 2026-09-05; the old two-call flow below is obsolete):

```
GET /anime/{malId}/characters?limit=100&fields=first_name,last_name,alternative_name,main_picture,num_favorites,biography
```

Returns `[{node:{id,first_name,last_name,alternative_name,main_picture,num_favorites,biography}, role}]`, paged. Gets everything the old flow needed **without** the per-character `GET /characters/{id}`.

Still call `/characters/{id}` only when you want `animeography` (the reverse character→anime index) — it is dropped on the roster sub-path.

### Voice Actors (VA) — blocked
No API path. MAL removed public VA linkage for client-id auth:
- `/anime/{id}/characters` has no VA sub-field.
- `/characters/{id}` drops `voice_actors`.
- `/people/{id}` exists but unreachable from anime (no `/anime/{id}/staff`), and drops `voice_acting_roles`/`anime`/`manga`.

Not fixable on our side — MAL policy. Would need OAuth2 (user token) or scraping, both out of scope.

### Anime staff / authors — blocked
`/anime/{id}/staff` 404. `staff` field on detail dropped.
Manga still has `authors` via `/manga/{id}` detail (`[{node:{id}, role:"Story & Art"}]` → resolve via `/people/{id}`).

### Impl implications (if we proceed, characters-only)
- New catalog entities: `Character` + join `anime_character` (carries `role`).
- New MAL interfaces: `MalAnimeCharacterRef` (`{id, role}`), `MalCharacterNode`.
- `MalClient.getAnimeCharacters(malId)` + `getCharacterDetails(charId, fields)`.
- Worker: `AnimeSyncService.fetchCharacterBackfill` — mirror `MangaSyncService.fetchDetailBackfill` shape: per-anime char list, then per-char detail, upsert, link. Guard against 429 (N sequential calls/anime).
- Migration: `character` table + `anime_character` join table.
- Skip VA entirely. Do not add `VoiceActor` entity — no data source.


---

## Undocumented ("hidden") fields — reverse-engineered 2026-09-05

Ref anime **49233** (Youjo Senki II) with `X-MAL-CLIENT-ID`. None of these appear in the [official beta docs](https://myanimelist.net/apiconfig/references/api/v2), whose maximal detail example returns exactly 31 keys.

### Method — the response-key oracle
MAL never errors on a bad `fields=` value; it silently returns the `{id,title,main_picture}` stub, so an unknown field and a *dropped* field look identical from the status code. The oracle is the **response key set**:

- field is real and populated → its key appears in the JSON
- field is unknown, or real-but-dropped-on-this-endpoint → no key

`fields=zzz_not_a_field` is the negative control; `fields=status,num_episodes` the positive. Subfield syntax (`videos{id}`) is **not** validated — `studios{zzz}` returns `studios` intact, so subfields give no oracle.

**Rate limit:** ~60 rapid probes triggered a `307` redirect to `https://api.myanimelist.net/error.json` on *every* request, including previously-working ones. It cleared after ~12s. Pace brute-force probing at ≥1.1s/request. A 307-to-error.json means throttled, not "field rejected" — do not read it as a result.

### Confirmed hidden fields

| endpoint | field | shape / notes |
|---|---|---|
| `/anime/{id}` | `videos` | `[{id,title,url,thumbnail,created_at,updated_at}]` — PVs & trailers as `https://youtu.be/<v>` + `i.ytimg.com` thumb. Timestamps are **epoch-milliseconds**, unlike every other MAL date (ISO strings). Dropped on list endpoints; `/anime/{id}/videos` sub-path is 404. |
| `/anime/{id}` | `opening_themes`, `ending_themes` | `[{id,anime_id,text}]`, already noted above. |
| `/anime/{id}`, `/manga/{id}`, `/characters/{id}`, `/people/{id}` | `num_favorites` | int. **The only hidden field that survives list/ranking/seasonal endpoints** — costs nothing to add to the existing sync `fields=`. |
| `/anime/{id}/characters` | `first_name`,`last_name`,`alternative_name`,`main_picture`,`num_favorites`,`biography` | Roster nodes are fully hydratable — see the correction above. |
| `/characters/{id}` | `biography` | HTML-ish: `<br />`, `&#039;`-style entities, `[spoiler]...[/spoiler]` bbcode. Sanitize before render. |
| `/characters/{id}` | `animeography` | `[{node:{id,title,main_picture}, role}]`. Reverse character→anime index — no equivalent exists for people. |
| `/people/{id}` | `birthday` | `"1960-01-02"` |
| `/people/{id}` | `more` | Biography blob. `[url=...]`/`[i]...[/i]` bbcode + `<br />` + entities. Often the person's Twitter/YouTube, and `Date of Death:` when applicable. |
| `/manga/{id}` | `authors{...}` | Person subfields resolve inline: `first_name,last_name,main_picture,birthday,num_favorites,more`. |

### Confirmed NOT to exist (~110 names swept)
`external`, `links`, `external_links`, `websites`, `streaming`, `streaming_platforms`, `available_at`, `where_to_watch`, `platforms`, `providers`, `services`, `simulcast`, `official_site`, `site`, `url`, `urls`, `resources`, `related_links`, `web_links`, `watch`, `watch_links`, `homepage`, `social`, `twitter`, `youtube_url`, `anidb`, `ann`, `external_ids`, `ids`, `mappings`, `source_links`, `tv_station`, `broadcaster`, `station`, `channel`, `network`, `networks`, `licensors`, `producers`, `trailer`, `trailer_url`, `promo`, `promos`, `preview`, `clips`, `theme_songs`, `themes`, `music`, `songs`, `ost`, `soundtrack`, `insert_songs`, `episodes`, `characters`/`staff` (on anime detail), `voice_actors`/`va`/`cast`/`seiyuu`/`actors`/`people` (everywhere), `mangaography`, `num_members`, `num_reviews`, `num_recommendations`, `num_favorites`-siblings (`num_watching` etc — use `statistics`), `title_english`, `synonyms`, `approved`, `trivia`, `spoilers`, `mal_url`, `slug`, `permalink`, `interest_stacks`.

### The "Available At / Resources / Streaming Platforms" block — no API path exists
The block MAL renders on `myanimelist.net/anime/49233` (Official Site, @youjosenki, AniDB, ANN, Wikipedia ×2, Syoboi, Crunchyroll, Ani-One Asia, Bahamut Anime Crazy) is **unreachable from the API by any key**. Verified: all link-shaped names above return nothing, `/anime/{id}/external` and `/streaming` are 404, and the HTML page exposes **no** AJAX/JSON route — the links are server-rendered inline inside `<div class="js-links" data-rel="resource" style="display:none">` (the "Less links" toggle is pure CSS over already-present markup).

HTML shape, if scraping is ever considered:
```
div.external_links a[data-ga-click-type="external-links-anime-pc-<slug>"]
  slug ∈ official-site | twitter | anidb | ann | wikipedia | syoboi | crunchyroll | youtube | bahamut
caption text in div.caption; three <h2> sections: Available At / Resources / Streaming Platforms
```
Resource URLs carry stable provider ids — `anidb.net/...aid=16392`, `animenewsnetwork.com/...id=24470`, `cal.syoboi.jp/tid/7890` — i.e. they could become additional `external_ref` sources.

**AniList covers ~60% of that block via the official API** (`externalLinks`) — see `anilist-api.md`. Missing there: the entire Resources row (AniDB/ANN/Wikipedia/Syoboi) and region-locked streamers like Bahamut.
