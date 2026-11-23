# Releasing

Versions are Go module tags `vX.Y.Z[-stage.N]`. Pushing a `v*` tag runs `.github/workflows/release.yml`: `gofmt`, `go vet`, `go test ./...`, then a GitHub release with notes from `scripts/release-notes.sh`.

## Version number

The core `X.Y.Z` is a semver bump from commit types since the last **stable** tag (no suffix); the highest bump wins:

| Commits since last stable | Bump |
|---|---|
| `type!:` or a `BREAKING CHANGE:` body line | major; while `v0`, minor (Go promises no compatibility before `v1`) |
| any `feat` | minor |
| only `fix`/`refactor`/`docs`/`chore`/... | patch |
| nothing since the last tag | no release |

No stable tag exists yet, so the first core is `v0.1.0`.

## Stages

Each core goes through optional pre-release stages, in this order only:

`-alpha.N` -> `-beta.N` -> `-rc.N` -> stable

- `N` starts at 1 and counts up within a core and stage.
- Moving to a later stage resets `N` to 1. A stage never goes back (no alpha after `rc` of the same core).
- Any stage may be skipped, including straight to stable.
- After stable `vX.Y.Z` ships, the next change starts a new core.

Example: `v0.1.0-alpha.25` -> `v0.1.0-rc.1` -> `v0.1.0`; then a `feat` lands -> `v0.2.0-alpha.1`.

## Cutting a release

```bash
scripts/next-version.sh alpha        # or beta | rc | stable; prints tag + deciding commit
git tag -a v0.1.0-rc.1 -m v0.1.0-rc.1   # annotated; gpg-signed when tag.gpgSign is set
git push origin v0.1.0-rc.1
```

State the chosen tag and the deciding commit before tagging. Never reuse, move or delete a pushed tag: the Go module proxy caches it forever. A bad release is fixed by a new tag.

## Release notes

`scripts/release-notes.sh <tag> [previous-tag]` groups commit subjects: `feat` -> Features, `fix` -> Fixes, `update`/`perf`/`refactor`/`style` -> Improvements, `chore`/`ci`/`docs`/`build`/`test` -> Maintenance; `type!:` is marked **breaking**. The range starts at:

- pre-release tag: the previous `v*` tag of any kind
- stable tag: the previous stable tag, so stable notes roll up all its pre-releases

Suffixed tags publish as GitHub pre-releases (not "Latest"). The job skips when the release exists. If it failed, publish by hand:

```bash
scripts/release-notes.sh vX.Y.Z > notes.md && gh release create vX.Y.Z --title vX.Y.Z --notes-file notes.md
```

Big release: `gh release edit vX.Y.Z --notes-file notes.md` to add a short **Highlights** list above "What's changed".

## Commit subjects

Subjects become the notes, so keep them user-facing:

```
feat|fix|refactor|docs|chore(scope)[!]: title

- item 1
- item N
```
