#!/usr/bin/env bash

pkill bsnake
go run bsnake >/tmp/log.txt 2>&1 &
sleep 2

if [ "$1" == "solo" ]; then
  battlesnake play -W 7 -H 7 --name you --url http://localhost:8080/monty/ -g solo -v
elif [ "$1" == "dual" ]; then
  battlesnake play -W 7 -H 7 --name 1 --url http://localhost:8080/monty/ --name 2 --url http://localhost:8080/basic/ -g standard
else
  echo "unknown $1"
fi
