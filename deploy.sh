#!/bin/bash

export CGO_ENABLED=0
export GOOS=linux

echo "building"
go build -o="/tmp/bsnake" .
scp run.sh root@droplet01:.

echo "stopping"
ssh root@droplet01 "pkill bsnake"

echo "copying"
scp /tmp/bsnake root@droplet01:.

echo "starting"
ssh root@droplet01 "./run.sh"

