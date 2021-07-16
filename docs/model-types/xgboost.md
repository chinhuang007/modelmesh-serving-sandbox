---
title: XGBoost
description: XGBoost gradient boosting models
---

## Format

XGBoost models must be serialised using the
[booster.save_model() method](https://xgboost.readthedocs.io/en/latest/tutorials/saving_model.html).
It can be serialized as JSON or in the binary `.bst` format.

## Storage Layout

**Simple**

The storage path can point directly to a serialized model

```
<storage-path/model-name.json>
```

## Example

**Storage Layout**

```
s3://wml-serving-examples/
└── xgboost-models/example.json
```

**Predictor**

```
apiVersion: wmlserving.ai.ibm.com/v1
kind: Predictor
metadata:
  name: xgboost-example
spec:
  modelType:
    name: xgboost
  path: xgboost-models/example.json
  storage:
    s3:
      secretKey: modelStorage
      bucket: wml-serving-examples
```
