# datami

Go 1.27 library - unified anime/manga metadata client across multiple sources.

```
go get github.com/Hanivan/datami
```

## Quick start

```go
c, _ := datami.New(datami.Config{MALClientID: "your-id"})
m, _ := c.MAL.Get(context.Background(), "1")
fmt.Println(m.Titles[0].Text)
```

Each provider is also callable directly:

| Provider | Get | Search | ListAll |
|----------|-----|--------|---------|
| MAL | [yes] | [yes] | - |
| AniList | [yes] | [yes] | - |
| AniDB | [yes] | [yes] | - |
| Kitsu | [yes] | [yes] | - |
| Hikka | - | [yes] | [yes] |
| Kaize | - | [yes] | [yes] |
| Nautiljon | - | [yes] | [yes] |
| OtakOtaku | [yes] | - | [yes] |

Methods that a source doesn't support return `types.ErrNotImplemented`.

### AniDB

AniDB needs a client registered on anidb.net. If `AniDBClientName` is empty, every AniDB call returns `ErrInvalidInput`.

```go
c, _ := datami.New(datami.Config{
    AniDBClientName:    "myclient",
    AniDBClientVersion: 1,
    AniDBCacheDir:      "/var/cache/datami",
})
m, _ := c.AniDB.Get(ctx, "1")
chars, _ := c.AniDB.GetCharacters(ctx, "1")
res, _ := c.AniDB.Search(ctx, "bebop", datami.SearchOpts{})
```

- `Search` uses the official anime-titles dump, cached in `AniDBCacheDir`, refreshed at most every 24h; a failed download is retried after 1h. `LoadTitles(io.Reader)` preloads it.
- `GetByMalID` returns `ErrNotImplemented`.
- `Character.Tags` is reserved for a future scraper.
- Rate limit is 1 request per 2s, no retries.

## Docs

- [Provider API references](docs/api-docs/)
