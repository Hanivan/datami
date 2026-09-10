# AniDB UDP API — AI Implementation Reference

> Source: the supplied AniDB wiki export (`anidb/3.md` — UDP API Definition, spec version
> `0.03.730`, 2015-03-25, by Exp & Epoximator & Ommina; `anidb/1.md` — API overview).
> This file is a normalized, AI-oriented reference derived from that source.
> **Do not invent undocumented commands, parameters, mask bits, enum values, or return codes.**
> Where the source is silent but its own tables or samples imply structure, this document marks it
> **`[INFERRED]`** or **`[HIDDEN]`** and states what was observed.

## Purpose

Use this document as implementation context for an AI coding agent building an AniDB UDP client.

The UDP API is the **primary** AniDB API: it covers files, episodes, anime, groups, characters,
creators, MyList, votes, notifications, and messaging. The HTTP API (see
`ANIDB_HTTP_API_AI_REFERENCE.md`) exposes only a small read-only subset.

The source is explicit about two things a crawler author must internalize:

> The UDP API is **not** an appropriate choice if you desire to download a local copy of the AniDB
> database.

> If you are mainly interested in notifications and private messaging, check out Jabber and RSSRDF
> support first.

## API Identity

| Property | Value |
|----------|-------|
| Host | `api.anidb.net` |
| Port | `9000/UDP` |
| `protover` for this spec | `3` |
| Default encoding | ASCII |
| Max datagram | 1400 bytes (PPPoE constraint) |

> Host and port **must not be hardcoded** in a frontend — read them from a configuration file.

Client registration is required, and is **separate** from HTTP-API client registration.

## Protocol Basics

### Framing

One command = exactly **one UDP packet** = exactly **one line**.
One reply = one packet, possibly multiple lines.

```
{three digit return code} {str return string}\n
{data field 0}|{data field 1}|…|{data field n}
```

- Lines end with `\n` only — **never** `\r\n`.
- Fields within a data line are separated by `|`.
- Multiple result entries = one entry per line.
- **Important data may be appended directly to the return-code line**, not only on data lines.
  Codes documented as doing this: `200`, `201`, `209`, `271`, `272`, `504`.

### Forward compatibility

> A client should **ignore any additional trailing data fields it doesn't expect**. Fields are
> appended to command output over time.

Parse positionally from the left, tolerate extra fields on the right, never assert on field count.

### Truncation

Replies never exceed 1400 bytes (or the negotiated `mtu`). **String fields are truncated silently
with no warning.** With `comp=1` the server compresses instead of truncating where it can.

Documented truncation *order* for `ANIME` and `FILE`: character id list is dropped first, then tag
list, then synonym list, then short-name list.

### Types

| Type | Meaning |
|------|---------|
| `int2` | 2-byte integer, string representation |
| `int4` | 4-byte integer, string representation |
| `int8` | 8-byte integer, string representation |
| `boolean` | `1` = true, `0` = false |
| `str` | String; UDP packet length caps it at 1400 bytes |
| `hexstr` | Hex representation, two chars per byte, byte one first |

Spec notation: `{vartype varname}` = substitute a value; `[...]` = optional.

### Content escaping

**Outbound (client → server):**
- HTML form encoding for option values. At minimum encode `&` as `&amp;`. Other HTML entities optional.
- Newlines must be sent as `<br />`.

**Inbound (server → client):**
- Newlines in content are encoded as `<br />`
- `'` (apostrophe) is encoded as `` ` `` (backtick)
- `|` (pipe) is encoded as `/` (forward slash)
- Dates are **Unix time** (seconds since 1970-01-01 00:00:00 UTC)

> The `'` → `` ` `` substitution is lossy in the other direction: a genuine backtick and an escaped
> apostrophe are indistinguishable. Same for a genuine `/` and an escaped `|`. This also explains the
> backticks visible in HTTP API description text.

**Collision hazard:** `'` is *also* the record separator inside several composite fields
(`CHARACTER` anime blocks, `GROUP` relations, synonym lists, multi-stream `FILE` audio fields). Since
content apostrophes are already escaped to `` ` ``, an unescaped `'` in those fields is always a
separator — but the two are the same glyph, so split before unescaping, never after.

### Data indexes (`fid`, `aid`, `eid`, `gid`, `lid`)

- All indexes start at **1**, not 0.
- A value of `0` in an id field means **NONE / NULL** (commonly seen on `gid`).
- **IDs are never reused.** A deleted entry's id is retired permanently.
- MyList ids (`lid`) are **globally unique**, not per-user unique.

### `tag` option — request/response correlation

Any command accepts `tag={str usertag}`. The server prefixes each reply line with the tag plus a space.

```
AUTH user=baka&pass=baka&protover=3&client=someclient&clientver=1&tag=abc123
abc123 200 Jxqxb LOGIN ACCEPTED
```

- Tags are **per-command, not persistent** — append to every command that needs one.
- Tags exist so a client can have multiple requests in flight.
- **`[HIDDEN]` constraint:** with `comp=1`, compressed datagrams always begin with two zero bytes, so
  **a tag must never start with `\0\0`**.
- **PUSH notifications (720–799) never carry tags.** That is the documented way to tell an
  unsolicited notification from a command reply.
- The tag option is valid on all commands and is therefore omitted from each command's parameter list
  below.

## Connection Management

### Session lifecycle

1. Optionally `ENCRYPT` (must precede `AUTH`).
2. `AUTH` → receive `{session_key}`.
3. Append `s={session_key}` to **every** subsequent command that requires login.
4. `LOGOUT` when done.

- `session_key` is 4–8 chars of `[a-zA-Z0-9]`.
- The virtual connection **times out after 35 minutes** with no client data.
- Send `UPTIME` every 30 minutes to keep it alive *only if you need it alive*.
- If not using notifications/push, **do not** keep the session alive — `LOGOUT` explicitly.
- If another command is likely within 20 minutes, **stay logged in**. Do **not** do an
  AUTH/command/LOGOUT cycle per command.

### Commands that do NOT require a session

`PING`, `ENCRYPT`, `ENCODING`, `AUTH`, `VERSION`.

Everything else requires `s=`.

### Local port discipline

- Pick a **fixed local port > 1024 at install time** and reuse it for all UDP sockets — including
  during testing and debugging.
- **If the API sees too many different UDP ports from one IP within ~1 hour, it bans the IP.**
- The port may be hardcoded, but an option to override it must be offered.

### NAT

A session is identified by **(client IP, client port)**. If either changes mid-session, the client
must re-authenticate — and repeated re-auth from new ports is exactly what triggers the ban above.

- Send `nat=1` with `AUTH`. The reply then includes the IP and port **as seen by the server**.
- If that port differs from the local socket's port, the client is behind NAT and must `PING`
  periodically to keep the router's mapping alive (routers typically deallocate after 5–15 min).
- **Do not ping more often than needed.**

### Login failure backoff

If `AUTH` gets **no reply at all**, back off exponentially:

> 30 s → 2 min → 5 min → 10 min → 30 min → … up to ~2–4 h.

### Flood protection — hard limits

| Window | Limit |
|--------|-------|
| Short term | **< 0.5 packets/second** (one packet per two seconds — *not* two per second). Enforced after the first 5 packets. |
| Long term | **< 1 packet per 4 seconds** sustained. "Extended amount of time" is undefined; use common sense. |

Once exceeded, **all further packets are dropped silently with no feedback**.

> **Dropped packets still count toward the packet rate.** A client that keeps sending during a
> drop will be **banned forever**.

Diagnostics:
- Sudden high packet loss where there was none = you have hit the flood ratio. Fix: enforce inter-command delay.
- No replies at all = down, or banned (bans typically last ~30 minutes).

### Caching — mandatory

> A client should locally cache FILE/EP/ANIME/GROUP/… info wherever possible for **extended amounts
> of time**.
>
> Later versions of the API might enforce this by **banning clients which ask for the same
> information more than once every week/month**.

Anti-leech protection: using the API to download AniDB gets you banned.

## Error Model

### Universal return codes

Possible on **every** command:

| Code | Meaning |
|------|---------|
| `505` | `ILLEGAL INPUT OR ACCESS DENIED` |
| `555` | `BANNED` — followed by `{str reason}` |
| `598` | `UNKNOWN COMMAND` |
| `600` | `INTERNAL SERVER ERROR` |
| `601` | `ANIDB OUT OF SERVICE - TRY AGAIN LATER` |
| `602` | `SERVER BUSY - TRY AGAIN LATER` |
| `604` | `TIMEOUT - DELAY AND RESUBMIT` |

Additionally, on every command that requires login:

| Code | Meaning | Client action |
|------|---------|---------------|
| `501` | `LOGIN FIRST` | Silently re-`AUTH`, then resend the failed command |
| `502` | `ACCESS DENIED` | Abort the current action, show the user a message |
| `506` | `INVALID SESSION` | Re-issue `AUTH` |

> A frontend shall expect `501` and `502` on **ANY** command.

Notes from the source:
- `555 BANNED` "should not be expected by the client" — it is enabled only temporarily to help
  developers diagnose misuse.
- On `601` (daily maintenance), **wait at least 30 minutes** before retrying.
- Any command with parameters may return `505`.

### Fatal errors (6xx)

At any point:

```
6xx INTERNAL SERVER ERROR
ERROR: {str errormessage}
```

or

```
6xx INTERNAL SERVER ERROR - {str errormessage}
```

`xx` is 00–99. All 6xx except `601` and `602` should be reported upstream.

> **`[HIDDEN]`** `6XX` messages **do not always return the tag** sent with the failing command. A
> tag-keyed request/response correlator must therefore have a fallback path for untagged 6xx replies,
> or it will leak pending requests.

---

# Command Reference

## Authing Commands

### `AUTH` — Authenticate

```
AUTH user={str username}&pass={str password}&protover={int4 apiversion}&client={str clientname}&clientver={int4 clientversion}[&nat=1&comp=1&enc={str encoding}&mtu={int4 mtu}&imgserver=1]
```

