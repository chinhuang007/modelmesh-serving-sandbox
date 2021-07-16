---
title: Predictor Spec
---

## Predictor Spec

Here is a complete example of a Predictor spec:

```yaml
apiVersion: wmlserving.ai.ibm.com/v1
kind: Predictor
metadata:
  name: my-mnist-predictor
spec:
  modelType:
    name: tensorflow
    version: "1.15" # Optional
  runtime: # Optional
    name: triton-2.x
  path: my_models/mnist-tf
  storage:
    s3:
      secretKey: my_storage
      bucket: my_bucket
```

Notes:

- Prior to version `0.3.0`, please use `apiVersion: ai.ibm.com/v1alpha1` in the Predictor yaml
- Prior to version `0.5.0`, please use `apiVersion: wmlserving.ai.ibm.com/v1alpha1` in the Predictor yaml
- `runtime` is optional. If included, the model will be loaded/served using the `ServingRuntime` with the specified name, and the predictors `modelType` must match an entry in that runtime's `supportedModels` list (see [runtimes](../runtimes))
- The CRD contains additional fields but they have been omitted here for now since they are not yet fully supported
