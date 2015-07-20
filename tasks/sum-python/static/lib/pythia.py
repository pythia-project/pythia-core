# -*- coding: utf-8 -*-
#
# Pythia library for unit testing-based tasks
# Author: Sébastien Combéfis <sebastien@combefis.be>
#
# Copyright (C) 2015, Computer Science and IT in Education ASBL
# Copyright (C) 2015, École Centrale des Arts et Métiers
#
# This program is free software: you can redistribute it and/or modify
# under the terms of the GNU General Public License as published by
# the Free Software Foundation, version 2 of the License, or
#  (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
# General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.

import ast
import csv
import os
import random
import re
import stat

def setupWorkingDirectory(dest):
  '''Setup working directory'''
  # Create working directory if not existing yet
  if not os.path.exists(dest):
    os.makedirs(dest)
  os.chmod(dest, stat.S_IRWXU | stat.S_IRWXG | stat.S_IRWXO)
  # Create directories for input and output data for unit tests
  inputdir = '{}/input'.format(dest)
  os.makedirs(inputdir)
  os.chmod(inputdir, stat.S_IRWXU | stat.S_IRWXG | stat.S_IRWXO)
  outputdir = '{}/output'.format(dest)
  os.makedirs(outputdir)
  os.chmod(outputdir, stat.S_IRWXU | stat.S_IRWXG | stat.S_IRWXO)


def fillSkeletons(src, dest, fields):
  '''Fill skeleton files containing placeholders with specified values'''
  for (root, dirs, files) in os.walk(src):
    dirdest = dest + root[len(src):]
    # Handle each file in the src directory
    for filename in files:
      filesrc = '{}/{}'.format(root, filename)
      filedest = '{}/{}'.format(dirdest, filename)
      # Open the file
      with open(filesrc, 'r', encoding='utf-8') as file:
        content = file.read()
      # Replace each placeholder with the specified value
      for field, value in fields.items():
        regex = re.compile('@([^@]*)@{}@([^@]*)@'.format(field))
        for prefix, postfix in set(regex.findall(content)):
          rep = '\n'.join([prefix + v + postfix for v in value.splitlines()])
          content = content.replace('@{}@{}@{}@'.format(prefix, field, postfix), rep)
      # Create the new file
      with open(filedest, 'w', encoding='utf-8') as file:
        file.write(content)
      os.chmod(filedest, stat.S_IRWXU | stat.S_IRWXG | stat.S_IROTH)


def generateTestData(dest, filename, config):
  '''Generate input data for unit tests'''
  # Open destination file
  filedest = '{}/{}'.format(dest, filename)
  with open(filedest, 'w', encoding='utf-8') as file:
    writer = csv.writer(file, delimiter=';', quotechar='"')
    # Write predefined tests to the specified file if any
    if 'predefined' in config:
      for data in config['predefined']:
        writer.writerow(ast.literal_eval(data['data']))
    # Create an array of generators as specified by configuration
    # and write random tests to the specified file if any
    if 'random' in config:
      random = config['random']
      generator = ArrayGenerator([RandomGenerator.build(descr) for descr in random['args']])
      for i in range(random['n']):
        writer.writerow(generator.generate())
  os.chmod(filedest, stat.S_IRWXU | stat.S_IRWXG | stat.S_IROTH)


class RandomGenerator:
  '''Class to generate random test data'''
  def generate(self):
    '''Generate one random value'''
    return None

  def build(description):
    '''Build a random generator according to a textual description'''
    # int(a,b)
    m = re.match('^int\((-{0,1}[1-9][0-9]*),(-{0,1}[1-9][0-9]*)\)', description)
    if not m is None:
      return IntRandomGenerator(int(m.group(1)), int(m.group(2)))
    # default case
    return RandomGenerator()


class ArrayGenerator(RandomGenerator):
  '''Class to generate an array of random values with specified generators'''
  def __init__(self, generators):
    self.generators = generators

  def generate(self):
    return [g.generate() for g in self.generators]


class IntRandomGenerator(RandomGenerator):
  '''Class to generate a random integer comprised between two bounds'''
  def __init__(self, lowerbound, upperbound):
    self.lowerbound = lowerbound
    self.upperbound = upperbound

  def generate(self):
    return random.randint(self.lowerbound, self.upperbound)


class TestSuite:
  '''Basic test suite'''
  def __init__(self, inputfile):
    self.inputfile = inputfile

  def check(self, data):
    try:
      answer = self.studentCode(data)
    except Exception as e:
      return 'exception:{}'.format(e)
    res = self.moreCheck(answer, data)
    if res != 'passed':
      return res
    return 'checked:{}'.format(answer)

  def studentCode(self, data):
    return None

  def moreCheck(self, answer, data):
    return 'passed'

  def parseTestData(self, data):
    return tuple(data)

  def run(self, dest, filename):
    # Create the results file
    with open('{}/{}'.format(dest, filename), 'w', encoding='utf-8') as result:
      # Read and run tests
      with open(self.inputfile, 'r', encoding='utf-8') as file:
        reader = csv.reader(file, delimiter=';', quotechar='"')
        for row in reader:
          res = self.check(self.parseTestData(row))
          result.write('{}\n'.format(res))


class FeedbackSuite:
  '''Basic feedback suite'''
  def __init__(self, inputfile, resultfile, config):
    self.inputfile = inputfile
    self.resultfile = resultfile
    self.config = config

  def check(self, data, actual):
    check = False
    try:
      # Compare student and teacher answers
      expected = self.teacherCode(data)
      if str(expected) == actual:
        check = True
    except Exception as e:
      expected = None
    return (check, expected)

  def teacherCode(self, data):
    return None

  def parseTestData(self, data):
    return tuple(data)

  def generate(self):
    verdict = True
    feedback = {}
    succeeded = 0
    total = 0
    with open(self.inputfile, 'r', encoding='utf-8') as file:
      reader = csv.reader(file, delimiter=';', quotechar='"')
      with open(self.resultfile, 'r', encoding='utf-8') as file:
        for row in reader:
          actual = file.readline().strip()
          tokens = actual.split(':')
          if tokens[0] == 'checked':
            input = self.parseTestData(row)
            (check, expected) = self.check(input, tokens[1])
            if check:
              succeeded += 1
            elif not 'example' in feedback:
              verdict = False
              feedback['example'] = {'input': input, 'expected': expected, 'actual': tokens[1]}
              if total < len(self.config):
                message = self.config[total]['feedback']
                if tokens[1] in message:
                  feedback['message'] = message[tokens[1]]
          else:
            verdict = False
          total += 1
    feedback['stats'] = {'succeeded': succeeded, 'total': total}
    return (verdict, feedback)
