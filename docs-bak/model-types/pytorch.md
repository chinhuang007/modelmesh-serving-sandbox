---
title: PyTorch
description: PyTorch ScriptModule
---

## Format

The serving runtime uses the C++ distribution of PyTorch called
[`LibTorch`](https://pytorch.org/cppdocs/installing.html)
to support high performance inference. This library requires that models be
serialized as a `ScriptModule` composed with
[`TorchScript`](https://pytorch.org/cppdocs/#torchscript).
Refer to PyTorch's documentation on
[loading a `TorchScript` model in C++](https://pytorch.org/tutorials/advanced/cpp_export.html)
for details on converting a PyTorch model in Python to an exported
`ScriptModule`.

## Configuration

A `pytorch` model currently requires configuration in a file called
`config.pbtxt`. The configuration must specify the shape of the input and
output tensors and the maximum batch size. Following is an example
of this configuration:

```
platform: "pytorch_libtorch"
max_batch_size: 1
input [
  {
    name: "INPUT__0"
    data_type: TYPE_FP32
    dims: [3,32,32]
  }
]
output [
  {
    name: "OUTPUT__0"
    data_type: TYPE_FP32
    dims: [10]
  }
]
```

For details on the configuration specification, refer to
[Triton's Model Configuration documentation](https://github.com/triton-inference-server/server/blob/r20.12/docs/model_configuration.md).

## Storage Layout

**Simple**

The serialized `ScriptModule` must be placed in a numeric sub-directory that
is a sibling of the required configuration file.

```
<storage-path>/
├── config.pbtxt
└── 1/
    └── model.pt
```

**Triton Native**

Currently, support for PyTorch is provided by the Triton runtime. Model-Mesh Serving
also supports the native Triton file layout with multiple version
directories.

```
<storage-path>/
├── config.pbtxt
├── <version>/
│   └── model.pt
└── <version>/
    └── model.pt
```

For details on Triton's Model Repository structure, refer to
[Triton's Model Repository documentation](https://github.com/triton-inference-server/server/blob/r20.12/docs/model_repository.md).

<InlineNotification>

**Note** Only one of the numeric version directories will be used because WML
Serving handles versioning in a higher layer.

</InlineNotification>

## Example

**Storage Layout**

```
s3://wml-serving-examples/
└── pytorch-model
    ├── config.pbtxt
    └── 1/
        └── model.pt
```

**Predictor**

```
apiVersion: wmlserving.ai.ibm.com/v1
kind: Predictor
metadata:
  name: pytorch-example
spec:
  modelType:
    name: pytorch
  path: pytorch-model
  storage:
    s3:
      secretKey: modelStorage
      bucket: wml-serving-examples
```
