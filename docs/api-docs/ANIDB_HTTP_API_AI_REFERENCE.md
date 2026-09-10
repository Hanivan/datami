# AniDB HTTP API — AI Implementation Reference

> Source: the supplied AniDB wiki exports (`anidb/1.md` — API overview, `anidb/2.md` — HTTP API Definition).
> This file is a normalized, AI-oriented reference derived from that source.
> **Do not invent undocumented endpoints, parameters, fields, enum values, or behavior.**
> Where the source is silent but a sample payload implies structure, this document marks it
> **`[INFERRED]`** and states what was observed. Treat `[INFERRED]` as a hypothesis to verify,
> never as contract.

## Purpose

Use this document as implementation context for an AI coding agent building an AniDB HTTP client,
catalog crawler, or ingestion service.

AniDB is **not** a modern REST/JSON API. It is a small set of **XML-over-HTTP** data pages, plus
daily bulk dumps. There is no OpenAPI schema, no HTTP status semantics, and no JSON by default.

## API Surface Overview

AniDB exposes four distinct surfaces. This document covers surfaces 1 and 2.

| # | Surface | Access | Covers |
|---|---------|--------|--------|
| 1 | **HTTP XML API** | Public, registered client | Non-file anime metadata, recommendations, per-user summaries |
| 2 | **Data dumps** | Public, unauthenticated | Full anime-title index |
| 3 | **UDP API** | Public, registered client, login required | Files, episodes, mylist, notifications — see `ANIDB_UDP_API_AI_REFERENCE.md` |
| 4 | **TCP API** | Private, on request only | Partial offline DB copies. Requires in-memory encryption; unencrypted data may never touch disk. Only known client: AniDB O'Matic (AOM) |

> The source explicitly asks clients **not to scrape AniDB HTML**. Excessive HTTP requests cause bans.

## API Identity

**Base URL**

`http://api.anidb.net:9001/httpapi`

All requests are **HTTP GET** with query parameters. There is no request body.

> **`[INFERRED]`** The source documents only `http://` on port 9001 — no HTTPS variant is mentioned.
> Do not assume TLS is available.

### Global required parameters

Every HTTP XML API request must carry all four:

| Parameter | Type | Required | Notes |
|-----------|------|----------|-------|
| `client` | string | yes | Registered client identifier, **lower-case** (convert if necessary). This is the registered HTTP-API client id, **not** your project name. Registration is separate from UDP-API client registration. |
| `clientver` | integer | yes | Registered client version number for that `client` string |
| `protover` | integer | yes | Protocol version. Documented possible value: `1` |
| `request` | string | yes | Which data page to fetch (see Endpoint Catalog) |

Canonical shape:

```
http://api.anidb.net:9001/httpapi?client={str}&clientver={int}&protover=1&request={str}
```

## Global API Rules

### Encoding and transfer

- All content is **UTF-8**.
- All content is **gzip compressed**. Your HTTP library may not decompress automatically — handle it.
- Transfer uses **HTTP/1.1 chunked encoding**, *even when the client requested uncompressed HTTP/1.0*.
  A naive `Content-Length`-based reader will break.

### Rate limiting and caching

Hard rules from the source:

- **Do not exceed one request per 2 seconds.**
- **Employ heavy local caching.** Requesting the same dataset multiple times in one day can get you banned.
- Request flooding gets you banned.
- The API must not be used to bulk-download AniDB (anti-leech). Detected attempts result in a ban.

A crawler must therefore persist every fetched payload and re-fetch on a schedule measured in days,
not minutes.

### Error handling — no HTTP status codes

**AniDB does not signal errors via HTTP status codes.** A failed request returns HTTP 200 with an
XML error document:

```xml
<error>Banned</error>
```

Implementation rule: after every request, parse the body and check whether the root element is
`<error>` **before** attempting to parse the expected root element. Never branch on
`response.status`.

> The source attributes this behavior to "ommina" and gives exactly one observed error string
> (`Banned`). The full set of `<error>` strings is **not documented**.

### Authentication

The HTTP XML API has **two separate auth models**:

1. **Client identification only** (`client` + `clientver` + `protover`) — for all anime/catalog pages.
   No user credentials.
2. **Username + plaintext password in the query string** — for the `mylistsummary` and `hints`
   user-data pages.

