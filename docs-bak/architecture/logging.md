---
title: Logging
description: Logging
---

## System Level Logging

Model-Mesh Serving does not implement system level logging as this is provided by the OpenShift cluster which hosts the Model-Mesh Serving components and the associated products. It is the responsibility of the OpenShift deployer to assure that the cluster has been deployed and configured properly in order to provide system level logging.

## Application Logging

Model-Mesh Serving components are always embedded into IBM products and are not deployable by end users in a standalone manner. Because of this, the notion of the 'application' is incomplete until the Model-Mesh Serving component has been incorporated into the target product. However, Model-Mesh Serving components do provide logging for sake of problem determination as well as on any user accessible boundaries.

For example, since Model-Mesh Serving provides Kubernetes control plane extensions through Custom Resource Definitions (CRDs) and these are available through the Kubernetes APIs, a user can directly access Model-Mesh Serving through this API. Conversely, the model serving HTTP and GRPC APIs exposed by the model server are designated product use only. These APIs do not provide audit logging by default, though additional access logging can be configured. Finally, some individual runtime components may provide additional auditing capabilities, but the application provider must assess whether these satisfy the overall logging requirements.

### Runtime Components

The following table describes the various runtime components which comprise the Model-Mesh Serving component and the audit logging concerns of the respective component.

| Runtime Component | Audit Logging                                                                                                                                                              | Client Content Access Logging                                                                                                         |
| ----------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------- |
| Operator          | The operator exposes it's API through the Kubernetes API server. As such, the auditing of all API accesses is available through the system level Kubernetes audit logs.    | The Operator does not store client content                                                                                            |
| Controller        | The controller exposes it's API through the Kubernetes API server. As such, the auditing of all API accesses is available through the system level Kubernetes audit logs.  | The controller does not store client content                                                                                          |
| Runtime           | The model serving runtime can be configured to log accesses, but embedding applications should consider auditing in the application layer instead                          | Client content is available only through the model serving API. Content accesses should be audited through the application API layer. |
| Etcd              | Etcd is only accessed by the controller and runtime components and not directly accessed. All audits are captured by the Kubernetes audit log or the runtime audit log     | The etcd server does not store client content                                                                                         |

### Access Logging

Individual model serving requests are not logged by default since depending upon the application, this may incur substantial performance overhead yet might not be required by a particular application.

This type of logging can be enabled through the configuration parameter `EnableAccessLogging`. See [configuration](../configuration) for futher details regarding how this and other configurable features are enabled.
