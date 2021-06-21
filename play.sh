#!/usr/bin/env bash

pkill bsnake
go run github.com/corverroos/bsnake >/tmp/log.txt 2>&1 &
sleep 2

if [ "$1" == "solo" ]; then
  battlesnake play -W 7 -H 7 --name you --url http://localhost:8080/boomboom/ -g solo -v
elif [ "$1" == "dual" ]; then
  battlesnake play -W 11 -H 11 --name 1 --url http://localhost:8080/boomboom/ --name 2 --url http://localhost:8080/basic/ -g standard -v
elif [ "$1" == "royale" ]; then
  battlesnake play -W 11 -H 11 --name 1 --url http://localhost:8080/boomboom/ --name 2 --url http://localhost:8080/basic/ -g royale -v
elif [ "$1" == "self" ]; then
  battlesnake play -W 11 -H 11 --name 1 --url http://localhost:8080/boomboom/ --name 2 --url http://localhost:8080/boomboom/ -g standard -v
else
  echo "unknown $1"
fi