For (2), the source is explicit and this matters for security review:

> The password is the user's **main AniDB website password**, *not* the API password from the
> profile page.

Consequences an implementer must handle:
- Credentials travel in a **URL query string over plain HTTP** — logged by proxies, caches, and
  server access logs.
- Never log the full request URL for these two endpoints.
- Prefer the UDP API's `ENCRYPT` + `AUTH` flow for anything user-scoped.

## Endpoint Catalog

| `request=` | Auth | Returns | Root element |
|------------|------|---------|--------------|
| `anime` | client only | Full anime record by AID | `<anime>` |
| `randomrecommendation` | client only | Random recommendations (front-page mirror) | `<randomrecommendation>` |
| `randomsimilar` | client only | Random similar-anime pairs (front-page mirror) | `<randomsimilar>` |
| `hotanime` | client only | Currently hot anime (front-page mirror) | `<hotanime>` |
| `main` | client only | Combined `hotanime` + `randomsimilar` + `randomrecommendation` | `<main>` |
| `mylistsummary` | client + user/pass | Per-user mylist summary | `<mylistsummary>` |
| `hints` | client + user/pass | Per-user recommendation hints | `<hints>` |

---

## `request=anime` — Anime by AID

Retrieves non-file, non-episode-detail information for one anime.

### Parameters

| Parameter | Type | Required | Notes |
|-----------|------|----------|-------|
| `request` | string | yes | `anime` |
| `aid` | integer | yes | AniDB anime id |

```
http://api.anidb.net:9001/httpapi?request=anime&client={str}&clientver={int}&protover=1&aid={int}
```

There is **no search-by-title parameter** on this endpoint. Title → AID resolution is done
client-side against the anime-titles dump (see Data Dumps).

### Documented caveats

- **`<tags>` carries both `spoiler` and `localspoiler` attributes with the same value.** Clients must
  read `localspoiler`; `spoiler` will eventually be removed.

  > **`[INFERRED]`** The supplied sample payload shows only `localspoiler` and `globalspoiler` on
  > `<tag>` — no bare `spoiler` attribute appears. The deprecated attribute may already be gone,
  > or may appear only on some records. Parse defensively: read `localspoiler`, fall back to
  > `spoiler` only if absent.

- **`<temporary>` vote semantics changed.** The node is retained for compatibility, but its value now
  represents **the mean of all votes (permanent and temporary combined)** — not temporary votes alone.
  Do not treat `permanent` and `temporary` as disjoint populations.

### Response schema — `<anime>`

Root: `<anime id="{int}" restricted="{bool}">`

| Attribute | Type | Notes |
|-----------|------|-------|
| `id` | int | AID |
| `restricted` | bool (`"true"`/`"false"`) | 18+ / adult gate. Observed values are lower-case strings. |

Child elements, in the order observed in the sample:

| Element | Cardinality | Type | Notes |
|---------|-------------|------|-------|
| `<type>` | 0..1 | string | Anime type. Observed values in source: `TV Series`, `OVA`. **The complete enum is not documented in the HTTP source.** |
| `<episodecount>` | 0..1 | int | |
| `<startdate>` | 0..1 | `YYYY-MM-DD` | May be partial in practice; the HTTP source documents no date-flag mechanism (the UDP API does — see `dateflags` there) |
| `<enddate>` | 0..1 | `YYYY-MM-DD` | Absent for still-airing titles |
| `<titles>` | 0..1 | container | |
| `<relatedanime>` | 0..1 | container | |
| `<similaranime>` | 0..1 | container | |
| `<recommendations>` | 0..1 | container | Has `total` attribute |
| `<url>` | 0..1 | string | Official site URL |
| `<creators>` | 0..1 | container | |
| `<description>` | 0..1 | string | Contains AniDB wiki-link markup — see Description markup below |
| `<ratings>` | 0..1 | container | |
| `<picture>` | 0..1 | string | **Bare filename**, e.g. `440.jpg` — see Images below |
| `<resources>` | 0..1 | container | External-database cross-references |
| `<tags>` | 0..1 | container | |
| `<characters>` | 0..1 | container | |
| `<episodes>` | 0..1 | container | |

#### `<titles>` → `<title>`

