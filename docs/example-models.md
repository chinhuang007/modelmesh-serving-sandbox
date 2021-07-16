---
title: Sample Models
---
# Sample Models
A set of example models are shared via an IBM Cloud COS instance to use when getting started with Model-Mesh Serving and experimenting with the provided runtimes.

## Predictors

The `config/example-predictors` directory contains Predictor manifests for many of the example models. Assuming that the entry specified below is added to the storage configuration secret, the Predictors can be deployed and used for experimentation.

## Naming Conventions

Models are organized into virtual directories based on model type:

```
s3://wml-serving-example-models-public/
└── <model-type>/
    ├── <model-name-file>
    └── <model-name>/
        └── <model-data>
```

## Available Models

### Virtual Object Tree

```
s3://wml-serving-example-models-public/
├── lightgbm
│   └── mushroom.bst
├── mleap
│   └── airbnb.model.zip
├── onnx
│   └── mnist.onnx
├── pmml
│   └── drug-prediction.xml
├── pytorch
│   └── cifar
│       ├── 1
│       │   └── model.pt
│       └── config.pbtxt
├── sklearn
│   └── mnist-svm.joblib
├── spark
│   └── simple-text-classification.tar.gz
├── tensorflow
│   └── mnist.savedmodel
│       ├── saved_model.pb
│       └── variables
│           ├── variables.data-00000-of-00001
│           └── variables.index
└── xgboost
    └── mushroom.json
```

## Storage Configuration

Below is the secret configuration that should be added to the `storage-config` secret to configure access to the example models bucket. The credentials provided have read-only access

The quick-start installation includes configuring access to the example models.

If adding a field to a Secret YAML manifest via `stringData`:

```yaml
wml-serving-example-models: |
  {
    "type": "s3",
    "access_key_id": "ecb983f11822423ca9e487f898d54a8f",
    "secret_access_key": "cdbeff6a32aef6c2374ae69eef503e6dd0c93d6a74bc2467",
    "endpoint_url": "https://s3.us-south.cloud-object-storage.appdomain.cloud",
    "region": "us-south",
    "default_bucket": "wml-serving-example-models-public"
  }
```

or to directly patch an existing secret:

```bash
kubectl patch secret/storage-config -p '{"data": {"wml-serving-example-models": "ewogICJ0eXBlIjogInMzIiwKICAiYWNjZXNzX2tleV9pZCI6ICJlY2I5ODNmMTE4MjI0MjNjYTllNDg3Zjg5OGQ1NGE4ZiIsCiAgInNlY3JldF9hY2Nlc3Nfa2V5IjogImNkYmVmZjZhMzJhZWY2YzIzNzRhZTY5ZWVmNTAzZTZkZDBjOTNkNmE3NGJjMjQ2NyIsCiAgImVuZHBvaW50X3VybCI6ICJodHRwczovL3MzLnVzLXNvdXRoLmNsb3VkLW9iamVjdC1zdG9yYWdlLmFwcGRvbWFpbi5jbG91ZCIsCiAgInJlZ2lvbiI6ICJ1cy1zb3V0aCIsCiAgImRlZmF1bHRfYnVja2V0IjogIndtbC1zZXJ2aW5nLWV4YW1wbGUtbW9kZWxzLXB1YmxpYyIKfQo="}}'
```
