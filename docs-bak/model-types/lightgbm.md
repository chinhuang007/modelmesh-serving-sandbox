---
title: LightGBM
description: LightGBM gradient boosting models
---

## Format

LightGBM models must be serialised using the
[Booster.save_model() method](https://lightgbm.readthedocs.io/en/latest/pythonapi/lightgbm.Booster.html#lightgbm.Booster.save_model).

## Storage Layout

**Simple**

The storage path can point directly to a serialized model

```
<storage-path/model-name.bst>
```

## Example

**Storage Layout**

```
s3://wml-serving-examples/
└── lightgbm-models/example.bst
```

**Predictor**

```
apiVersion: wmlserving.ai.ibm.com/v1
kind: Predictor
metadata:
  name: lightgbm-example
spec:
  modelType:
    name: lightgbm
  path: lightgbm-models/example.bst
  storage:
    s3:
      secretKey: modelStorage
      bucket: wml-serving-examples
```
