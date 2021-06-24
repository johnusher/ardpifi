#! /usr/bin/python3

import base64
import sys

import numpy as np

for line in sys.stdin:
    if not line.strip() or line.strip().startswith('#'):
        continue
    data = base64.b64decode(line.strip())
    array = np.frombuffer(data, dtype=np.uint8).reshape((28, 28))
    print('array', array)
    sums = array.sum(axis=0, dtype=np.int64)
    print('sums', sums)
    print(np.min(sums))
