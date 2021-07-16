---
title: Upgrade and Rollback
---

To understand the upgrade and rollback procedures, it is important to understand how components of Model-Mesh Serving are versioned. There is a distinction between the OLM operator that abstracts the installation and management of the components from the operand that is the components themselves. The version and configuration of the components are controlled by the `WmlServing` resource.

The operator and operand are versioned independently. Each version of the operator can install a range of versions of the operand. The desired version of the operand is set in the `spec.version` field of the `WmlServing` resource, the status field `versions.reconciled` indicates the instantiated version in the cluster, and the status field `versions.available.versions` lists the versions that can be selected.

### Upgrade

The versions of the operator and operand are decoupled. Upgrading Model-Mesh Serving to the latest version requires updates to each of them.

The upgrade is non-disruptive as changes are made using rolling updates of Kubernetes Deployments. The model mesh container in the Deployments coordinates the rollout to ensure that evicted models are loaded in other pods before shutting down in order to prevent any major disruption in capacity of any particular model.

#### Procedure

1. Cluster administrator pulls in an update to the OLM catalog of IBM provided operators, including updates to the catalog for Model-Mesh Serving
2. Based on the OLM Subscription, the Model-Mesh Serving operator is upgraded either automatically or after manual approval
   - This will roll out an update to the `wmlserving-operator` pod, but the Model-Mesh Serving components will not be changed
3. The namespace administrator chooses an available version and updates the `spec.version` field on the `WmlServing` custom resource
4. The change to the desired version triggers reconciliation of the `WmlServing` resource and upgrades the Model-Mesh Serving components
   - This may roll out updates to the `wmlserving-controller` pod and to `ServingRuntimes`

### Rollback

If an issue is encountered after an upgrade, the previous stable state can be restored by changing the `WmlServing` resource's `spec.version` back to the previous value. OLM does not support reverting the version of the installed operator, but this should not be necessary to restore a stable state to the Model-Mesh Serving components.

#### Procedure

1. The namespace administrator sets the `spec.version` field on the `WmlServing` custom resource back to its previous value
2. The change to the desired version triggers reconciliation of the Model-Mesh Serving components
   - This may roll out updates to the `wmlserving-controller` pod and to `ServingRuntimes`

### Example

To validate the procedures, upgrade testing was performed during the development of Model-Mesh Serving. The tests below were completed on April 30, 2021. This details the steps in upgrading the operator from a v1.0.0 release to a mock v1.0.1 release and upgrades the Model-Mesh Serving components from `0.4.2` to a mock `0.4.3-rc` release. The mock versions were built from a branch with an additional environment variable to cause an update to every pod.

---

A functioning installation of the operator and operand is assumed as a prerequisite.

1.  Enable manual approval of InstallPlans in the namespace with Model-Mesh Serving installed to better control the upgrade

    ```
    $ oc patch subscription ibm-ai-wmlserving --type json \
    --patch '[{"op": "replace", "path": "/spec/installPlanApproval", "value": "Manual"}]'
    ```

