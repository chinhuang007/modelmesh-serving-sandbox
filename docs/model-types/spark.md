---
title: Spark
description: Spark model
---

## Format

Spark model is serialized using the **save** method as shown below.

```
> model.save("/tmp/spark-model-folder")
```

For more information on ML persistence: https://spark.apache.org/docs/latest/ml-pipeline.html#ml-persistence-saving-and-loading-pipelines

Here is saved spark model directory structure:

```
> tree /tmp/spark-model-folder
/tmp/spark-model-folder
├── metadata
│   ├── _SUCCESS
│   └── part-00000
└── stages
    ├── 0_tok_45438196e5d4
    │   └── metadata
    │       ├── _SUCCESS
    │       └── part-00000
    ├── 1_hashingTF_7b2df2567a35
    │   └── metadata
    │       ├── _SUCCESS
    │       └── part-00000
    └── 2_logreg_4b033827b878
        ├── data
        │   ├── _SUCCESS
        │   └── part-00000-43ed38b0-ec72-45ec-a1e9-f7094f31f3fc-c000.snappy.parquet
        └── metadata
            ├── _SUCCESS
            └── part-00000

9 directories, 10 files
```

## Storage Layout

### 1. Simple

storage path can point directly to the model folder

```
<storage-path/model-folder>
├── metadata
└── stages
```

### Example

**Storage Layout**

```
s3://wml-serving-examples/
└── spark-models
    └── example-model-folder
        └── metadata
        └── stages
```

**Predictor**

```
apiVersion: wmlserving.ai.ibm.com/v1
kind: Predictor
metadata:
  name: spark-example
spec:
  modelType:
    name: spark
  path: spark-models/example-model-folder
  storage:
    s3:
      secretKey: modelStorage
      bucket: wml-serving-examples
```

### 2. Archived

storage path can also point to archived model file in gz (or) zip format.

```
<storage-path/model-file.gz>
```

Following are the steps to create an archive model file using saved model folder:

1. Create archive model file in gz format:

```
  cd /tmp/spark-model-folder
  tar -czvf spark-model-file.gz metadata stages
```

2. Create archive model file in zip format:

```
  cd /tmp/spark-model-folder
  zip -r spark-model-file.zip metadata stages
```

### Example

**Storage Layout**

```
s3://wml-serving-examples/
└── spark-models
    └── example-model-file.gz
```

**Predictor**

```
apiVersion: wmlserving.ai.ibm.com/v1
kind: Predictor
metadata:
  name: spark-example
spec:
  modelType:
    name: spark
  path: spark-models/example-model-file.gz
  storage:
    s3:
      secretKey: modelStorage
      bucket: wml-serving-examples
```
