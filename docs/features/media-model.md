# Media Model

## What

`types.Media` is the canonical response every provider maps to (`types/media.go`). Related entities: `Character`, `VoiceActor`, `CharacterRole`, `Person` (`types/character.go`), `TagsLinks`/`Tag`/`ExternalLink` (`types/tag.go`), `SearchOpts` (`types/search.go`). Typed enums live in `types/enums.go`. Root package re-exports aliases so consumers write `datami.Media`.

## How it works

Each provider's `mapping.go` converts raw payloads to `Media`. Fields a source does not supply stay zero/nil. `Status`, `MediaType`, `Raw` carry `json:"-"` (not serialized); `Title.Kind` has no json tag.

`Media` fields:

| Group | Fields |
| --- | --- |
| Titles/text | `Titles []Title`, `Synopsis`, `Background` |
| Dates | `StartDate`, `EndDate` (`*Date`: `Year`, `Month`, `Day` as `*int`), `Season`, `Year` |
| Classification | `Status`, `MediaType`, `Genres []string`, `Nsfw`, `Rating`, `SourceField`, `MediaTypeField`, `StatusField` (verbatim source strings) |
| Counts | `EpisodeCount`, `ChapterCount`, `VolumeCount`, `DurationMin`, `AverageEpDur` |
| Score/stats | `Score *Score` (`Mean`, `Rank`, `Popularity`), `NumListUsers`, `NumScoringUsers`, `NumFavorites` |
| Relations | `Studios []Studio{ID,Name}`, `Relations []Relation`, `ExternalIDs []ExternalID{Source,ID,ResourceType}`, `Authors []Author`, `Magazines []Magazine` |
| Images | `Images map[ImageSize]Image{URL,Width,Height}` |
| Raw blobs | `Broadcast`, `Pictures`, `AlternativeTitles`, `Rankings`, `Stats` (`json.RawMessage`), `Raw map[string]any` |

`Relation`: `TargetID`, `TargetSource`, `TargetType`, `Kind`, `TargetMalID`, `TargetFormat`, `TargetTitle` (display only).

Enums (string types, `types/enums.go`):

| Type | Values |
| --- | --- |
| `MediaStatus` | `StatusFinished`, `StatusOngoing`, `StatusPlanned`, `StatusCancelled`, `StatusHiatus`, `StatusDiscontinued` |
| `MediaType` | `TypeTV`, `TypeTVShort`, `TypeMovie`, `TypeSpecial`, `TypeOVA`, `TypeONA`, `TypeMusic`, `TypeManga`, `TypeNovel`, `TypeOneShot`, `TypeDoujinshi`, `TypeManhwa`, `TypeManhua`, `TypeOEL`, `TypeLightNovel` |
| `TitleKind` | `KindCanonical`, `KindRomaji`, `KindEnglish`, `KindNative`, `KindSynonym`, `KindShort` |
| `Season` | `SeasonWinter`, `SeasonSpring`, `SeasonSummer`, `SeasonFall` |
| `RelationKind` | `RelationSequel`, `Prequel`, `SideStory`, `ParentStory`, `Summary`, `Alternative`, `SpinOff`, `Adaptation`, `Character`, `Other`, `Source`, `Compilation`, `Contains` (prefix `Relation`) |
| `ImageSize` | `SizeMedium`, `SizeLarge`, `SizeOriginal` |
| `ResourceType` | `ResourceAnime`, `ResourceManga` |
| `CharacterRoleType` | `RoleMain`, `RoleSupporting` |

Other types:

- `Character`: `ID`, `Names`, `Image`, `ImageLarge`, `Biography`, `FirstName`, `LastName`, `NativeName`, `AlternativeName`, `NumFavorites *int`, `Roles []CharacterRole`, `VoiceActors []VoiceActor`, `Gender`, `Tags` (reserved, filled by no current source).
- `VoiceActor`: `ID`, `Name`, `NativeName`, `Language`, `Image`, `ImageLarge`. Per code comments only AniList fills them on characters.
- `CharacterRole`: `MediaID`, `Role CharacterRoleType`.
- `Person`: `ID`, `Name`, `Image`, `ImageLarge`, `Birthday`, `Biography`, `NativeName`, `Role`, `PrimaryOccupations`, `NumFavorites *int`.
- `TagsLinks{Tags []Tag, ExternalLinks []ExternalLink}`; `Tag`: `Name`, `Category`, `Rank *int`, `IsGeneralSpoiler`, `IsMediaSpoiler`, `IsAdult`; `ExternalLink`: `ID`, `URL`, `Site`, `SiteID *int`, `Type`, `Language`, `Notes`, `IsDisabled`.
- `SearchOpts`: `Limit`, `MediaType string`, `Year`, `Status string`, `Page` (used by `ListAll`).

Root aliases (`config.go`): `Media`, `Title`, `Date`, `Score`, `Studio`, `Image`, `ExternalID`, `Relation`, `SearchOpts`, `Error`, `MetricsHook`, `TagsLinks`, `Tag`, `ExternalLink`. Not aliased: `Character`, `VoiceActor`, `CharacterRole`, `Person`, `Author`, `Magazine`, enums; import `types` for those. Providers: [mal.md](mal.md), [anilist.md](anilist.md), [anidb.md](anidb.md), [kitsu.md](kitsu.md), [hikka.md](hikka.md), [kaize.md](kaize.md), [nautiljon.md](nautiljon.md), [otakotaku.md](otakotaku.md).

## Code

```go
m, err := c.AniList.Get(ctx, "1")
if m.Status == types.StatusFinished && m.MediaType == types.TypeTV {
    img := m.Images[types.SizeLarge].URL
    _ = img
}
res, _ := c.Kitsu.Search(ctx, "naruto", datami.SearchOpts{Limit: 5, Page: 1})
```

## Test

```bash
go test ./types/
```

## Gotchas

- `Status` and `MediaType` are not JSON-serialized; use `StatusField`/`MediaTypeField` for source strings in JSON output.
- `Title.Kind` has no JSON tag, serialized as `Kind`.
- `SearchOpts.MediaType` and `Status` are plain strings, not the enums.
- `Date` parts are pointers: nil means unknown, 0 is not used as absent.
- Pointer counts (`NumFavorites`) distinguish 0 from absent on `Character`/`Person`; `Media.NumFavorites` is a plain int.
- Errors: [errors.md](errors.md).
