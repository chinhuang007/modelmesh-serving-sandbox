---
title: Model Schema (alpha)
---

The input and output schema of ML models can be provided via the `Predictor` CR along with the model files themselves. This must be a JSON file in the standard format described below, which currently must reside in the same storage instance as the corresponding model.

<InlineNotification kind="warning">

**Warning**: The generic model schema should be considered alpha. Breaking changes to how the schema is used are expected. Do not rely on this schema in production.

</InlineNotification>

### Schema Format

The JSON for schema should be in **_KFS V2 format_**, fields are mapped to tensors.

```json
{
        "inputs": [{
                "name": "Tensor name",
                "datatype": "Tensor data type",
                "shape": [Dimension of the tensor]
        }],
        "outputs": [{
                "name": "Tensor name",
                "datatype": "Tensor data type",
                "shape": [Dimension of the tensor]
        }]
}
```

Refer to the [KFServing V2 Protocol](https://github.com/kubeflow/kfserving/blob/master/docs/predict-api/v2/required_api.md#tensor-data) docs for tensor data representation and supported tensor data types.

### Sample schema

```json
{
  "inputs": [
    {
      "name": "INPUT0",
      "datatype": "STRING",
      "shape": [1, 28, 28]
    }
  ],
  "outputs": [
    {
      "name": "OUTPUT0",
      "datatype": "STRING",
      "shape": [10]
    }
  ]
}
```

The `schemaPath` field of the `Predictor` custom resource should be set to point to this JSON file within the predictor's specified storage instance.

#### Example `Predictor` CR

```yaml
apiVersion: wmlserving.ai.ibm.com/v1
kind: Predictor
metadata:
  name: tensorflow-pbnoschema
spec:
  modelType:
    name: tensorflow
  path: tfmnist-pb-noschema
  schemaPath: schema/tf-schema.json
  storage:
    s3:
      secretKey: myStorage
      bucket: wml-serving-schema
```

Note that this field is optional. Not all model types require a schema to be provided - for example when the model serialization format incorporates equivalent schema information or it is otherwise not required by the corresponding runtime. In some cases the schema isn't required but will be used for additional payload validation when it is.
