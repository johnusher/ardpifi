#! /usr/bin/python3
"""For each input line, decode into an array and classify the letter.

Lines are assumed to be 28x28 uint8 arrays, base64 encoded.

This script requires the tflite runtime:
https://www.tensorflow.org/lite/guide/python
"""

import base64
import os
import sys

import numpy as np
import tflite_runtime.interpreter as tflite


def main():
    interpreter = tflite.Interpreter(model_path())
    classifier = Classifier(interpreter)

    for line in sys.stdin:
        if not line.strip() or line.strip().startswith('#'):
            continue
        data = base64.b64decode(line.strip())
        
        array = np.frombuffer(data, dtype=np.uint8).reshape((28, 28)).transpose()
        output = classifier.classify(array)
        print(output)


def model_path():
    script_dir = os.path.dirname(__file__)
    return os.path.join(script_dir, 'model.tflite')
   
# Order of letters should match the one in train.py
_LETTERS = ['Other'] + list('CDMNOS')

class Classifier:
    def __init__(self, interpreter):
        interpreter.allocate_tensors()
        self._input_details = interpreter.get_input_details()
        self._output_details = interpreter.get_output_details()
        self._interpreter = interpreter

    def classify(self, array):
        # TODO: The data might have to be transposed.
        # TODO: The image should be blurred a bit to be more similar to the
        # training dataset.
        # Get the values between 0 and 1.
        normalized = array.astype(np.float32) / array.max()
        # Stretch shape to [1, 28, 28, 1]
        normalized = np.expand_dims(normalized, 2)
        normalized = np.expand_dims(normalized, 0)
        self._interpreter.set_tensor(self._input_details[0]['index'], normalized)
        self._interpreter.invoke()

        logits = self._interpreter.get_tensor(self._output_details[0]['index'])[0]
        probs = _softmax(logits)
        prob, letter = max(zip(probs, _LETTERS))
        return prob, letter
       

def _softmax(x):
    return np.exp(x) / sum(np.exp(x))

if __name__ == '__main__':
    main()

