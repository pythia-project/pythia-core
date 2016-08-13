#!/bin/sh
mkdir /tmp/work
cp /task/TestMono.cs /tmp/work/TestMono.cs
cd /tmp/work
mcs TestMono.cs
mono TestMono.exe
