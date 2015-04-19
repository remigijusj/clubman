#!/bin/bash
set -e

if [ "$1" != "run" ]; then
  git pull
fi

patch < secrets.diff
go build
patch -R < secrets.diff

./nk-fitness 2>&1 | tee -a ./nk-fitness.log
