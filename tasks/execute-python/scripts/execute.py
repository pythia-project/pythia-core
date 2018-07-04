#!/usr/bin/python3
# -*- coding: utf-8 -*-

import json
import subprocess

# Execute script.py
proc = subprocess.run(['python3', '/tmp/work/script.py'], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
result = {
    'returncode': proc.returncode,
    'stdout': proc.stdout.decode('utf-8'),
    'stderr': proc.stderr.decode('utf-8')
}
print(json.dumps(result))
