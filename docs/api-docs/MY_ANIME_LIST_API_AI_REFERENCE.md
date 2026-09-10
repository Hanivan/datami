# MyAnimeList API — AI Implementation Reference

> Source: the supplied MyAnimeList API (beta ver.) documentation.
> This file is a normalized, AI-oriented reference derived from that source.
> **Do not invent undocumented endpoints, parameters, fields, enum values, or behavior.**
> When the source is ambiguous, preserve the source behavior and flag the ambiguity rather than guessing.

## Purpose

Use this document as the implementation context for an AI coding agent building a MyAnimeList API client, crawler/ingestion layer, or integration service.

The API base URL documented by the source is:

`https://api.myanimelist.net/v2`

The supplied documentation describes these resource groups:

- `anime`
- `user animelist`
- `forum`
- `manga`
- `user mangalist`
- `user`

## Global API Rules

### Versioning

The API version is part of the request URI. The documented version is `v2`.

A new API version is introduced for backwards-incompatible changes such as:

- removing endpoints
- removing response fields
- changing the meaning of existing fields

Backwards-compatible changes include:

- adding endpoints
- adding optional request parameters
- adding response fields
- changing response field order
- changing fields suffixed with `_formatted`

### Authentication

Two authentication mechanisms are documented:

1. **OAuth2 (`main_auth`)**
   - Authorization URL: `https://myanimelist.net/v1/oauth2/authorize`
   - Scope documented in the supplied source: `write:users`
   - Used for endpoints requiring user authorization.
   - Send the resulting token as:
     `Authorization: Bearer YOUR_TOKEN`

2. **Client ID (`client_auth`)**
   - Header: `X-MAL-CLIENT-ID: YOUR_CLIENT_ID`
   - Intended for endpoints where user login is not required.

Some endpoints list both auth mechanisms; follow the endpoint's documented authorization requirement.

### Pagination

List endpoints use:

- `limit`
- `offset`

Typical response envelope:

```json
{
  "data": [],
  "paging": {
    "previous": "https://...",
    "next": "https://..."
  }
}
```

Prefer following the returned `paging.next` URL when crawling rather than reconstructing pagination URLs manually.

### Field Selection

MAL does **not** return every field by default.

Use the `fields` query parameter to request fields explicitly.

Nested fields are supported, e.g.:

`fields=synopsis,my_list_status{priority,comments}`

For ingestion/crawling, request only the fields actually needed.

### NSFW

Some APIs omit NSFW content by default.

Use:

`nsfw=true|false`

when the endpoint supports it.

### Dates / Times

- `date-time`: ISO 8601, e.g. `2015-03-02T06:03:11+00:00`
- `date`: `YYYY-MM-DD`, `YYYY-MM`, or `YYYY`
- `time`: `HH:MM`

### Common HTTP Errors

| Status | Meaning |
|---|---|
| `400` | Invalid parameters |
| `401` | Invalid/expired access token (`invalid_token`) |
| `403` | DoS detected etc. |
| `404` | Resource not found |

Error shape:

```json
{
  "error": "invalid_token",
  "message": "token is invalid"
}
```

## Endpoint Catalog

| Resource | Method | Endpoint | Auth | Main purpose |
|---|---:|---|---|---|
| anime | GET | `/anime` | OAuth / Client ID | Search anime |
| anime | GET | `/anime/{anime_id}` | OAuth / Client ID | Anime details |
| anime | GET | `/anime/ranking` | OAuth / Client ID | Anime rankings |
| anime | GET | `/anime/season/{year}/{season}` | OAuth / Client ID | Seasonal anime |
| anime | GET | `/anime/suggestions` | OAuth | User-specific suggestions |
| user animelist | PUT | `/anime/{anime_id}/my_list_status` | OAuth | Add/update user's anime list item |
| user animelist | DELETE | `/anime/{anime_id}/my_list_status` | OAuth | Delete user's anime list item |
| user animelist | GET | `/users/{user_name}/animelist` | OAuth / Client ID | Read a user's anime list |
| forum | GET | `/forum/boards` | OAuth / Client ID | Forum board hierarchy |
| forum | GET | `/forum/topic/{topic_id}` | OAuth / Client ID | Forum topic + posts |
| forum | GET | `/forum/topics` | OAuth / Client ID | Search/list forum topics |
| manga | GET | `/manga` | OAuth / Client ID | Search manga |
| manga | GET | `/manga/{manga_id}` | OAuth / Client ID | Manga details |
| manga | GET | `/manga/ranking` | OAuth / Client ID | Manga rankings |
| user mangalist | PUT | `/manga/{manga_id}/my_list_status` | OAuth | Add/update user's manga list item |
| user mangalist | DELETE | `/manga/{manga_id}/my_list_status` | OAuth | Delete user's manga list item |
| user mangalist | GET | `/users/{user_name}/mangalist` | OAuth / Client ID | Read a user's manga list |
| user | GET | `/users/@me` | OAuth | Current authenticated user |

