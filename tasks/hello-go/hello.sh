#!/bin/sh
PATH=/usr/bin:$PATH
mkdir /tmp/work
GOPATH=/tmp/work
cp /task/hello.go /tmp/work/hello.go
cd /tmp/work
go build
./work
