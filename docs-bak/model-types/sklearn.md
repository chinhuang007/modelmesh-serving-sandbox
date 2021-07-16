---
title: scikit-learn
description: Scikit-learn Joblib models
---

## Format

Scikit-learn model serialized using [joblib.dump](https://joblib.readthedocs.io/en/latest/generated/joblib.dump.html).
See the [model persistence](https://scikit-learn.org/stable/modules/model_persistence.html)
Scikit-learn documentation for details.

Scikit-learn models serialized using "pickle" library and "joblib" library are supported.

## Configuration

Model configuration for an `sklearn` model is optional, but it may be used
for advanced use-cases. If configuration is provided, it must be in a file
called `model-settings.json`. The complete specification of this
configuration is defined by
[MLServer's ModelSettings class](https://github.com/SeldonIO/MLServer/blob/0.2.1/mlserver/settings.py#L49).

<InlineNotification>

**Note** When using a configuration file, the `name` field will be ignored by Model-Mesh Serving.

</InlineNotification>

## Storage Layout

**Simple**

The storage path can point directly to an Sklearn model serialized using
`joblib`.

```
<storage-path/model.joblib>
```

The file does not need to be called `model.joblib`, it can have any name.

**Directory**

The storage path can point to a directory containing a single file that is
the Sklearn model serialized using `joblib`.

```
<storage-path>/
└── model.joblib
```

The file does not need to be called `model.joblib`, it can have any name.

**Explicit Configuration**

If the `model-settings.json` configuration file is provided, it must be in
the directory pointed to by the `Predictor`'s path. The model files must also
be contained under this path.

```
<storage-path>/
├── model-settings.json
└── <model-files>
```

## Example

**Storage Layout**

```
s3://wml-serving-examples/sklearn-model/model.joblib
```

**Predictor**

```
apiVersion: wmlserving.ai.ibm.com/v1
kind: Predictor
metadata:
  name: sklearn-example
spec:
  modelType:
    name: sklearn
  path: sklearn-model/model.joblib
  storage:
    s3:
      secretKey: modelStorage
      bucket: wml-serving-examples
```
