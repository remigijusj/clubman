#!/bin/bash
set -e

replaces ()
{
  local name="config.go"
  local pass="$1"
  local reload="$2"

  echo "SUBST $name WITH reload=$reload"
  sed -i -e "s/emailsPass = \".*\"/emailsPass = \"$pass\"/g" -e "s/reloadTmpl = [a-z]*/reloadTmpl = $reload/g" "$name"
}

git pull

read -e -s -p "Enter the password: " password
replaces $password "false"
go build
replaces "" "true"

./nk-fitness
