# baseline cnn model for mnist
from numpy import mean
from numpy import std
from matplotlib import pyplot
from sklearn.model_selection import KFold
from tensorflow.keras.utils import to_categorical
from tensorflow.keras.models import Sequential
from tensorflow.keras import layers
import tensorflow as tf
import tensorflow_datasets as tfds

_LETTERS = 'CDMNOS'
_LABELS = [ord(c) - ord('A') + 1 for c in _LETTERS]
_OTHER_LABEL = 0
_NUM_CLASSES = 37
_BATCH_SIZE = 32


# load train and test dataset
def load_dataset():
  # load dataset
  (train_ds, test_ds), ds_info = tfds.load(
      name='emnist/letters',
      split=['train', 'test'],
      shuffle_files=True,
      with_info=True,
      as_supervised=True)

  train_ds = prepare(train_ds, ds_info)
  test_ds = prepare(test_ds, ds_info)

  return (train_ds, test_ds), ds_info

def prepare(dataset, ds_info):
  wanted_labels = tf.constant(_LABELS, dtype=tf.int64)

  @tf.function
  def is_wanted(image, label):
    return tf.math.reduce_any(label == wanted_labels)

  @tf.function
  def map_label(image, label):
    return image, tf.cond(
      tf.math.reduce_any(label == wanted_labels),
      lambda: tf.argmax(tf.equal(wanted_labels, label)),
      lambda: tf.constant(_OTHER_LABEL, dtype=tf.int64))

  return (dataset
    # .filter(is_wanted)
    .cache()
    .shuffle(1000)
    .map(map_label)
    .map(_prep_pixels))

@tf.function
def _prep_pixels(image, label):
  image = tf.cast(image, tf.float32)
  image = image / 255.0
  image = tf.transpose(image, perm=[1, 0, 2])
  return image, label

# define cnn model
def define_model():
  data_augmentation = tf.keras.Sequential([
    layers.experimental.preprocessing.RandomRotation(1),
  ])
  model = Sequential()
  model.add(data_augmentation)
  model.add(layers.Conv2D(16, (3, 3), activation='relu', kernel_initializer='he_uniform', input_shape=(28, 28, 1)))
  model.add(layers.MaxPooling2D((2, 2)))
  model.add(layers.Flatten())
  model.add(layers.Dense(50, activation='relu', kernel_initializer='he_uniform'))
  model.add(layers.Dense(_NUM_CLASSES))
  # compile model
  model.compile(
      optimizer=tf.keras.optimizers.Adam(0.001),
      loss=tf.keras.losses.SparseCategoricalCrossentropy(from_logits=True),
      metrics=['accuracy'],
  )
  return model

# evaluate a model using k-fold cross-validation
def evaluate_model(train_ds, test_ds):
  train_ds = train_ds.batch(256).prefetch(tf.data.experimental.AUTOTUNE)
  test_ds = test_ds.batch(256).prefetch(tf.data.experimental.AUTOTUNE)
  scores, histories = [], []
  # define model
  model = define_model()
  # fit model
  history = model.fit(
    train_ds,
    epochs=50,
    validation_data=test_ds,
    class_weight=_label_weights(_NUM_CLASSES, _LABELS),
    verbose=0)
  # evaluate model
  _, acc = model.evaluate(test_ds, verbose=0)
  print('> %.3f' % (acc * 100.0))
  # stores scores
  scores.append(acc)
  histories.append(history)
  return scores, histories

# plot diagnostic learning curves
def summarize_diagnostics(histories):
    for i in range(len(histories)):
        # plot loss
        pyplot.subplot(2, 1, 1)
        pyplot.title('Cross Entropy Loss')
        pyplot.plot(histories[i].history['loss'], color='blue', label='train')
        pyplot.plot(histories[i].history['val_loss'], color='orange', label='test')
        # plot accuracy
        pyplot.subplot(2, 1, 2)
        pyplot.title('Classification Accuracy')
        pyplot.plot(histories[i].history['accuracy'], color='blue', label='train')
        pyplot.plot(histories[i].history['val_accuracy'], color='orange', label='test')
    pyplot.show()

# summarize model performance
def summarize_performance(scores):
    # print summary
    print('Accuracy: mean=%.3f std=%.3f, n=%d' % (mean(scores)*100, std(scores)*100, len(scores)))
    # box and whisker plots of results
    pyplot.boxplot(scores)
    pyplot.show()

def show_dataset():
  (train_ds, test_ds), ds_info = load_dataset()
  tfds.visualization.show_examples(train_ds, ds_info)
  
# run the test harness for evaluating a model
def run_test_harness():
    # load dataset
    (train_ds, test_ds), ds_info = load_dataset()
    # evaluate model
    scores, histories = evaluate_model(train_ds, test_ds)
    # learning curves
    summarize_diagnostics(histories)
    # summarize estimated performance
    summarize_performance(scores)

def _label_weights(total_num_labels, wanted_labels):
  """Reweight the label costs to account over represented OTHER label."""
  label_weights = {
    l: 1.0 / (len(wanted_labels) + 1)
    for l in range(len(wanted_labels) + 1)
  }
  label_weights[_OTHER_LABEL] =  1.0 / ((len(wanted_labels) + 1) * (total_num_labels - len(_LABELS)))
  return label_weights

# entry point, run the test harness
if __name__ == "__main__":
  run_test_harness()
