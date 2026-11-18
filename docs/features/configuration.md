# Configuration

## What

`datami.Config` (`config.go`) is the single struct passed to `datami.New` (`client.go`). `New` builds all eight provider clients and returns `*datami.Client` with fields `MAL`, `AniList`, `AniDB`, `Kitsu`, `Hikka`, `Kaize`, `Nautiljon`, `Otakotaku`. Each provider can also be built directly via its own package.

## How it works

`New` resolves `Timeout` (0 becomes 10s), passes it plus `Metrics` to every provider, and wraps provider constructor errors as `fmt.Errorf("<name>: %w", err)` (AniDB error is returned unwrapped). Providers with no credentials (Kitsu, Hikka, Kaize, Nautiljon, Otakotaku) take only `Timeout` and `Metrics`.

## Config

| Field | Default | Feeds |
| --- | --- | --- |
| `MALClientID` | "" | `mal.Config.ClientID`, sent as `X-MAL-CLIENT-ID` header |
| `AniListToken` | "" | `anilist.Config.Token`, sent as `Authorization: Bearer` when non-empty |
| `AniListDescriptionAsHTML` | nil (= true, HTML) | `anilist.Config.DescriptionAsHTML`; `*bool`, false gives AniList markdown |
| `AniDBClientName` | "" | `anidb.Config.ClientName`; empty leaves AniDB unconfigured (`anidb.Unconfigured()`, every call returns `ErrInvalidInput`, no network) |
| `AniDBClientVersion` | 0 | `anidb.Config.ClientVersion`; with a name set, `anidb.New` rejects `<= 0` with `ErrInvalidInput` |
| `AniDBCacheDir` | "" (becomes `os.UserCacheDir()/datami/anidb`) | `anidb.Config.CacheDir`, titles dump cache |
| `Timeout` | 10s | all providers; per-request `http.Client.Timeout` |
| `Metrics` | nil | all providers |

`MetricsHook` is `func(source string, op string, dur time.Duration, err error)` (`types/media.go`). `source` is one of `types.SourceMAL` ("mal"), `SourceAniList` ("anilist"), `SourceAniDB` ("anidb"), `SourceKitsu`, `SourceHikka`, `SourceKaize`, `SourceNautiljon`, `SourceOtakotaku` (lowercase names). `op` is the provider operation name (the provider `Op*` const, e.g. `mal.OpGet`, `anilist.OpSearch`). Rates are hardcoded per provider, see [http-client.md](http-client.md). Fires per attempt, see [http-client.md](http-client.md).

Per-provider direct constructors (packages under `provider/`): `New(cfg) (*Client, error)` hits the real endpoint; `NewWithTestServer(baseURL, cfg)` points at a test server (AniDB: `NewWithTestServer(apiURL, titlesURL, cfg)`, empty `ClientName` becomes "test"/1). Each provider `Config` has `Timeout` and `Metrics` (plain func type, same signature); mal and anilist also take `UserAgent` (sent on every request; "" sends `datami (+https://github.com/Hanivan/datami)`, never Go's anonymous default); extras: mal `ClientID`, anilist `Token`/`DescriptionAsHTML`, anidb `ClientName`/`ClientVersion`/`CacheDir`. Provider docs: [mal.md](mal.md), [anilist.md](anilist.md), [anidb.md](anidb.md), [kitsu.md](kitsu.md), [hikka.md](hikka.md), [kaize.md](kaize.md), [nautiljon.md](nautiljon.md), [otakotaku.md](otakotaku.md).

## Code

```go
c, err := datami.New(datami.Config{
    MALClientID:        "abc",
    AniDBClientName:    "myapp",
    AniDBClientVersion: 1,
    Timeout:            15 * time.Second,
    Metrics: func(source, op string, d time.Duration, err error) {
        log.Printf("%s %s %v err=%v", source, op, d, err)
    },
})
m, err := c.AniList.Get(ctx, "1")
```

Direct: `k, _ := kitsu.New(kitsu.Config{Timeout: 5 * time.Second})`.

## Test

```bash
go test ./...
go test ./anilist/ ./provider/...
```

Use `NewWithTestServer` with `httptest.NewServer`.

## Gotchas

- `datami.MetricsHook` is a named type; provider `Config.Metrics` is the unnamed func type, assignable from it.
- Missing `MALClientID` is not validated at `New`; MAL will answer 401/403.
- AniDB cache dir lookup failure makes `datami.New` return an error (only when a client name is set).
