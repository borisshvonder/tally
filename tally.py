#!/usr/bin/env python3
# -*- coding: UTF-8 -*-

import unittest
import sys

class World:
  def __init__(self):
    self._reset()

  def _reset(self):
    self.__console = Console()


class Console:

  def cmdline_args(self):
    "Get startup command line args"
    return sys.argv

  def out(self, *objects, **kwargs):
    print(*objects, **kwargs)
      

global world
world=World()

def print_hello():
  world.console.out("Hello,", " world!")

class Test_print_hello(unittest.TestCase):

  def setUp(self):
    self.console_out=[]
    world.console.out = lambda *args, **kwargs: self._out(*args, **kwargs)

  def _out(self, *args, **kwargs):
    self.console_out.extend(args)
    if kwargs: self.console_out.append(kwargs)

  def tearDown(self):
    world._reset()

  def test_hello(self):
    print_hello()
    self.assertEqual(["Hello,", " world!"], self.console_out)

def main():
  print_hello()

if __name__=='__main__':
  main()
