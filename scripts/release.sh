#!/usr/bin/env bash

set -e

if [ -n "$TRAVIS_TAG" ]; then
  curl -sL https://git.io/goreleaser | bash
  SHA=$(cat dist/checksums.txt | grep Darwin_x86_64 | awk '{ print $1}')
  go run scripts/proctor_template.go $TRAVIS_TAG $SHA
  rm -rf homebrew-gojek
  git clone "https://$GITHUB_TOKEN:@github.com/gojektech/homebrew-gojek.git"
  cp scripts/proctor.rb homebrew-gojek/Formula/proctor.rb
  cd homebrew-gojek
  git add .
  git commit -m "[TravisCI] updating brew formula for release $TRAVIS_TAG"
  git push --force --quiet "https://$GITHUB_TOKEN:@github.com/gojektech/homebrew-gojek.git"
fi