```xml
<title xml:lang="x-jat" type="main">Seikai no Monshou</title>
```

| Attribute | Type | Notes |
|-----------|------|-------|
| `xml:lang` | BCP-47-ish | Observed: `x-jat` (romanized Japanese), `ja`, `en`, `fr`, `pl`, `cs`, `zh-Hans`. Not an enum — treat as open. |
| `type` | enum | Observed: `main`, `official`, `synonym`, `short` |

There is exactly **one** `type="main"` title per anime in the sample. `official` is per-language.

#### `<relatedanime>` → `<anime>`

```xml
<anime id="4" type="Sequel">Seikai no Senki</anime>
```

| Field | Type | Notes |
|-------|------|-------|
| `id` (attr) | int | Related AID |
| `type` (attr) | string | Observed in HTTP sample: `Sequel`, `Prequel`, `Summary`. **The full relation vocabulary is not documented in the HTTP source.** The UDP API documents the numeric equivalent set (`1`=sequel, `2`=prequel, `11`=same setting, `12`=alternative setting, `32`=alternative version, `41`=music video, `42`=character, `51`=side story, `52`=parent story, `61`=summary, `62`=full story, `100`=other) — **`[INFERRED]`** the HTTP strings are the display labels of that same set. |
| text | string | Related anime's main title |

#### `<similaranime>` → `<anime>`

```xml
<anime id="584" approval="75" total="89">Ginga Eiyuu Densetsu</anime>
```

| Field | Type | Notes |
|-------|------|-------|
| `id` (attr) | int | |
| `approval` (attr) | int | Users who approved this similarity |
| `total` (attr) | int | Total users who voted on it |
| text | string | Main title |

**`[INFERRED]`** `approval/total` is a ratio; the source never states a normalization or a minimum
`total` threshold. Do not render a bare percentage without a vote-count floor.

#### `<recommendations total="{int}">` → `<recommendation>`

```xml
<recommendation type="Recommended" uid="567190">free text…</recommendation>
```

| Field | Type | Notes |
|-------|------|-------|
| `total` (attr, on parent) | int | Total recommendations. **May exceed the number of `<recommendation>` children returned** — the sample shows `total="20"` with fewer inline entries. |
| `type` (attr) | enum | Observed: `Recommended`, `Must See`. **Full enum not documented.** |
| `uid` (attr) | int | Author's AniDB user id |
| text | string | Free-form user prose, multi-line, apostrophes escaped as `` ` `` |

#### `<creators>` → `<name>`

```xml
<name id="4303" type="Music">Hattori Katsuhisa</name>
```

| Field | Type | Notes |
|-------|------|-------|
| `id` (attr) | int | Creator id (`creatorid` in UDP terms) |
| `type` (attr) | string | Observed: `Music`, `Direction`, `Character Design`, `Series Composition`, `Original Work`. **Full role vocabulary not documented.** Treat as open string. |
| text | string | Creator name (transcription) |

> Note the element name is `<name>`, not `<creator>`. It is also **not** the same shape as the UDP
> `CREATOR` reply, which carries kanji name, type, picture, and wiki URLs.

#### `<description>`

Plain text containing **AniDB's own inline link markup**, not HTML:

```
* Based on the sci-fi novel series by http://anidb.net/cr4495 [Morioka Hiroyuki].
```

Pattern: `http://anidb.net/{prefix}{id} [Display Text]`

Observed prefixes:

| Prefix | Entity | Example |
|--------|--------|---------|
| `cr` | creator | `http://anidb.net/cr4495` |
| `ch` | character | `http://anidb.net/ch4081` |
| `t` | tag | `http://anidb.net/t2324` |

> **`[INFERRED]`** The prefix vocabulary is not documented; the three above are the only ones in the
> sample. Others (`a` for anime, `ep`, `g`) are plausible but unverified. A parser should match a
> generic `https?://anidb\.net/([a-z]*)(\d+) \[([^\]]*)\]` and keep unknown prefixes as-is rather
> than dropping the link.

