---
title: Known Issues
---

<InlineNotification>

**NOTE:**

These issues apply to the Model-Mesh Serving `0.5.0` release and most if not all should be addressed in the subsequent release.

</InlineNotification>

## Operator

### Model-Mesh Serving operator won't install with cluster scope

`Model-Mesh Serving` operator is supported ONLY at namespace scope, cluster scope will be supported soon.

### `WmlServing` CR shows empty status during installation

During initial installation there isn't very clear indication that things are "working", it can be up to 10 mins later that the `WmlServing` CR gets into `READY` state.

**Workaround**

Check the `WmlServing` CR status as shown below until it shows `True`.

```
oc get wmlserving
NAME                READY
wmlserving-sample   True
```

### Update/Patch of `WmlServing` CR has no effect

Currently `WmlServing` CR doesn't support patch/update.

**Workaround**

The workaround is to delete and recreate the `WmlServing` CR after making the necessary change.

### Deleted `ServingRuntime` CR reappears when operator reconciles

When the `WmlServing` resource is reconciled, the built-in `ServingRuntime`s will be recreated.

**Workaround**

Disable the `ServingRuntime` instead of deleting it. As an example, the below command will disable the MSP MLServer `ServingRuntime`:

```
oc patch servingruntime msp-ml-server-0.x  -p '[{"op": "replace", "path": "/spec/disabled", "value": true}]'  --type json
```

### Operator pod restarts due to leader election

We have observed unexpected restarts of the operator pod due to failures in leader election. The logs will indicate this with an error like:

```
{"error": "leader election lost"}
```

As long as these restarts are not occurring frequently, they should not affect the functionality of the Operator.

## Predictor

### `Predictor` CR takes long time to get into `Loaded` state

There can be a delay before `Predictor` CR reaches Pending/Loading state if it is the first `Predictor` for a given `ServingRuntime`, even once the corresponding Pods are in Running state.

### Delay in loading the first Predictor assigned to the Triton runtime

Predictors assigned to the Triton runtime may be stuck in the Pending state for some time while the Triton pods are being created. The Triton image is large and may take a while to download.

### Predictors with an unrecognized model type can remain in `Pending` state if there are no other valid Predictors created

Predictors might get stuck in `Pending` state if they do not have recognized model type or explicit runtime assignment and there are no other valid Predictors.

**Workaround**

Ensure that you specify a supported model type and/or runtime name in the `Predictor` CR.
