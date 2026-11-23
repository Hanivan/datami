#!/usr/bin/env bash
# Release notes for a tag, grouped by commit type (type(scope): title).
# Pre-release tag (vX.Y.Z-pre.N): changes since the previous v* tag of any kind.
# Stable tag (vX.Y.Z): changes since the previous stable tag, so it rolls up its pre-releases.
# Usage: scripts/release-notes.sh <tag> [previous-tag]
set -euo pipefail
tag=${1:?usage: release-notes.sh <tag> [previous-tag]}
if [ -z "${2:-}" ]; then
  exclude=()
  [[ $tag == *-* ]] || exclude=(--exclude 'v*-*')
  prev=$(git describe --tags --abbrev=0 --match 'v*' "${exclude[@]}" "$tag^" 2>/dev/null || true)
else
  prev=$2
fi
repo=${GITHUB_REPOSITORY:-$(gh repo view --json nameWithOwner -q .nameWithOwner)}
range=$tag
[ -n "$prev" ] && range="$prev..$tag"

# @words get backticks (no GitHub mentions)
log=$(git log --no-merges --format='%s%x09%h' "$range")

section() { # title, type regex
  local lines
  lines=$(printf '%s\n' "$log" | grep -E "^($2)(\(|!|:)" | sed -E 's/^[a-z]+(\(([^)]*)\))?(!)?: (.*)\t(.*)$/- **\2** \3\4 (\5)/; s/\*\* !/** **breaking:** /; s/^- \*\*\*\* \*\*breaking/- **breaking/; s/^- \*\*\*\* /- /; s/(^|[^`[:alnum:]])@([[:alnum:]_-]+)/\1`@\2`/g') || true
  [ -n "$lines" ] && printf '### %s\n\n%s\n\n' "$1" "$lines"
  return 0
}

echo "## What's changed"
echo
section 'Features' 'feat|feature'
section 'Fixes' 'fix'
section 'Improvements' 'update|perf|refactor|style'
section 'Maintenance' 'chore|ci|docs|build|test|init'
if [ -n "$prev" ]; then
  echo "**Full changelog**: https://github.com/$repo/compare/$prev...$tag"
else
  echo "**Full changelog**: https://github.com/$repo/commits/$tag"
fi
