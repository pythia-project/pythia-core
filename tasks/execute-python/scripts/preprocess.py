#!/usr/bin/python3
# -*- coding: utf-8 -*-

import os
import stat
import sys

# Setup working directory
workdir = '/tmp/work'
if not os.path.exists(workdir):
    os.makedirs(workdir)
os.chmod(workdir, stat.S_IRWXU | stat.S_IRWXG | stat.S_IRWXO)

# Read input data
data = sys.stdin.read().rstrip('\0')

# Create python script file
scriptfile = '/tmp/work/script.py'
with open(scriptfile, 'w', encoding='utf-8') as file:
    file.write(data)
os.chmod(scriptfile, stat.S_IRWXU | stat.S_IRWXG | stat.S_IRWXO)
