#!/bin/sh
mkdir /tmp/work
cp /task/test-c.c /tmp/work/test-c.c
cd /tmp/work
make --silent test-c
./test-c
