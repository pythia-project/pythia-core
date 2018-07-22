#!/bin/sh
mkdir /tmp/work
cp /task/hello.c /tmp/work/hello.c
cd /tmp/work
gcc hello.c -o hello
./hello
