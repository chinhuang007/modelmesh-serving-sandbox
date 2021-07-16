---
title: Operator-based Installation
---

The recommended approach to installing Model-Mesh Serving is to use OLM and the operator. If you are installing in a development environment and wish to skip the operator, see [No Operator Installation](install-script).

Download the case bundle from https://github.com/IBM/cloud-pak/raw/master/repo/case/ibm-ai-wmlserving-1.0.0.tgz

## Additional prerequisite steps for non-release version

To use the the latest development version of the operator instead of a released one, perform these additional steps before continuing to the "Set the namespace" section.

### 1. Clone the case bundle repo

```shell
git clone git@github.ibm.com:ai-foundation/ibm-ai-wmlserving-case-bundle.git
```

### 2. Setup a registry mirror

Create an image mirror on the cluster which points to the staging registry.

```yaml
apiVersion: operator.openshift.io/v1alpha1
kind: ImageContentSourcePolicy
metadata:
  name: mirror-config
spec:
  repositoryDigestMirrors:
    - mirrors:
        - cp.stg.icr.io/cp/ai
      source: cp.icr.io/cp/ai
    - mirrors:
        - cp.stg.icr.io/cp
      source: icr.io/cpopen
```

This is will restart the nodes of the cluster one by one. Wait for all the nodes to be restarted.

### 3. Setup a global pull secret

#### a) Take a backup of the existing global pull secret.

```shell
oc get secret/pull-secret -n openshift-config -o yaml > pull-secret-bk.yaml
```

#### b) Update it to include credentials for `cp.stg.icr.io`.

```shell
pull_secret=$(echo -n "iamapikey:<APIKEY>" | base64 -w0)

oc get secret/pull-secret -n openshift-config -o jsonpath='{.data.\.dockerconfigjson}' | base64 -d | sed -e 's|"auths":{|"auths":{"cp.stg.icr.io":{"auth":"'$pull_secret'"\},|' > /tmp/dockerconfig.json

oc set data secret/pull-secret -n openshift-config --from-file=.dockerconfigjson=/tmp/dockerconfig.json
```

Follow the below links to get your own api key:

- https://playbook.cloudpaklab.ibm.com/entitled-registry-access/
- https://playbook.cloudpaklab.ibm.com/entitled-image-registry/#Read_Only_Access

## Install Model-Mesh Serving operator

### 1. Set the namespace

```shell
export NAMESPACE=<project-name>
```

### 2. Configure image pull secret

