#!/bin/sh
PATH=/usr/lib/jvm/java-7-openjdk-i386/bin:$PATH
mkdir /tmp/work
cp /task/Hello.java /tmp/work/Hello.java
cd /tmp/work
javac Hello.java
java Hello
