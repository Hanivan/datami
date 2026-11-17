# datami

Go 1.27 library: one client for anime/manga metadata across MAL, AniList, AniDB, Kitsu, Hikka, Kaize, Nautiljon and OtakOtaku. Every source maps to one `types.Media` shape.

```
go get github.com/Hanivan/datami
```

## Quick start

```go
c, _ := datami.New(datami.Config{MALClientID: "your-id"})
m, _ := c.MAL.Get(context.Background(), "1")
fmt.Println(m.Titles[0].Text)
```

| Provider | Get | Search | ListAll | Needs |
|----------|-----|--------|---------|-------|
| [MAL](docs/features/mal.md) | [yes] | [yes] | - | client id |
| [AniList](docs/features/anilist.md) | [yes] | [yes] | - | - |
| [AniDB](docs/features/anidb.md) | [yes] | [yes] | - | registered client |
| [Kitsu](docs/features/kitsu.md) | [yes] | [yes] | - | - |
| [Hikka](docs/features/hikka.md) | - | [yes] | [yes] | - |
| [Kaize](docs/features/kaize.md) | - | [yes] | [yes] | - |
| [Nautiljon](docs/features/nautiljon.md) | - | [yes] | [yes] | - |
| [OtakOtaku](docs/features/otakotaku.md) | [yes] | - | [yes] | - |

Unsupported methods return `types.ErrNotImplemented`. Errors are `*types.Error`; match with `errors.Is` ([errors](docs/features/errors.md)).

## Commands

| Command | Does |
|---------|------|
| `go test ./...` | all tests (httptest, no network) |
| `go vet ./...` | vet |
| `gofmt -l .` | must print nothing (CI) |

## Docs

Per-feature details (config, rate limits, error mapping, gotchas) are in [docs/features](docs/features/README.md): [configuration](docs/features/configuration.md), [errors](docs/features/errors.md), [HTTP client](docs/features/http-client.md), [media model](docs/features/media-model.md). Upstream API references: [docs/api-docs](docs/api-docs/).
