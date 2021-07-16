---
title: Configuration
---
# Configuration

System-wide configuration parameters can be set by creating a ConfigMap with name `model-serving-config`. It should contain a single key named `config.yaml`, whose value is a yaml doc containing the configuration. All parameters have defaults and are optional. If the ConfigMap does not exist, all parameters will take their defaults.

The configuration can be updated at runtime and will take effect immediately. Be aware however that certain changes could cause temporary disruption to the service - in particular changing the service name, port, TLS configuration and/or headlessness.

Example:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: model-serving-config
data:
  config.yaml: |
    #Sample config overrides
    inferenceServiceName: "wml-serving"
    inferenceServicePort: 8033
    podsPerRuntime: 2
    metrics:
      enabled: true
```

The following parameters are currently supported. _Note_ the keys are expressed here in camel case but are in fact case-insensitive.

| Variable                         | Description                                                                                             | Default              |
| -------------------------------- | ------------------------------------------------------------------------------------------------------- | -------------------- |
| `etcdSecretName`                 | DEPRECATED - The secret which contains etcd connection information (\* see below)                       | `model-serving-etcd` |
| `inferenceServiceName`           | The service name which is used for communication with the serving server                                | `wml-serving`        |
| `inferenceServicePort`           | The port number for communication with the inferencing service                                          | `8033`               |
| `storageSecretName`              | The secret containing entries for each storage backend from which models can be loaded (\*\* see below) | `storage-config`     |
| `podsPerRuntime`                 | Number of server Pods to run per enabled Serving Runtime (\*\*\* see below)                             | `2`                  |
| `tls.secretName`                 | Kubernetes TLS type secret to use for securing the Service; no TLS if empty (\*\*\*\* see below)        |                      |
| `tls.clientAuth`                 | Enables mutual TLS authentication. Supported values are `required` and `optional`, disabled if empty    |                      |
| `headlessService`                | Whether the Service should be headless (recommended)                                                    | `true`               |
| `enableAccessLogging`            | Enables logging of each request to the model server                                                     | `false`              |
| `serviceAccountName`             | The service account to use for runtime Pods                                                             | `wmlserving`         |
| `metrics.enabled`                | Enables serving of Prometheus metrics                                                                   | `true`               |
| `metrics.port`                   | Port on which to serve metrics via the `/metrics` endpoint                                              | `2112`               |
| `metrics.scheme`                 | Scheme to use for the `/metrics` endpoint (`http` or `https`)                                           | `https`              |
| `scaleToZero.enabled`            | Whether to scale down Serving Runtimes that have no Predictors                                          | `true`               |
| `scaleToZero.gracePeriodSeconds` | The number of seconds to wait after Predictors are deleted before scaling to zero                       | `60`                 |

(\*) The `etcdSecretName` variable has been deprecated and moved to an environment variable set by the operator. The proper way to modify this value is to set `spec.etcd.externalConnection.secretName` in the `WmlServing` resource.
The etcd configuration secret must contain a key named `etcd_connection` and have contents that conforms to https://github.com/IBM/etcd-java/blob/master/etcd-json-schema.md. If the etcd server is shared between namespaces and/or environments, ensure that the `root_prefix` json attribute is set to something unique per namespace.

(\*\*) Currently requires a controller restart to take effect

(\*\*\*) This parameter will likely be removed in a future release; the Pod replica counts will become more dynamic.

(\*\*\*\*) The TLS configuration secret allows for keys:

- `tls.crt` - path to TLS secret certificate
- `tls.key` - path to TLS secret key
- `ca.crt` (optional) - single path or comma-separated list of paths to trusted certificates

## Defining roles based on personas

1. The `Cluster-Admin`

   Installs wml-serving operator. This is top level role with widest privileges among all the personas.

2. The `Namespace-Admin` or `WMLServing-Admin`

   This persona can install and configure Model-Mesh Serving, `ServingRuntime`s, `ConfigMap`s and storage `Secrets`.

3. The `Model-Deployer`

   User pertaining to this persona can create, update and delete `Predictor`s.

There could be more personas or change in scope of existing personas as wml-serving operator evolves in future.

## Generating TLS Certificates for Dev/Test

TLS is enabled through adding a value for `tls.secretName` in the user's ConfigMap that points to an existing kube secret with TLS key/cert details.

To create a SAN key/cert for TLS, use command:

```sh
$ openssl req -x509 -newkey rsa:4096 -sha256 -days 3560 -nodes -keyout example.key -out example.crt -subj '/CN=wml-serving' -extensions san -config openssl-san.config
```

Where the contents of `openssl-san.config` look like:

```
[ req ]
distinguished_name = req
[ san ]
subjectAltName = DNS:wml-serving,DNS:localhost,IP:0.0.0.0
```

With the generated key/cert, create a kube secret with contents like:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: <NAME_OF_SECRET>
type: kubernetes.io/tls
stringData:
  tls.crt: <contents-of-example.crt>
  tls.key: <contents-of-example.key>
  ca.crt: <contents-of-example.crt>
```

For basic TLS, only the fields `tls.crt` and `tls.key` are needed in the kube secret. For mutual TLS, add `ca.crt` in the kube secret and set the configuration `tls.clientAuth` to `require` in the ConfigMap `model-serving-config`.