1.  Record the initial state of the Operator and the Model-Mesh Serving components from the install

    OpenShift UI shows version 1.0.0 is installed.
    ![image](https://media.github.ibm.com/user/9339/files/75de2200-a94d-11eb-882e-c6fba4bcd5b0)

    State of the cluster after the install:

    ```
    $ oc get csv
    NAME                       DISPLAY             VERSION   REPLACES   PHASE
    ibm-ai-wmlserving.v1.0.0   Ibm Ai Wmlserving   1.0.0                Succeeded
    ibm-etcd-operator.v0.0.1   Etcd                0.0.1                Succeeded

    $ oc describe wmlserving wmlserving
    ...
      Versions:
        Available:
          Versions:
            Name:    0.3.0
            Name:    0.4.2
        Reconciled:  0.4.2
    Events:          <none>

    $ oc get po,predictors
    NAME                                               READY   STATUS    RESTARTS   AGE
    pod/ibm-etcd-operator-564864f5f6-rmw7j             1/1     Running   0          3m15s
    pod/wmlserving-msp-ml-server-0.x-7f65d687c-c9p6x   3/3     Running   0          8m45s
    pod/wmlserving-msp-ml-server-0.x-7f65d687c-xf7xj   3/3     Running   0          8m46s
    pod/wmlserving-mlserver-0.x-6b57549cc4-5qcfk       3/3     Running   0          8m45s
    pod/wmlserving-mlserver-0.x-6b57549cc4-6m8ml       3/3     Running   0          8m45s
    pod/wmlserving-triton-2.x-65f7c67c48-b2wzt         3/3     Running   0          8m45s
    pod/wmlserving-triton-2.x-65f7c67c48-vtzb6         3/3     Running   0          8m44s
    pod/wmlserving-controller-7c84b744b5-lhqhd         1/1     Running   0          9m13s
    pod/wmlserving-etcd-0                              1/1     Running   0          9m5s
    pod/wmlserving-etcd-1                              1/1     Running   0          8m50s
    pod/wmlserving-etcd-2                              1/1     Running   0          8m34s
    pod/wmlserving-operator-58779759cf-79t7v           1/1     Running   0          9m46s

    NAME                                                       TYPE      AVAILABLE   ACTIVEMODEL   TARGETMODEL   TRANSITION   AGE
    predictor.wmlserving.ai.ibm.com/example-xgboost-mushroom   xgboost   true        Loaded                      UpToDate     3m28s
    ```

1.  In this test, pulling in an updated to the catalog of operators is simulated by manually updating the image referenced in the `ibm-ai-wmlserving-catalog` CatalogSource to a new image digest.

    ```
    $ oc -n openshift-marketplace patch catalogsource ibm-ai-wmlserving-catalog --type json \
      --patch '[{"op": "replace", "path": "/spec/image", "value": "icr.io/cpopen/ibm-ai-wmlserving-operator-catalog@sha256:f3b9b65c0f0b0b0255502154df0b47f46fa37c8ba651116c47c1839431b56da9"}]'
    ```

    After the catalog source pod was refreshed, the UI showed that operator version `1.0.1` is the latest version and that an upgrade was available to the installed operator:

    ![image](https://media.github.ibm.com/user/9339/files/da997c80-a94d-11eb-883b-656e23cad105)

    ![image](https://media.github.ibm.com/user/9339/files/074d9400-a94e-11eb-8a9c-9c2825bd202e)

1.  Approve the InstallPlan to trigger an upgrade the operator and watch the changes. The `wmlserving-operator` pod is updated, but no other pods changed.

    ```
    $ oc get po -w
    NAME                                     READY   STATUS              RESTARTS   AGE
    ...
    wmlserving-operator-58779759cf-79t7v     1/1     Running             0          16m
    wmlserving-operator-6c55ddd6f-6phml      0/1     Pending             0          0s
    wmlserving-operator-6c55ddd6f-6phml      0/1     Pending             0          0s
    wmlserving-operator-6c55ddd6f-6phml      0/1     ContainerCreating   0          0s
    wmlserving-operator-6c55ddd6f-6phml      0/1     ContainerCreating   0          2s
    wmlserving-operator-6c55ddd6f-6phml      0/1     Running             0          7s
    wmlserving-operator-6c55ddd6f-6phml      1/1     Running             0          18s
    wmlserving-operator-58779759cf-79t7v     1/1     Terminating         0          16m
    wmlserving-operator-58779759cf-79t7v     0/1     Terminating         0          16m
    wmlserving-operator-58779759cf-79t7v     0/1     Terminating         0          16m
    wmlserving-operator-58779759cf-79t7v     0/1     Terminating         0          16m
    ```

1.  Check that the update to the operator image added `0.4.3-rc` to the list of available versions on the `WmlServing` resource, but has not changed the reconciled version

    ```
    $ oc describe wmlserving wmlserving
    ...
      Versions:
        Available:
          Versions:
            Name:    0.3.0
            Name:    0.4.2
            Name:    0.4.3-rc
        Reconciled:  0.4.2
    Events:          <none>
    ```

1.  Update the `spec.version` field to trigger the upgrade of the operand to trigger a rollout of controller resources including the `wmlserving-controller` pod, ServingRuntimes, and the runtime pods

    ```
    $ oc patch wmlserving wmlserving --type json \
      --patch '[{"op": "replace", "path": "/spec/version", "value": "0.4.3-rc"}]'
    ```

1.  Check the end state of the cluster after the rollout of component changes is complete

    ```
    $ oc get csv
    NAME                       DISPLAY             VERSION   REPLACES                   PHASE
    ibm-ai-wmlserving.v1.0.1   Ibm Ai Wmlserving   1.0.1     ibm-ai-wmlserving.v1.0.0   Succeeded
    ibm-etcd-operator.v0.0.1   Etcd                0.0.1                                Succeeded

    $ oc get po,predictors
    NAME                                                READY   STATUS    RESTARTS   AGE
    pod/ibm-etcd-operator-564864f5f6-rmw7j              1/1     Running   0          20m
    pod/wmlserving-msp-ml-server-0.x-5757bcd744-cw6pw   3/3     Running   0          7m34s
    pod/wmlserving-msp-ml-server-0.x-5757bcd744-qmlpg   3/3     Running   0          7m32s
    pod/wmlserving-mlserver-0.x-76fb4fb76b-bjzzj        3/3     Running   0          7m33s
    pod/wmlserving-mlserver-0.x-76fb4fb76b-hlt49        3/3     Running   0          7m32s
    pod/wmlserving-triton-2.x-7879fbdc79-dj6rr          3/3     Running   0          7m31s
    pod/wmlserving-triton-2.x-7879fbdc79-tmlml          3/3     Running   0          7m32s
    pod/wmlserving-controller-6bf74d884b-pdqsm          1/1     Running   0          7m35s
    pod/wmlserving-etcd-0                               1/1     Running   0          26m
    pod/wmlserving-etcd-1                               1/1     Running   0          26m
    pod/wmlserving-etcd-2                               1/1     Running   0          26m
    pod/wmlserving-operator-6c55ddd6f-6phml             1/1     Running   0          10m

    NAME                                                       TYPE      AVAILABLE   ACTIVEMODEL   TARGETMODEL   TRANSITION   AGE
    predictor.wmlserving.ai.ibm.com/example-xgboost-mushroom   xgboost   true        Loaded                      UpToDate     20m

    ```

    And the `WmlServing` status shows that the reconciled version is now at `0.4.3-rc`

    ```
    $ oc describe wmlserving wmlserving
    ...
      Versions:
        Available:
          Versions:
            Name:    0.3.0
            Name:    0.4.2
            Name:    0.4.3-rc
        Reconciled:  0.4.3-rc
    Events:          <none>
    ```

1.  If any issue were to be encountered with version `0.4.3-rc`, the previous state could be restored by reverting the `spec.version` field back to `0.4.2`

    ```
    $ oc patch wmlserving wmlserving --type json \
      --patch '[{"op": "replace", "path": "/spec/version", "value": "0.4.2"}]'
    ```
