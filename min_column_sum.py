#! /usr/bin/python3
"""For each input line, decode into an array and print the smallest column sum.

Lines are assumed to be 28x28 uint8 arrays, base64 encoded.
"""

import base64
import sys

import numpy as np

for line in sys.stdin:
    if not line.strip() or line.strip().startswith('#'):
        continue
    data = base64.b64decode(line.strip())
    array = np.frombuffer(data, dtype=np.uint8).reshape((28, 28))
    sums = array.sum(axis=0, dtype=np.int64)
    print(np.min(sums))
