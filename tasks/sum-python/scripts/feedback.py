#!/usr/bin/python3
# -*- coding: utf-8 -*-

import ast
import csv
import json
import sys

sys.path.append('/task/static')
from lib import pythia

def sum(a, b):
  return a + b

class TaskFeedbackSuite(pythia.FeedbackSuite):
  def __init__(self, config):
    pythia.FeedbackSuite.__init__(self, '/tmp/work/input/data.csv', '/tmp/work/output/data.res', config)

  def teacherCode(self, data):
    return sum(*data)

  def parseTestData(self, data):
    return tuple(int(x) for x in data)

# Read test configuration
config = []
with open('/task/config/test.json', 'r', encoding='utf-8') as file:
  content = file.read()
  config = json.loads(content)
  config = config['predefined']
(verdict, feedback) = TaskFeedbackSuite(config).generate()

# Retrieve task id
with open('/tmp/work/tid', 'r', encoding='utf-8') as file:
  tid = file.read()
print({'tid': tid, 'status': 'success' if verdict else 'failed', 'feedback': feedback})
