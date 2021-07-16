# Model-Mesh Serving

Controller for common modelmesh-serving component

---

## Getting Started

Please see our [getting started documentation](./docs/install/index.md) for details of how to obtain, install and use Model-Mesh Serving.

For help, please open an issue in this repository.

## Components and their Repositories

Model-Mesh Serving currently comprises components spread over a number of repositories.

![Architecture Image](./docs/architecture/images/0.2.0-highlevel.png)

Issues across all components are tracked centrally in this repo.

#### Core Components

- https://github.com/kserve/modelmesh-serving (this repo) - the model serving controller
- https://github.com/kserve/modelmesh - the model-mesh containers used for orchestrating model placement and routing

#### Runtime Adapters

- [modelmesh-runtime-adapter](https://github.com/kserve/modelmesh-runtime-adapter) - the containers which run in each model serving pod and act as an intermediary between model-mesh and third-party model-server containers. Its build produces a single "multi-purpose" image which can be used as an adapter to work with each of the model servers used for the built-in runtimes. It also incorporates the "puller" logic which is responsible for retrieving the models from storage before handing over to the respective adapter logic to load the model (and to delete after unloading). This image is also used for a container in the load/unload path of custom ServingRuntime Pods, as a "standalone" puller.

#### Model Servers

These are third party images which we repackage as UBI-based images for RedHat OCP compliance.

- [triton-inference-server](https://github.com/triton-inference-server/server) - Nvidia's Triton
- [seldon-mlserver](https://github.com/SeldonIO/MLServer) - Seldon's python MLServer which is part of [KFServing](https://github.com/kubeflow/kfserving)

#### Libraries

These are internally-developed Java libraries used by the model-mesh component.

- [kv-utils](https://github.com/IBM/kv-utils) - Useful KV store recipes abstracted over etcd and Zookeeper
- [litelinks-core](https://github.com/IBM/litelinks-core) - RPC/service discovery library based on Apache Thrift, used only for communications internal to model-mesh.

## Contributing

Please read our [contributing guide](./CONTRIBUTING.md) for details on contributing, how to setup your environment to develop locally, and details on submitting pull requests.

### Building

Sample build:
```bash
GIT_COMMIT=$(git rev-parse HEAD)
BUILD_ID=$(date '+%Y%m%d')-$(git rev-parse HEAD | cut -c -5)
BASE_IMAGE_TAG=$(cat BASE_IMAGE_TAG | awk -F= '{print $2}')
IMAGE_TAG_VERSION=0.0.1
IMAGE_TAG=${IMAGE_TAG_VERSION}-$(git branch --show-current)_${BUILD_ID}
DEV_IMAGE=modelmesh-serving-controller:develop

docker build -f Dockerfile.develop \
    -t ${DEV_IMAGE} \
    --build-arg BASE_IMAGE_TAG=${BASE_IMAGE_TAG} .

docker build -t modelmesh-serving-controller:${IMAGE_TAG} \
    --build-arg IMAGE_VERSION=${IMAGE_TAG} \
    --build-arg COMMIT_SHA=${GIT_COMMIT} \
    --build-arg DEV_IMAGE=${DEV_IMAGE} \
    --build-arg BASE_IMAGE_TAG=${BASE_IMAGE_TAG} .
```
