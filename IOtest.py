#! /usr/bin/python3
"""
python -m py_compile IOtest.py
"""

import base64
import sys

import numpy as np

from PIL import Image



# for line in sys.stdin:
#     # sys.stdout.write(line.strip())
#     data = base64.b64decode(line.strip())
#     sys.stdout.write(data)

#     # base64_bytes = base64.b64encode(data)
#     # sys.stdout.write(data)
#     # base64_message = base64_bytes.encode('ascii')

#     # sys.stdout.write("Hello, %s. \n" %base64_message)
#     # print(base64_message)
#     # print("ff")

    
    
for line in sys.stdin:
    
    if not line.strip() or line.strip().startswith('#'):
        continue
    
    # sys.stdout.write(line.strip())

    data = base64.b64decode(line.strip())
    
    # print(data)
    # sys.stdout.w
    array = np.frombuffer(data, dtype=np.uint8).reshape((28, 28))
    sums = array.sum(axis=0, dtype=np.int64)
    # print(np.min(sums))
    print(np.max(sums))
