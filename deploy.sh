#!/bin/bash
set -e

replaces ()
{
  sed -i -e "s/emailsPass = \".*\"/emailsPass = \"$1\"/g" "config.go"
}

if [ "$1" != "run" ]; then
  git pull
fi

read -e -s -p "Enter the password: " password
patch < secrets.diff
replaces $password
go build
replaces ""
patch -R < secrets.diff

./nk-fitness 2>&1 | tee -a ./nk-fitness.log
