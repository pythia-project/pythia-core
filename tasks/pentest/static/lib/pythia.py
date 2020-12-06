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


def clean(string):
    string = string.replace('&', '&#38;')
    string = string.replace('>', '&#62;')
    string = string.replace('<', '&#60;')
    string = string.replace('"', '&#34;')
    return string


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


class Preprocessor:
    '''Class to preprocess student's answers'''
    def __init__(self, fields):
        self.__fields = fields

    def run(self, dest, filename):
        result = self.preprocess(self.__fields)
        if result != None:
            with open('{}/{}'.format(dest, filename), 'w', encoding='utf-8') as file:
                file.write(result)

    def preprocess(self, fields):
        return None


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
        intpattern = '-{0,1}[1-9][0-9]*'
        floatpattern = '-{0,1}[1-9][0-9]*(?:\.[0-9]*[1-9]){0,1}'
        # int(a,b)
        m = re.match('^int\(({0}),({0})\)$'.format(intpattern), description)
        if m is not None:
            return IntRandomGenerator(int(m.group(1)), int(m.group(2)))
        # bool
        if 'bool' == description:
            return BoolRandomGenerator()
        # float(a,b)
        m = re.match('^float\(({0}),({0})\)$'.format(floatpattern), description)
        if m is not None:
            return FloatRandomGenerator(float(m.group(1)), float(m.group(2)))
        # str(a,b)
        m = re.match('^str\(({0}),({0})\)$'.format(intpattern), description)
        if m is not None:
            return StringRandomGenerator(int(m.group(1)), int(m.group(2)))
        # enum(list)
        m = re.match('^enum\((.+)\)$', description)
        if m is not None:
            return EnumRandomGenerator(m.group(1).split(','))
        # set(a,b)[config]
        m = re.match('^set\(({0}),({0})\)\[(.+)\]$'.format(intpattern), description)
        if m is not None:
            return SetRandomGenerator(int(m.group(1)), int(m.group(2)), m.group(3))
        # default case
        return RandomGenerator()


class ArrayGenerator(RandomGenerator):
    '''Class to generate an array of random values with specified generators'''
    def __init__(self, generators):
        self.__generators = generators

    def generate(self):
        return [g.generate() for g in self.__generators]


class IntRandomGenerator(RandomGenerator):
    '''Class to generate a random integer comprised between two bounds'''
    def __init__(self, lowerbound, upperbound):
        self.__lowerbound = lowerbound
        self.__upperbound = upperbound

    def generate(self):
        return random.randint(self.__lowerbound, self.__upperbound)


class BoolRandomGenerator(RandomGenerator):
    '''Class to generate a random boolean value'''
    def __init__(self):
        pass

    def generate(self):
        return random.randint(0, 1) == 0


class FloatRandomGenerator(RandomGenerator):
    '''Class to generate a random float comprised between two bounds'''
    def __init__(self, lowerbound, upperbound):
        self.__lowerbound = lowerbound
        self.__upperbound = upperbound

    def generate(self):
        return random.uniform(self.__lowerbound, self.__upperbound)


class StringRandomGenerator(RandomGenerator):
    '''Class to generate a random string with a specified number of characters'''
    def __init__(self, minchars, maxchars):
        self.__minchars = minchars
        self.__maxchars = maxchars

    def generate(self):
        letters = 'abcdefghijklmnopqrst0123456789'
        n = random.randint(self.__minchars, self.__maxchars)
        result = ''
        for i in range(n):
            result += letters[random.randint(0, len(letters) - 1)]
        return result


class EnumRandomGenerator(RandomGenerator):
    '''Class to generate a random value from an enumeration'''
    def __init__(self, values):
        self.__values = values

    def generate(self):
        return self.__values[random.randint(0, len(self.__values) - 1)]


class SetRandomGenerator(RandomGenerator):
    '''Class to generate a random set with a specified number of elements of a specified type'''
    def __init__(self, minlen, maxlen, config):
        self.__minlen = minlen
        self.__maxlen = maxlen
        self.__config = config

    def generate(self):
        n = random.randint(self.__minlen, self.__maxlen)
        generator = RandomGenerator.build(self.__config)
        result = set()
        for i in range(n):
            result.add(generator.generate())
        return result


class NoAnswerException(Exception):
    '''Exception representing the case where no answer was provided'''
    pass


class UndeclaredException(Exception):
    '''Exception representing the case where a variable was not declared'''
    def __init__(self, name):
        self.__name = name

    @property
    def name(self):
        return self.__name