You must create a pull secret called `ibm-entitlement-key` in the target namespace. For more information see the [CloudPak Playbook](https://playbook.cloudpaklab.ibm.com/entitled-image-registry/#Utilize_entitled_registry_images_in_a_helm_chart)

```shell
oc create secret docker-registry ibm-entitlement-key -n $NAMESPACE \
  --docker-server=cp.icr.io \
  --docker-username="cp" \
  --docker-password="<APIKEY>"
```

Get the apikey from here: https://myibm.ibm.com/products-services/containerlibrary

### 3. Setup etcd

etcd is required to run Model-Mesh Serving. You can either bring your own etcd or install the `ibm-etcd-operator-catalog` so that the `wmlserving-operator` can create etcd for you.

#### Bring your own etcd

Create a file named `etcd-config.json`, populating the values based upon your etcd server. The same etcd server can be shared between environments and/or namespaces, but in this case _the `root_prefix` must be set differently in each namespace's respective secret_. The complete json schema for this configuration is documented [here](https://github.com/IBM/etcd-java/blob/master/etcd-json-schema.md).

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
oc create secret generic model-serving-etcd --from-file=etcd_connection=etcd-config.json
```

Place this secret name in the WmlServing CR which you will create in a later section.

```yaml
apiVersion: ai.ibm.com/v1
kind: WmlServing
metadata:
  name: wmlserving-sample
  namespace: $NAMESPACE
spec:
  etcd:
    externalConnection:
      secretName: model-serving-etcd
```

### 4. Install the ibm-etcd-operator-catalog

Deploy etcd catalog from https://github.ibm.com/CloudPakOpenContent/ibm-etcd-operator/

```yaml
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: ibm-etcd-operator-catalog
  namespace: openshift-marketplace
spec:
  displayName: IBM etcd operator Catalog
  publisher: IBM
  sourceType: grpc
  image: icr.io/cpopen/ibm-etcd-operator-catalog@sha256:dd1687d02b8fb35bc08464330471d6f6f6541c36ca56d2716d02fa071a5cc754
  updateStrategy:
    registryPoll:
      interval: 45m
```

**Note:**
The etcd-operator requires dynamic storage provisioning in order to create persistent volume claims.
For Fyre clusters, you can ssh to the `inf` node and run [this script](https://github.ibm.com/dacleyra/ocp-on-fyre/blob/master/nfs-storage-provisioner-ocpplus.sh) to setup an nfs provisioner.
Other dynamic storage providers are available but out of scope of this document.

### 5. Install the opencloud-operators catalog

Deploy the IBM Cloud Pak foundational services catalog source which is used to install cert-manager.

```yaml
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: opencloud-operators
  namespace: openshift-marketplace
spec:
  displayName: IBMCS Operators
  publisher: IBM
  sourceType: grpc
  image: icr.io/cpopen/ibm-common-service-catalog:latest
  updateStrategy:
    registryPoll:
      interval: 45m
```

### 6. Deploy the Model-Mesh Serving Operator catalog

```shell
cloudctl case launch --case ibm-ai-wmlserving-1.0.0.tgz --namespace openshift-marketplace --inventory wmlservingOperatorSetup --action install-catalog --args "--registry icr.io --inputDir /tmp/saved --recursive" --tolerance 1
```

### 7. Install the Model-Mesh Serving Operator

```shell
cloudctl case launch --case ibm-ai-wmlserving-1.0.0.tgz --tolerance 1 -n $NAMESPACE --action install-operator --inventory wmlservingOperatorSetup
```

All three operator pods (wmlserving, etcd and ibm-common-service) should be up and running.

```shell
$ oc get pods
NAME                                           READY   STATUS    RESTARTS   AGE
ibm-common-service-operator-8669686f89-64kcn   1/1     Running   0          5m20s
ibm-etcd-operator-65d44f5fb7-5fkwq             1/1     Running   0          5m22s
wmlserving-operator-74ffcb9976-hf7lf           1/1     Running   0          5m18s
```

### 8. Create a `WmlServing` Instance

To create an instance of Model-Mesh Serving and start the controller, you must create a `WmlServing` resource. The following sample uses the default config:

<InlineNotification>

**Note:**

- It is mandatory to provide the `storageClass` of your dynamic storage provisioner for etcd. For details of the supported storage classes refer to the etcd operator [documentation](https://github.ibm.com/CloudPakOpenContent/ibm-etcd-operator/blob/master/stable/ibm-etcd-operator-case-bundle/case/ibm-etcd-operator/README.md).
- Currently, Model-Mesh Serving CR doesn't support Update/Patch. If any change should be applied on the CR, it should be deleted and recreated again.

</InlineNotification>

```shell
oc apply -f - <<EOF
apiVersion: ai.ibm.com/v1
kind: WmlServing
metadata:
  name: wmlserving-sample
  namespace: $NAMESPACE
spec:
  etcd:
    storageClass: "managed-nfs-storage"
EOF
```

### 9. Check `WmlServing` CR status

```
oc get wmlserving wmlserving-sample
NAME                READY
wmlserving-sample   True
```

When the installation is complete, you should see the following pods:

```
$ oc get pods
NAME                                           READY   STATUS    RESTARTS   AGE
ibm-common-service-operator-8669686f89-64kcn   1/1     Running   0          5m20s
ibm-etcd-operator-65d44f5fb7-5fkwq             1/1     Running   0          5m22s
wmlserving-operator-74ffcb9976-hf7lf           1/1     Running   0          5m18s
wmlserving-controller-7dd947d49c-7x4sr         1/1     Running   0          3m18s
wmlserving-sample-etcd-0                       1/1     Running   0          3m9s
wmlserving-sample-etcd-1                       1/1     Running   0          2m58s
wmlserving-sample-etcd-2                       1/1     Running   0          2m39s
```

By default, built-in runtime pods will not be up as the `ScaleToZero` configuration parameter is enabled. The corresponding runtime pod spins up when ever a Predictor CR gets submitted.
Alternatively, `ScaleToZero` configuration parameter can also be disabled (see [configuration](../configuration)) so that Model-Mesh Serving will spin up all the built-in runtime pods as shown below:

```
wml-serving-mlserver-0.x-8499d6f846-kccf2          3/3     Running   0          7m59s
wml-serving-mlserver-0.x-66cd794bd5-hvpkp          3/3     Running   0          7m59s
wml-serving-msp-ml-server-0.x-67ddb449bb-tmhnc     3/3     Running   0          7m58s
wml-serving-msp-ml-server-0.x-67ddb449bb-xx27p     3/3     Running   0          7m58s
wml-serving-triton-2.x-7cf854dbf4-ft9qd            3/3     Running   0          7m58s
wml-serving-triton-2.x-7cf854dbf4-mkmxc            3/3     Running   0          7m58s
```

**Note:**
The video demonstration on how to install Model-Mesh Serving operator is available [here](OperatorInstallDemo.mp4).
