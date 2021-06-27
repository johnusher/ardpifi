#! /usr/bin/python3
""" echo "mu" | python IOtest.py  
"""

import sys
import base64

for line in sys.stdin:
    # sys.stdout.write(line)
    data = base64.b64decode(line.strip())
    # data = base64.b64decode(line)

    # sys.stdout.write("Hello, %s. \n" % line)
    # sys.stdout.write("Hello, %s. \n" % data)

    base64_bytes = base64.b64encode(data)

    dataOut = base64.b64encode(base64_bytes)

    # sys.stdout.write("Hello, %s. \n" % dataOut)
    
    base64_message = base64_bytes.decode('ascii')

    # print(base64_message)

    sys.stdout.write(base64_message)
    


# name = sys.stdin.readline("")
#         print("Hello, %s. \n" % name)
#         sys.stdout.flush() # there's a stdout buffer
