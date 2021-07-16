---
title: TensorFlow
description: TensorFlow v1 or v2 models saved as a SavedModel or GraphDef
---

## Format

Both v1 and v2 TensorFlow models are supported as either a SavedModel or a
serialized `GraphDef`. The SavedModel format is the default for TF2.x and is
preferred because it includes the model weights and configuration. Refer to
the TensorFlow documentation on
["Using the SavedModel format"](https://www.tensorflow.org/guide/saved_model)
or the [specification of the `GraphDef` protocol buffer message](https://www.tensorflow.org/api_docs/python/tf/compat/v1/GraphDef)
for details on these formats and their serialization.

## Configuration

**SavedModel Format**

A SavedModel contains all needed configuration to serve the model. The
`config.pbtxt` file is not necessary, but can be provided for additional
configuration. For details on the configuration specification, refer to
[Triton's Model Configuration documentation](https://github.com/triton-inference-server/server/blob/r20.12/docs/model_configuration.md).

**`GraphDef` Format**

A `GraphDef` model currently requires configuration in a file called
`config.pbtxt`. The configuration must specify the shape of the input and
output tensors and the maximum batch size. Following is an example
of this configuration:

```
platform: "tensorflow_graphdef"
max_batch_size: 4
input [
  {
    name: "input0"
    data_type: TYPE_FP32
    dims: [3,32,32]
  }
]
output [
  {
    name: "output0"
    data_type: TYPE_FP32
    dims: [10]
  }
]
```

For details on the configuration specification, refer to
[Triton's Model Configuration documentation](https://github.com/triton-inference-server/server/blob/r20.12/docs/model_configuration.md).

## Storage Layout - SavedModel

If the optional configuration file is not needed, there are multiple
supported storage layouts as referenced below.

**Simple**

The storage path can point directly to the contents of the `SavedModel`:

```
<storage-path>/
└── <saved-model-files>
```

**Directory**

The storage path can also point to a directory containing the `SavedModel` directory:

```
<storage-path>/
└── model.savedmodel/
    └── <saved-model-files>
```

The directory does not have to be called `model.savedmodel`.

**Triton Native**

Currently, support for TensorFlow is provided by the Triton runtime. WML
Serving also supports the native Triton file layout with multiple version
directories.

```
<storage-path>/
├── [config.pbtxt] (optional)
├── <version>/
│   └── model.savedmodel
│       └── <saved-model-files>
└── <version>/
    └── model.savedmodel
        └── <saved-model-files>
```

For details on Triton's Model Repository structure, refer to
[Triton's Model Repository documentation](https://github.com/triton-inference-server/server/blob/r20.12/docs/model_repository.md).

<InlineNotification>

**Note** Only one of the numeric version directories will be used because WML
Serving handles versioning in a higher layer.

</InlineNotification>

## Storage Layout - GraphDef

The configuration file is required for a `GraphDef` model, so the repository
structure options are more limited.

**Simple**

The simple layout for a `GraphDef` model currently requires the
`config.pbtxt` to live next to a numeric sub-directory containing the
serialized `GraphDef`.

```
<storage-path>/
├── config.pbtxt
└── 1/
    └── model.graphdef
```

**Triton Native**

Currently, support for TensorFlow is provided by the Triton runtime. WML
Serving also supports the native Triton file layout with multiple version
directories.

```
<storage-path>/
├── config.pbtxt
├── <version>/
│   └── model.graphdef
└── <version>/
    └── model.graphdef
```

For details on Triton's Model Repository structure, refer to
[Triton's Model Repository documentation](https://github.com/triton-inference-server/server/blob/r20.12/docs/model_repository.md).

<InlineNotification>

**Note** Only one of the numeric version directories will be used because WML
Serving handles versioning in a higher layer.

</InlineNotification>

## Example Predictor

The following example is using the `SavedModel` format with the simple
repository layout.

**Storage Layout**

```
s3://wml-serving-examples/tensorflow-model/
├── variables/
└── saved_model.pb
```

**Predictor**

```
apiVersion: wmlserving.ai.ibm.com/v1
kind: Predictor
metadata:
  name: tensorflow-example
spec:
  modelType:
    name: tensorflow
  path: tensorflow-model
  storage:
    s3:
      secretKey: modelStorage
      bucket: wml-serving-examples
```
