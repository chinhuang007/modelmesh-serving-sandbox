---
title: Using Model-Mesh Serving
---

Trained models are deployed in Model-Mesh Serving via `Predictor`s. These represent a stable service endpoint behind which the underlying model can change.

Models must reside on shared storage. Currently, only S3-based storage is supported but support for other types will follow. Note that model data residing at a particular path within a given storage instance is **assumed to be immutable**. Different versions of the same logical model are treated at the base level as independent models and must reside at different paths. In particular, where a given model server/runtime natively supports the notion of versioning (such as Nvidia Triton, TensorFlow Serving, etc), the provided path should not point to the top of a (pseudo-)directory structure containing multiple versions. Instead, point to the subdirectory which corresponds to a specific version.

## Deploying a scikit-learn model

### Prerequisites

The Model-Mesh Serving instance should be installed in the desired namespace. See [install docs](install) for more details.

### Deploy a sample model directly from our shared object storage

1. Check the `storage-config` secret for access to shared COS instance

A set of example models are shared via an IBM Cloud COS instance to use when getting started with Model-Mesh Serving and experimenting with the provided runtimes. Access to this COS instance is set up in the `storage-config` secret.

If you used quick start install then there will be a key within the storage-config secret already configured with the name `wml-serving-example-models`. If you installed Model-Mesh Serving using the operator, you will have to configure the `storage-config` secret for access:

```shell
$ kubectl patch secret/storage-config -p '{"data": {"wml-serving-example-models": "ewogICJ0eXBlIjogInMzIiwKICAiYWNjZXNzX2tleV9pZCI6ICJlY2I5ODNmMTE4MjI0MjNjYTllNDg3Zjg5OGQ1NGE4ZiIsCiAgInNlY3JldF9hY2Nlc3Nfa2V5IjogImNkYmVmZjZhMzJhZWY2YzIzNzRhZTY5ZWVmNTAzZTZkZDBjOTNkNmE3NGJjMjQ2NyIsCiAgImVuZHBvaW50X3VybCI6ICJodHRwczovL3MzLnVzLXNvdXRoLmNsb3VkLW9iamVjdC1zdG9yYWdlLmFwcGRvbWFpbi5jbG91ZCIsCiAgInJlZ2lvbiI6ICJ1cy1zb3V0aCIsCiAgImRlZmF1bHRfYnVja2V0IjogIndtbC1zZXJ2aW5nLWV4YW1wbGUtbW9kZWxzLXB1YmxpYyIKfQo="}}'
```

For reference the contents of the secret value for the `wml-serving-example-models` entry looks like:

```json
{
  "type": "s3",
  "access_key_id": "ecb983f11822423ca9e487f898d54a8f",
  "secret_access_key": "cdbeff6a32aef6c2374ae69eef503e6dd0c93d6a74bc2467",
  "endpoint_url": "https://s3.us-south.cloud-object-storage.appdomain.cloud",
  "region": "us-south",
  "default_bucket": "wml-serving-example-models-public"
}
```

<InlineNotification>

**Note** After updating the storage config secret, there may be a delay of up to 2 minutes until the change is picked up. You should take this into account when creating/updating Predictors that use storage keys which have just been added or updated - they may fail to load otherwise.

</InlineNotification>

For more details of configuring model storage, see the [Setup Storage](/predictors/setup-storage) page.

2. Create a Predictor Custom Resource to serve the sample model