---

## Get anime list
**Method:** `GET`  **Path:** `/anime?q=one&limit=4`  **Authorization:** OAuth (`write:users`) / Client ID
### Parameters


**query parameters**
| q | string  |
| - | ------- |
| limit | integer |  Default:  100 |
The maximum value is 100.
| offset | integer |  Default:  0 |
| fields | string  |              |

### Request example

```bash
curl 'https://api.myanimelist.net/v2/anime?q=one&limit=4' \
-H 'Authorization: Bearer YOUR_TOKEN'
```

---

## Get anime details
**Method:** `GET`  **Path:** `/anime/30230?fields=id,title,main_picture,alternative_titles,start_date,end_date,synopsis,mean,rank,popularity,num_list_users,num_scoring_users,nsfw,created_at,updated_at,media_type,status,genres,my_list_status,num_episodes,start_season,broadcast,source,average_episode_duration,rating,pictures,background,related_anime,related_manga,recommendations,studios,statistics`  **Authorization:** OAuth (`write:users`) / Client ID
### Parameters


**path parameters**
| anime_id |  required  | integer  |

**query parameters**
| fields | string  |

### Request example

```bash
curl 'https://api.myanimelist.net/v2/anime/30230?fields=id,title,main_picture,alternative_titles,start_date,end_date,synopsis,mean,rank,popularity,num_list_users,num_scoring_users,nsfw,created_at,updated_at,media_type,status,genres,my_list_status,num_episodes,start_season,broadcast,source,average_episode_duration,rating,pictures,background,related_anime,related_manga,recommendations,studios,statistics' \
-H 'Authorization: Bearer YOUR_TOKEN'
```

---

## Get anime ranking
**Method:** `GET`  **Path:** `/anime/ranking?ranking_type=all&limit=4`  **Authorization:** OAuth (`write:users`) / Client ID
### Behavior

The returned anime contains the `ranking` field.
### Parameters


**query parameters**
| ranking_type |  required  | string  |
| value        |                         |
| all          | Top Anime Series        |
| airing       | Top Airing Anime        |
| upcoming     | Top Upcoming Anime      |
| tv           | Top Anime TV Series     |
| ova          | Top Anime OVA Series    |
| movie        | Top Anime Movies        |
| special      | Top Anime Specials      |
| bypopularity | Top Anime by Popularity |
| favorite     | Top Favorited Anime     |
| limit | integer |  Default:  100 |
The maximum value is 500.
| offset | integer |  Default:  0 |
| fields | string  |              |

### Request example

```bash
curl 'https://api.myanimelist.net/v2/anime/ranking?ranking_type=all&limit=4' \
-H 'Authorization: Bearer YOUR_TOKEN'
```

---

## Get seasonal anime
**Method:** `GET`  **Path:** `/anime/season/2017/summer?limit=4`  **Authorization:** OAuth (`write:users`) / Client ID
### Behavior

Get seasonal anime.
| Season name Months  |                             |
| ------------------- | --------------------------- |
| winter              | January, February, March    |
| spring              | April, May, June            |
| summer              | July, August, September     |
| fall                | October, November, December |
### Parameters


