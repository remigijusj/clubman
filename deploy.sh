#!/bin/bash
set -e

NAME="$1"
MODE="$2"

if [ "$MODE" != "run" ]; then
  git pull
fi

go build -o "$NAME" ./src

if [ "$MODE" == "run" ]; then
  "./$NAME" 2>&1 | tee -a "./$NAME.log"
fi
