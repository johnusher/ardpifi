#! /usr/bin/python3
"""For each input line, decode into an array and print the smallest column sum.

Lines are assumed to be 28x28 uint8 arrays, base64 encoded.
"""

import base64
import sys



for line in sys.stdin:
    if not line.strip() or line.strip().startswith('#'):
        continue
    data = base64.b64decode(line.strip())

    print(data.decode("utf-8"))
