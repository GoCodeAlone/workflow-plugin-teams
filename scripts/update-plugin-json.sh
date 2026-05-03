#!/usr/bin/env sh
# Updates plugin.json with the release version and platform download URLs.
# Called by .goreleaser.yaml before hooks.
# Usage: sh scripts/update-plugin-json.sh <version>
# Requires: jq (standard on Linux/macOS)
set -eu

VERSION="${1:?version argument required}"
BASE="https://github.com/GoCodeAlone/workflow-plugin-teams/releases/download/v${VERSION}/workflow-plugin-teams-"

jq --arg v "$VERSION" --arg b "$BASE" '
  .version = $v |
  .downloads = [
    {"os":"linux",  "arch":"amd64","url":($b+"linux-amd64.tar.gz")},
    {"os":"linux",  "arch":"arm64","url":($b+"linux-arm64.tar.gz")},
    {"os":"darwin", "arch":"amd64","url":($b+"darwin-amd64.tar.gz")},
    {"os":"darwin", "arch":"arm64","url":($b+"darwin-arm64.tar.gz")}
  ]
' plugin.json > plugin.json.tmp && mv plugin.json.tmp plugin.json