**path parameters**
| year   |  required  | integer  |
| season |  required  | string   |

**query parameters**
| sort | string  |
| Value Order            |            |
| `anime_score`          | Descending |
| `anime_num_list_users` | Descending |
| limit | integer |  Default:  100 |
The maximum value is 500.
| offset | integer |  Default:  0 |
| fields | string  |              |

### Request example

```bash
curl 'https://api.myanimelist.net/v2/anime/season/2017/summer?limit=4' \
-H 'Authorization: Bearer YOUR_TOKEN'
```

---

## Get suggested anime
**Method:** `GET`  **Path:** `/anime/suggestions?limit=4`  **Authorization:** OAuth (`write:users`)
### Behavior

Returns suggested anime for the authorized user.
If the user is new comer, this endpoint returns an empty list.
### Parameters


**query parameters**
| limit | integer |  Default:  100 |
The maximum value is 100.
| offset | integer |  Default:  0 |
| fields | string  |              |

### Request example

```bash
curl 'https://api.myanimelist.net/v2/anime/suggestions?limit=4' \
-H 'Authorization: Bearer YOUR_TOKEN'
```

---

## Update my anime list status
**Method:** `PUT`  **Path:** `/anime/17074/my_list_status`  **Authorization:** OAuth (`write:users`)
### Behavior

Add specified anime to my anime list.
If specified anime already exists, update its status.
This endpoint updates only values specified by the parameter.
### Parameters


**path parameters**
| anime_id |  required  | integer  |

**Request Body schema: application/x-www-form-urlencoded**
| status | string  |
- watching
- completed
- on_hold
- dropped
- plan_to_watch
| is_rewatching | boolean  |
| score          | integer  |
0-10
| num_watched_episodes | integer  |
| priority               | integer  |
0-2
| num_times_rewatched | integer  |
| rewatch_value        | integer  |
0-5
| tags     | string  |
| comments | string  |

### Request example

```bash
curl 'https://api.myanimelist.net/v2/anime/17074/my_list_status' \
-X PUT \
-d status=completed \
-d score=8 \
-d num_watched_episodes=3 \
-H 'Authorization: Bearer YOUR_TOKEN'
```

---

## Delete my anime list item.
**Method:** `DELETE`  **Path:** `/anime/21/my_list_status`  **Authorization:** OAuth (`write:users`)
### Behavior

If the specified anime does not exist in user's anime list, this endpoint does nothing
and returns `404 Not Found`.
So be careful when retrying.
### Parameters


**path parameters**
| anime_id |  required  | integer  |

### Request example

```bash
curl 'https://api.myanimelist.net/v2/anime/21/my_list_status' \
-X DELETE \
-H 'Authorization: Bearer YOUR_TOKEN'
```

---

## Get user anime list
**Method:** `GET`  **Path:** `/users/@me/animelist?fields=list_status&limit=4`  **Authorization:** OAuth (`write:users`) / Client ID
### Parameters


**path parameters**
| user_name |  required  | string  |

**query parameters**
| status | string  |
- watching
- completed
- on_hold
- dropped
- plan_to_watch
| sort | string  |
| Value Order                    |            |
| `list_score`                   | Descending |
| `list_updated_at`              | Descending |
| `anime_title`                  | Ascending  |
| `anime_start_date`             | Descending |
| `anime_id` (Under Development) | Ascending  |
| limit | integer |  Default:  100 |
The maximum value is 1000.
| offset | integer |  Default:  0 |

### Request example

```bash
curl 'https://api.myanimelist.net/v2/users/@me/animelist?fields=list_status&limit=4' \
-H 'Authorization: Bearer YOUR_TOKEN'
```

---

## Get forum boards
**Method:** `GET`  **Path:** `/forum/boards`  **Authorization:** OAuth (`write:users`) / Client ID
### Request example

```bash
curl 'https://api.myanimelist.net/v2/forum/boards' \
-H 'Authorization: Bearer YOUR_TOKEN'
```

---

## Get forum topic detail
**Method:** `GET`  **Path:** `/forum/topic/481`  **Authorization:** OAuth (`write:users`) / Client ID
### Parameters


