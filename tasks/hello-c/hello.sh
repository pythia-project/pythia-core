#!/bin/sh
mkdir /tmp/work
cp /task/hello.c /tmp/work/hello.c
cd /tmp/work
make --silent hello
./hello
