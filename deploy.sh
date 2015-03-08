#!/bin/bash
set -e

replaces ()
{
  sed -i -e "s/emailsPass = \".*\"/emailsPass = \"$1\"/g" "config.go"
}

git pull

read -e -s -p "Enter the password: " password
patch < secrets.diff
replaces $password
go build
replaces ""
patch -R < secrets.diff

./nk-fitness
