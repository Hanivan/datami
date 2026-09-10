# datami

Go 1.27 library - unified anime/manga metadata client across multiple sources.

```
go get github.com/linkeunid/datami
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

## Docs

- [Provider API references](docs/api-docs/)
