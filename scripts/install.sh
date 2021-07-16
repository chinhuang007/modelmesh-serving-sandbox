#!/usr/bin/env bash
# Copyright 2021 IBM Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.#

# Install Model-Mesh Serving CRDs, controller, and built-in runtimes into specified Kubernetes namespaces.
# Expect cluster-admin authority and Kube cluster access to be configured prior to running.

set -Eeuo pipefail

namespace=
delete=false
quickstart=false

function showHelp() {
  echo "usage: $0 [flags]"
  echo
  echo "Flags:"
  echo "  -n, --namespace                (required) Kubernetes namespace to deploy Model-Mesh Serving to."
  echo "  -d, --delete                   Delete any existing instances of Model-Mesh Serving in Kube namespace before running install, including CRDs, RBACs, controller, older CRD with ai.ibm.com api group name, etc."
  echo "  --quickstart                   Install and configure required supporting datastores in the same namespace (etcd and MinIO) - for experimentation/development"
  echo
  echo "Installs Model-Mesh Serving CRDs, controller, and built-in runtimes into specified"
  echo "Kubernetes namespaces."
  echo
  echo "Expects cluster-admin authority and Kube cluster access to be configured prior to running."
  echo "Also requires Etcd secret 'model-serving-etcd' to be created in namespace already."
}

die() {
  color_red='\e[31m'
  color_yellow='\e[33m'
  color_reset='\e[0m'
  printf "${color_red}FATAL:${color_yellow} $*${color_reset}\n" 1>&2
  exit 10
}

info() {
  color_blue='\e[34m'
  color_reset='\e[0m'
  printf "${color_blue}$*${color_reset}\n" 1>&2
}

success() {
  color_green='\e[32m'
  color_reset='\e[0m'
  printf "${color_green}$*${color_reset}\n" 1>&2
}

check_pod_status() {
  local -r JSONPATH="{range .items[*]}{'\n'}{@.metadata.name}:{@.status.phase}:{range @.status.conditions[*]}{@.type}={@.status};{end}{end}"
  local -r pod_selector="$1"
  local pod_status
  local pod_entry

  pod_status=$(kubectl get pods $pod_selector -o jsonpath="$JSONPATH") || kubectl_exit_code=$? # capture the exit code instead of failing

  if [[ $kubectl_exit_code -ne 0 ]]; then
    # kubectl command failed. print the error then wait and retry
    echo "Error running kubectl command."
    echo $pod_status
    return 1
  elif [[ ${#pod_status} -eq 0 ]]; then
    echo -n "No pods found with selector $pod_selector. Pods may not be up yet."
    return 1
  else
    # split string by newline into array
    IFS=$'\n' read -r -d '' -a pod_status_array <<<"$pod_status"

    for pod_entry in "${pod_status_array[@]}"; do
      local pod=$(echo $pod_entry | cut -d ':' -f1)
      local phase=$(echo $pod_entry | cut -d ':' -f2)
      local conditions=$(echo $pod_entry | cut -d ':' -f3)
      if [ "$phase" != "Running" ] && [ "$phase" != "Succeeded" ]; then
        return 1
      fi
      if [[ $conditions != *"Ready=True"* ]]; then
        return 1
      fi
    done
  fi
  return 0
}

wait_for_pods_ready() {
  local -r JSONPATH="{.items[*]}"
  local -r pod_selector="$1"
  local wait_counter=0
  local kubectl_exit_code=0
  local pod_status

  while true; do
    pod_status=$(kubectl get pods $pod_selector -o jsonpath="$JSONPATH") || kubectl_exit_code=$? # capture the exit code instead of failing

    if [[ $kubectl_exit_code -ne 0 ]]; then
      # kubectl command failed. print the error then wait and retry
      echo $pod_status
      echo -n "Error running kubectl command."
    elif [[ ${#pod_status} -eq 0 ]]; then
      echo -n "No pods found with selector '$pod_selector'. Pods may not be up yet."
    elif check_pod_status "$pod_selector"; then
      echo "All $pod_selector pods are running and ready."
      return
    else
      echo -n "Pods found with selector '$pod_selector' are not ready yet."
    fi

    if [[ $wait_counter -ge 60 ]]; then
      echo
      kubectl get pods $pod_selector
      die "Timed out after $((10 * wait_counter / 60)) minutes waiting for pod with selector: $pod_selector"
    fi

    wait_counter=$((wait_counter + 1))
    echo " Waiting 10 secs..."
    sleep "10s"
  done
}

while (($# > 0)); do
  case "$1" in
  -h | --h | --he | --hel | --help)
    showHelp
    exit 2
    ;;
  -n | --n | -namespace | --namespace)
    shift
    namespace="$1"
    ;;
  -d | --d | -delete | --delete)
    delete=true
    ;;
  --quickstart)
    quickstart=true
    ;;
  -*)
    die "Unknown option: '${1}'"
    ;;
  esac
  shift
done

#################      PREREQUISITES      #################
if [[ -z $namespace ]]; then
  showHelp
  die "Kubernetes namespace needs to be set."
fi

# /dev/null will hide output if it exists but show errors if it does not.
if ! type kustomize >/dev/null; then
  die "kustomize is not installed. Go to https://kubectl.docs.kubernetes.io/installation/kustomize/ to install it."
fi

if ! kubectl get namespaces $namespace >/dev/null; then
  die "Kube namespace does not exist: $namespace"
fi

info "Setting kube context to use namespace: $namespace"
kubectl config set-context --current --namespace="$namespace"

# Ensure the namespace is overridden for all the resources
cd config/default
kustomize edit set namespace "$namespace"
cd ..

# Clean up previous instances but do not fail if they do not exist
if [[ $delete == "true" ]]; then
  info "Deleting any previous Model-Mesh Serving instances and older CRD with ai.ibm.com api group name"
  kubectl delete crd/predictors.ai.ibm.com --ignore-not-found=true
  kubectl delete crd/servingruntimes.ai.ibm.com --ignore-not-found=true
  kustomize build default | kubectl delete -f - --ignore-not-found=true
  kubectl delete -f dependencies/quickstart.yaml --ignore-not-found=true
fi

# Quickstart resources
if [[ $quickstart == "true" ]]; then
  info "Deploying quickstart resources for etcd and minio"
  kubectl apply -f dependencies/quickstart.yaml

  info "Waiting for dependent pods to be up..."
  wait_for_pods_ready "--field-selector metadata.name=etcd"
  wait_for_pods_ready "--field-selector metadata.name=minio"
fi

if ! kubectl get secret model-serving-etcd >/dev/null; then
  die "Could not find Etcd kube secret 'model-serving-etcd'. This is a prerequisite for running Model-Mesh Serving install."
else
  echo "model-serving-etcd secret found"
fi

info "Creating storage-config secret if it does not exist"
kubectl create -f default/storage-secret.yaml 2>/dev/null || :
kubectl get secret storage-config

info "Installing Model-Mesh Serving CRDs, RBACs, and controller"
kustomize build default | kubectl apply -f -

info "Waiting for Model-Mesh Serving controller pod to be up..."
wait_for_pods_ready "-l control-plane=modelmesh-controller"

info "Installing Model-Mesh Serving built-in runtimes"
kustomize build runtimes --load-restrictor LoadRestrictionsNone | kubectl apply -f -

success "Successfully installed Model-Mesh Serving!"