**path parameters**
| topic_id |  required  | integer  |

**query parameters**
| limit  | integer  <= 100  |  Default:  100 |
| offset | integer          |  Default:  0   |

### Request example

```bash
curl 'https://api.myanimelist.net/v2/forum/topic/481' \
-H 'Authorization: Bearer YOUR_TOKEN'
```

---

## Get forum topics
**Method:** `GET`  **Path:** `/forum/topics?q=love&subboard_id=2&limit=10`  **Authorization:** OAuth (`write:users`) / Client ID
### Parameters


**query parameters**
| board_id    | integer          |                     |
| subboard_id | integer          |                     |
| limit        | integer  <= 100  |  Default:  100      |
| offset       | integer          |  Default:  0        |
| sort         | string           |  Default:  "recent" |
| q                 | string  |
| topic_user_name | string  |
| user_name        | string  |

### Request example

```bash
curl 'https://api.myanimelist.net/v2/forum/topics?q=love&subboard_id=2&limit=10' \
-H 'Authorization: Bearer YOUR_TOKEN'
```

---

## Get manga list
**Method:** `GET`  **Path:** `/manga?q=berserk`  **Authorization:** OAuth (`write:users`) / Client ID
### Parameters


**query parameters**
| q | string  |
| - | ------- |
| limit | integer |  Default:  100 |
The maximum value is 100.
| offset | integer |  Default:  0 |
| fields | string  |              |

### Request example

```bash
curl 'https://api.myanimelist.net/v2/manga?q=berserk' \
-H 'Authorization: Bearer YOUR_TOKEN'
```

---

## Get manga details
**Method:** `GET`  **Path:** `/manga/2?fields=id,title,main_picture,alternative_titles,start_date,end_date,synopsis,mean,rank,popularity,num_list_users,num_scoring_users,nsfw,created_at,updated_at,media_type,status,genres,my_list_status,num_volumes,num_chapters,authors{first_name,last_name},pictures,background,related_anime,related_manga,recommendations,serialization{name}`  **Authorization:** OAuth (`write:users`) / Client ID
### Parameters


**path parameters**
| manga_id |  required  | integer  |

**query parameters**
| fields | string  |

### Request example

```bash
curl 'https://api.myanimelist.net/v2/manga/2?fields=id,title,main_picture,alternative_titles,start_date,end_date,synopsis,mean,rank,popularity,num_list_users,num_scoring_users,nsfw,created_at,updated_at,media_type,status,genres,my_list_status,num_volumes,num_chapters,authors{first_name,last_name},pictures,background,related_anime,related_manga,recommendations,serialization{name}' \
-H 'Authorization: Bearer YOUR_TOKEN'
```

---

## Get manga ranking
**Method:** `GET`  **Path:** `/manga/ranking?ranking_type=all&limit=4`  **Authorization:** OAuth (`write:users`) / Client ID
### Behavior

The returned manga contains the `ranking` field.
### Parameters


**query parameters**
| ranking_type |  required  | string  |
| value        |                |
| all          | All            |
| manga        | Top Manga      |
| novels       | Top Novels     |
| oneshots     | Top One-shots  |
| doujin       | Top Doujinshi  |
| manhwa       | Top Manhwa     |
| manhua       | Top Manhua     |
| bypopularity | Most Popular   |
| favorite     | Most Favorited |
| limit | integer |  Default:  100 |
The maximum value is 500.
| offset | integer |  Default:  0 |
| fields | string  |              |

### Request example

```bash
curl 'https://api.myanimelist.net/v2/manga/ranking?ranking_type=all&limit=4' \
-H 'Authorization: Bearer YOUR_TOKEN'
```

---

## Update my manga list status
**Method:** `PUT`  **Path:** `/manga/2/my_list_status`  **Authorization:** OAuth (`write:users`)
### Behavior

Add specified manga to my manga list.
If specified manga already exists, update its status.
This endpoint updates only values specified by the parameter.
### Parameters


**path parameters**
| manga_id |  required  | integer  |

