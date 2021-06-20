#!/bin/bash

export CGO_ENABLED=0
export GOOS=linux

TARGET=root@battle01

echo "building"
go build -o="/tmp/bsnake" .
scp run.sh $TARGET:.

echo "stopping"
ssh $TARGET "pkill bsnake"

echo "copying"
scp /tmp/bsnake $TARGET:.

echo "starting"
ssh $TARGET "./run.sh"

