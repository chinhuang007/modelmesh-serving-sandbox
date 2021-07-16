---
title: Developer Quick Start
---

To get started quickly with the controller functions you can follow this walkthrough. This assumes a cluster without OLM installed, you can likely skip these steps on OpenShift.

## Prerequisites

You need a running etcd instance and the endpoint configuration embedded into a secret.

Create a file named etcd-connection:

```
{
  "endpoints": "http://example-client.default.svc.cluster.local:2379"
}
```

Then create the secret:

```
kubectl create secret generic etcd-connection --from-file=etcd_connection=etcd-config.json
```

## Running the controller

git clone this repository, then run:

```
make generate
make install
go run ./main.go
```

## Applying resources

Once the controller is running, deploy the example runtime:

```
kubectl apply -f config/samples/servingruntime.yaml
```

Then deploy a predictor:

```
kubectl apply -f config/samples/predictor.yaml
```

A few minutes after applying, you should see a number of pods related to the deployment, for example:

```
✗ kubectl get pods
NAME                             READY   STATUS     RESTARTS   AGE
etcd-operator-67b5648d4f-gsqhf   3/3     Running    0          72m
tensorflow-bmlk8spfz9            0/1     Init:0/1   0          115s
tensorflow-d4b584fbf-55kzz       2/2     Running    0          115s
tensorflow-d4b584fbf-dmf7d       2/2     Running    0          115s
```