**Request Body schema: application/x-www-form-urlencoded**
| status | string  |
- reading
- completed
- on_hold
- dropped
- plan_to_read
| is_rereading | boolean  |
| score         | integer  |
0-10
| num_volumes_read  | integer  |
| num_chapters_read | integer  |
| priority            | integer  |
0-2
| num_times_reread | integer  |
| reread_value      | integer  |
0-5
| tags     | string  |
| comments | string  |

### Request example

```bash
curl 'https://api.myanimelist.net/v2/manga/2/my_list_status' \
-X PUT \
-d status=completed \
-d score=8 \
-H 'Authorization: Bearer YOUR_TOKEN'
```

---

## Delete my manga list item.
**Method:** `DELETE`  **Path:** `/manga/2/my_list_status`  **Authorization:** OAuth (`write:users`)
### Behavior

If the specified manga does not exist in user's manga list, this endpoint does nothing
and returns `404 Not Found`.
So be careful when retrying.
### Parameters


**path parameters**
| manga_id |  required  | integer  |

### Request example

```bash
curl 'https://api.myanimelist.net/v2/manga/2/my_list_status' \
-X DELETE \
-H 'Authorization: Bearer YOUR_TOKEN'
```

---

## Get user manga list
**Method:** `GET`  **Path:** `/users/@me/mangalist?fields=list_status&limit=4`  **Authorization:** OAuth (`write:users`) / Client ID
### Parameters


**path parameters**
| user_name |  required  | string  |

**query parameters**
| status | string  |
- reading
- completed
- on_hold
- dropped
- plan_to_read
| sort | string  |
| Value Order                    |            |
| `list_score`                   | Descending |
| `list_updated_at`              | Descending |
| `manga_title`                  | Ascending  |
| `manga_start_date`             | Descending |
| `manga_id` (Under Development) | Ascending  |
| limit | integer |  Default:  100 |
The maximum value is 1000.
| offset | integer |  Default:  0 |

### Request example

```bash
curl 'https://api.myanimelist.net/v2/users/@me/mangalist?fields=list_status&limit=4' \
-H 'Authorization: Bearer YOUR_TOKEN'
```

---

## Get my user information
**Method:** `GET`  **Path:** `/users/@me?fields=anime_statistics`  **Authorization:** OAuth (`write:users`)
### Parameters


**path parameters**
| user_id |  required  | string  |

**query parameters**
| fields | string  |

### Request example

```bash
curl 'https://api.myanimelist.net/v2/users/@me?fields=anime_statistics' \
-H 'Authorization: Bearer YOUR_TOKEN'
```

---

## Implementation Guidance for an AI Agent

When implementing a MAL client from this reference:

1. Use `https://api.myanimelist.net/v2` as the base URL.
2. Keep authentication configurable; do not hard-code credentials.
3. Use `Authorization: Bearer <token>` for OAuth endpoints.
4. Use `X-MAL-CLIENT-ID: <client-id>` where the endpoint permits client authentication.
5. Treat `fields` as an important optimization. Do not assume full objects are returned.
6. Treat list responses as paginated and follow `paging.next`.
7. Do not assume `limit` can exceed the documented maximum for an endpoint.
8. Preserve MAL IDs as stable external IDs.
9. Do not infer that a missing field means the value is empty; it may simply not have been requested through `fields`.
10. For user list mutation endpoints, only send fields that should actually change. The source explicitly says updates only apply to values specified by the request.
11. For delete-list-item endpoints, note the documented `404` behavior when the item does not exist; retries should therefore be handled carefully.
12. Preserve raw MAL payloads when building an ingestion/crawler pipeline if downstream normalization may need fields that are not currently modeled.
13. Do not treat the sample data values as current MAL data. They are documentation examples.
14. The supplied source is a beta-version API document; verify against the live API before relying on undocumented behavior.

## Source Coverage Note

This reference is intentionally based on the supplied documentation rather than silently adding information from other MAL API versions or third-party clients.

The original supplied document contains the detailed ReDoc material, including endpoint-specific request/response samples and field examples. Use this normalized reference for AI context and keep the original document available when exact response examples are needed.
