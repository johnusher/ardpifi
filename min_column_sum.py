#! /usr/bin/python3
"""For each input line, decode into an array and print the smallest column sum.

Lines are assumed to be 28x28 uint8 arrays, base64 encoded.
"""

import base64
import sys

import numpy as np

from PIL import Image
# import matplotlib.pyplot as plt


for line in sys.stdin:
    if not line.strip() or line.strip().startswith('#'):
        continue
    data = base64.b64decode(line.strip())
    
    array = np.frombuffer(data, dtype=np.uint8).reshape((28, 28)).transpose()
    sums = array.sum(axis=0, dtype=np.int64)
    
    print(np.min(sums))    


    # Image.fromarray(array,mode="1").save("fromPy3.png") # doesn't show image
    # print(array)
    # print("fin1")    


    Image.fromarray(array * 255, mode='L').save('asdf.bmp')
    # plt.imshow(array) #Needs to be in row,col order
    # plt.axis('off')
    # plt.savefig("fromPy4.png")   # works! but takes ages, and is wrong size, and is rotated with row, col in wrong order
    print("fin2")    
   
