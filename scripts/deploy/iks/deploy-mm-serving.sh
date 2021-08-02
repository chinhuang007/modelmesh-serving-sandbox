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

# Need the following env
# - SERVING_KUBERNETES_CLUSTER_NAME:   kube cluster name
# - SERVING_NS:                        namespace for modelmesh-serving, defulat: modelmesh-serving

MAX_RETRIES="${MAX_RETRIES:-5}"
SLEEP_TIME="${SLEEP_TIME:-10}"
EXIT_CODE=0

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


C_DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$C_DIR" ]]; then C_DIR="$PWD"; fi
source "${C_DIR}/helper-functions.sh"
echo "C_DIR=${C_DIR}"
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

retry 3 3 ibmcloud login --apikey "${IBM_CLOUD_API_KEY}" --no-region
retry 3 3 ibmcloud target -r "$REGION" -o "$ORG" -s "$SPACE" -g "$RESOURCE_GROUP"
retry 3 3 ibmcloud ks cluster config -c "$SERVING_KUBERNETES_CLUSTER_NAME"

kubectl create ns "$SERVING_NS"

wait_for_namespace "$SERVING_NS" "$MAX_RETRIES" "$SLEEP_TIME" || EXIT_CODE=$?

if [[ $EXIT_CODE -ne 0 ]]
then
  echo "Deploy unsuccessful. \"${SERVING_NS}\" not found."
  exit $EXIT_CODE
fi


# Update kustomize
curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"  | bash
mv kustomize /usr/local/bin/kustomize

# Check if all pods are running - allow 60 retries (10 minutes)
./scripts/install.sh --namespace "$SERVING_NS" --fvt
wait_for_pods "$SERVING_NS" 60 "$SLEEP_TIME" || EXIT_CODE=$?

if [[ $EXIT_CODE -ne 0 ]]
then
  echo "Deploy unsuccessful. Not all pods running."
  exit 1
fi

echo "Finished modelmesh-serving deployment."

#echo "=========================================================="
#echo "Copy and prepare artificates for subsequent stages"
#if [[ -z "$ARCHIVE_DIR" || "$ARCHIVE_DIR" == "." ]]; then
#  echo -e "Build archive directory contains entire working directory."
#else
#  echo -e "Copying working dir into build archive directory: ${ARCHIVE_DIR} "
#  mkdir -p "$ARCHIVE_DIR"
#  find . -mindepth 1 -maxdepth 1 -not -path "./$ARCHIVE_DIR" -exec cp -R '{}' "${ARCHIVE_DIR}/" ';'
#fi

#cp build.properties "${ARCHIVE_DIR}/" || :

#{
#  echo "SERVING_NS=${SERVING_NS}"
#  echo "MANIFEST=${MANIFEST}"
#} >> "${ARCHIVE_DIR}/build.properties"
