#!/bin/sh
mkdir /tmp/work
cp /task/Hello.cs /tmp/work/Hello.cs
cd /tmp/work
mcs Hello.cs
mono Hello.exe