Apostrophes in description text appear as backticks (`` ` ``) — the same escape the UDP API
documents for its own transport. Leading `* ` denotes a bullet line.

#### `<ratings>`

```xml
<ratings>
  <permanent count="4430">8.16</permanent>
  <temporary count="4460">8.23</temporary>
  <review count="12">8.70</review>
</ratings>
```

| Element | Attribute | Notes |
|---------|-----------|-------|
| `<permanent>` | `count` = vote count | Value is a decimal 0–10 |
| `<temporary>` | `count` | **Now the mean of all votes** (see caveat above) |
| `<review>` | `count` | Mean review score |

**`[INFERRED]` — scale inconsistency across endpoints.** In `request=anime` the values are decimals
on a 0–10 scale (`8.16`). In `request=randomrecommendation` the sample shows
`<permanent count="781">623</permanent>` — an integer that looks like a 0–1000 scale (the UDP API's
native rating scale, where `950` = 9.50). **The two endpoints appear to use different rating scales.**
Normalize per-endpoint; do not assume one parser fits both.

`<recommendations>` also appears **inside `<ratings>`** in the `randomrecommendation` payload as a
bare count (`<recommendations>8</recommendations>`), which is a different element from the top-level
`<recommendations total="…">` container in `request=anime`. Do not share a parser between them.

#### `<picture>` and images

`<picture>` (and `<picurl>` on tags, `picture` attributes on characters and seiyuu) contain a
**bare filename** — `440.jpg`, `14304.jpg`, `184301.jpg`. No base URL is given anywhere in the HTTP
source.

**`[INFERRED]` — where the base URL comes from.** The UDP API's `AUTH` command accepts `imgserver=1`,
which makes the login reply include an **image server name** on a second line. That is the only
documented mechanism in either source for obtaining the image host. An HTTP-only client has **no
documented way** to construct image URLs and must either hardcode a host (fragile — the option exists
precisely because it changes) or perform one UDP `AUTH` with `imgserver=1` to discover it.

#### `<resources>` → `<resource>` → `<externalentity>`

Cross-references to external databases:

```xml
<resource type="1">
  <externalentity>
    <identifier>14</identifier>
  </externalentity>
</resource>
```

| Field | Type | Notes |
|-------|------|-------|
| `type` (attr) | int | External-database discriminator |
| `<externalentity>` | 1..n | One per linked entity |
| `<identifier>` | 0..n | **Repeatable** — some types use a multi-part key |
| `<url>` | 0..1 | Used instead of `<identifier>` for some types |

**`[HIDDEN]` — the `type` → database mapping is documented nowhere in the source.** These are the
type ids observed in the single sample payload, with the identifier shape seen:

| `type` | Identifier shape observed | Notes |
|--------|---------------------------|-------|
| `1` | `14` (int) | |
| `2` | `290` (int) | |
| `3` | `376`, `xzudnt` (int + slug, two `<identifier>` children) | Multi-part key |
| `4` | `<url>http://www.sunrise-inc.co.jp/seikai/</url>` | Uses `<url>`, not `<identifier>`. Matches the top-level `<url>` element value in the same record. |
| `6` | `Crest_of_the_Stars` (underscore slug) | |
| `7` | `星界の紋章` (Japanese text) | |
| `8` | `1333` (int) | |
| `9` | `90291` (int) | |
| `10` | `1869` (int) | |
| `14` | `11753`, `v` (int + single char, two `<identifier>` children) | Multi-part key |
| `17` | `seikai01` (slug) | |
| `31` | `6674` (int) | |

Do **not** hardcode a `type` → provider mapping from memory. The safe ingestion strategy is to
persist `(type, [identifiers…], url)` verbatim and resolve providers later against AniDB's own
documentation or the web UI. Note that ids `5`, `11`, `12`, `13`, `15`, `16`, `18`–`30` are absent
from the sample — the space is sparse and open-ended.

The UDP API's `amask` exposes three of these as first-class fields (`int ANN id`, `int allcinema id`,
`str AnimeNfo id`), which is a partial cross-check available to an implementer.

#### `<tags>` → `<tag>`

```xml
<tag id="36" parentid="2607" weight="300" localspoiler="false"
     globalspoiler="false" verified="true" update="2018-01-21">
  <name>military</name>
  <description>…</description>
  <picurl>212184.jpg</picurl>
</tag>
```

