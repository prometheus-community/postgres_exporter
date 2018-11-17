#!/bin/bash
# Script to setup the assets clone of the repository using GIT_ASSETS_BRANCH and
# GIT_API_KEY.

[ ! -z "$GIT_ASSETS_BRANCH" ] || exit 1

setup_git() {
  git config --global user.email "travis@travis-ci.org" || exit 1
  git config --global user.name "Travis CI" || exit 1
}

# Constants
ASSETS_DIR=".assets-branch"

# Clone the assets branch with the correct credentials
git clone --single-branch -b "$GIT_ASSETS_BRANCH" \
    "https://${GIT_API_KEY}@github.com/${TRAVIS_REPO_SLUG}.git" "$ASSETS_DIR" || exit 1

