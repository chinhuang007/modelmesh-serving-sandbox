---
title: MLeap
description: MLeap Bundle model
---

## Format

> MLeap is a common serialization format and execution engine for machine learning pipelines. It supports Spark, Scikit-learn and Tensorflow for training pipelines and exporting them to an MLeap Bundle. Serialized pipelines (bundles) can be deserialized [...] to power realtime API services.

([source](https://mleap-docs.combust.ml/))

Model-Mesh Serving supports the execution of Spark and Scikit-learn machine learning
pipelines Bundles. The convenient Bundle format includes all necessary
configuration to load and run the pipeline. For information on creating and
serializing a Bundle, refer to the
[MLeap documentation](https://mleap-docs.combust.ml/mleap-runtime/bundle.html).

## Storage Layout

**Simple**

For a Bundle archive, the storage path can point directly to the file

```
<storage-path/model-bundle.zip>
```

**Extracted**

An extracted Bundle archive is also supported:

```
<storage-path>/
├── bundle.json
└── root
    └── <model-data>
```

## Example

**Storage Layout**

```
s3://wml-serving-examples/
└── mleap-models/example.bundle.zip
```

**Predictor**

```
apiVersion: wmlserving.ai.ibm.com/v1
kind: Predictor
metadata:
  name: mleap-example
spec:
  modelType:
    name: mleap
  path: mleap-models/example.bundle.zip
  storage:
    s3:
      secretKey: modelStorage
      bucket: wml-serving-examples
```
