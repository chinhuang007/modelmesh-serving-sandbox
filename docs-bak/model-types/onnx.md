---
title: ONNX - Open Neural Network Exchange
description: Open Neural Network Exchange model format
---

## Format

ONNX is an open format built to represent machine learning models. ONNX defines a common set of operators - the building blocks of machine learning and deep learning models - and a common file format to enable AI developers to use models with a variety of frameworks, tools, runtimes, and compilers.

ONNX defines a common file format that abstracts the building blocks of machine
learning and deep learning models. It is possible to convert models trained from
many different frameworks/tools to the ONNX format. See the
[ONNX tutorial documentation](https://github.com/onnx/tutorials#converting-to-onnx-format)
for some examples.

## Storage Layout

ONNX models may consist of a single file or a directory, both are supported.

**Simple**

```
<storage-path/model-name>
```

## Example Predictor

**Storage Layout**

```
s3://wml-serving-examples/
  onnx-models/example.onnx
```

**Predictor**

```
apiVersion: wmlserving.ai.ibm.com/v1
kind: Predictor
metadata:
  name: onnx-example
spec:
  modelType:
    name: onnx
  path: onnx-models/example.onnx
  storage:
    s3:
      secretKey: modelStorage
      bucket: wml-serving-examples
```
