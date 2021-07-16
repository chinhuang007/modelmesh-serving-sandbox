---
title: Minikube Installation
---

# Minikube

This guide provides instructions for using Model-Mesh Serving with minikube as well as some common development workflows.

### Start minikube

```
minikube start --vm-driver=virtualbox
```

### Setup Namespaces

```
kubectl create namespace wml-serving
kubectl config set-context --current --namespace=wml-serving
```

### Start and configure Etcd

Startup a new terminal session where the etcd server will be hosted.

Run the command:

```
etcd --listen-client-urls http://0.0.0.0:2379 --advertise-client-urls http://0.0.0.0:2379
```

Etcd will be listening on your local host address, but since minikube will be running in a virtual machine you will need to know the host address to setup the communication. On mac systems you can find your local address with the command:
`ifconfig | grep -A 1 'en0' | tail -1 | cut -d ':' -f 2 | cut -d ' ' -f 2`

Next, create a file named etcd-config.json with the contents, making sure to enter your local address:

```
{
  "endpoints": "http://<host ip>:2379",
  "root_prefix":  "model-mesh"
}
```

Use that file to create a secret:

```
kubectl create secret generic model-serving-etcd --from-file=etcd_connection=etcd-config.json
```

### Override Configuration Defaults

To override configuration defaults, create a config map in the current namespace:

```
apiVersion: v1
kind: ConfigMap
metadata:
  name: model-serving-config
data:
  config.yaml: |
# Sample config overrides
#inferenceServiceName: "model-mesh"
#inferenceServicePort: 8033
#podsPerDeployment: 2
#modelMesh:
#  image: us.icr.io/model-mesh/model-mesh
#  tag: main-20201029-3
#puller:
#  image: us.icr.io/model-mesh/model-serving-puller
#  tag: develop-20201103-3
```

# Workflows

## Building the image and running the deployment

```
#install the CRDs, just do this once on this cluster unless you are modifying CRDs
make install

# set the docker env to point to minikube
eval $(minikube -p minikube docker-env)

# build the image into the minikube docker env
make build

# deploy
make deploy

# If you are on minikube without a registry, you need to set the image pull policy to not pull
kubectl patch deployment wmlserving-controller --type=json -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/imagePullPolicy", "value": "Never"}]'
```

## Building an image with a registry

Set the IMG environment variable to build the image with the target registry name:

```
IMG=myregistry/myimage:tag make build
```

Push it

```
docker push myregistry/myimage:tag
```

Then deploy it:

```
IMG=myregistry/myimage:tag make deploy
```

## Run the controller in terminal

Note that this will run the controller in your current kube context so the behavior will differ slightly from the deployment which runs on cluster.

```
# Scale down the deployment to 0
kubectl scale --replicas=0 deployment/wmlserving-controller

# Run the controller
ETCD_SECRET_NAME=model-mesh-etcd go run *.go
```

## Modifying CRDs

Regenerate the CRDs:

```
make manifests generate
```

Assure they install:

```
make install
```

## Pre-integration verification

```
# Run formatter, linter, and tests, build the image
# prefixing task list with 'run' task will run these in a docker container
make run build test
```
