#!/usr/bin/python3
# -*- coding: utf-8 -*-

import json
import sys

sys.path.append('/task/static')
from lib import pythia

# Setup working directory
pythia.setupWorkingDirectory('/tmp/work')

# Read input data and fill skeleton files
data = sys.stdin.read().rstrip('\0')
input = json.loads(data)
pythia.fillSkeletons('/task/skeleton', '/tmp/work', input['fields'])

# Save task id
with open('/tmp/work/tid', 'w', encoding='utf-8') as file:
  file.write(input['tid'])
