#!/usr/bin/env bash
# Next release tag. Version core (X.Y.Z) = semver bump from commit types since
# the last stable tag; the stage picks the suffix (alpha -> beta -> rc -> stable).
# Usage: scripts/next-version.sh alpha|beta|rc|stable
set -euo pipefail
stage=${1:?usage: next-version.sh alpha|beta|rc|stable}
case $stage in alpha) rank=1 ;; beta) rank=2 ;; rc) rank=3 ;; stable) rank=4 ;; *) echo "unknown stage: $stage" >&2; exit 2 ;; esac

latest=$(git describe --tags --abbrev=0 --match 'v*' 2>/dev/null || true)
if [ -n "$latest" ] && [ "$(git rev-list -n1 "$latest")" = "$(git rev-parse HEAD)" ]; then
  echo "nothing new since $latest: no release" >&2; exit 1
fi

stable=$(git describe --tags --abbrev=0 --match 'v*' --exclude 'v*-*' 2>/dev/null || true)
range=HEAD
[ -n "$stable" ] && range="$stable..HEAD"
IFS=. read -r major minor patch <<<"${stable:-v0.0.0}"
major=${major#v}

subjects=$(git log --no-merges --format='%h %s' "$range")
breaking=$(printf '%s\n' "$subjects" | grep -E '^[0-9a-f]+ [a-z]+(\([^)]*\))?!:' | head -1 || true)
[ -z "$breaking" ] && breaking=$(git log --no-merges --format='%h %s%n%b%x00' "$range" | awk -v RS='\0' '/\nBREAKING CHANGE:/ {print $1, $2; exit}' | head -1 || true)
feat=$(printf '%s\n' "$subjects" | grep -E '^[0-9a-f]+ feat(\(|:)' | head -1 || true)
if [ -n "$breaking" ]; then
  why="breaking: $breaking"
  # v0: breaking bumps minor (Go makes no compatibility promise before v1)
  if [ "$major" = 0 ]; then minor=$((minor + 1)); patch=0; else major=$((major + 1)); minor=0; patch=0; fi
elif [ -n "$feat" ]; then
  why="feat: $feat"; minor=$((minor + 1)); patch=0
else
  why="patch: $(printf '%s\n' "$subjects" | tail -1)"; patch=$((patch + 1))
fi
core="v$major.$minor.$patch"

git rev-parse -q --verify "refs/tags/$core" >/dev/null && { echo "$core already released" >&2; exit 1; }
# Highest existing pre-release of this core: stages only move forward.
cur=$(git tag -l "$core-*" | sed -nE 's/^.*-(alpha|beta|rc)\.([0-9]+)$/\1 \2/p' |
  awk '{r=($1=="alpha")?1:($1=="beta")?2:3; print r, $2}' | sort -k1,1n -k2,2n | tail -1)
cur_rank=${cur% *}; cur_n=${cur#* }
if [ -n "$cur" ] && [ "$rank" -lt "$cur_rank" ]; then
  echo "$core is already past $stage" >&2; exit 1
fi
if [ "$stage" = stable ]; then
  next=$core
elif [ -n "$cur" ] && [ "$rank" = "$cur_rank" ]; then
  next="$core-$stage.$((cur_n + 1))"
else
  next="$core-$stage.1"
fi
echo "$next"
echo "decided by $why" >&2