| Parameter | Req | Notes |
|-----------|-----|-------|
| `user` | yes | AniDB username. Charset is restricted to `a-z A-Z 0-9 _ -` |
| `pass` | yes | **The normal AniDB website password.** The profile's "API Password" is used *only* for the optional `ENCRYPT` feature. |
| `protover` | yes | The API spec version the client was written against — **`3`** for this document. A client **MAY NOT** send anything else. |
| `client` | yes | Registered client name, lower-case `a-z`, 4–16 chars |
| `clientver` | yes | Starts at 1, **must** be incremented on every release |
| `nat` | no | `1` → reply includes server-observed IP:port |
| `comp` | no | `1` → client supports DEFLATE; server compresses instead of truncating |
| `enc` | no | Session encoding. Silently ignored if unsupported — use `ENCODING` to test. |
| `mtu` | no | Default 1400, min 400, max 1400 |
| `imgserver` | no | `1` → reply includes the image server name on a second line |

**Replies**

| Code | Shape |
|------|-------|
| `200` | `200 {str session_key} LOGIN ACCEPTED` |
| `201` | `201 {str session_key} LOGIN ACCEPTED - NEW VERSION AVAILABLE` |
| `500` | `LOGIN FAILED` |
| `503` | `CLIENT VERSION OUTDATED` |
| `504` | `CLIENT BANNED - {str reason}` |
| `505` | `ILLEGAL INPUT OR ACCESS DENIED` |
| `601` | `ANIDB OUT OF SERVICE - TRY AGAIN LATER` |

With `nat=1`:
```
200 {str session_key} {str ip}:{int2 port} LOGIN ACCEPTED
201 {str session_key} {str ip}:{int2 port} LOGIN ACCEPTED - NEW VERSION AVAILABLE
```

With `imgserver=1`:
```
200 {str session_key} LOGIN ACCEPTED
{str image server name}
```

> **`[HIDDEN]` — `imgserver=1` is the only documented way, across both APIs, to discover the image
> host.** Both the UDP `picname`/`pic` fields and the HTTP API's `<picture>`/`<picurl>` elements
> contain bare filenames with no base URL anywhere else in either source. An HTTP-only client has no
> documented path to full image URLs.
>
> **`[INFERRED]`** the two reply forms are not documented as combinable; the source shows `nat=1` and
> `imgserver=1` reply shapes separately and never together.

**Client obligations**

- Handle **all** AUTH return codes before releasing any version.
- `500` → prompt the user for credentials again.
- `503` → the server no longer supports this `protover`. Pop up a message telling the user to update.
- `504` → **this client version is banned, not the user.** Pop up a message.
- `201` → notify the user a newer client exists (old one still works).
- **Do not reuse an existing client's `clientname`.** Fork the code, get your own name.

### `LOGOUT`

```
LOGOUT
```

| Code | Meaning |
|------|---------|
| `203` | `LOGGED OUT` |
| `403` | `NOT LOGGED IN` |

Always issue on exit, or when no traffic is expected for ≥ 30 minutes.

### `ENCRYPT` — Start encrypted session

```
ENCRYPT user={str name}&type={int2 type}
```

| Code | Meaning |
|------|---------|
| `209` | `{str salt} ENCRYPTION ENABLED` |
| `309` | `API PASSWORD NOT DEFINED` |
| `509` | `NO SUCH ENCRYPTION TYPE` |
| `394` | `NO SUCH USER` |

Mechanics:

- `type` = `1` → 128-bit AES. **Only type defined.**
- Encryption key = `md5(api_key_of_user + salt)`, where `api_key_of_user` is the **UDP API Key from
  the user's profile** (distinct from the website password used by `AUTH`).
- Padding: **PKCS5Padding**.
- Every message after the `ENCRYPT` reply is encrypted, in both directions. **Non-encrypted messages
  are discarded by the server.**
- A normal `AUTH` still follows and must itself be encrypted.
- Encryption **cannot be disabled while logged in** — only `LOGOUT` (correctly encrypted) or timeout
  ends it.
- Should be **off by default** and opt-in via configuration, to limit server load.

---

## Notify / Push Commands

### Concepts

Three event types produce notifications:

1. **New file** — anime-scoped only. *Group- and producer-scoped are defined in the protocol but not
   implemented.*
2. **New private message**
3. **Buddy event**

Only new-file and new-message events are **persistent** (they survive until acknowledged). Buddy and
going-down events are transient.

Two delivery models, not mutually exclusive:

**Polling** — easy, blocking sockets suffice, but introduces latency and staleness. Note that
notifications cleared by a client are also cleared from the website; users must be told.

```
NOTIFY  (≤ once / 20 min)  →  NOTIFYLIST  →  NOTIFYGET  →  NOTIFYACK
```

**Server PUSH** — lower latency, but requires non-blocking sockets and the ability to receive
unsolicited packets. Notifications go to the IP:port the `AUTH` came from.

```
PUSH  →  listen for 720–799  →  PUSHACK  →  NOTIFYGET  →  NOTIFYACK
       +  UPTIME every 30–35 min
```

> Listen for `720–799` **NOTIFICATION** packets, **not** `290` (which is the `NOTIFY` reply).
> Use `PUSHACK` with the **`nid` from the packet**, but `NOTIFYGET`/`NOTIFYACK` with the **`relid`
> from the notification body** — the packet id and the entity id are different values. Getting this
> wrong is the most likely PUSH bug.

### `PUSH` — Register for push notifications

```
PUSH notify={boolean on_new_file}&msg={boolean on_new_private_message}[&buddy={boolean on_buddy_event}]
```

| Code | Meaning |
|------|---------|
| `270` | `NOTIFICATION ENABLED` |
| `370` | `NOTIFICATION DISABLED` — returned when both values are 0 |

Obligations once registered:
- `PUSHACK` every notification, using its `notify_packet_id`.
- `PING` to keep NAT alive (< 30 min).
- `UPTIME` to keep the session valid (≥ 30 min).
- **Every notification is resent 3 times unless acknowledged. After that the client is logged out.**

### `NOTIFY` — Pending counts

```
NOTIFY [buddy=1]
```

```
290 NOTIFICATION
{int4 pending_file_notifications}|{int4 number_of_unread_messages}
```

With `buddy=1`, a third field is appended: `{int4 number_of_online_buddies}`.

> **MUST NOT be issued more than once every 20 minutes.**

**`[HIDDEN]` diagnostic:** if a `NOTIFY` succeeded within the last 35 minutes and the next `NOTIFY`
returns `501 LOGIN FIRST`, that specifically means **AniDB logged the client out for failing to
respond to a PUSH notification**. There is no command to retrieve missed PUSH notifications.

### `NOTIFYLIST` — List pending ids

```
NOTIFYLIST
```

```
291 NOTIFYLIST
{str type}|{int4 id}
…
```

| `type` | Meaning | `id` is |
|--------|---------|---------|
| `M` | Message | The actual message id |
| `N` | New-file notification | The id of the **related type** — currently always an `aid`, potentially a `gid` or producer id in future |

If nothing is pending, only the first line is returned.

> **MUST NOT be issued regularly.** Trigger it only from a push message or a non-zero `NOTIFY` reply.

### `NOTIFYGET` — Fetch a notification or message

```
NOTIFYGET type={str type}&id={int4 id}
```

For `type=M`:
```
292 NOTIFYGET
{int4 id}|{int4 from_user_id}|{str from_user_name}|{int4 date}|{int4 type}|{str title}|{str body}
392 NO SUCH ENTRY
```
Message `type`: `0`=normal, `1`=anonymous, `2`=system, `3`=mod.

For `type=N`:
```
293 NOTIFYGET
{int4 relid}|{int4 type}|{int2 count}|{int4 date}|{str relidname}|{str fids}
393 NO SUCH ENTRY
```

| Field | Notes |
|-------|-------|
| `relid` | Id of the related type (anime) |
| `relidname` | Name of the related type. **`[HIDDEN]`** the field is called `relidname` in the reply format but `relname` in the prose — same field. |
| `type` | Notification type: `0`=all, `1`=new, `2`=group, `3`=complete |
| `count` | Events pending for this subscription |
| `date` | Unix time |
| `fids` | **Comma-separated** list of affected file ids (note: comma, not the usual `'`) |

`id` accepts the value from `NOTIFYLIST`, or `relid` from a `271` notification, or `mid` from a `272`
notification.

> **`[HIDDEN]` — codes `271` and `272` are named here as notification-bearing codes** (and listed
> among codes that append data to the return line), but **neither appears in the Return Codes table
> nor has a documented packet format.** The documented push packets use `720`/`794`/`753`/`799`.
> Treat `271`/`272` as undocumented legacy; log rather than assume.

### `NOTIFYACK` — Acknowledge

```
NOTIFYACK type={str type}&id={int4 id}
```

| `type` | Success | Failure |
|--------|---------|---------|
| `M` | `281 NOTIFYACK SUCCESSFUL` | `381 NO SUCH ENTRY` |
| `N` | `282 NOTIFYACK SUCCESSFUL` | `382 NO SUCH ENTRY` |

Marks a message read / clears a new-file notification. **Buddy events are not acknowledgeable.**

### Notification packet formats (unsolicited, 720–799)

**New file — `720`**
```
720 {int4 notify_packet_id} NOTIFICATION - NEW FILE
{int4 fidlist}|{int2 reltype}|{int2 priority}
```
- `fidlist` — **comma-separated** file ids despite the `int4` annotation. New files arrive in batches,
  so multiple fids for one anime at once is normal.
- `reltype` — `1`=anime, `2`=group, `3`=producer. *Group and producer are not implemented yet.*
- `priority` — `0`=low, `1`=medium, `2`=high

**New private message — `794`**
```
794 {int4 notify_packet_id} NOTIFICATION - NEW MESSAGE
{int2 type}|{int4 date}|{int4 sent_by_uid}|{str sent_by_name}|{str subject}|{str body}|{int mid}
```
- `type` — `0`=normal, `1`=anonymous, `2`=system, `3`=mod
- `body` **can be truncated**
- `mid` is usable with `NOTIFYACK`

**Buddy event — `753`**
```
753 {int4 notify_packet_id} NOTIFICATION - BUDDY
{int4 buddy uid}|{int2 event type}
```
- event type: `0`=LOGIN, `1`=LOGOUT, `2`=ACCEPTED, `3`=ADDED

**Going down — `799`**
```
799 {int4 notify_packet_id} NOTIFICATION - SHUTTING DOWN
{int4 time offline}|{int4 comment}
```
- `time offline` in **minutes**; `0` = indefinite (direct the user to the website).
- **`[HIDDEN]`** `comment` is typed `int4` in the format line but described in prose as "a short
  explanation for the downtime" — i.e. **it is a string**. The type annotation is a source error.
