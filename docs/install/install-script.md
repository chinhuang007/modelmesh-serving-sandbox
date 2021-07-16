---
title: Non-Operator Installation
---

# Installation
## Prerequisites

- **Kubernetes cluster** - A Kubernetes cluster is required. You will need `cluster-admin` authority in order to complete all of the prescribed steps.

  - To set up Model-Mesh Serving for local minikube, review the [minikube instructions](minikube.md).

- **Kubectl and Kustomize** - The installation will occur via the terminal using [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) and [kustomize](https://kubectl.docs.kubernetes.io/installation/kustomize/).

- **etcd** - Model-Mesh Serving requires an [etcd](https://etcd.io/) server in order to coordinate internal state which can be either dedicated or shared. More on this later.

We provide an install script to quickly run Model-Mesh Serving without the operator. This may be useful for experimentation or development but should not be used in production.

## Installation

Download and extract the latest modelmesh-serving from [Github](https://github.ibm.com/kserve/modelmesh-serving):

```shell
git clone git@github.ibm.com:kserve/modelmesh-serving.git
cd modelmesh-serving
```

Run the script to install Model-Mesh Serving CRDs, controller, and built-in runtimes into the specified Kubernetes namespaces.

```shell
kubectl create namespace modelmesh-serving
./scripts/install.sh --namespace modelmesh-serving --quickstart --delete
```

See the installation help below for detail:
```shell
./scripts/install.sh --help
usage: ./scripts/install.sh [flags]

Flags:
  -n, --namespace                (required) Kubernetes namespace to deploy Model-Mesh Serving to.
  -d, --delete                   Delete any existing instances of Model-Mesh Serving in Kube namespace before running install, including CRDs, RBACs, controller, older CRD with ai.ibm.com api group name, etc.
  --quickstart                   Install and configure required supporting datastores in the same namespace (etcd and MinIO) - for experimentation/development

Installs Model-Mesh Serving CRDs, controller, and built-in runtimes into specified
Kubernetes namespaces.

Expects cluster-admin authority and Kube cluster access to be configured prior to running.
Also requires Etcd secret 'model-serving-etcd' to be created in namespace already.
```
A Kubernetes `--namespace` is required, which must already exist. You must also have cluster-admin authority and cluster access must be configured prior to running the install script.

The `--quickstart` option can be specified to install and configure supporting datastores in the same namespace (etcd and MinIO) for experimental/development use - in particular the connections to these datastores are not secure and the etcd cluster is a single member which is not highly available. 

Use of `--quickstart` also configures the `storage-config` secret to be able to pull from the [Model-Mesh Serving example models bucket](../example-models.md) which contains the model data for the sample Predictors. For complete details on the manfiests applied with `--quickstart` see [config/dependencies/quickstart.yaml](https://github.ibm.com/kserve/modelmesh-serving/blob/main/config/dependencies/quickstart.yaml). 

If `--quickstart` is not chosen, the namespace provided must have an Etcd secret named `model-serving-etcd` created which provides access to the Etcd cluster. See the [instructions](#setup-the-etcd-connection-information) on this step.

You can also optionally use `--delete` to delete any existing instances of Model-Mesh Serving in the designated Kube namespace before running the install.

## Setup the etcd connection information

If the `--quickstart` install option is **not** being used, details of an existing etcd cluster must be specified prior to installation. Otherwise, please skip this step and proceed to [Installation](#installation).

Create a file named etcd-config.json, populating the values based upon your etcd server. The same etcd server can be shared between environments and/or namespaces, but in this case _the `root_prefix` must be set differently in each namespace's respective secret_. The complete json schema for this configuration is documented [here](https://github.com/IBM/etcd-java/blob/master/etcd-json-schema.md).

```json
{
  "endpoints": "https://etcd-service-hostame:2379",
  "userid": "userid",
  "password": "password",
  "root_prefix": "unique-chroot-prefix"
}
```

Then create the secret using the file (note that the key name within the secret must be `etcd_connection`):

```shell
kubectl create secret generic model-serving-etcd --from-file=etcd_connection=etcd-config.json
```

A secret named `model-serving-etcd` will be created and passed to the controller.
## Delete the installation

To wipe out the Model-Mesh Serving CRDs, controller, and built-in runtimes from the specified Kubernetes namespaces:

```shell
./scripts/install.sh --namespace modelmesh-serving -delete
```



