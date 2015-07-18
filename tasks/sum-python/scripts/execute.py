#!/usr/bin/python3
# -*- coding: utf-8 -*-

import sys

sys.path.append('/task/static')
from lib import pythia

sys.path.append('/tmp/work')
import q1

class TaskTestSuite(pythia.TestSuite):
  def __init__(self):
    pythia.TestSuite.__init__(self, '/tmp/work/input/data.csv')

  def studentCode(self, data):
    return q1.sum(*data)

  def parseTestData(self, data):
    return tuple(int(x) for x in data)

TaskTestSuite().run('/tmp/work/output', 'data.res')