- Sent to any client with notifications enabled. **Only one datagram is sent and the server will not
  listen for replies** — do not `PUSHACK` this one.

### `PUSHACK`

```
PUSHACK nid={int4 notify_packet_id}
```

| Code | Meaning |
|------|---------|
| `701` | `PUSHACK CONFIRMED` |
| `702` | `NO SUCH PACKET PENDING` |

> A client **must be prepared to issue this before using `PUSH`.**

---

## Notification Subscription Commands

### `NOTIFICATIONADD` — Subscribe to an anime or group

```
NOTIFICATIONADD aid={int4 aid}&type={int2 type}&priority={int2 priority}
NOTIFICATIONADD gid={int4 gid}&type={int2 type}&priority={int2 priority}
```

| Code | Payload |
|------|---------|
| `246` | `NOTIFICATION ITEM ADDED` + `{int4 notification id}` |
| `248` | `NOTIFICATION ITEM UPDATED` + `{int4 notification id}` |
| `399` | `NO CHANGES` |

- `priority`: `0`=low, `1`=medium, `2`=high
- `type`: `0`=all, `1`=new, `2`=group, `3`=complete

### `NOTIFICATIONDEL` — Unsubscribe

```
NOTIFICATIONDEL aid={int4 aid}
NOTIFICATIONDEL gid={int4 gid}
```

| Code | Payload |
|------|---------|
| `247` | `NOTIFICATION ITEM DELETED` + `{int4 notification_table}|{int4 relid}` |
| `324` | `NO SUCH NOTIFICATION ITEM` |

- `notification_table`: `1`=anime, `2`=group
- `relid` echoes the supplied `aid`/`gid`

---

## Buddy Commands

### `BUDDYADD`

```
BUDDYADD uid={int4 buddy uid}
BUDDYADD uname={int4 buddy name}
```

> **`[HIDDEN]`** `uname` is typed `{int4 buddy name}` in the source — clearly a **`str`**. Type
> annotation error.

| Code | Meaning |
|------|---------|
| `255` | `BUDDY ADDED` |
| `355` | `BUDDY ALREADY ADDED` |
| `394` | `NO SUCH USER` |

### `BUDDYDEL`

```
BUDDYDEL uid={int4 buddy uid}
```
`256 BUDDY DELETED` / `356 NO SUCH BUDDY`

### `BUDDYACCEPT`

```
BUDDYACCEPT uid={int4 user uid}
```
`257 BUDDY ACCEPTED` / `357 BUDDY ALREADY ACCEPTED` / `356 NO SUCH BUDDY`

### `BUDDYDENY`

```
BUDDYDENY uid={int4 buddy uid}
```
`258 BUDDY DENIED` / `358 BUDDY ALREADY DENIED` / `394 NO SUCH USER`

### `BUDDYLIST`

```
BUDDYLIST startat={int2 start at #}
```
```
253 {int2 start} {int2 end} {int2 total} BUDDY LIST
{int4 uid}|{str username}|{int2 state}
…
```

`state` is a **16-bit bitfield, lowest bit first**:

| Bit | Dec | Meaning |
|-----|-----|---------|
| 1 | 1 | This user is in your buddylist |
| 2 | 2 | This user has accepted you |
| 3 | 4 | This user is waiting for your approval |

> **`[INFERRED]`** the source gives bit positions but no decimal column for this field; the decimals
> above follow from "lowest bit first" and the parallel `BUDDYSTATE` example.

### `BUDDYSTATE`

```
BUDDYSTATE startat={int2 start at #}
```
```
254 {int2 start} {int2 end} {int2 total} BUDDY STATE
{int4 uid}|{int2 onlinestate}
…
```

`onlinestate` bitfield:

| Bit | Dec | Meaning |
|-----|-----|---------|
| 1 | 1 | HTTP online (HTTP request within last 10 min) |
| 2 | 2 | UDP API online (currently connected to the UDP server) |

Source's own examples: `0`=offline, `1`=http, `2`=udp, `3`=both.

---

## Data Commands

### `ANIME` — Retrieve anime data

```
ANIME aid={int4 id}[&amask={hexstr}]
ANIME aname={str anime name}[&amask={hexstr}]
```

| Code | Meaning |
|------|---------|
| `230` | `ANIME` + data line |
| `330` | `NO SUCH ANIME` |

**Default reply (no `amask`)** — 19 fields:

```
{int4 aid}|{int4 eps}|{int4 ep count}|{int4 special cnt}|{int4 rating}|{int4 votes}|{int4 tmprating}|{int4 tmpvotes}|{int4 review rating average}|{int4 reviews}|{str year}|{str type}|{str romaji}|{str kanji}|{str english}|{str other}|{str short names}|{str synonyms}|{str category list}
```

Rules:

- **With `amask`, fields are returned in amask table order: byte 1 / bit 7 first**, top-to-bottom,
  left-to-right. The mask determines both *which* fields and *what order*.
- Synonyms and short names are separated by `'`.
- Category list is comma-separated, ordered by weight descending. **Categories are deprecated** — no
  longer used or updated internally, returned only for old-client compatibility. **Use tags instead.**
- `aname` lookup requires a **perfect match** of romaji / kanji / english / other / synonym / short name.
- **Selecting an `unused` or `reserved` bit returns `505 ILLEGAL INPUT`.**
- Ratings are on a **0–1000 integer scale** (see `VOTE`: `950` = 9.50).

#### `amask` bit tables

**Byte 1**

| Bit | Dec | Field |
|-----|-----|-------|
| 7 | 128 | `int aid` |
| 6 | 64 | `int dateflags` |
| 5 | 32 | `str year` |
| 4 | 16 | `str type` |
| 3 | 8 | `str related aid list` |
| 2 | 4 | `str related aid type` |
| 1 | 2 | retired |
| 0 | 1 | retired |

**Byte 2**

| Bit | Dec | Field |
|-----|-----|-------|
| 7 | 128 | `str romaji name` |
| 6 | 64 | `str kanji name` |
| 5 | 32 | `str english name` |
| 4 | 16 | `str other name` |
| 3 | 8 | `str short name list` |
| 2 | 4 | `str synonym list` |
| 1 | 2 | retired |
| 0 | 1 | retired |

**Byte 3**

| Bit | Dec | Field |
|-----|-----|-------|
| 7 | 128 | `int4 episodes` |
| 6 | 64 | `int4 highest episode number` |
| 5 | 32 | `int4 special ep count` |
| 4 | 16 | `int air date` |
| 3 | 8 | `int end date` |
| 2 | 4 | `str url` |
| 1 | 2 | `str picname` |
| 0 | 1 | retired |

**Byte 4**

| Bit | Dec | Field |
|-----|-----|-------|
| 7 | 128 | `int4 rating` |
| 6 | 64 | `int vote count` |
| 5 | 32 | `int4 temp rating` |
| 4 | 16 | `int temp vote count` |
| 3 | 8 | `int4 average review rating` |
| 2 | 4 | `int review count` |
| 1 | 2 | `str award list` |
| 0 | 1 | `bool is 18+ restricted` |

**Byte 5**

| Bit | Dec | Field |
|-----|-----|-------|
| 7 | 128 | retired |
| 6 | 64 | `int ANN id` |
| 5 | 32 | `int allcinema id` |
| 4 | 16 | `str AnimeNfo id` |
| 3 | 8 | `str tag name list` |
| 2 | 4 | `int tag id list` |
| 1 | 2 | `int tag weight list` |
| 0 | 1 | `int date record updated` |

**Byte 6**

| Bit | Dec | Field |
|-----|-----|-------|
| 7 | 128 | `int character id list` |
| 6 | 64 | retired |
| 5 | 32 | retired |
| 4 | 16 | retired |
| 3 | 8 | unused |
| 2 | 4 | unused |
| 1 | 2 | unused |
| 0 | 1 | unused |

**Byte 7**

| Bit | Dec | Field |
|-----|-----|-------|
| 7 | 128 | `int4 specials count` |
| 6 | 64 | `int4 credits count` |
| 5 | 32 | `int4 other count` |
| 4 | 16 | `int4 trailer count` |
| 3 | 8 | `int4 parody count` |
| 2 | 4 | unused |
| 1 | 2 | unused |
| 0 | 1 | unused |

> **`[HIDDEN]` — the retired byte-6 bits are the old producer bits.** The source states elsewhere:
> "'Producer id list' and 'producer name list' match the data returned by the former producer bits
> (but using the revised creator ids)". Those field *names* appear nowhere in the byte tables above,
> so the mapping from retired bits to current fields is not recoverable from this source.
>
> **`[HIDDEN]`** Byte 5's `ANN id`, `allcinema id`, and `AnimeNfo id` are the UDP-side equivalents of
> three entries in the HTTP API's undocumented `<resource type="…">` set — a partial cross-check for
> anyone trying to decode that mapping.

#### `dateflags` bitfield

Used by `ANIME` and `CALENDAR`:

| Bit | Meaning |
|-----|---------|
| 0 | Startdate — unknown day |
| 1 | Startdate — unknown month **and** day |
| 2 | Enddate — unknown day |
| 3 | Enddate — unknown month **and** day |
| 4 | **Air date is in the past / anime has ended** |
| 5 | Startdate — unknown year |
| 6 | Enddate — unknown year |

> Bit 4 is not a precision flag like the others — it is a status flag riding in the same field.

#### Related aid type (byte 1, bit 2)

| Value | Meaning |
|-------|---------|
| 1 | sequel |
| 2 | prequel |
| 11 | same setting |
| 12 | alternative setting |
| 32 | alternative version |
| 41 | music video |
| 42 | character |
| 51 | side story |
| 52 | parent story |
| 61 | summary |
| 62 | full story |
| 100 | other |

> Values `3`–`10`, `13`–`31`, `33`–`40`, `43`–`50`, `53`–`60`, `63`–`99` are unassigned in the source.
> Do not assume the gaps are meaningless — treat unknown values as "other" rather than erroring.

#### Truncation priority

> The character id list is the first data to be truncated if needed. Then: tag list, synonym list,
> short name list. **This applies to `FILE` too.**

