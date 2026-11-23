# datami features

One file per feature. Each covers **what** it does, **how** it works, **config**, **where** the code lives, **how to test**, and **gotchas**.

| Feature | Summary |
|---------|---------|
| [Configuration](configuration.md) | Root `Config`, per-provider constructors, metrics hook |
| [Errors](errors.md) | `types.Error`, sentinels, HTTP status mapping |
| [HTTP client](http-client.md) | Rate limits, retries, Retry-After, body size cap |
| [Media model](media-model.md) | `Media`, enums, `Character`, `Person`, tags/links |
| [MAL](mal.md) | MyAnimeList v2 API: anime, manga, seasonal, ranking, characters |
| [AniList](anilist.md) | GraphQL: media, characters, staff, tags/links, `*ByMalID` |
| [AniDB](anidb.md) | HTTP API + titles-dump search, registered client, cache |
| [Kitsu](kitsu.md) | JSON:API: get, search, manga |
| [Hikka](hikka.md) | Search and list, JSON |
| [Kaize](kaize.md) | Search and list, HTML scrape |
| [Nautiljon](nautiljon.md) | Search and list, HTML scrape |
| [OtakOtaku](otakotaku.md) | Get and list, JSON |
| [Releasing](releasing.md) | Versioning, alpha/beta/rc stages, tagging, release notes |

Install and commands: [project README](../../README.md).
