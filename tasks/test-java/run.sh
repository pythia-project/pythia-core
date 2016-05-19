#!/bin/sh
PATH=/usr/lib/jvm/java-7-openjdk-i386/bin:$PATH
mkdir /tmp/work
cp /task/TestJava.java /tmp/work/TestJava.java
cd /tmp/work
javac TestJava.java
java TestJava
