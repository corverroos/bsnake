#!/usr/bin/env bash

pkill bsnake
go run github.com/corverroos/bsnake >/tmp/log.txt 2>&1 &
sleep 2

if [ "$1" == "solo" ]; then
  battlesnake play -W 7 -H 7 --name you --url http://localhost:8080/latest/ -g solo -v
elif [ "$1" == "heur" ]; then
  battlesnake play -W 11 -H 11 --name M3 --url http://localhost:8080/mx3/ --name M4 --url http://localhost:8080/mx4/ -g standard -v
elif [ "$1" == "dual" ]; then
  battlesnake play -W 11 -H 11 --name L --url http://localhost:8080/latest/ --name M3 --url http://localhost:8080/mx3/ -g standard -v
elif [ "$1" == "trip" ]; then
  battlesnake play -W 11 -H 11 --name L --url http://localhost:8080/latest/ --name 1 --url http://localhost:8080/v1/ --name M3 --url http://localhost:8080/mx3/ -g standard -v
elif [ "$1" == "quad" ]; then
  battlesnake play -W 11 -H 11 --name V3 --url http://localhost:8080/v3/ --name V2 --url http://localhost:8080/v2/ --name M3 --url http://localhost:8080/mx3/ --name M2 --url http://localhost:8080/mx2/ -g standard -v
elif [ "$1" == "royale" ]; then
  battlesnake play -W 11 -H 11 --name L --url http://localhost:8080/latest/ --name 1 --url http://localhost:8080/v1/ --name 0 --url http://localhost:8080/v0/ -g royale -v
elif [ "$1" == "self" ]; then
  battlesnake play -W 11 -H 11 --name L1 --url http://localhost:8080/latest/ --name L2 --url http://localhost:8080/latest/ -g standard -v
else
  echo "unknown $1"
fi