class BadTypeException(Exception):
    '''Exception representing the case where a variable has the wrong type'''
    def __init__(self, name, actualtype, expectedtype):
        self.__name = name
        self.__actualtype = actualtype
        self.__expectedtype = expectedtype

    @property
    def name(self):
        return self.__name

    @property
    def actualtype(self):
        return self.__actualtype

    @property
    def expectedtype(self):
        return self.__expectedtype


class TestSuite:
    '''Basic test suite'''
    def __init__(self, inputfile):
        self.__inputfile = inputfile

    def check(self, data):
        try:
            answer = self.studentCode(data)
        except NoAnswerException as e:
            return 'exception:noanswer'
        except UndeclaredException as e:
            return 'exception:undeclared:{}'.format(e.name)
        except BadTypeException as e:
            return 'exception:badtype:{}:{}:{}'.format(e.name, e.actualtype, e.expectedtype)
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
            with open(self.__inputfile, 'r', encoding='utf-8') as file:
                reader = csv.reader(file, delimiter=';', quotechar='"')
                for row in reader:
                    res = self.check(self.parseTestData(row))
                    result.write('{}\n'.format(res))


class FeedbackSuite:
    '''Basic feedback suite'''
    def __init__(self, stderr, stdout, inputfile=None, resultfile=None, config=None):
        self.__stderr = stderr
        self.__stdout = stdout
        self.__inputfile = inputfile
        self.__resultfile = resultfile
        self.__config = config

    def check(self, data, actual):
        check = False
        try:
            # Compare student and teacher answers
            expected = self.teacherCode(data)
            if self.compare(expected, actual):
                check = True
        except Exception as e:
            expected = None
        return (check, expected)

    def compare(self, expected, actual):
        return str(expected) == actual

    def teacherCode(self, data):
        return None

    def parseTestData(self, data):
        return tuple(data)

    def checkStdout(self, out):
        return (True, None)

    def generate(self):
        verdict = True
        feedback = {}
        # Check stderr
        if os.path.exists(self.__stderr):
            with open(self.__stderr, 'r', encoding='utf-8') as file:
                content = file.read().strip()
                if content != '':
                    verdict = False
                    feedback['message'] = '<p>Une erreur s\'est produite lors de l\'exécution de votre programme :</p><pre>' + clean(content) + '</pre>'
                    feedback['score'] = 0
        # Check stdout
        if verdict and self.__stdout != None and os.path.exists(self.__stdout):
            with open(self.__stdout, 'r', encoding='utf-8') as file:
                (check, msg) = self.checkStdout(file.read())
                if not check:
                    verdict = False
                    feedback['message'] = clean(msg)
                    feedback['score'] = 0
        # Check the unit testing-based tests
        if verdict and self.__inputfile != None and self.__resultfile != None:
            succeeded = 0
            total = 0
            with open(self.__inputfile, 'r', encoding='utf-8') as file:
                reader = csv.reader(file, delimiter=';', quotechar='"')
                with open(self.__resultfile, 'r', encoding='utf-8') as file:
                    for row in reader:
                        actual = file.readline().strip()
                        tokens = actual.split(':')
                        # Student code produced an answer
                        if tokens[0] == 'checked':
                            input = self.parseTestData(row)
                            (check, expected) = self.check(input, tokens[1])
                            if check:
                                succeeded += 1
                            elif 'example' not in feedback:
                                verdict = False
                                feedback['example'] = {'input': str(input), 'expected': str(expected), 'actual': tokens[1]}
                                if total < len(self.__config) and 'feedback' in self.__config[total]:
                                    message = self.__config[total]['feedback']
                                    if tokens[1] in message:
                                        feedback['message'] = message[tokens[1]]
                                    elif '*' in message:
                                        feedback['message'] = message['*']
                        # An exception occured
                        elif tokens[0] == 'exception':
                            verdict = False
                            feedback['message'] = 'An error occured with your code'
                            if tokens[1] == 'noanswer':
                                feedback['message'] += ': You did not answer the question'
                            elif tokens[1] == 'undeclared':
                                feedback['message'] += ': Missing variable <code>{}</code>'.format(tokens[2])
                            elif tokens[1] == 'badtype':
                                feedback['message'] += ': Bad type for variable <code>{}</code> : found <code>{}</code>, but <code>{}</code> expected'.format(tokens[2], clean(tokens[3]), clean(tokens[4]))
                            else:
                                feedback['message'] += ': {}'.format(tokens[1])
                        # Unexpected error
                        else:
                            verdict = False
                            feedback['message'] = 'An error occured with your code: {}'.format(tokens[0])
                        total += 1
            feedback['stats'] = {'succeeded': succeeded, 'total': total}
            feedback['score'] = succeeded / total
        return (verdict, feedback)