Consequence: never treat a short character-id list as authoritative. If you need complete character
data, request it with a narrower mask.

**Example**
```
> ANIME aid=1&amask=b2f0e0fc000000&s=xxxxx
< 230 ANIME
  1|1999-1999|TV Series|Space,Future,…|Seikai no Monshou|星界の紋章|Crest of the Stars||13|13|3|853|3225|756|110|875|11
```

### `ANIMEDESC` — Retrieve anime description

```
ANIMEDESC aid={int4 id}&part={int4 partno}
```

| Code | Payload |
|------|---------|
| `233` | `{int4 current part}|{int4 max parts}|{str description}` |
| `330` | `NO SUCH ANIME` |
| `333` | `NO SUCH DESCRIPTION` |

- Descriptions run to ~5000 chars but a packet caps at 1400 bytes → **multiple calls required**.
- `part` is **zero-based**.
- **No lookup by title — `aid` only.**

At 1 request per 2 s, a full description costs up to ~8 s. Cache it.

### `CALENDAR` — Upcoming titles

```
CALENDAR
```

```
297 CALENDAR
{int aid}|{int startdate}|{int dateflags}
… repeated
397 CALENDAR EMPTY
```

- Returns the **25 most recently aired** anime followed by the **next 25 due**, ordered by start date.
- **Do not depend on exactly 50 results.** The count varies with the season and with the user's
  adult-content preference.
- **Results are filtered by the logged-in user's "show adult" preference.** The same command returns
  different data for different users.
- Takes no parameters beyond `s=`.
- `dateflags` — same bitfield as `ANIME`.

> **`[HIDDEN]`** The reply format line in the source terminates rows with `/n` — a typo for `\n`.
> Rows are newline-separated as everywhere else.

### `CHARACTER` — Character information

```
CHARACTER charid={int characterid}
```

| Code | Payload |
|------|---------|
| `235` | see below |
| `335` | `NO SUCH CHARACTER` |

```
{int charid}|{str character name kanji}|{str character name transcription}|{str pic}|{anime blocks}|{int episode list}|{int last update date}|{int2 type}|{str gender}
```

**Anime blocks** — a composite field. Each block is:

```
{int anime id},{int appearance},{int creatorid},{boolean is_main_seiyuu}
```

repeated, **separated by a single quote (`'`)**, comma-separated within a block.

- If no seiyuu is associated for a given `aid`, `creatorid` and `is_main_seiyuu` are **empty but the
  commas remain** → `2644,0,,`. A naive split on `,` must not drop empties.
