#!/usr/bin/env bash
set -euxo pipefail

main() {

  echo "Hello World!
I received secrets: $SAMPLE_SECRET_ONE and $SAMPLE_SECRET_TWO
I received arguments: $SAMPLE_ARG_ONE and $SAMPLE_ARG_TWO"

}

main

