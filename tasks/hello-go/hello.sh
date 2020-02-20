#!/bin/sh
mkdir /tmp/work
cp /task/hello.go /tmp/work/hello.go
cd /tmp/work
go build -o hello
./hello