- An empty episode list means "undefined" — no episode values have been added (not "appears in no
  episodes").

`appearance` values: `0`=appears in, `1`=cameo appearance in, `2`=main character in,
`3`=secondary cast in.

`type` values — the source explicitly warns **"this is subject to changes so don't rely on this
mapping"**:

| Value | Meaning |
|-------|---------|
| 1 | Character |
| 2 | Mecha |
| 3 | Organisation |
| 4 | Vessel |

`gender` values (same warning):

| Value | Meaning |
|-------|---------|
| `M` | Male |
| `F` | Female |
| `I` | Intersexual |
| `D` | Dimorphic |
| `-` | none / does not apply |
| `?` | unknown |

> **`[HIDDEN]`** The example reply has **only 7 fields**, ending at `last update date` — `type` and
> `gender` are absent:
> ```
> 488|ニコ・ロビン|Nico Robin|14789.jpg|4097,2,1900,1'69,2,…|​|1236938094
> ```
> This is the source's own illustration of why a client must tolerate a **short** reply as well as
> extra trailing fields. **`[INFERRED]`** the example predates the addition of `type`/`gender`.
>
> **`[HIDDEN]`** The episode list is typed `{int episode list}` — singular `int` for a list field.
> The separator is undocumented; **`[INFERRED]`** comma, by analogy with `fids` and the anime-block
> internals. Verify against live data.

### `CREATOR` — Creator information

```
CREATOR creatorid={int creatorid}
```

| Code | Payload |
|------|---------|
| `245` | see below |
| `345` | `NO SUCH CREATOR` |

```
{int creatorid}|{str creator name kanji}|{str creator name transcription}|{int type}|{str pic_name}|{str url_english}|{str url_japanese}|{str wiki_url_english}|{str wiki_url_japanese}|{int last update date}
```

`type` values: `1`=person, `2`=company, `3`=collaboration.

Cross-reference note from the source: **"ANIME AMASK byte6, bit5"** — i.e. one of the retired byte-6
bits was the producer/creator list that this command supersedes.

**Example** (source's own, with `|` visible as the escaped `%7C`):
```
245 CREATOR
718|GAINAX|Gainax|2|10092.png||http://www.gainax.co.jp/%7CGainax%7CGainax%7C1237048093
```

> **`[HIDDEN]`** The example shows `%7C` mid-field where field separators should be. The last four
> logical fields (`wiki_url_english`, `wiki_url_japanese`, `last update date`) appear URL-encoded
> into one run. **`[INFERRED]`** this is a wiki-transcription artifact rather than protocol behavior —
> but it is a warning that `|` inside URL values is a real hazard. Recall the documented inbound
> escape: a content `|` arrives as `/`.

### `EPISODE` — Retrieve episode data

```
EPISODE eid={int eid}
EPISODE aname={str anime name}&epno={int4 episode number}
EPISODE aid={int anime id}&epno={int4 episode number}
```

| Code | Payload |
|------|---------|
| `240` | see below |
| `340` | `NO SUCH EPISODE` |

```
{int eid}|{int aid}|{int4 length}|{int4 rating}|{int votes}|{str epno}|{str eng}|{str romaji}|{str kanji}|{int aired}|{int type}
```

- `length` is in **minutes**.
- `epno` **includes the special-character prefix** (only for specials) and **zero padding** (only for
  normal episodes) — e.g. `01`, `S2`. Contrast the HTTP API, whose `<epno>` text carries the bare
  number and puts the kind in a `type` attribute.
- Special prefix characters: `S`(special), `C`(credits), `T`(trailer), `P`(parody), `O`(other).
- `type` is the numeric form of that prefix: `1`=regular (no prefix), `2`=`S`, `3`=`C`, `4`=`T`,
  `5`=`P`, `6`=`O`.

> **`[HIDDEN]`** The first example reply has **9 fields**, missing `aired` and `type`:
> `1|1|24|400|4|01|Invasion|shinryaku|??` — while the second has all 11. Another concrete case where
> field count varies between records. Index defensively.

### `FILE` — Retrieve file data

```
FILE fid={int4 id}&fmask={hexstr}&amask={hexstr}
FILE size={int8 size}&ed2k={str ed2khash}&fmask={hexstr}&amask={hexstr}
FILE aname={str}&gname={str}&epno={int4}&fmask={hexstr}&amask={hexstr}
FILE aname={str}&gid={int4}&epno={int4}&fmask={hexstr}&amask={hexstr}
FILE aid={int4}&gname={str}&epno={int4}&fmask={hexstr}&amask={hexstr}
FILE aid={int4}&gid={int4}&epno={int4}&fmask={hexstr}&amask={hexstr}
```

| Code | Payload |
|------|---------|
| `220` | `{int4 fid}|…(data list per masks)` |
| `322` | `MULTIPLE FILES FOUND` + `{int4 fid 0}|…|{int4 fid n}` |
| `320` | `NO SUCH FILE` |

Rules:

- Data is returned in table order: **`fmask` fields first, then `amask` fields**, top-to-bottom,
  left-to-right.
- **`fid` is always returned first regardless of the masks.**
- **There is no way to request all fields.** Requesting an `unused` or `reserved` bit → `505`.
- With `aname`/`gname`/`epno`, **only the first matching file is returned**.
- `audio codec` and `audio bitrate` return **multiple `'`-separated values** when the file has
  multiple audio streams — these two fields are variable-arity within a fixed-position layout.
- **Byte 1, bit 2 (`other episodes`)**: each episode id is followed by an **integer percentage** of
  that episode the file covers. Usually 100, but 50/30/other occur. **The percentage says nothing
  about *which* part of the episode is covered.**

#### `state` bitfield

| Bit | Dec | Constant | Meaning |
|-----|-----|----------|---------|
| 1 | 1 | `FILE_CRCOK` | Matched official CRC (green in AniDB) |
| 2 | 2 | `FILE_CRCERR` | Did **not** match official CRC (red in AniDB) |
| 3 | 4 | `FILE_ISV2` | Version 2 |
| 4 | 8 | `FILE_ISV3` | Version 3 |
| 5 | 16 | `FILE_ISV4` | Version 4 |
| 6 | 32 | `FILE_ISV5` | Version 5 |
| 7 | 64 | `FILE_UNC` | Uncensored |
| 8 | 128 | `FILE_CEN` | Censored |

Source examples: `9` = CRCOK+ISV3; `0` = not CRC-checked, version 1; `34` = CRCERR+ISV5; `1` = CRCOK
version 1; `8` = not checked, version 3.

> **`[HIDDEN]`** Version 1 has **no bit** — it is encoded by the *absence* of bits 3–6. Likewise "not
> CRC checked" is the absence of both bits 1 and 2, which is distinct from "CRC failed". Three-state
> logic in a bitfield.

#### `fmask` bit tables

**Byte 1**

| Bit | Dec | Field |
|-----|-----|-------|
| 7 | 128 | unused |
| 6 | 64 | `int4 aid` |
| 5 | 32 | `int4 eid` |
| 4 | 16 | `int4 gid` |
| 3 | 8 | `int4 mylist id` |
| 2 | 4 | `list other episodes` |
| 1 | 2 | `int2 IsDeprecated` |
| 0 | 1 | `int2 state` |

**Byte 2**

| Bit | Dec | Field |
|-----|-----|-------|
| 7 | 128 | `int8 size` |
| 6 | 64 | `str ed2k` |
| 5 | 32 | `str md5` |
| 4 | 16 | `str sha1` |
| 3 | 8 | `str crc32` |
| 2 | 4 | unused |
| 1 | 2 | `video colour depth` |
| 0 | 1 | reserved |

**Byte 3**

| Bit | Dec | Field |
|-----|-----|-------|
| 7 | 128 | `str quality` |
| 6 | 64 | `str source` |
| 5 | 32 | `str audio codec list` |
| 4 | 16 | `int4 audio bitrate list` |
| 3 | 8 | `str video codec` |
| 2 | 4 | `int4 video bitrate` |
| 1 | 2 | `str video resolution` |
| 0 | 1 | `str file type (extension)` |

**Byte 4**

| Bit | Dec | Field |
|-----|-----|-------|
| 7 | 128 | `str dub language` |
| 6 | 64 | `str sub language` |
| 5 | 32 | `int4 length in seconds` |
| 4 | 16 | `str description` |
| 3 | 8 | `int4 aired date` |
| 2 | 4 | unused |
| 1 | 2 | unused |
| 0 | 1 | `str anidb file name` |

**Byte 5** — MyList fields (present in a `FILE` reply, not only `MYLIST`)

| Bit | Dec | Field |
|-----|-----|-------|
| 7 | 128 | `int4 mylist state` |
| 6 | 64 | `int4 mylist filestate` |
| 5 | 32 | `int4 mylist viewed` |
| 4 | 16 | `int4 mylist viewdate` |
| 3 | 8 | `str mylist storage` |
| 2 | 4 | `str mylist source` |
| 1 | 2 | `str mylist other` |
| 0 | 1 | unused |

> **`[HIDDEN]`** `fmask` byte 5 makes `FILE` a **combined file + mylist lookup** — a client that
> needs both can skip the separate `MYLIST` call and halve its packet budget. The source never says
> this outright.
>
> **`[HIDDEN]`** `video colour depth` (byte 2, bit 1) is the only field in either mask table given
> **no type prefix**. Arity/type unknown.
>
> **`[HIDDEN]`** `IsDeprecated` (byte 1, bit 1) is typed `int2` but named as a boolean, and is
> documented nowhere in prose. **`[INFERRED]`** it marks a file superseded by a newer version.

#### `FILE` `amask` bit tables

Note: this is a **different mask from the `ANIME` command's `amask`**, despite the shared name. Only
4 bytes, and the field sets differ.

**Byte 1**

| Bit | Dec | Field |
|-----|-----|-------|
| 7 | 128 | `int4 anime total episodes` |
| 6 | 64 | `int4 highest episode number` |
| 5 | 32 | `str year` |
| 4 | 16 | `str type` |
| 3 | 8 | `str related aid list` |
| 2 | 4 | `str related aid type` |
| 1 | 2 | `str category list` |
| 0 | 1 | reserved |

**Byte 2**

| Bit | Dec | Field |
|-----|-----|-------|
| 7 | 128 | `str romaji name` |
| 6 | 64 | `str kanji name` |
| 5 | 32 | `str english name` |
| 4 | 16 | `str other name` |
| 3 | 8 | `str short name list` |
| 2 | 4 | `str synonym list` |
| 1 | 2 | retired |
| 0 | 1 | retired |

**Byte 3**

| Bit | Dec | Field |
|-----|-----|-------|
| 7 | 128 | `str epno` |
| 6 | 64 | `str ep name` |
| 5 | 32 | `str ep romaji name` |
| 4 | 16 | `str ep kanji name` |
| 3 | 8 | `int4 episode rating` |
| 2 | 4 | `int4 episode vote count` |
| 1 | 2 | unused |
| 0 | 1 | unused |

**Byte 4**

| Bit | Dec | Field |
|-----|-----|-------|
| 7 | 128 | `str group name` |
| 6 | 64 | `str group short name` |
| 5 | 32 | unused |
| 4 | 16 | unused |
| 3 | 8 | unused |
| 2 | 4 | unused |
| 1 | 2 | unused |
| 0 | 1 | `int4 date aid record updated` |

> **`[HIDDEN]`** `FILE` `amask` byte 1 bit 1 is a **live `category list`** field, whereas the `ANIME`
> command's own documentation calls categories deprecated and directs clients to tags. The `FILE`
> mask has **no tag fields at all**. To get tags you must issue a separate `ANIME` call.

**Example**
```
> FILE size=177747474&ed2k=70cd93fd…&fmask=7FF8FEF8&amask=C000F0C0&s=xxxxx
< 220 FILE
312498|4688|69260|4243|0||0|1|177747474|70cd93fd…|b2a7c7d5…|7af9b962…|a200fe73|high|DTV|Vorbis (Ogg Vorbis)|104|H264/AVC|800|704x400|japanese|english'english'english|1560||1175472000|26|26|01|The Wings to the Sky|Sora he no Tsubasa|????|#nanoha-DamagedGoodz|Nanoha-DGz
```

Note `english'english'english` — three subtitle streams in one field, `'`-separated.

### `GROUP` — Retrieve group data

```
GROUP gid={int gid}
GROUP gname={str group name}
```

| Code | Payload |
|------|---------|
| `250` | see below |
| `350` | `NO SUCH GROUP` |

```
{int gid}|{int4 rating}|{int votes}|{int4 acount}|{int fcount}|{str name}|{str short}|{str irc channel}|{str irc server}|{str url}|{str picname}|{int4 foundeddate}|{int4 disbandeddate}|{int2 dateflags}|{int4 lastreleasedate}|{int4 lastactivitydate}|{list grouprelations}
```

- Requires login.
- `gname` is an exact match on either name or short name; both are unique, so a name hit is the only hit.
- **`releasedate` ≠ `activitydate`**: `lastreleasedate` is when the group actually released a file;
  `lastactivitydate` is when a file was **added to AniDB**. Last release can be far older than last
  activity.

`dateflags` here is a **different bitfield from `ANIME`'s** — same bit positions, different subjects,
and **bit 4 is absent**:

| Bit | Meaning |
|-----|---------|
| 0 | Foundeddate — unknown day |
| 1 | Foundeddate — unknown month and day |
| 2 | Disbandeddate — unknown day |
| 3 | Disbandeddate — unknown month and day |
| 5 | Foundeddate — unknown year |
| 6 | Disbandeddate — unknown year |

> **`[HIDDEN]`** Bit 4 is defined for `ANIME`/`CALENDAR` ("air date in the past") but simply **not
> listed** for `GROUP`. Whether it is unused or undocumented here is unknown — do not reuse one
> `dateflags` parser across both commands without accounting for it.

`grouprelations` is a list of `'`-separated pairs, each `{int4 othergroupid},{int2 relationtype}`:

| Value | Meaning |
|-------|---------|
| 1 | Participant in |
| 2 | Parent of |
| 4 | Merged from |
| 5 | Now known as |
| 6 | Other |

> **`[HIDDEN]`** `3` is **missing** from the relation enum — the source jumps 2 → 4. Presumably a
> retired relation type. Handle unknown values gracefully.

**Example**
```
> GROUP gid=7091&s=bunny
< 250 GROUP
7091|832|1445|43|566|Frostii|Frostii|#frostii|irc.rizon.net|http://frostii.com|15844.jpg|1228089600|0|1|1301875200|1304222640|7255,1'3097,4'748,4'8106,1'8159,2'8402,1'8696,1'9022,1
```

Note `disbandeddate=0` for an active group — `0` means NULL here as it does for ids.

### `GROUPSTATUS` — Completed episodes per group

```
GROUPSTATUS aid={int animeid}[&state={int completion_state}]
```

| Code | Payload |
|------|---------|
| `225` | rows, one per group |
| `325` | `NO SUCH GROUPS FOUND` |
| `330` | `NO SUCH ANIME` |

```
{int group id}|{str group name}|{int completion state}|{int last episode number}|{int rating}|{int votes}|{str episode range}\n
… repeated
```

- **Groups are filtered by the languages selected in the user's profile** — user-dependent results.
- Without `state`, groups with completion `ongoing`, `finished`, or `complete` are returned.
- Ordered by **descending episode count**.
- **If more groups exist than fit in a packet, the extras are silently discarded.** No indication is
  given. `GROUPSTATUS` on a popular anime is inherently lossy.

`state` values (as documented, by name):

`ongoing`, `stalled`, `complete`, `dropped`, `finished`, `specials only`

> **`[HIDDEN]`** The parameter is typed `{int completion_state}` but the documented values are
> **strings**. The example reply shows a completion state of `3` in the data column. The numeric
> mapping is **not given anywhere in the source** — the six names are listed without numbers.
> **`[INFERRED]`** the reply's `3` corresponds to one of them (the example's groups are all
> `1-25` of a 25-episode anime, so plausibly `complete`), but the full name→int mapping is
> unrecoverable from this source. Do not guess it.

**Example**
```
> GROUPSTATUS aid=8692&s=vLl1N
< 225 GROUPSTATUS
7407|Coalgirls|3|25|839|12|1-25
9863|Hadena Subs|3|25|374|1|1-25
11951|ChaosBlades|3|25|0|0|1-25
```

### `UPDATED` — Recently updated anime ids

```
UPDATED entity=1&[age={int4 id}|time={int4 date}]
```

| Code | Payload |
|------|---------|
| `243` | `{int entity}|{int total count}|{int last update date}|{list aid}` |
| `343` | `NO UPDATES` |

- `entity` is **always 1**.
- Supply **either** `age` **or** `time`, never both.
  - `age` in **days** (`age=30` → changed in the last 30 days)
  - `time` in **Unix time** (`time=1264982400` → changed since 2010-02-01)
- **Maximum 200 values returned.**
- `count` is the **total matching**, not the number returned — **if `count > 200` you have not
  retrieved everything.** The source offers no pagination parameter.
- `last update date` is the AniDB update time of the last aid in the list — **`[INFERRED]`** this is
  the intended cursor: re-issue with `time` set to it to walk forward through a backlog >200.
- Each aid appears **once**; its age reflects the most recent change.
- Ordered oldest-change to newest.

An anime counts as updated when:

- the anime record itself changes
- a **main or official** title is added/edited/deleted (**not** short or synonym)
- an episode is **added or deleted** — **NOT edited**
- an episode title is added/edited/deleted
- an anime relation is added or deleted

> **`[HIDDEN]` — the incremental-sync trap.** Episode *edits* (e.g. correcting a runtime or an air
> date) do **not** mark the anime as updated, and short/synonym title changes do not either. A
> crawler that trusts `UPDATED` alone will silently drift on those fields. Periodic full re-fetch is
> required regardless.
>
> **`[HIDDEN]`** The `age` parameter is typed `{int4 id}` in the command string — a copy-paste error;
> prose makes clear it is a day count.

---

## MyList Commands

### `MYLIST` — Retrieve MyList data

```
MYLIST lid={int4 lid}
MYLIST fid={int4 fid}
MYLIST size={int4 size}&ed2k={str ed2khash}
MYLIST aname={str}[&gname={str}&epno={int4}]
MYLIST aname={str}[&gid={int4}&epno={int4}]
MYLIST aid={int4}[&gname={str}&epno={int4}]
MYLIST aid={int4}[&gid={int4}&epno={int4}]
```

| Code | Payload |
|------|---------|
| `221` | single entry (see below) |
| `312` | `MULTIPLE MYLIST ENTRIES` — **a completely different, variable-width shape** |
| `321` | `NO SUCH ENTRY` |

**`221` single entry:**
```
{int4 lid}|{int4 fid}|{int4 eid}|{int4 aid}|{int4 gid}|{int4 date}|{int2 state}|{int4 viewdate}|{str storage}|{str source}|{str other}|{int2 filestate}
```

**`312` multiple entries — variable field count:**
```
{str anime title}|{int episodes}|{str eps with state unknown}|{str eps with state on hhd}|{str eps with state on cd}|{str eps with state deleted}|{str watched eps}|{str group 1 short name}|{str eps for group 1}|…|{str group N short name}|{str eps for group N}
```

> **`[HIDDEN]`** After the first 7 fixed fields, `312` emits **(group name, episode range) pairs
> repeating to the end of the line**. The field count depends on how many groups released the anime.
> Parse the head positionally, then consume the tail two at a time.
>
> **`[HIDDEN]`** `size` is typed `int4` here but `int8` in the `FILE` command, for the same value.
> Use 64-bit.
>
> **`[HIDDEN]`** The source's own `MYLIST` example shows the command returning **`322 MULTIPLE FILES
> FOUND`** — the `FILE` command's code — while the documented multi-result code for `MYLIST` is
> `312 MULTIPLE MYLIST ENTRIES`. The example's payload shape matches the `312` layout. Either the
> example's code is a transcription error or `MYLIST` can genuinely return `322`. **Handle both.**

Episode ranges use the form `1-12,14-16,T1` — comma-separated runs, with special-type prefixes.

#### MyList `state` values

| Value | Meaning |
|-------|---------|
| 0 | unknown — state unknown, or user declines to say |
| 1 | internal storage — on HDD, not shared |
| 2 | external storage — on CD/DVD/… |
| 3 | deleted — deleted or otherwise unavailable (e.g. reencoded) |
| 4 | remote storage — on NAS/cloud/… |

> When adding files after hashing, a client **should default to `1` (internal storage)** unless the
> user explicitly chooses otherwise.

#### MyList `filestate` values (normal, i.e. non-generic files)

| Value | Meaning |
|-------|---------|
| 0 | normal / original |
| 1 | corrupted version / invalid CRC |
| 2 | self edited |
| 10 | self ripped |
| 11 | on DVD |
| 12 | on VHS |
| 13 | on TV |
| 14 | in theaters |
| 15 | streamed |
| 100 | other |

> Gaps at `3`–`9` and `16`–`99` are unassigned. **`[HIDDEN]`** the "(for normal files, i.e. not
> generic)" qualifier implies **generic files use a different filestate vocabulary**, which the
> source never provides.

### `MYLISTADD` — Add or edit a MyList entry

The command is one **file-identifying block** plus optional fields.

**Single file (`fileinfo` block):**
```
MYLISTADD fid={int4 fid}
MYLISTADD size={int4 size}&ed2k={str ed2khash}
MYLISTADD lid={int4 lid}&edit=1          (edit only)
```

**One or more files (`animeinfo` + `groupinfo` + `episodeinfo`):**
```
MYLISTADD aid={int4}&gid={int}&epno={int4}
MYLISTADD aid={int4}&gname={str}&epno={int4}
MYLISTADD aid={int4}&generic=1&epno={int4}
MYLISTADD aname={str}&gid={int}&epno={int4}
MYLISTADD aname={str}&gname={str}&epno={int4}
MYLISTADD aname={str}&generic=1&epno={int4}
```

**Optional fields (append to any of the above):**

| Parameter | Type | Notes |
|-----------|------|-------|
| `state` | int2 | MyList state (see above) |
| `viewed` | boolean | `0` unwatched, `1` watched |
| `viewdate` | int4 | Unix time. **Defaults to now if omitted. Disregarded when `viewed=0`.** |
| `source` | str | |
| `storage` | str | |
| `other` | str | **The only field that may contain newlines** — encode them as `<br />` |
| `edit` | `1` | Edit instead of create |

**Replies**

| Code | Condition | Payload |
|------|-----------|---------|
| `210` | added by `fid` / `size`+`ed2k` | `{int4 mylist id of new entry}` |
| `210` | added by `aname` / `aid` | `{int4 number of entries added}` |
| `310` | `FILE ALREADY IN MYLIST` | The existing record, **in `MYLIST` `221` format** |
| `311` | `MYLIST ENTRY EDITED` | `{int4 number of entries edited}` |
| `322` | `MULTIPLE FILES FOUND` | `{int4 fid 1}|…|{int4 fid n}` |
| `320` / `330` / `350` | no such file / anime / group | |
| `411` | `NO SUCH MYLIST ENTRY` | |

> **`[HIDDEN]` — `210` is polymorphic.** Adding by file id returns a **`lid`**; adding by anime name
> or id returns a **count**. Same code, different semantics, and both are bare integers — you cannot
> tell them apart from the reply alone. **Disambiguate by which parameters you sent.**

Other documented behavior:

- `epno=0` means **all episodes** (the default). **Negative means "up to"** — `-12` = episodes 1–12.
- On edit, **fields you don't supply retain their existing values** — they are not reset to defaults.
- `group` is optional **only when `edit=1`**; you cannot add every file for an anime in one go.
- **If you know the `lid`, use the `lid` form** — it puts less load on the server.
- On `310`, the current record is returned so the client can reconcile without a second lookup.

### `MYLISTDEL` — Remove from MyList

```
MYLISTDEL lid={int4 lid}
MYLISTDEL fid={int4 fid}
MYLISTDEL size={int4 size}&ed2k={str ed2k hash}
MYLISTDEL aname={str}[&gname={str}&epno={int4}]
MYLISTDEL aname={str}[&gid={int4}&epno={int4}]
MYLISTDEL aid={int4}[&gname={str}&epno={int4}]
MYLISTDEL aid={int4}[&gid={int4}&epno={int4}]
```

| Code | Payload |
|------|---------|
| `211` | `MYLIST ENTRY DELETED` + `{int4 number of entries}` |
| `411` | `NO SUCH MYLIST ENTRY` |

> `group` is optional, and **the command deletes ALL entries found**. Combined with the optional
> group and optional `epno`, `MYLISTDEL aid=…` alone will delete the user's entire MyList for that
> anime. There is no dry-run and no undo.

### `MYLISTSTATS` — MyList statistics

```
MYLISTSTATS
```
```
222 MYLIST STATS
{animes}|{eps}|{files}|{size of files}|{added animes}|{added eps}|{added files}|{added groups}|{leech %}|{glory %}|{viewed % of db}|{mylist % of db}|{viewed % of mylist}|{number of viewed eps}|{votes}|{reviews}|{viewed length in minutes}
```

All 17 fields are `int`. **`[HIDDEN]`** the percentage fields are integers, so their scale (0–100 vs
0–1000 vs basis points) is undocumented. The source gives no example reply for this command.

### `VOTE` — Vote on anime / episode / group

```
VOTE type={int2 type}&id={int4 id}[&value={int4 vote value}&epno={int4|str episode number}]
VOTE type={int2 type}&name={string name}[&value={int4 vote value}&epno={int4|str episode number}]
```

| Code | Payload |
|------|---------|
| `260` | `VOTED` + `{str entity name}|{vote value}|{vote type}|{entity id}` |
| `261` | `VOTE FOUND` + same shape |
| `262` | `VOTE UPDATED` + `{str entity name}|{old vote value}|{vote type}|{entity id}` |
| `263` | `VOTE REVOKED` + `{str entity name}|{revoked vote value}|{vote type}|{entity id}` |
| `360` | `NO SUCH VOTE` + `{str entity name}|0|{vote type}|{entity id}` |
| `361` | `INVALID VOTE TYPE` |
| `362` | `INVALID VOTE VALUE` |
| `363` | `PERMVOTE NOT ALLOWED` + `{str aname}|{vote value}|{type}|{entity id}` |
| `364` | `ALREADY PERMVOTED` + `{str entity name}|{existing vote value}|{vote type}|{entity id}` |

`type` values: `1`=anime, `2`=anime tmpvote, `3`=group, `6`=episode.

> **`[HIDDEN]`** `4` and `5` are missing from the vote-type enum.

`value` semantics — **one parameter, four behaviors**:

| Value | Effect |
|-------|--------|
| negative | **Revoke** the vote |
| `0` (default) | **Retrieve** the existing vote — a read, not a write |
| `100`–`1000` | **Cast/update** the vote (`950` = 9.50) |
| anything else | Illegal |

> **`[HIDDEN]`** Omitting `value` performs a **read**, because the default is `0`. This makes `VOTE`
> a dual-purpose read/write command; a client that omits `value` expecting a no-op gets `261 VOTE
> FOUND` or `360 NO SUCH VOTE` instead.

Other behavior:
- Votes are **updated automatically, no confirmation asked**.
- **Temp-voting is not possible when a permanent vote exists** (`363`/`364`).
- To vote on an episode: either `type=1` with an `epno`, or `type=6` with a known `eid`.
- `epno` accepts a string, so special episodes (`S2`) can be voted on.

Source examples show `VOTE type=1&id=5101&value=950&epno=S2` returning the **episode's** name and id
(`91981`) — i.e. supplying `epno` redirects an anime-typed vote to the episode entity.

### `RANDOM` / `RANDOMANIME` — Random anime

> **`[HIDDEN]` — the section heading says `RANDOM` but the command string is `RANDOMANIME`.** Use
> `RANDOMANIME`. (The return-code table also lists `RANDOM_ANIME = 232`, which the command
> documentation does not mention — it reuses `230 ANIME`.)

```
RANDOMANIME type={int4 type}
```

Reply: `230 ANIME …` — identical shape to the `ANIME` command.

| `type` | Pool |
|--------|------|
| 0 | from db |
| 1 | watched |
| 2 | unwatched |
| 3 | all mylist |

> **`[INFERRED]`** The command documentation does not mention `amask`, though the reply is an `ANIME`
> reply. Whether `RANDOMANIME` accepts `amask` is undocumented.

---

## Misc Commands

### `MYLISTEXPORT` — Queue a MyList export

```
MYLISTEXPORT [template={str template_name}|cancel=1]
```

| Code | Meaning |
|------|---------|
| `217` | `EXPORT QUEUED` |
| `218` | `EXPORT CANCELLED` |
| `317` | `EXPORT NO SUCH TEMPLATE` |
| `318` | `EXPORT ALREADY IN QUEUE` |
| `319` | `EXPORT NO EXPORT QUEUED OR IS PROCESSING` |

- Exports run only during **low server load** — **may take up to 24 hours**.
- **Only one export may be queued at a time**, per user, across both UDP and the website.
- `template_name` must match a template on the MyList Export page.
- `cancel=1` cancels a pending request queued by **either** channel.
- The client is notified by an **AniDB message** when the export is ready — subscribe to message
  notifications to catch it.

> This is the sanctioned bulk-extraction path for a user's own data. Prefer it over crawling
> `MYLIST` file-by-file.

### `PING`

```
PING [nat=1]
```
```
300 PONG
{int4 port}   (only when nat=1)
```

- **Does not require a session** — usable before login.
- Use to check API availability, and to keep NAT mappings alive.
- `nat=1` reveals whether the router changed the outgoing port.

### `VERSION`

```
VERSION
```
```
998 VERSION
{str server version}
```
Does not require a session.

### `UPTIME`

```
UPTIME
```
```
208 UPTIME
{int4 udpserver uptime in milliseconds}
```

> **The preferred way to check that the session is still OK.** Requires a session (unlike `PING`) —
> that is precisely what makes it a session probe.

### `ENCODING`

```
ENCODING name={str encoding name}
```

| Code | Meaning |
|------|---------|
| `219` | `ENCODING CHANGED` |
| `519` | `ENCODING NOT SUPPORTED` |

- Does not require a session.
- Default ASCII; **resets to default on logout**.
- **The preferred way to set encoding is `AUTH`'s `enc` parameter.** `ENCODING` exists mainly for
  testing which encodings are supported — `AUTH` ignores an unsupported `enc` silently.
- Supported encodings are the Java 1.5 charset names.

### `SENDMSG`

```
SENDMSG to={str username}&title={str title}&body={str body}
```

| Code | Meaning |
|------|---------|
| `294` | `SENDMSG SUCCESSFUL` |
| `394` | `NO SUCH USER` |
| `501` | `LOGIN FIRST` |

Hard limits: **`title` ≤ 50 chars, `body` ≤ 900 chars.**

### `USER` — Resolve user id / name

```
USER [user={str user name}|uid={int user id}]
```

| Code | Payload |
|------|---------|
| `295` | `USER` + `{int4 uid}|{str username}` |
| `394` | `NO SUCH USER` |

Resolves in either direction.

### `STATS` *(disabled)*

```
STATS
```
```
206 STATS
{int4 animes}|{int4 eps}|{int4 files}|{int4 groups}|{int4 users}|{int8 total file size in bytes}|{int4 open creqs}
```

> Marked **`[disabled]`** in the source. Documented for completeness; do not build against it.

### `TOP` *(disabled)*

```
TOP
```
```
207 TOP
{str longest mylist}|{int count}
{str largest mylist}|{int count}
{str most lame files}|{int count}
{str most indep. user}|{int count}
{str biggest leecher}|{int count}
{str most anime added}|{int count}
{str most eps added}|{int count}
{str most files added}|{int count}
{str most groups added}|{int count}
{str most votes}|{int count}
{str most reviews}|{int count}
```

All strings are usernames. The profile setting "Hide myself in IRC stats" applies here too.

> Marked **`[disabled]`** in the source.

---

## Return Codes

> The names below are the actual code strings returned by the server, but **clients are encouraged to
> work with the numeric codes instead.**

### 2xx — Success

| Code | Name |
|------|------|
| 200 | `LOGIN_ACCEPTED` |
| 201 | `LOGIN_ACCEPTED_NEW_VERSION` |
| 203 | `LOGGED_OUT` |
| 205 | `RESOURCE` |
| 206 | `STATS` |
| 207 | `TOP` |
| 208 | `UPTIME` |
| 209 | `ENCRYPTION_ENABLED` |
| 210 | `MYLIST_ENTRY_ADDED` |
| 211 | `MYLIST_ENTRY_DELETED` |
| 214 | `ADDED_FILE` |
| 215 | `ADDED_STREAM` |
| 217 | `EXPORT_QUEUED` |
| 218 | `EXPORT_CANCELLED` |
| 219 | `ENCODING_CHANGED` |
| 220 | `FILE` |
| 221 | `MYLIST` |
| 222 | `MYLIST_STATS` |
| 223 | `WISHLIST` |
| 224 | `NOTIFICATION` |
| 225 | `GROUP_STATUS` |
| 226 | `WISHLIST_ENTRY_ADDED` |
| 227 | `WISHLIST_ENTRY_DELETED` |
| 228 | `WISHLIST_ENTRY_UPDATED` |
| 229 | `MULTIPLE_WISHLIST` |
| 230 | `ANIME` |
| 231 | `ANIME_BEST_MATCH` |
| 232 | `RANDOM_ANIME` |
| 233 | `ANIME_DESCRIPTION` |
| 234 | `REVIEW` |
| 235 | `CHARACTER` |
| 236 | `SONG` |
| 237 | `ANIMETAG` |
| 238 | `CHARACTERTAG` |
| 240 | `EPISODE` |
| 243 | `UPDATED` |
| 244 | `TITLE` |
| 245 | `CREATOR` |
| 246 | `NOTIFICATION_ENTRY_ADDED` |
| 247 | `NOTIFICATION_ENTRY_DELETED` |
| 248 | `NOTIFICATION_ENTRY_UPDATE` |
| 249 | `MULTIPLE_NOTIFICATION` |
| 250 | `GROUP` |
| 251 | `CATEGORY` |
| 253 | `BUDDY_LIST` |
| 254 | `BUDDY_STATE` |
| 255 | `BUDDY_ADDED` |
| 256 | `BUDDY_DELETED` |
| 257 | `BUDDY_ACCEPTED` |
| 258 | `BUDDY_DENIED` |
| 260 | `VOTED` |
| 261 | `VOTE_FOUND` |
| 262 | `VOTE_UPDATED` |
| 263 | `VOTE_REVOKED` |
| 265 | `HOT_ANIME` |
| 266 | `RANDOM_RECOMMENDATION` |
| 267 | `RANDOM_SIMILAR` |
| 270 | `NOTIFICATION_ENABLED` |
| 281 | `NOTIFYACK_SUCCESSFUL_MESSAGE` |
| 282 | `NOTIFYACK_SUCCESSFUL_NOTIFICATION` |
| 290 | `NOTIFICATION_STATE` |
| 291 | `NOTIFYLIST` |
| 292 | `NOTIFYGET_MESSAGE` |
| 293 | `NOTIFYGET_NOTIFY` |
| 294 | `SENDMESSAGE_SUCCESSFUL` |
| 295 | `USER_ID` |
| 297 | `CALENDAR` |

### 3xx — Negative / not-found

| Code | Name |
|------|------|
| 300 | `PONG` |
| 301 | `AUTHPONG` |
| 305 | `NO_SUCH_RESOURCE` |
| 309 | `API_PASSWORD_NOT_DEFINED` |
| 310 | `FILE_ALREADY_IN_MYLIST` |
| 311 | `MYLIST_ENTRY_EDITED` |
| 312 | `MULTIPLE_MYLIST_ENTRIES` |
| 313 | `WATCHED` |
| 314 | `SIZE_HASH_EXISTS` |
| 315 | `INVALID_DATA` |
| 316 | `STREAMNOID_USED` |
| 317 | `EXPORT_NO_SUCH_TEMPLATE` |
| 318 | `EXPORT_ALREADY_IN_QUEUE` |
| 319 | `EXPORT_NO_EXPORT_QUEUED_OR_IS_PROCESSING` |
| 320 | `NO_SUCH_FILE` |
| 321 | `NO_SUCH_ENTRY` |
| 322 | `MULTIPLE_FILES_FOUND` |
| 323 | `NO_SUCH_WISHLIST` |
| 324 | `NO_SUCH_NOTIFICATION` |
| 325 | `NO_GROUPS_FOUND` |
| 330 | `NO_SUCH_ANIME` |
| 333 | `NO_SUCH_DESCRIPTION` |
| 334 | `NO_SUCH_REVIEW` |
| 335 | `NO_SUCH_CHARACTER` |
| 336 | `NO_SUCH_SONG` |
| 337 | `NO_SUCH_ANIMETAG` |
| 338 | `NO_SUCH_CHARACTERTAG` |
| 340 | `NO_SUCH_EPISODE` |
| 343 | `NO_SUCH_UPDATES` |
| 344 | `NO_SUCH_TITLES` |
| 345 | `NO_SUCH_CREATOR` |
| 350 | `NO_SUCH_GROUP` |
| 351 | `NO_SUCH_CATEGORY` |
| 355 | `BUDDY_ALREADY_ADDED` |
| 356 | `NO_SUCH_BUDDY` |
| 357 | `BUDDY_ALREADY_ACCEPTED` |
| 358 | `BUDDY_ALREADY_DENIED` |
| 360 | `NO_SUCH_VOTE` |
| 361 | `INVALID_VOTE_TYPE` |
| 362 | `INVALID_VOTE_VALUE` |
| 363 | `PERMVOTE_NOT_ALLOWED` |
| 364 | `ALREADY_PERMVOTED` |
| 365 | `HOT_ANIME_EMPTY` |
| 366 | `RANDOM_RECOMMENDATION_EMPTY` |
| 367 | `RANDOM_SIMILAR_EMPTY` |
| 370 | `NOTIFICATION_DISABLED` |
| 381 | `NO_SUCH_ENTRY_MESSAGE` |
| 382 | `NO_SUCH_ENTRY_NOTIFICATION` |
| 392 | `NO_SUCH_MESSAGE` |
| 393 | `NO_SUCH_NOTIFY` |
| 394 | `NO_SUCH_USER` |
| 397 | `CALENDAR_EMPTY` |
| 399 | `NO_CHANGES` |

### 4xx — MyList / session

| Code | Name |
|------|------|
| 403 | `NOT_LOGGED_IN` |
| 410 | `NO_SUCH_MYLIST_FILE` |
| 411 | `NO_SUCH_MYLIST_ENTRY` |
| 412 | `MYLIST_UNAVAILABLE` |

### 5xx — Client error / auth

| Code | Name |
|------|------|
| 500 | `LOGIN_FAILED` |
| 501 | `LOGIN_FIRST` |
| 502 | `ACCESS_DENIED` |
| 503 | `CLIENT_VERSION_OUTDATED` |
| 504 | `CLIENT_BANNED` |
| 505 | `ILLEGAL_INPUT_OR_ACCESS_DENIED` |
| 506 | `INVALID_SESSION` |
| 509 | `NO_SUCH_ENCRYPTION_TYPE` |
| 519 | `ENCODING_NOT_SUPPORTED` |
| 555 | `BANNED` |
| 598 | `UNKNOWN_COMMAND` |

### 6xx — Server error

| Code | Name |
|------|------|
| 600 | `INTERNAL_SERVER_ERROR` |
| 601 | `ANIDB_OUT_OF_SERVICE` |
| 602 | `SERVER_BUSY` |
| 603 | `NO_DATA` |
| 604 | `TIMEOUT - DELAY AND RESUBMIT` |
| 666 | `API_VIOLATION` |

### 7xx — Push

| Code | Name |
|------|------|
| 701 | `PUSHACK_CONFIRMED` |
| 702 | `NO_SUCH_PACKET_PENDING` |

Notification packets occupy `720`–`799`; documented formats exist for `720`, `753`, `794`, `799`.

### 9xx

| Code | Name |
|------|------|
| 998 | `VERSION` |

---

## `[HIDDEN]` — Codes with no documented command

These appear in the return-code table but **no command in the source produces them**. They indicate
either removed features, undocumented commands, or planned functionality. Listing them so a client's
code-handling switch is exhaustive rather than crashing on an unknown code.

| Code(s) | Name | Note |
|---------|------|------|
| `205` / `305` | `RESOURCE` / `NO_SUCH_RESOURCE` | No `RESOURCE` command documented |
| `214`, `215` | `ADDED_FILE`, `ADDED_STREAM` | No file/stream-adding command documented |
| `223`, `226`–`229`, `323` | `WISHLIST*`, `NO_SUCH_WISHLIST` | **An entire wishlist feature** with add/delete/update/multiple codes — no commands documented |
| `224` | `NOTIFICATION` | Distinct from `290 NOTIFICATION_STATE`; no command emits it |
| `231` | `ANIME_BEST_MATCH` | Suggests a fuzzy `ANIME` lookup mode; `ANIME aname=` is documented as **exact match only** |
| `232` | `RANDOM_ANIME` | `RANDOMANIME` is documented as returning `230` instead |
| `234` / `334` | `REVIEW` / `NO_SUCH_REVIEW` | No `REVIEW` command; review *counts* are exposed via `ANIME` amask |
| `236` / `336` | `SONG` / `NO_SUCH_SONG` | No `SONG` command |
| `237`, `238` / `337`, `338` | `ANIMETAG`, `CHARACTERTAG` / their `NO_SUCH_*` | No tag-lookup commands; tags come only as amask lists on `ANIME` |
| `244` / `344` | `TITLE` / `NO_SUCH_TITLES` | No `TITLE` command — titles come from the HTTP data dump |
| `249` | `MULTIPLE_NOTIFICATION` | |
| `251` / `351` | `CATEGORY` / `NO_SUCH_CATEGORY` | Categories are deprecated in favour of tags |
| `265`, `266`, `267` / `365`, `366`, `367` | `HOT_ANIME`, `RANDOM_RECOMMENDATION`, `RANDOM_SIMILAR` + empties | **UDP equivalents of the HTTP API's `hotanime` / `randomrecommendation` / `randomsimilar` endpoints** — codes reserved, commands undocumented |
| `271`, `272` | — | Named in prose (as `NOTIFICATION` codes carrying `relid`/`mid` and as codes that append data to the return line) but **absent from the code table** |
| `301` | `AUTHPONG` | An authenticated `PING`? No command documented |
| `313` | `WATCHED` | |
| `314` | `SIZE_HASH_EXISTS` | |
| `315` | `INVALID_DATA` | |
| `316` | `STREAMNOID_USED` | Pairs with `215 ADDED_STREAM` |
| `410` | `NO_SUCH_MYLIST_FILE` | Distinct from `411 NO_SUCH_MYLIST_ENTRY`; no command documents it |
| `412` | `MYLIST_UNAVAILABLE` | |
| `603` | `NO_DATA` | Not in the universal-6xx list given in the Basics section |
| `666` | `API_VIOLATION` | Not in the universal-6xx list either. **`[INFERRED]`** a ban-adjacent code — treat like `555`: stop, do not retry. |

---

## AI Coding Rules

1. **Rate limit hard: one packet per 2 seconds minimum, one per 4 seconds sustained.** Dropped packets
   still count. A retry loop against a silent server is how clients get permanently banned.
2. **Reuse one fixed local UDP port > 1024**, in production *and* in tests. Port churn from one IP is
   a ban trigger.
3. **Never hardcode `api.anidb.net:9000`** — read from config.
4. **Cache everything.** The source warns that repeat lookups may become a bannable offence.
5. **Parse positionally, tolerate extra trailing fields, and tolerate missing ones.** The source's own
   examples for `CHARACTER` and `EPISODE` are shorter than their documented formats.
6. **Never assert on field count.**
7. **Unescape in the right order:** split on separators first (`|`, then `'` or `,` inside composite
   fields), then decode `` ` ``→`'`, `/`→`|`, `<br />`→newline.
