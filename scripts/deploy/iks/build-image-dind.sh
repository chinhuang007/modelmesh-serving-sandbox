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
  make build
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
  echo "=======================Push image to Docker Hub==============================="
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