| Field | Type | Notes |
|-------|------|-------|
| `id` (attr) | int | Tag id |
| `parentid` (attr) | int, **optional** | Absent on root tags. The tag tree is reconstructed client-side by joining `parentid` → `id`. |
| `weight` (attr) | int | Observed `0`–`500`. `0` appears on structural/branch tags (`themes`, `fetishes`, `content indicators`) that carry no relevance signal. **Filter `weight="0"` out of user-facing tag lists** — they are taxonomy nodes, not descriptors. |
| `infobox` (attr) | bool, **optional** | **`[HIDDEN]` — never mentioned in the prose.** Observed as `infobox="true"` on a subset of tags. **`[INFERRED]`** it marks tags AniDB surfaces in the anime info box (the sample's `infobox` tags sit under parents `2846`, `2609`, `2611`). Absent = not in infobox. |
| `localspoiler` (attr) | bool | **Use this one.** |
| `globalspoiler` (attr) | bool | Distinct from `localspoiler` — a tag can spoil universally vs. only for this anime |
| `spoiler` (attr) | bool | **Deprecated.** Documented as present and duplicating `localspoiler`; not observed in the sample. |
| `verified` (attr) | bool | Moderator-verified |
| `update` (attr) | `YYYY-MM-DD` | Last change to the tag definition (not to this anime's use of it) |
| `<name>` | string | |
| `<description>` | 0..1 string | Tag definition prose, shared across all anime using the tag. Large. Apostrophes escaped as `` ` ``. |
| `<picurl>` | 0..1 string | Bare filename |

**Ingestion note:** tag `<description>` is per-tag, not per-anime, and is repeated in full on every
anime payload that uses the tag. It dominates response size. Store tags in a separate table keyed by
tag `id` and deduplicate.

Known structural parent tag ids from the sample: `2604` content indicators, `2605` dynamic,
`2607` themes, `2608` fetishes, `2609`, `2611`, `2846`, `2891`.

#### `<characters>` → `<character>`

```xml
<character id="28" type="main character in" update="2012-07-25">
  <rating votes="1196">9.15</rating>
  <name>Abriel Nei Debrusc Borl Paryun Lafiel</name>
  <gender>female</gender>
  <charactertype id="1">Character</charactertype>
  <description>…</description>
  <picture>14304.jpg</picture>
  <seiyuu id="12" picture="184301.jpg">Kawasumi Ayako</seiyuu>
</character>
```

| Field | Type | Notes |
|-------|------|-------|
| `id` (attr) | int | Character id (`charid` in UDP terms) |
| `type` (attr) | enum | **Appearance role.** Observed: `main character in`, `secondary cast in`, `appears in`. The UDP `CHARACTER` command documents the numeric set as `0`=appears in, `1`=cameo appearance in, `2`=main character in, `3`=secondary cast in — **`[INFERRED]`** `cameo appearance in` is the fourth HTTP string, unobserved in the sample. |
| `update` (attr) | `YYYY-MM-DD` | |
| `<rating votes="{int}">` | decimal | 0–10 scale here |
| `<name>` | string | Transcription. **No kanji name in the HTTP payload** — the UDP `CHARACTER` command returns both kanji and transcription. |
| `<gender>` | enum | Observed: `male`, `female`. UDP documents the fuller set as single letters: `M`, `F`, `I` (intersexual), `D` (dimorphic), `-` (n/a), `?` (unknown). **`[INFERRED]`** the HTTP strings are the expanded labels of that set; the source warns this mapping "is subject to changes so don't rely on this mapping". |
| `<charactertype id="{int}">` | enum | Observed pairs: `1`=`Character`, `3`=`Organization`, `4`=`Vessel`. UDP documents `1`=Character, `2`=Mecha, `3`=Organisation, `4`=Vessel. **Note the spelling differs between surfaces** (`Organization` in HTTP XML vs `Organisation` in UDP docs) and `2`=Mecha is unobserved in the HTTP sample. Key off the numeric `id`, never the label. |
| `<description>` | 0..1 string | |
| `<picture>` | 0..1 string | Bare filename |
| `<seiyuu>` | 0..n | Voice actor. `id` = creator id, `picture` = bare filename, text = name. Absent when no seiyuu is credited. |

#### `<episodes>` → `<episode>`

```xml
<episode id="1" update="2011-07-01">
  <epno type="1">1</epno>
  <length>25</length>
  <airdate>1999-01-03</airdate>
  <rating votes="28">3.31</rating>
  <title xml:lang="ja">侵略</title>
  <title xml:lang="en">Invasion</title>
</episode>
```

| Field | Type | Notes |
|-------|------|-------|
| `id` (attr) | int | Episode id (`eid`) |
| `update` (attr) | `YYYY-MM-DD` | |
| `<epno type="{int}">` | string | **Text is the episode number; `type` is the episode kind.** Observed types in the sample: `1` and `3`. UDP documents the full set: `1`=regular, `2`=special (`S` prefix), `3`=credit (`C`), `4`=trailer (`T`), `5`=parody (`P`), `6`=other (`O`). The HTTP `<epno>` text is the **bare number without the prefix character**; UDP's `epno` string includes the prefix. Reconstruct the display form as `{prefix}{number}` using `type`. |
| `<length>` | int | Minutes |
| `<airdate>` | 0..1 `YYYY-MM-DD` | |
| `<rating votes="{int}">` | decimal | 0–10 |
| `<title>` | 1..n | `xml:lang` per title. No `type` attribute here (unlike anime titles). |

**`[HIDDEN]` — episodes are not returned in episode-number order.** The sample returns `id=1` (ep 1),
`id=2` (ep 2), then `id=1009` (ep **10**), `id=1010` (ep **5**), `id=1011` (ep **4**)… Ordering
appears to follow `eid` ascending, which does not track `epno`. **Always sort client-side by
`(type, numeric epno)`.**

---

## `request=randomrecommendation`

Mirrors the random-recommendation block on the AniDB front page. Results are genuinely random and
will **not** match what the website shows at the same moment.

### Parameters

Global four only (`request=randomrecommendation`).

### Response

```xml
<randomrecommendation>
  <recommendation>
    <anime id="7899" restricted="false">
      <type>TV Series</type>
      <episodecount>13</episodecount>
      <startdate>2010-10-04</startdate>
      <enddate>2010-12-27</enddate>
      <title xml:lang="x-jat" type="main">Arakawa Under the Bridge 2</title>
      <picture>54734.jpg</picture>
      <ratings>
        <permanent count="781">623</permanent>
        <recommendations>8</recommendations>
      </ratings>
    </anime>
  </recommendation>
</randomrecommendation>
```

Differences from `request=anime` that a shared parser will get wrong:

- `<anime>` is **nested** under `<recommendation>`, not the root.
- `<title>` is a **single flat element**, not wrapped in `<titles>`.
- `<ratings>` contains `<recommendations>` as a **bare integer count**, and `<permanent>` carries an
  **integer** value (`623`) rather than the decimal (`8.16`) form used by `request=anime`.

> The source's closing tag in the sample is written `<randomrecommendation>` rather than
> `</randomrecommendation>` — a transcription error in the wiki, not a protocol feature.

---

## `request=randomsimilar`

Mirrors the front-page "random similar" block. Also genuinely random; will not match the site.

### Response

```xml
<randomsimilar>
  <similar>
    <source aid="7056" restricted="false">
      <title xml:lang="x-jat" type="main">Aoi Bungaku Series</title>
      <picture>37124.jpg</picture>
    </source>
    <target aid="3654" restricted="false">
      <title xml:lang="x-jat" type="main">Ayakashi: Japanese Classic Horror</title>
      <picture>44085.jpg</picture>
    </target>
  </similar>
</randomsimilar>
```

**`[HIDDEN]` — attribute name changes here.** `<source>` and `<target>` use **`aid`**, while
`<anime>` elsewhere uses **`id`** for the same concept. A generic "read the anime id" helper must
handle both.

The pair is **directed** (`source` → `target`); the source does not state whether the inverse
relation is also returned independently.

---

## `request=hotanime`

Mirrors the front-page hot-anime block. **Unlike the two random endpoints, these results do match the
website**, modulo cache lag of a few hours.

### Response

```xml
<hotanime>
  <anime id="8556" restricted="false">
    <episodecount>12</episodecount>
    <startdate>2012-01-10</startdate>
    <title xml:lang="x-jat" type="main">Another</title>
    <ratings>
      <permanent count="248">6.61</permanent>
      <temporary count="261">7.88</temporary>
      <review count="1">4.33</review>
    </ratings>
    <picture>79963.jpg</picture>
  </anime>
</hotanime>
```

- `<anime>` sits directly under the root here (no wrapper element).
- Uses `id`, not `aid`.
- Ratings are **decimal** here — matching `request=anime`, unlike `randomrecommendation`.
- No `<type>` and no `<enddate>` in the sample (still-airing titles).
- The source does not document how many entries are returned.

---

## `request=main`

One-stop command returning all three front-page blocks. Use it instead of issuing two or three
separate requests.

### Response

```xml
<main>
  <hotanime>…</hotanime>
  <randomsimilar>…</randomsimilar>
  <randomrecommendation>…</randomrecommendation>
</main>
```

Each child block has exactly the shape of its standalone endpoint. Given the one-request-per-2-seconds
limit, `main` is strictly better than three calls whenever more than one block is needed.

---

## `request=mylistsummary` — User MyList summary

**Requires user credentials in the query string.** See the Authentication warning above.

### Parameters

| Parameter | Type | Required | Notes |
|-----------|------|----------|-------|
| `request` | string | yes | `mylistsummary` |
| `user` | string | yes | AniDB username |
| `pass` | string | yes | **Main website password**, not the profile API password |
| `output` | enum | no | **`[HIDDEN]`** — `xml` or `json`. Appears only in the Base URL line as `[&output=xml|json]`; never explained in prose, and no JSON sample is given. Default is XML. |

```
http://api.anidb.net:9001/httpapi?client={str}&clientver={int}&protover=1&request=mylistsummary&user={str}&pass={str}
```

### Response

```xml
<mylistsummary uid="40237">
  <mylistitem aid="5841" state="2">
    <episodecount>22</episodecount>
    <seencount>22</seencount>
    <specialcount>9</specialcount>
    <seenspecialcount>4</seenspecialcount>
  </mylistitem>
</mylistsummary>
```

| Field | Type | Notes |
|-------|------|-------|
| `uid` (attr, root) | int | Resolved user id |
| `aid` (attr) | int | |
| `state` (attr) | int | **`[HIDDEN]` — not documented on this endpoint.** **`[INFERRED]`** it is the UDP MyList state enum: `0`=unknown, `1`=internal storage (HDD), `2`=external storage (CD/DVD), `3`=deleted, `4`=remote storage. Sample value `2` is consistent with that. Verify before relying. |
| `<episodecount>` | int | |
| `<seencount>` | int | |
| `<specialcount>` | int | |
| `<seenspecialcount>` | int | |

> The sample's closing tag is written `<mylistsummary>` instead of `</mylistsummary>` — again a wiki
> transcription artifact.

---

## `request=hints` — User recommendation hints

Returns unwatched titles with a calculated predicted rating, derived from the user's existing votes.

Behavior:
- **Recalculated once per day** — polling more often than daily is pure waste.
- The list may be **empty** if the user is new, has too few votes, or has stale (months-old) data.

### Parameters

| Parameter | Type | Required | Notes |
|-----------|------|----------|-------|
| `request` | string | yes | `hints` |
| `user` | string | yes | AniDB username |
| `pass` | string | yes | Main website password |
| `type` | integer | yes | **Always `1`, but must be included.** No other value is documented. |
| `output` | enum | no | **`[HIDDEN]`** — `xml` or `json`, same undocumented option as `mylistsummary` |

### Response

```xml
<hints uid="40237" type="1">
  <hint adddate="2015-03-23">
    <relid>5</relid>
    <calculated>8.09</calculated>
  </hint>
</hints>
```

| Field | Type | Notes |
|-------|------|-------|
| `uid` (attr) | int | |
| `type` (attr) | int | Echoes the request `type` |
| `adddate` (attr) | `YYYY-MM-DD` | **`[INFERRED]`** — undocumented. Given `hints` are recalculated daily, this is most plausibly when the hint entered the user's list, not when the anime was added to AniDB. |
| `<relid>` | int | **`[INFERRED]`** — undocumented. Named like the UDP notification `relid` ("id of the related type"). Sample value `5` is a valid AID. Treat as an AID; verify. |
| `<calculated>` | decimal | Predicted rating, 0–10 scale |

---

## Data Dumps

Unauthenticated bulk files. **Not** part of the `httpapi` endpoint.

| URL | Format |
|-----|--------|
| `http://anidb.net/api/anime-titles.dat.gz` | Gzipped flat file |
| `http://anidb.net/api/anime-titles.xml.gz` | Gzipped XML |

Rules, stated in the source in capitals:

> **YOU DO NOT REQUEST THIS FILE MORE THAN ONCE PER DAY**

Both files are updated daily and contain **all anime titles**. Intended uses:

- Client-side anime search (there is no search endpoint on the HTTP API)
- **Title → AID lookup**, which is the required precursor to `request=anime`

A third-party hosted search built on these dumps exists: `http://anisearch.outrance.pl/`.

> The source documents neither file's internal schema. An implementer must inspect the actual dump.
> Do not guess the column layout of the `.dat` variant.

---

## Additional Resources (not HTTP XML API)

Listed in the source's API overview for completeness. Each has its own separate documentation not
included in the supplied export.

| Resource | Purpose |
|----------|---------|
| **MyList Export** | User-initiated export of an entire MyList in several formats including XML. Recommended as the *first* step for obtaining a bulk local dataset of a user's files/episodes/anime, rather than crawling. Can be queued programmatically via the UDP `MYLISTEXPORT` command. |
| **SigServer** | Dynamic forum-signature images; underlying data also available raw |
| **RSSRDF** | RSS feeds for recently added files and per-user file notifications |
| **IRC bots** | `Chii` and `Lafiel` on SynIRC — real-time DB change notifications |
| **Jabber** | Recommended over the UDP API for notifications and private messaging |

---

## AI Coding Rules

1. **Never branch on HTTP status.** Check for a root `<error>` element on every response.
2. **Rate limit at 1 request / 2 s minimum**, and cache aggressively. Same-day repeat fetches of the
   same AID risk a ban.
3. **Never request the anime-titles dump more than once per 24 h.**
4. **Do not scrape AniDB HTML.** The source asks explicitly; do the `main` request instead.
5. **Resolve titles locally** against the dump. There is no search endpoint.
6. **Use `localspoiler`, not `spoiler`.**
7. **Treat `<temporary>` as "mean of all votes"**, not as temporary-only.
8. **Do not assume one rating scale.** `request=anime` and `hotanime` return decimals (0–10);
   `randomrecommendation` returns integers that look like 0–1000. Normalize per endpoint.
9. **Handle both `id` and `aid`** when reading an anime identifier — the attribute name varies by
   endpoint.
10. **Sort episodes yourself.** Payload order tracks `eid`, not `epno`.
11. **Deduplicate tag descriptions** into a tag table; they are repeated verbatim per anime and
    dominate payload size.
12. **Persist `<resources>` verbatim** as `(type, identifiers[], url)`. The provider mapping is
    undocumented — do not fabricate it.
13. **Filter `weight="0"` tags** out of user-facing lists; they are taxonomy branch nodes.
14. **Never log the request URL** for `mylistsummary` or `hints` — it contains the user's main
    website password in plaintext over HTTP.
15. **Images need a host from elsewhere.** `<picture>`/`<picurl>` are bare filenames; the only
    documented discovery mechanism is UDP `AUTH` with `imgserver=1`.
16. **Handle chunked transfer and gzip explicitly**, including on HTTP/1.0 requests.

---

## Source Boundary

This reference is derived solely from the supplied AniDB wiki exports (`anidb/1.md`, `anidb/2.md`).

Every field, attribute, and enum value above is either (a) stated in that source, or (b) observed in
the source's own sample payloads and marked **`[INFERRED]`** / **`[HIDDEN]`**. Nothing has been added
from third-party clients or model memory.

Substantial areas the supplied source does **not** cover, and where the live AniDB documentation must
be consulted before implementation:

- The complete `<type>` enum for anime (only `TV Series` and `OVA` observed)
- The complete `<resource type="…">` → external-database mapping
- The `<error>` string vocabulary (only `Banned` observed)
- The anime-titles dump file schemas (both `.dat` and `.xml`)
- The `output=json` response shapes for `mylistsummary` and `hints`
- The image server base URL
- Response sizes / entry counts for `hotanime`, `randomsimilar`, `randomrecommendation`
- Whether `spoiler` (deprecated) is still emitted