8. **Treat `0` in an id field as NULL**, not as a valid id.
9. **Use `int64` for file sizes** regardless of the `int4` annotation on `MYLIST`.
10. **Handle `501` by silently re-`AUTH`ing and resending**; handle `502` by aborting the action.
11. **Back off exponentially (30 s → … → 2–4 h) when `AUTH` gets no reply at all.**
12. **`LOGOUT` when idle > 30 min; stay logged in when another command is likely within 20 min.**
    Never do AUTH/command/LOGOUT per command.
13. **Selecting an `unused`/`reserved` mask bit returns `505`.** Build masks from the tables, never
    from `0xFF` fills.
14. **`ANIME`'s `amask` and `FILE`'s `amask` are different masks.** Do not share a builder.
15. **Mask order determines field order.** Decode by iterating the same bit order you encoded.
16. **Expect silent truncation.** For `ANIME`/`FILE`, character ids drop first, then tags, synonyms,
    short names. For `GROUPSTATUS`, whole groups are dropped with no marker.
17. **`UPDATED` caps at 200 and `count` tells you if you missed some** — page forward using
    `last update date` as a `time` cursor.
18. **`UPDATED` does not fire on episode edits or short/synonym title changes.** Schedule periodic
    full re-fetches anyway.
19. **`CALENDAR` and `GROUPSTATUS` are filtered by the logged-in user's profile settings** (adult
    content, languages). Results are not user-independent — do not cache them as global data.
