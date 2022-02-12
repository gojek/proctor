#!/usr/bin/env bash

set -e

curl -sL https://git.io/goreleaser | bash
SHA=$(cat dist/checksums.txt | grep Darwin_x86_64 | awk '{ print $1}')
echo " REF_NAME => $GITHUB_REF_NAME"
go run scripts/proctor_template.go "$GITHUB_REF_NAME" "$SHA"
rm -rf homebrew-gojek
git clone "https://$GITHUB_TOKEN:@github.com/gojek/homebrew-gojek.git"
cp scripts/proctor.rb homebrew-gojek/Formula/proctor.rb
cd homebrew-gojek
git add .
git commit -m "[GH Actions] updating brew formula for release $GITHUB_REF_NAME"
#git push --force --quiet "https://$GITHUB_TOKEN:@github.com/gojektech/homebrew-gojek.git"

