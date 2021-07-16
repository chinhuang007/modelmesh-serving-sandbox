---
title: Getting Started
---

# Prerequisites

- **Kubernetes cluster** - An Openshift cluster running on Openshift 4.6+. You will need `cluster-admin` authority in order to complete all of the prescribed steps.

- **cloudctl** - Download latest version of cloudctl CLI from [Cloud Pak CLI](https://github.com/IBM/cloud-pak-cli/releases).

- **oc** - Install the latest version of OpenShift CLI from [oc CLI](https://docs.openshift.com/container-platform/4.7/cli_reference/openshift_cli/getting-started-cli.html).

- **etcd** - Model-Mesh Serving requires an [etcd](https://etcd.io/) server in order to coordinate internal state which can be either dedicated or shared. More on this later.

- **S3-compatible object storage** - Before models can be deployed, a remote S3-compatible datastore is needed from which to pull the model data. This could be for example an [IBM Cloud Object Storage](https://www.ibm.com/cloud/object-storage) instance, or a locally running [MinIO](https://github.com/minio/minio) deployment. Note that this is not required to be in place prior to the initial installation.

# Namespace Scope

Model-Mesh Serving is namespace scoped, meaning all of its components must exist within a single namespace and only one instance of Model-Mesh Serving can be installed per namespace. This applies to the Model-Mesh Serving Operator as well. Multiple Model-Mesh Serving instances can be installed in separate namespaces within the cluster.

Start your installation by creating a new namespace.

```
oc new-project <namespace-name>
```

# Deployed Components

|            | Type             | Pod                         | Count   | Default CPU request/limit per-pod | Default mem request/limit per-pod |
| ---------- | ---------------- | --------------------------- | ------- | --------------------------------- | --------------------------------- |
| 1          | Operator         | Common service operator     | 1       | 100m / 500m                       | 200Mi / 512Mi                     |
| 2          | Operator         | ETCD operator               | 1       | 100m / 500m                       | 100Mi / 500Mi                     |
| 3          | Operator         | Model-Mesh Serving operator        | 1       | 50m / 500m                        | 40Mi / 40Mi                       |
| 4          | Controller       | Model-Mesh Serving controller      | 1       | 50m / 1                           | 96Mi / 512Mi                      |
| 5          | Metastore        | ETCD pods                   | 3       | 200m / 200m                       | 512Mi / 512Mi                     |
| 6          | Built-in Runtime | Nvidia Triton runtime Pods  | 0 \(\*) | 850m / 10                         | 1568Mi / 1984Mi                   |
| 7          | Built-in Runtime | The MSP Server runtime Pods | 0 \(\*) | 850m / 10                         | 1568Mi / 1984Mi                   |
| 8          | Built-in Runtime | The MLServer runtime Pods   | 0 \(\*) | 850m / 10                         | 1568Mi / 1984Mi                   |
| **totals** |                  |                             | 7       | 900m / 3.1                        | 1.925Gi / 3.02Gi                  |

(\*) [`ScaleToZero`](production-use/scaling#scale-to-zero) is enabled by default, so runtimes will have 0 replicas until a Predictor is created that uses that runtime. Once a Predictor is assigned, the runtime pods will scale up to 2.

When Model-Mesh Serving operator is installed, pods shown in 1 - 3 are created.

When Model-Mesh Serving instance is created with default, internal ETCD connection, pods shown in 4 and 5 are created. When Model-Mesh Serving instance is created with external ETCD connection, pod shown in 4 is created.

When `ScaleToZero` **is enabled**, deployments for runtime pods will be scaled to 0 when there are no Predictors for that runtime. When `ScaletoZero` is enabled and first predictor CR is submitted, Model-Mesh Serving will spin up the corresponding built-in runtime pods.

When `ScaletoZero` is **disabled**, pods shown in 6 to 8 are created, with a total CPU(request/limit) of 6/63.1 and total memory(request/limit) of 11.11Gi/14.652Gi.

The deployed footprint can be significantly reduced in the following ways:

- Individual built-in runtimes can be disabled by setting `disabled: true` in their corresponding `ServingRuntime` resource - if the corresponding model types aren't used.

- The number of Pods per runtime can be changed from the default of 2 (e.g. down to 1), via the `podsPerRuntime` global configuration parameter (see [configuration](configuration)). It is recommended for this value to be a minimum of 2 for production deployments.

- Memory and/or CPU resource allocations can be reduced (or increased) on the primary model server container in either of the built-in `ServingRuntime` resources (container name `triton` or `mleap`). This has the effect of adjusting the total capacity available for holding served models in memory.

```shell
> oc edit servingruntime triton-2.x
> oc edit servingruntime msp-ml-server-0.x
> oc edit servingruntime mlserver-0.x
```

<InlineNotification>

**Note** The MLServer runtime is only present in Model-Mesh Serving versions >= 0.3.0

</InlineNotification>

Please be aware that:

1. Changes made to the _built-in_ runtime resources will likely be reverted when upgrading/re-installing
2. Most of this resource allocation behaviour/config will change in future versions to become more dynamic - both the number of pods deployed and the system resources allocated to them

In addition, the following resources will be created in the same namespace:

- `model-serving-defaults` - ConfigMap holding default values tied to a release, should not be modified. Configuration can be overriden by creating a user ConfigMap, see [configuration](configuration)
- `tc-config` - ConfigMap used for some internal coordination
- `storage-config` - Secret holding config for each of the storage backends from which models can be loaded - see [the example](predictors)

# Next Steps

- [Install using OLM and operator](install/install)

- See [the configuration page](configuration) for details of how to configure system-wide settings via a ConfigMap, either before or after installation.

- See this [example walkthrough](predictors) of deploying a TensorFlow model as a `Predictor`.