20. **`VOTE` with no `value` is a read, not a no-op.**
21. **`MYLISTDEL` with only `aid` deletes every matching entry.** No dry-run, no undo.
22. **In PUSH mode, `PUSHACK` uses the packet `nid`; `NOTIFYGET`/`NOTIFYACK` use the body's `relid`.**
    Do not cross them.
23. **Unacknowledged push notifications are resent 3× and then log you out.**
24. **Untagged 6xx replies exist** — a tag-keyed correlator needs a fallback path or it will leak
    pending requests.
25. **Never let a tag begin with two zero bytes** when `comp=1` is in use.
26. **Ratings are 0–1000 integers on this API** (`950` = 9.50) — unlike the HTTP API's `request=anime`,
    which returns 0–10 decimals.
27. **`AUTH` uses the website password; `ENCRYPT` uses the profile's separate UDP API Key.** They are
    different secrets.
28. **`dateflags` bit 4 means "ended", not a precision flag** — and `GROUP`'s `dateflags` omits bit 4
    entirely.
29. **Handle unknown enum values as "other"** rather than erroring. Several enums have documented gaps
    (`related aid type`, group `relationtype` missing `3`, vote `type` missing `4`/`5`).
30. **Have a default branch for every return code**, including the ones with no documented command.

---

## Source Boundary

