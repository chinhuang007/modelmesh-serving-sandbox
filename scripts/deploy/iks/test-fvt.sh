#!/bin/bash

# Remove the x if you need no print out of each command
set -ex

# These env vars should come from the build.properties that `run-setup.sh` generates
echo "BUILD_NUMBER=${BUILD_NUMBER}"
echo "ARCHIVE_DIR=${ARCHIVE_DIR}"
echo "GIT_BRANCH=${GIT_BRANCH}"
echo "GIT_COMMIT=${GIT_COMMIT}"
echo "GIT_COMMIT_SHORT=${GIT_COMMIT_SHORT}"
echo "REGION=${REGION}"
echo "ORG=${ORG}"
echo "SPACE=${SPACE}"
echo "RESOURCE_GROUP=${RESOURCE_GROUP}"

# These env vars should come from the pipeline run environment properties
echo "SERVING_KUBERNETES_CLUSTER_NAME=$SERVING_KUBERNETES_CLUSTER_NAME"
echo "SERVING_NS=$SERVING_NS"

retry() {
  local max=$1; shift
  local interval=$1; shift

  until "$@"; do
    echo "trying.."
    max=$((max-1))
    if [[ "$max" -eq 0 ]]; then
      return 1
    fi
    sleep "$interval"
  done
}

# Run fvt tests and look for PASS
run_fvt() {
  local REV=1
  local RUN_STATUS="FAILED"

  echo " =====   run standard fvt   ====="
  kubectl get all -n "$SERVING_NS"
  export KUBECONFIG=~/.kube/config
    
  go test -v ./fvt -ginkgo.v -ginkgo.progress -test.timeout 40m > fvt.out
  cat fvt.out
  RUN_STATUS=$(cat fvt.out | awk '{ print $1}' | grep PASS)

  if [[ "$RUN_STATUS" == "PASS" ]]; then
    REV=0
  fi

  return "$REV"
}

retry 3 3 ibmcloud login --apikey "${IBM_CLOUD_API_KEY}" --no-region
retry 3 3 ibmcloud target -r "$REGION" -o "$ORG" -s "$SPACE" -g "$RESOURCE_GROUP"
retry 3 3 ibmcloud ks cluster config -c "$SERVING_KUBERNETES_CLUSTER_NAME"

echo "this is a test for ax  trigger"
RESULT=0
STATUS_MSG=PASSED

run_fvt || RESULT=$?

if [[ "$RESULT" -ne 0 ]]; then
  STATUS_MSG=FAILED
  echo "FVT test ${STATUS_MSG}"
  exit 1
fi

echo "FVT test ${STATUS_MSG}"
