#!/bin/bash
#
# Copyright 2021 kubeflow.org
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Remove the x if you need no print out of each command
set -ex

# These env vars should come from the build.properties that `build-image.sh` generates
echo "REGISTRY_URL=${REGISTRY_URL}"
echo "REGISTRY_NAMESPACE=${REGISTRY_NAMESPACE}"
echo "BUILD_NUMBER=${BUILD_NUMBER}"
echo "ARCHIVE_DIR=${ARCHIVE_DIR}"
echo "GIT_BRANCH=${GIT_BRANCH}"
echo "GIT_COMMIT=${GIT_COMMIT}"
echo "GIT_COMMIT_SHORT=${GIT_COMMIT_SHORT}"
echo "REGION=${REGION}"
echo "ORG=${ORG}"
echo "SPACE=${SPACE}"
echo "RESOURCE_GROUP=${RESOURCE_GROUP}"

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

# fvt
run_fvt() {
  local REV=1
  local RUN_STATUS="FAILED"
  shift

  echo " =====   run standard fvt   ====="
  #kubectl config set-context --current --namespace=modelmesh-serving
  #kubectl create ns "$SERVING_NS"
  #kubectl get all

  # Update kustomize
  #curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"  | bash
  #mv kustomize /usr/local/bin/kustomize

  # Check if all pods are running - allow 60 retries (10 minutes)
  #./scripts/install.sh --namespace "$SERVING_NS" --fvt
  #wait_for_pods "$SERVING_NS" 60 "$SLEEP_TIME" || EXIT_CODE=$?

  kubectl get all -n "$SERVING_NS"
  kubectl get servingruntimes -n "$SERVING_NS"
  cat  ~/.kube/config
  export KUBECONFIG=~/.kube/config
    
  #go test -v ./fvt -ginkgo.v -ginkgo.progress -test.timeout 40m
  RUN_STATUS=$(go test -v ./fvt -ginkgo.v -ginkgo.progress -test.timeout 40m | awk '{ print $1}' | grep PASS)

  if [[ "$RUN_STATUS" == "PASS" ]]; then
    REV=0
    echo " =====   modelmesh-serving fvt PASSED ====="
  else
    echo " =====   modelmesh-serving fvt FAILED ====="
  fi

  return "$REV"
}

retry 3 3 ibmcloud login --apikey "${IBM_CLOUD_API_KEY}" --no-region
retry 3 3 ibmcloud target -r "$REGION" -o "$ORG" -s "$SPACE" -g "$RESOURCE_GROUP"
retry 3 3 ibmcloud ks cluster config -c "$SERVING_KUBERNETES_CLUSTER_NAME"

RESULT=0
STATUS_MSG=PASSED

run_fvt || RESULT=$?

if [[ "$RESULT" -ne 0 ]]; then
  STATUS_MSG=FAILED
fi

echo "FVT test ${STATUS_MSG}"