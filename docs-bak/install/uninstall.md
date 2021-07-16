---
title: Uninstalling Model-Mesh Serving Operator/Catalog
---

### Uninstallation of WMLServing Operator/Catalog

**Cleanup existing Predictors**

Execute the below command to delete all predictors at once:

`oc delete predictor --all -n $NAMESPACE`

**Delete WMLServing instance**

`oc delete wmlserving <wmlserving_CR_name> -n $NAMESPACE`

**Note**: The above command deletes the underlying `etcdCluster` instance, but the corresponding etcd PVCs are not deleted to avoid data loss.
It is recommended to take a backup if required and delete them to avoid reusing with new WMLServing instance of same name. The existing metadata in pvc might create conflict with new data that will be generated.

Following naming convention is used for etcd PVCs:

`data-<wmlserving_CR_name>-etcd-<replica no>`

For more details, refer to etcd [documentation](https://github.ibm.com/CloudPakOpenContent/ibm-etcd-operator/blob/master/README.md)

**Uninstall WMLServing Operator**

```
cloudctl case launch  \
   --case ibm-ai-wmlserving-1.0.0.tgz  \
   --tolerance 1  \
   --namespace $NAMESPACE  \
   --action uninstall-operator  \
   --inventory wmlservingOperatorSetup
```

**Uninstall WMLServing Catalog**

```
cloudctl case launch  \
   --case ibm-ai-wmlserving-1.0.0.tgz  \
   --namespace openshift-marketplace  \
   --inventory wmlservingOperatorSetup  \
   --action uninstall-catalog  \
   --tolerance 1
```

**Uninstall the ibm-common-service/ibm-etcd operators**

1. Delete the operator subscription:

Run the following command to check the subscriptions:

`oc get subscription -n $NAMESPACE`

Delete the subscriptions for both `ibm-common-service-operator` and `ibm-etcd-operator` using the below command:

`oc delete subscription <subscription_name> -n $NAMESPACE`

2. Delete Cluster Service Version (CSV):

Run the following command to check the CSVs:

`oc get csv -n $NAMESPACE`

Delete the CSVs for both `ibm-common-service-operator` and `ibm-etcd-operator` using the below command:

`oc delete csv <csv_name> -n $NAMESPACE`

**Delete Custom Resource Definitions**

```
oc delete CustomResourceDefinition wmlservings.ai.ibm.com
oc delete CustomResourceDefinition servingruntimes.wmlserving.ai.ibm.com
oc delete CustomResourceDefinition predictors.wmlserving.ai.ibm.com
```
