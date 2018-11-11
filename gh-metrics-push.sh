#!/bin/bash
# Script to copy and push new metric versions to the assets branch.

[ ! -z "$GIT_ASSETS_BRANCH" ] || exit 1
[ ! -z "$GIT_API_KEY" ] || exit 1

version=$(git describe HEAD) || exit 1

# Constants
ASSETS_DIR=".assets-branch"
METRICS_DIR="$ASSETS_DIR/metriclists"

# Ensure metrics dir exists
mkdir -p "$METRICS_DIR/"

# Remove old files so we spot deletions
rm -f "$METRICS_DIR/.*.unique"

# Copy new files
cp -f -t "$METRICS_DIR/" ./.metrics.*.prom.unique || exit 1

# Enter the assets dir and push.
cd "$ASSETS_DIR" || exit 1

git add "metriclists" || exit 1
git commit -m "Added unique metrics for build from $version" || exit 1
git push origin "$GIT_ASSETS_BRANCH" || exit 1

exit 0