This reference is derived solely from the supplied AniDB wiki export (`anidb/3.md`, UDP API spec
version `0.03.730`, dated **2015-03-25**, plus `anidb/1.md`).

Every command, parameter, mask bit, enum value, and return code above is either (a) stated in that
source, or (b) observed in the source's own examples and marked **`[INFERRED]`** / **`[HIDDEN]`**.
Nothing has been added from third-party clients or model memory.

**The spec is dated 2015.** Fields have almost certainly been appended since — which is precisely why
the source mandates tolerating unknown trailing fields.

Areas the supplied source does **not** resolve, and where the live AniDB documentation must be
consulted before implementation:

- The numeric mapping for `GROUPSTATUS` completion states (names given, integers not)
- Which retired `amask` bits corresponded to the former producer fields
- Commands for the reserved codes: wishlist (`223`/`226`–`229`/`323`), `REVIEW`, `SONG`,
  `ANIMETAG`/`CHARACTERTAG`, `TITLE`, `RESOURCE`, `HOT_ANIME`/`RANDOM_RECOMMENDATION`/`RANDOM_SIMILAR`
- The packet formats for codes `271` and `272`
- The `filestate` vocabulary for **generic** (non-normal) files
- The scale of the percentage fields in `MYLISTSTATS`
- The separator used inside `CHARACTER`'s episode list
- The type and arity of `video colour depth` (`fmask` byte 2, bit 1)
- Whether `RANDOMANIME` accepts an `amask`
- Whether `nat=1` and `imgserver=1` may be combined on `AUTH`
