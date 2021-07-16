---
title: Airgap Installation
---

### Step-by-Step Instructions to setup and install wml-serving operator in Airgap environment

#### 0. Setup ENV

```
export CASENAME=ibm-ai-wmlserving
export OFFLINEDIR=ibm-ai-wmlserving-archive
export NS=airgap-test
export ITEM=wmlservingOperatorSetup
export ACTION=install-catalog
export SOURCE_REGISTRY=cp.stg.icr.io
export SOURCE_REGISTRY_USER=xxxxxxxxx
export SOURCE_REGISTRY_PASS=xxxxxxxxx
export TARGET_REGISTRY=hyc-wml-devops-team-air-gap-test-docker-local.artifactory.swg-devops.com
export TARGET_REGISTRY_USER=xxxxxxxxx
export TARGET_REGISTRY_PASS=xxxxxxxxx
oc new-project $NS
```

#### 1. Compress the wml-serving case bundle and upload it to airgap cluster

```
git clone git@github.ibm.com:ai-foundation/ibm-ai-wmlserving-case-bundle.git --recursive
tar -czvf ibm-ai-wmlserving-case-bundle.tar ibm-ai-wmlserving-case-bundle
```

#### 2. Extract the latest ibm-ai-wmlserving-case-bundle.tar file

`tar xzvf ibm-ai-wmlserving-case-bundle.tar`

Note: To copy case bundle tarball to airgap cluster , Use either one of the approach

- Use a mount point where the tarball is present
- Use USBDrive which has the required tarball and copy it to airgap cluster

#### 3. Create a temp docker registry - this act as a customer registry. In this case we use `hyc-wml-devops-team-air-gap-test-docker-local.artifactory.swg-devops.com` as a customer registry

#### 4. Download the case using cloudctl command

cd to case directory

```
cloudctl case save -c ibm-ai-wmlserving  -t 1 -o ibm-ai-wmlserving-archive
```

#### 5. Replace the prod resgitry to the internal registry that contain the paid content image

Update `ibm-ai-wmlserving-archive/ibm-ai-wmlserving-1.0.0-images.csv` file pointing
registry to `cp.stg.icr.io`
image_name to `cp/ai-foundation/<image_name>`

#### 6. Set up temporary registry service (You might need to install this package: `sudo yum install httpd-tools -y` if it is not installed yet)

```
ibm-ai-wmlserving/inventory/wmlservingOperatorSetup/files/airgap.sh registry service init
```

#### 7. Create a source registry credential

```
cloudctl case launch \
--case $CASENAME \
--namespace $NS \
--inventory $ITEM \
--action configure-creds-airgap \
--args "--registry $SOURCE_REGISTRY --user $SOURCE_REGISTRY_USER --pass $SOURCE_REGISTRY_PASS" \
--tolerance 1
```

#### 8. Create a destination registry credential

```
cloudctl case launch  \
  --case $CASENAME    \
  --namespace $NS     \
  --inventory $ITEM   \
  --action configure-creds-airgap  \
  --args "--registry $TARGET_REGISTRY --user $TARGET_REGISTRY_USER --pass $TARGET_REGISTRY_PASS" \
  --tolerance 1
```

#### 9. Mirror the images

```
cloudctl case launch  \
  --case $CASENAME    \
  --namespace $NS     \
  --inventory $ITEM   \
  --action mirror-images  \
  --args "--registry $TARGET_REGISTRY --inputDir $OFFLINEDIR" \
  --tolerance 1
```

#### 10. Configure cluster with `ImageContentSourcePolicy`

```
cloudctl case launch  \
  --case $CASENAME    \
  --namespace $NS     \
  --inventory $ITEM   \
  --action configure-cluster-airgap  \
  --args "--registry $TARGET_REGISTRY --inputDir $OFFLINEDIR " \
  --tolerance 1
```

#### 11. Update the `ImageContentSourcePolicy`

- oc edit ImageContentSourcePolicy ibm-ai-wmlserving
- Add the below entries under `spec`--`repositoryDigestMirrors`

```
  - mirrors:
    - hyc-wml-devops-team-air-gap-test-docker-local.artifactory.swg-devops.com/cp/ai-foundation
    source: cp.icr.io/cp/ai
  - mirrors:
    - hyc-wml-devops-team-air-gap-test-docker-local.artifactory.swg-devops.com/cp/ai-foundation
    source: icr.io/cpopen
```

**_Please note that this would restart all the nodes , please wait till all the nodes are restarted and back to Ready state_**

#### 12. Run the Catalog installation test

```
cloudctl case launch  \
  --case $CASENAME    \
  --namespace openshift-marketplace   \
  --inventory $ITEM   \
  --action $ACTION \
  --args "--registry icr.io --inputDir $OFFLINEDIR --recursive" \
  --tolerance 1
```

#### 13. Run the Operator installation test

`export ACTION=install-operator`

```
cloudctl case launch  \
  --case $CASENAME  \
  --namespace $NS     \
  --action $ACTION \
  --inventory $ITEM   \
  --tolerance 1
```