The `config/example-predictors` directory contains Predictor manifests for many of the example models. For a list of available models, see the [example models documentation](example-models#available-models).

Here we are deploying an sklearn model located at `sklearn/mnist-svm.joblib` within the shared COS storage.

```shell
# Pulled from sample config/example-predictors/example-mlserver-sklearn-mnist-predictor.yaml
$ kubectl apply -f - <<EOF
apiVersion: wmlserving.ai.ibm.com/v1
kind: Predictor
metadata:
  name: example-mnist-predictor
spec:
  modelType:
    name: sklearn
  path: sklearn/mnist-svm.joblib
  storage:
    s3:
      secretKey: wml-serving-example-models
EOF
predictor.wmlserving.ai.ibm.com/example-mnist-predictor created
```

Note that `wml-serving-example-models` is the name of the secret key created/verified in the previous step.

For more details go to the [Predictor Spec page](/predictors/predictor-spec).

Once the `Predictor` is created, mlserver runtime pods are automatically started to load and serve it.

```shell
$ kubectl get pods

NAME                                         READY   STATUS              RESTARTS   AGE
wml-serving-mlserver-0.x-658b7dd689-46nwm    0/3     ContainerCreating   0          2s
wml-serving-mlserver-0.x-658b7dd689-46nwm    0/3     ContainerCreating   0          2s
wmlserving-controller-568c45b959-nl88c       1/1     Running             0          11m
```

3. Check the status of your Predictor:

```shell
$ kubectl get predictors
NAME                      TYPE      AVAILABLE   ACTIVEMODEL   TARGETMODEL   TRANSITION   AGE
example-mnist-predictor   sklearn   true        Loading                     UpToDate     60s

$ kubectl get predictor example-mnist-predictor -o=jsonpath='{.status.grpcEndpoint}'
grpc://wml-serving:8033
```

The states should reflect immediate availability, but may take some seconds to move from `Loading` to `Loaded`.
Inferencing requests for this Predictor received prior to loading completion will block until it completes.

<InlineNotification>

**Note:** When `ScaleToZero` is enabled, the first Predictor assigned to the Triton runtime may be stuck in the `Pending` state for some time while the Triton pods are being created. The Triton image is large and may take a while to download.

</InlineNotification>

## Using the deployed model

Configure your gRPC client to point to address `wml-serving:8033`. Use the protobuf-based gRPC inference service defined [here](https://github.com/kubeflow/kfserving/blob/master/docs/predict-api/v2/required_api.md#grpc)
to make inference requests to the model using the `ModelInfer` RPC, setting the name of the Predictor as the `model_name` field in the `ModelInferRequest` message.

Here is an example of how to do this using the command-line based [grpcurl](https://github.com/fullstorydev/grpcurl):

Port-forward to access the runtime service:

```shell
# access via localhost:8033
$ kubectl port-forward service/wml-serving 8033
Forwarding from 127.0.0.1:8033 -> 8033
Forwarding from [::1]:8033 -> 8033
```

In a separate terminal window, send an inference request using the proto file from `fvt/proto` or one that you have locally. Note that you have to provide the `model_name` in the data load, which is the name of the Predictor deployed.

```shell
$ grpcurl -plaintext -proto fvt/proto/kfs_inference_v2.proto localhost:8033 list
inference.GRPCInferenceService

# run inference
# with below input, expect output to be 8
$ grpcurl -plaintext -proto fvt/proto/kfs_inference_v2.proto -d '{ "model_name": "example-mnist-predictor", "inputs": [{ "name": "predict", "shape": [1, 64], "datatype": "FP32", "contents": { "fp32_contents": [0.0, 0.0, 1.0, 11.0, 14.0, 15.0, 3.0, 0.0, 0.0, 1.0, 13.0, 16.0, 12.0, 16.0, 8.0, 0.0, 0.0, 8.0, 16.0, 4.0, 6.0, 16.0, 5.0, 0.0, 0.0, 5.0, 15.0, 11.0, 13.0, 14.0, 0.0, 0.0, 0.0, 0.0, 2.0, 12.0, 16.0, 13.0, 0.0, 0.0, 0.0, 0.0, 0.0, 13.0, 16.0, 16.0, 6.0, 0.0, 0.0, 0.0, 0.0, 16.0, 16.0, 16.0, 7.0, 0.0, 0.0, 0.0, 0.0, 11.0, 13.0, 12.0, 1.0, 0.0] }}]}' localhost:8033 inference.GRPCInferenceService.ModelInfer

{
  "modelName": "example-mnist-predictor-725d74f061",
  "outputs": [
    {
      "name": "predict",
      "datatype": "FP32",
      "shape": [
        "1"
      ],
      "contents": {
        "fp32Contents": [
          8
        ]
      }
    }
  ]
}
```

## Updating the model

Changes can be made to the Predictor's Spec, such as changing the target storage and/or model, without interrupting the inferencing service.
The predictor will continue to use the prior spec/model until the new one is loaded and ready.

Below, we are changing the Predictor to use a completely different model, in practice the schema of the Predictor's model would be consistent across updates even if the type of model or ML framework changes.

```shell
$ kubectl apply -f - <<EOF
apiVersion: wmlserving.ai.ibm.com/v1
kind: Predictor
metadata:
  name: example-mnist-predictor
spec:
  modelType:
    name: tensorflow
  # Note updated model type and location
  path: tensorflow/mnist.savedmodel
  storage:
    s3:
      secretKey: wml-serving-example-models
EOF
predictor.wmlserving.ai.ibm.com/example-mnist-predictor configured

$ kubectl get predictors
NAME                      TYPE      AVAILABLE   ACTIVEMODEL   TARGETMODEL   TRANSITION   AGE
example-mnist-predictor   sklearn   true        Loaded        Loading       InProgress    10m
```

The "transition" state of the Predictor will be `InProgress` while waiting for the new backing model to be ready,
and return to `UpToDate` once the transition is complete.

```shell
$ kubectl get predictors
NAME                      TYPE      AVAILABLE   ACTIVEMODEL   TARGETMODEL   TRANSITION   AGE
example-mnist-predictor   sklearn   true        Loaded                      UpToDate      11m
```

If there is a problem loading the new model (for example it does not exist at the specified path), the transition state will
change to `BlockedByFailedLoad`, but the service will remain available. The active model state will still show as `Loaded`, and the
Predictor remains available.

```shell
$ kubectl get predictors
NAME                      TYPE      AVAILABLE   ACTIVEMODEL   TARGETMODEL   TRANSITION   AGE
example-mnist-predictor   sklearn   true        Loaded        Failed        BlockedByFailedLoad    20m
```

## For More Details

- [Setup Storage](/predictors/setup-storage)
- [Inferencing](/predictors/run-inference)
- [Predictor Spec](/predictors/predictor-spec)

A [Jupyter Notebook](https://github.ibm.com/ai-foundation/wml-serving/blob/main/docs/demo/model_serve_post-install.ipynb) of the example can also be found in wml-serving repo.
