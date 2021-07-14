#!/usr/bin/env bash
set -e

GITCOMMIT=$(git rev-parse --short HEAD 2>/dev/null)
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
VERSION=$(git branch --show-current)

go build -ldflags "-X speedrun/cmd.version=$VERSION -X speedrun/cmd.commit=$GITCOMMIT -X speedrun/cmd.date=$DATE"
