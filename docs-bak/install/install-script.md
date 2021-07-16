---
title: Non-Operator Installation
---

## Prerequisites

- **Kubernetes cluster** - A Kubernetes cluster is required. You will need `cluster-admin` authority in order to complete all of the prescribed steps.

  - To set up Model-Mesh Serving for local minikube, review the [minikube instructions](minikube).

- **Kubectl and Kustomize** - The installation will occur via the terminal using [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) and [kustomize](https://kubectl.docs.kubernetes.io/installation/kustomize/).

- **etcd** - Model-Mesh Serving requires an [etcd](https://etcd.io/) server in order to coordinate internal state which can be either dedicated or shared. More on this later.

- **S3-compatible object storage** - Before models can be deployed, a remote S3-compatible datastore is needed from which to pull the model data. This could be for example an [IBM Cloud Object Storage](https://www.ibm.com/cloud/object-storage) instance, or a locally running [MinIO](https://github.com/minio/minio) deployment. Note that this is not required to be in place prior to the initial installation.

We provide an install script to quickly run Model-Mesh Serving without the operator. This may be useful for experimentation or development but should not be used in production.

The install script has a `--quickstart` option for setting up a self-contained Model-Mesh Serving instance. This will deploy and configure local etcd and MinIO servers in the same Kubernetes namespace. Note that this is only for experimentation and/or development use - in particular the connections to these datastores are not secure and the etcd cluster is a single member which is not highly available. Use of `--quickstart` also configures the `storage-config` secret to be able to pull from the [Model-Mesh Serving example models bucket](../example-models) which contains the model data for the sample Predictors. For complete details on the manfiests applied with `--quickstart` see [config/dependencies/quickstart.yaml](https://github.ibm.com/ai-foundation/wml-serving/blob/main/config/install/dependencies/quickstart.yaml).

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

## Installation

Download and extract the latest install script from Artifactory using the latest release from [Github](https://github.ibm.com/ai-foundation/wml-serving/releases) (use your IBM credentials to log in):

```shell
VERSION=0.4.2_106
DOWNLOAD_URL=https://na.artifactory.swg-devops.com/artifactory/wcp-ai-foundation-team-generic-virtual/model-serving/wml-serving-${VERSION}-install.sh

wget --user=you@us.ibm.com --ask-password ${DOWNLOAD_URL}

# If you don't have wget you can also use curl
curl -u ${artifactory_user}:${artifactory_apikey} -o wml-serving-install.sh ${DOWNLOAD_URL}
```

Run the downloaded script to install Model-Mesh Serving CRDs, controller, and built-in runtimes into the specified Kubernetes namespaces, after reviewing the command line flags below.

An Artifactory user and API key are required to pull the Model-Mesh Serving tar file and images. These can be passed in as flags or set as environment variables.

A Kubernetes `--namespace` is required, which must already exist. You must also have cluster-admin authority and cluster access must be configured prior to running the install script.

An Artifactory user and API key is required if you are pulling the Model-Mesh Serving tar file and images. These can be passed in as flags or set as environment variables.

The `--quickstart` option can be specified to install and configure supporting datastores in the same namespace (etcd and MinIO) for experimental/development use. If this is not chosen, the namespace provided must have an Etcd secret named `model-serving-etcd` created which provides access to the Etcd cluster. See the [instructions above](#setup-the-etcd-connection-information) on this step.

```shell
$ ./wml-serving-install.sh --help
usage: ./wml-serving-install.sh [flags]

Flags:
  -n, --namespace                (required) Kubernetes namespace to deploy Model-Mesh Serving to.
  -u, --artifactory-user         Artifactory username to pull Model-Mesh Serving tarfile and images, can also set with env var ARTIFACTORY_USER.
  -a, --artifactory-apikey       Artifactory API key to pull Model-Mesh Serving tarfile and images, can also set with env var ARTIFACTORY_APIKEY.
  -v, --model-serve-version      Model-Mesh Serving version to pull and use. Example: wml-serving-0.3.0_165
  -p, --install-config-path      Path to local model serve installation configs. Can be Model-Mesh Serving tarfile or directory.
  -d, --delete                   Delete any existing instances of Model-Mesh Serving in Kube namespace before running install, including CRDs, RBACs, controller, older CRD with ai.ibm.com api group name, etc.
  --quickstart                   Install and configure required supporting datastores in the same namespace (etcd and MinIO) - for experimentation/development
  --redsonja-images              Use images pulled from redsonja IBM Container registry, requires redsonja user and apikey.
  --redsonja-apikey              IBM container registry apikey that has access to redsonja account, can also set with env var REDSONJA_APIKEY.
```

As you can see, you can optionally provide an explicit `--model-serve-version` for the Model-Mesh Serving tar file, which will be pulled down from Artifactory via the install script. If not provided, the version installed will be the one corresponding to the version of the install script. Alternatively, you can provide a local `--install-config-path` that points to a local Model-Mesh Serving tar file or directory containing Model-Mesh Serving configs to deploy.

You can also optionally use `--delete` to delete any existing instances of Model-Mesh Serving in the designated Kube namespace before running the install.

The installation will create a secret named `storage-config` if it does not already exist. If the `--quickstart` option was chosen, this will be populated with the connection details for the example models bucket in IBM Cloud Object Storage and the local MinIO; otherwise, it will be empty and ready for you to add your own entries.

To deploy Model-Mesh Serving with images from IBM Container Registry (ICR) instead of using the default Artifactory images, add flag `--redsonja-images`. Note that the Kube dockerconfig secret `ibm-entitlement-key` in the deployed namespace must have a username (likely `iamapikey`) and api key that has access to the `redsonja_hyboria/ai-foundation` namespace in the redsonja ICR account. If the Kube secret does not already exist, you can provide your `--redsonja-apikey <APIKEY>` and the secret will be created in the install script.

## Next Steps

- Continue with the "Deployed Components" section of the [Getting Started](../install#deployed-components) page.
