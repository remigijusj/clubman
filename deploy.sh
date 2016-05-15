#!/bin/bash
set -e

NAME="$1"
MODE="$2"

if [ "$MODE" != "run" ]; then
  git pull
fi

patch -p1 < secrets.diff
go build -o "$NAME" ./src
patch -p1 -R < secrets.diff

if [ "$MODE" == "run" ]; then
  "./$NAME" 2>&1 | tee -a "./$NAME.log"
fi
