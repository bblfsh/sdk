#!/usr/bin/env python3

# This is a python script used to generate test data for the diff library. It uses pairs of files
# to get diffed acquired from
# https://github.com/vmarkovtsev/treediff/blob/49356e7f85c261ed88cf46326791765c58c22b5b/dataset/flask.tar.xz
# It uses https://github.com/bblfsh/client-go#Installation to convert python sources into
# uast yaml files.
# It needs to be configured with proper DATASET_PATH which is a path to an unpacked
# flask.tar.xz file.

import os
from glob import glob


DATASET_PATH = "~/data/sourced/treediff/python-dataset"
pwd = os.path.expanduser(DATASET_PATH)

testnames = [e for e in open("smalltest.txt", "r").read().split() if e]


def get_src(name):
    return glob("{}/{}_before*.src".format(pwd, name))[0]


def get_dst(name):
    return glob("{}/{}_after*.src".format(pwd, name))[0]

i = 0
for src, dst in ((get_src(name), get_dst(name)) for name in testnames):
    print(i, src)
    os.system("bblfsh-cli -l python {} -o yaml > {}_src.uast".format(src, i))
    os.system("bblfsh-cli -l python {} -o yaml > {}_dst.uast".format(dst, i))
    i += 1
