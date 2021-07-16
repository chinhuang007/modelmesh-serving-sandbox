---
title: Guidelines and Best Practices
description: Guidelines and Best Practices for deploying WML Core Serving
---

## Deployment of WML Core Serving

## Considerations

##### Resources considerations

- Always consider creating a new namespace while installing wml-serving operator, that would help to create and manage resources specific to serving.

- Usage Considerations:

  - Enable only serving runtimes required for your environment to reduce the footprint of resources
  - Memory and/or CPU resource allocations can be reduced (or increased)

    _refer to_ [Deployed Components](../install#deployed-components)

##### Security considerations

- ###### Secured communication

  - We recommend to enable TLS for secured communication between your application and wml-serving, _refer to_ [Configuration](../configuration) on how to enable TLS.

  - The wml-serving by default configured with restricted scc, you could also create your own custom scc to introduce more restrictions.refer to [Configuration](../configuration)

- ###### Securing and encrypting data

  Passive Encryption refers to encryption that is implemented within a storage device and outside of the application. Active Encryption refers to encryption that is implemented by the application. It is recommended to configure the cluster with cluster-wide passive encryption, and add active encryption where applicable.

  The etcd which is a key/value data store configured with wml-serving, used for storing internal meta data about the runtimes and predictors. It does not contain any sensitive data other than the meta data, but still you could consider encrypting the backing storage used for etcd cluster.

  **Note:** The wml-serving operator can deploy and configure a local etcd cluster using the etcd operator _or_ can configure wml-serving to use a (possibly shared) existing etcd service. See [the install documentation](<../install/#installing-with-the-operator-(recommended)>) for more details.

##### Performance and Scaling considerations

- The number of serving runtime PODs can be adjusted to control footprint/capacity for model deployments.

- Based on the inference requests for a model, number of copies will be increased to accommodate the load and maintain performance. Since there will be at most one copy of each model per deployed Pod, you can increase the number of runtime PODs to achieve a greater maximum request throughput

  _refer to_ `podsPerRuntime` in [Configuration](../configuration)

##### Backup considerations

- _refer to_ [Backup and Restore](backup-and-restore)

##### Upgrade considerations

- _refer to_ [Upgrade and Rollback](upgrade-and-rollback)

## Other considerations

##### Securing container content

See [OpenShift doc](https://docs.openshift.com/container-platform/4.7/security/container_security/security-container-content.html).

##### Using container registries securely

See [OpenShift doc](https://docs.openshift.com/container-platform/4.7/security/container_security/security-registries.html).

##### Securing the container platform

See [OpenShift doc](https://docs.openshift.com/container-platform/4.7/security/container_security/security-platform.html).

##### Securing attached storage

See [OpenShift doc](https://docs.openshift.com/container-platform/4.7/security/container_security/security-storage.html).

##### Managing security context constraints

See [OpenShift doc](https://docs.openshift.com/container-platform/4.7/authentication/managing-security-context-constraints.html).

##### Scanning pods for vulnerabilities

See [OpenShift doc](https://docs.openshift.com/container-platform/4.7/security/pod-vulnerability-scan.html).

##### Monitoring cluster events and logs

See [OpenShift doc](https://docs.openshift.com/container-platform/4.7/security/container_security/security-monitoring.html).
