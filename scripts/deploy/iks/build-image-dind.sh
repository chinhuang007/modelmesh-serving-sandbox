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
  echo "=========================================================="
  echo "Copy and prepare artificates for subsequent stages"

  echo "Checking archive dir presence"
  if [[ -z "$ARCHIVE_DIR" || "$ARCHIVE_DIR" == "." ]]; then
    echo -e "Build archive directory contains entire working directory."
  else
    echo -e "Copying working dir into build archive directory: ${ARCHIVE_DIR} "
    mkdir -p "$ARCHIVE_DIR"
    find . -mindepth 1 -maxdepth 1 -not -path "./${ARCHIVE_DIR}" -exec cp -R '{}' "${ARCHIVE_DIR}/" ';'
  fi

  # Persist env variables into a properties file (build.properties) so that all pipeline stages consuming this
  # build as input and configured with an environment properties file valued 'build.properties'
  # will be able to reuse the env variables in their job shell scripts.

  # If already defined build.properties from prior build job, append to it.
  cp build.properties "${ARCHIVE_DIR}/" || :

  echo "=======================Build dev image ================================"
  ls -lrt
  make build.develop
  docker images
  echo "=======================Build runtime image ================================"
  make build
  docker images
  docker inspect "kserve/modelmesh-controller:latest"
}

push_image() {
  echo "=======================Push image ================================"  
}
case "$RUN_TASK" in
  "build")
    build_image
    ;;

  "build_push")
    build_image
    push_image
    ;;

  *)
    echo "please specify RUN_TASK=build|build_push"
    ;;
esac
