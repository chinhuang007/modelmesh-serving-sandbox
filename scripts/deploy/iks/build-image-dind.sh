#!/bin/bash

# Remove the x if you need no print out of each command
set -xe

# Environment variables needed by this script:
# - RUN_TASK:             execution task:
#                           - `build`: build the image
#                           - `build_push`: build and push the image
#

# The following envs could be loaded from `build.properties` that
# `run-setup.sh` generates.
# - REGION:               cloud region (us-south as default)
# - ORG:                  target organization (dev-advo as default)
# - SPACE:                target space (dev as default)
# - GIT_BRANCH:           git branch
# - GIT_COMMIT:           git commit hash
# - GIT_COMMIT_SHORT:     git commit hash short

REGION=${REGION:-"us-south"}
ORG=${ORG:-"dev-advo"}
SPACE=${SPACE:-"dev"}
RUN_TASK=${RUN_TASK:-"build"}

MAX_RETRIES="${MAX_RETRIES:-5}"
SLEEP_TIME="${SLEEP_TIME:-10}"
EXIT_CODE=0
DOCKER_TAG="$(git rev-parse --abbrev-ref HEAD)-$(date +"%Y%m%dT%H%M%S%Z")"

C_DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$C_DIR" ]]; then C_DIR="$PWD"; fi
source "${C_DIR}/helper-functions.sh"

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

######################################################################################
# Build image                                                                        #
######################################################################################
build_image() {
  echo "=======================Build modelmesh controller image======================="
  # Will build develop and then runtime images.

  echo "==============================Build dev image ================================"
  make build.develop
  docker images
  docker inspect "kserve/modelmesh-controller-develop:latest"
  echo "==========================Build runtime image ================================"
  echo $DOCKER_TAG
  #make build
  ./scripts/build_docker.sh --target runtime --tag $DOCKER_TAG
  docker images
  docker inspect "kserve/modelmesh-controller:latest"
}

######################################################################################
# Push image to Docker Hub                                                           #
######################################################################################
push_image() {
  echo "=======================Push image to Docker Hub==============================="
  # login dockerhub
  echo $DOCKERHUB_USERNAME
  echo $DOCKERHUB_NAMESPACE
  echo $PUBLISH_TAG
  set +x
  docker login -u "$DOCKERHUB_USERNAME" -p "$DOCKERHUB_TOKEN"
  set -x
  docker tag "kserve/modelmesh-controller:latest" "${DOCKERHUB_NAMESPACE}/modelmesh-controller:${PUBLISH_TAG}"
  docker push "${DOCKERHUB_NAMESPACE}/modelmesh-controller:${PUBLISH_TAG}"
}

test_image() {
  echo "=======================Test using the new image==============================="
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
  docker images

  sed -i 's/newTag:.*$/newTag: '"$DOCKER_TAG"'/' config/manager/kustomization.yaml
  cat config/manager/kustomization.yaml
  
  # Install modelmesh serving
  ./scripts/install.sh --namespace "$SERVING_NS" --fvt

  wait_for_pods "$SERVING_NS" 60 "$SLEEP_TIME" || EXIT_CODE=$?

  if [[ $EXIT_CODE -ne 0 ]]
  then
    echo "Deploy unsuccessful. Not all pods running."
    exit $EXIT_CODE
  fi

  export KUBECONFIG=~/.kube/config
  
  # Run fvt
  go test -v ./fvt -ginkgo.v -ginkgo.progress -test.timeout 40m > fvt.out
  cat fvt.out

  RUN_STATUS=$(cat fvt.out | awk '{ print $1}' | grep PASS)
  
  if [[ "$RUN_STATUS" != "PASS" ]]; then
    echo "FVT test failed"
    exit 1
  fi
}

case "$RUN_TASK" in
  "build")
    build_image
    ;;

  "build_push")
    build_image
    push_image
    ;;

  "build_test")
    build_image
    test_image
    ;;

  *)
    echo "please specify RUN_TASK=build|build_push|build_test"
    ;;
esac
