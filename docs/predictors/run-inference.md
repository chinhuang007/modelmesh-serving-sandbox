---
title: Inferencing
---

## Send an inference request to your Predictor

### Configure gRPC client

Configure your gRPC client to point to address `wml-serving:8033`, which is based on the kube-dns address and port corresponding to the service. Use the protobuf-based gRPC inference service defined [here](https://github.com/kubeflow/kfserving/blob/master/docs/predict-api/v2/required_api.md#grpc) to make inference requests to the model using the `ModelInfer` RPC, setting the name of the Predictor as the `model_name` field in the `ModelInferRequest` message.

Configure the gRPC clients which talk to your service to explicitly use:

- The `round_robin` loadbalancer policy
- A target URI string starting with `dns://` and based on the kube-dns address and port corresponding to the service, for example `dns:///model-mesh-test.wml-serving:8033` where `wml-serving` is the namespace, or just `dns:///model-mesh-test:8033` if the client resides in the same namespace. Note that you end up needing three consecutive `/`'s in total.

Not all languages have built-in support for this but most of the primary ones do. It's recommended to use the latest version of gRPC regardless. Here are some examples for specific languages (note other config such as TLS is omitted):

#### Java

```java
ManagedChannel channel = NettyChannelBuilder.forTarget("wml-serving:8033")
    .defaultLoadBalancingPolicy("round_robin").build();
```

Note that this was done differently in earlier versions of grpc-java - if this does not compile ensure you upgrade.

#### Go

```go
ctx, cancel := context.WithTimeout(context.Background(), 5  * time.Second)
defer cancel()
grpc.DialContext(ctx, "wml-serving:8033", grpc.WithBlock(), grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`))
```

#### Python

```python
credentials = grpc.ssl_channel_credentials(certificate_data_bytes)
channel_options = (("grpc.lb_policy_name", "round_robin"),)
channel = grpc.secure_channel(target, credentials, options=channel_options)
```

#### NodeJS

Using: https://www.npmjs.com/package/grpc

```javascript
// Read certificate
const cert = readFileSync(sslCertPath);

credentials = grpc.credentials.createSsl(cert);
// For insecure
credentials = grpc.credentials.createInsecure();

// Create client
const clientOptions = {
  "grpc.lb_policy_name": "round_robin",
};
// Get ModelMeshClient from grpc protobuf file
const client = ModelMeshClient(model_mesh_uri, credentials, clientOptions);

// Get rpc prototype for server
const response = await rpcProtoType.call(client, message);
```

### How to access service from outside the cluster without a NodePort

Using [`kubectl port-forward`](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/):

```shell
kubectl port-forward wml-serving 8033:8033
```

This assumes you are using port 8033, change the source and/or destination ports as appropriate.

Then change your client target string to localhost:8033, where 8033 is the chosen source port.

### grpcurl Example

Here is an example of how to do this using the command-line based [grpcurl](https://github.com/fullstorydev/grpcurl):

1. Install grpcurl:

```shell
$ grpcurl --version
grpcurl 1.8.1

# If it doesn't exist
$ brew install grpcurl
```

2. Port-forward to access the runtime service:

```shell
# access via localhost:8033
$ kubectl port-forward service/wml-serving 8033
Forwarding from 127.0.0.1:8033 -> 8033
Forwarding from [::1]:8033 -> 8033
```

3. In a separate terminal window, send an inference request using the proto file from `fvt/proto` or one that you have locally:

```shell
$ grpcurl -plaintext -proto fvt/proto/kfs_inference_v2.proto localhost:8033 list
inference.GRPCInferenceService

# run inference
# with below input, expect output to be 8
$ grpcurl \
  -plaintext \
  -proto fvt/proto/kfs_inference_v2.proto \
  -d '{ "model_name": "example-mnist-predictor", "inputs": [{ "name": "predict", "shape": [1, 64], "datatype": "FP32", "contents": { "fp32_contents": [0.0, 0.0, 1.0, 11.0, 14.0, 15.0, 3.0, 0.0, 0.0, 1.0, 13.0, 16.0, 12.0, 16.0, 8.0, 0.0, 0.0, 8.0, 16.0, 4.0, 6.0, 16.0, 5.0, 0.0, 0.0, 5.0, 15.0, 11.0, 13.0, 14.0, 0.0, 0.0, 0.0, 0.0, 2.0, 12.0, 16.0, 13.0, 0.0, 0.0, 0.0, 0.0, 0.0, 13.0, 16.0, 16.0, 6.0, 0.0, 0.0, 0.0, 0.0, 16.0, 16.0, 16.0, 7.0, 0.0, 0.0, 0.0, 0.0, 11.0, 13.0, 12.0, 1.0, 0.0] }}]}' \
  localhost:8033 \
  inference.GRPCInferenceService.ModelInfer

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

Note that you have to provide the `model_name` in the data load, which is the name of the Predictor deployed.

If a custom serving runtime which doesn't use the KFS V2 API is being used, the `mm-vmodel-id` header must be set to the Predictor name.

If you are sure the requests from your client are being routed in such a way that balances evenly across the cluster (as described [above](#configure-grpc-client)), you should include an additional metadata parameter `mm-balanced = true`. This allows some internal performance optimizations but should not be included if the source if the requests is not properly balanced.

For example adding these headers to the above grpcurl command:

```shell
grpcurl \
  -plaintext \
  -proto fvt/proto/kfs_inference_v2.proto \
  -rpc-header mm-vmodel-id:example-sklearn-mnist-svm \
  -rpc-header mm-balanced:true \
  -d '{ "model_name": "example-sklearn-mnist-svm", "inputs": [{ "name": "predict", "shape": [1, 64], "datatype": "FP32", "contents": { "fp32_contents": [0.0, 0.0, 1.0, 11.0, 14.0, 15.0, 3.0, 0.0, 0.0, 1.0, 13.0, 16.0, 12.0, 16.0, 8.0, 0.0, 0.0, 8.0, 16.0, 4.0, 6.0, 16.0, 5.0, 0.0, 0.0, 5.0, 15.0, 11.0, 13.0, 14.0, 0.0, 0.0, 0.0, 0.0, 2.0, 12.0, 16.0, 13.0, 0.0, 0.0, 0.0, 0.0, 0.0, 13.0, 16.0, 16.0, 6.0, 0.0, 0.0, 0.0, 0.0, 16.0, 16.0, 16.0, 7.0, 0.0, 0.0, 0.0, 0.0, 11.0, 13.0, 12.0, 1.0, 0.0] }}]}' \
  localhost:8033 \
  inference.GRPCInferenceService.ModelInfer
```

A [Jupyter Notebook](https://github.ibm.com/kserve/modelmesh-serving/blob/main/docs/demo/model_serve_post-install.ipynb) of the example with an additional Tensorflow example can also be found in wml-serving repo.
