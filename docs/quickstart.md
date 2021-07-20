# Quick Start Guide

To quickly get started using Model-Mesh Serving, here is a brief guide.

## Prerequisites

- A Kubernetes cluster where you cluster administrative privileges
- [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl) and [kustomize](https://kubectl.docs.kubernetes.io/installation/kustomize/)

## 1. Install Model-Mesh Serving

### Clone the repository

```shell
git clone git@github.ibm.com:kserve/modelmesh-serving.git
cd modelmesh-serving
```

### Run install script

```shell
kubectl create namespace modelmesh-serving
./scripts/install.sh --namespace modelmesh-serving --quickstart
```

This will install Model-Mesh serving in the `modelmesh-serving` namespace, along with an etcd and MinIO instances.
Eventually after running this script, you should see a `Successfully installed Model-Mesh Serving!` message.

### Verify installation

Check that the pods are running:

```shell
kubectl get pods

NAME                                        READY   STATUS    RESTARTS   AGE
pod/etcd                                    1/1     Running   0          5m
pod/minio                                   1/1     Running   0          5m
pod/modelmesh-controller-547bfb64dc-mrgrq   1/1     Running   0          5m
```

To see more detailed instructions and information, click [here](./install/install-script.md).

## 2. Deploy a model

With Model-Mesh Serving now installed, try deploying a model. Here we have an
SKLearn MNIST model which is served from the local MinIO container:

```shell
kubectl apply -f - <<EOF
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
      secretKey: localMinIO
EOF
```

After applying this predictor, you should see it in the `Loading` state:

```
kubectl get predictors

NAME                      TYPE      AVAILABLE   ACTIVEMODEL   TARGETMODEL   TRANSITION   AGE
example-mnist-predictor   sklearn   false       Loading                     UpToDate     7s
```

Eventually, you should see the ServingRuntime pods that will hold the SKLearn model become `Running`.

```shell
kubectl get pods

modelmesh-serving-mlserver-0.x-7db675f677-twrwd   3/3     Running   0          2m
modelmesh-serving-mlserver-0.x-7db675f677-xvd8q   3/3     Running   0          2m
```

Then, checking on the `predictors` again, you should see that it is now available:

```shell
kubectl get predictors

NAME                      TYPE      AVAILABLE   ACTIVEMODEL   TARGETMODEL   TRANSITION   AGE
example-mnist-predictor   sklearn   true        Loaded                      UpToDate     2m
```

ServingRuntimes are automatically provisioned based on the `modelType` specified in the predictor.
Two ServingRuntimes are included with Model-Mesh Serving by default. The current mappings for these
are:

```
triton-2.x    -> tensorflow, pytorch, onnx, tensorrt
mlserver-0.x  -> sklearn, xgboost, lightgbm
```

To see more detailed instructions and information, click [here](./predictors/index.md).

## 3. Perform a gRPC inference request

Now that a model is loaded and available, you can then perform inference.
Currently, only gRPC inference requests are supported. By default, Model-Mesh Serving uses a
[headless Service](https://kubernetes.io/docs/concepts/services-networking/service/#headless-services)
since a normal Service has issues load balancing gRPC requests. See more info
[here](https://kubernetes.io/blog/2018/11/07/grpc-load-balancing-on-kubernetes-without-tears/).

To test out inference requests, you can port-forward the headless service *in a separate terminal window*:

```shell
kubectl port-forward --address 0.0.0.0 service/modelmesh-serving  8033 -n modelmesh-serving
```

Then a gRPC client generated from the KFServing [grpc_predict_v2.proto](https://github.com/kubeflow/kfserving/blob/master/docs/predict-api/v2/grpc_predict_v2.proto)
file can be used with `localhost:8033`. A ready-to-use Python example of this can be found [here](https://github.com/pvaneck/model-serving-sandbox/tree/main/grpc-predict).

Alternatively, you can test inference with [grpcurl](https://github.com/fullstorydev/grpcurl). This can easily be installed with `brew install grpcurl` if on macOS.

With `grpcurl`, a request can be sent to the SKLearn MNIST model like the following:

```shell
grpcurl \
  -plaintext \
  -proto fvt/proto/kfs_inference_v2.proto \
  -d '{ "model_name": "example-mnist-predictor", "inputs": [{ "name": "predict", "shape": [1, 64], "datatype": "FP32", "contents": { "fp32_contents": [0.0, 0.0, 1.0, 11.0, 14.0, 15.0, 3.0, 0.0, 0.0, 1.0, 13.0, 16.0, 12.0, 16.0, 8.0, 0.0, 0.0, 8.0, 16.0, 4.0, 6.0, 16.0, 5.0, 0.0, 0.0, 5.0, 15.0, 11.0, 13.0, 14.0, 0.0, 0.0, 0.0, 0.0, 2.0, 12.0, 16.0, 13.0, 0.0, 0.0, 0.0, 0.0, 0.0, 13.0, 16.0, 16.0, 6.0, 0.0, 0.0, 0.0, 0.0, 16.0, 16.0, 16.0, 7.0, 0.0, 0.0, 0.0, 0.0, 11.0, 13.0, 12.0, 1.0, 0.0] }}]}' \
  localhost:8033 \
  inference.GRPCInferenceService.ModelInfer
```

This should give you output like the following:

```shell
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

To see more detailed instructions and information, click [here](./predictors/run-inference.md).

## 4. (Optional) Deleting your Model-Mesh Serving installation

To delete all Model-Mesh Serving resources that were installed, run the following from the root of the project:

```shell
./scripts/delete.sh --namespace modelmesh-serving
```
