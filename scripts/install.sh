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
# Optional env vars (can also be passed as flags): ARTIFACTORY_USER, ARTIFACTORY_APIKEY

set -Eeuo pipefail

base_artifactory_path="https://na.artifactory.swg-devops.com/artifactory"
artifactory_model_serve_path="wcp-ai-foundation-team-generic-virtual/model-serving"
artifactory_user="${ARTIFACTORY_USER:-}"
artifactory_apikey="${ARTIFACTORY_APIKEY:-}"
redsonja_apikey="${REDSONJA_APIKEY:-}"
model_serve_version=REPLACE
install_local_path=
namespace=
delete=false
quickstart=false
redsonja_images=false

function showHelp() {
  echo "usage: $0 [flags]"
  echo
  echo "Flags:"
  echo "  -n, --namespace                (required) Kubernetes namespace to deploy Model-Mesh Serving to."
  echo "  -u, --artifactory-user         Artifactory username to pull Model-Mesh Serving tarfile and images, can also set with env var ARTIFACTORY_USER."
  echo "  -a, --artifactory-apikey       Artifactory API key to pull Model-Mesh Serving tarfile and images, can also set with env var ARTIFACTORY_APIKEY."
  echo "  -v, --model-serve-version      Model-Mesh Serving version to pull and use. Example: wml-serving-0.3.0_165"
  echo "  -p, --install-config-path      Path to local model serve installation configs. Can be Model-Mesh Serving tarfile or directory."
  echo "  -d, --delete                   Delete any existing instances of Model-Mesh Serving in Kube namespace before running install, including CRDs, RBACs, controller, older CRD with ai.ibm.com api group name, etc."
  echo "  --quickstart                   Install and configure required supporting datastores in the same namespace (etcd and MinIO) - for experimentation/development"
  echo "  --redsonja-images              Use images pulled from redsonja IBM Container registry, requires redsonja user and apikey."
  echo "  --redsonja-apikey              IBM container registry apikey that has access to redsonja account, can also set with env var REDSONJA_APIKEY."
  echo
  echo "Installs Model-Mesh Serving CRDs, controller, and built-in runtimes into specified"
  echo "Kubernetes namespaces. If a --model-serve-version is given, will try to pull that"
  echo "version from Artifactory. If local --install-config-path will try to install configs"
  echo "at that given path. If neither are given, will pull default latest version."
  echo
  echo "Expects cluster-admin authority and Kube cluster access to be configured prior to running."
  echo "Also requires Etcd secret 'model-serving-etcd' to be created in namespace already."
  echo
  echo "Requires either an Artifactory user and api key or Redsonja user and api key in"
  echo "a Kube secret named 'ibm-entitlement-key' to pull the necessary images."
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

replace_artifactory_registry() {
  local -r filename="$1"
  local -r artifactory_registry="wcp-ai-foundation-team-docker-virtual.artifactory.swg-devops.com"
  local -r redsonja_registry="us.icr.io\/redsonja_hyboria\/ai-foundation"

  if ! [[ -f $filename ]]; then
    die "File does not exist: $filename, failed to replace artifactory registry with redsonja registry"
  fi

  perl -pi -e "s/${artifactory_registry}/${redsonja_registry}/g" $filename
}

while (($# > 0)); do
  case "$1" in
  -h | --h | --he | --hel | --help)
    showHelp
    exit 2
    ;;
  -a | --a | -apikey | --apikey | -artifactory-apikey | --artifactory-apikey)
    shift
    artifactory_apikey="$1"
    ;;
  -u | --u | -user | --user | -artifactory-user | --artifactory-user)
    shift
    artifactory_user="$1"
    ;;
  -redsonja-apikey | --redsonja-apikey)
    shift
    redsonja_apikey="$1"
    ;;
  -redsonja-images | --redsonja-images)
    redsonja_images=true
    ;;
  -n | --n | -namespace | --namespace)
    shift
    namespace="$1"
    ;;
  -v | --v | -version | --version | -model-serve-version | --model-serve-version)
    shift
    model_serve_version="$1"
    ;;
  -p | --p | -install-path | --install-path | -install-config-path | --install-config-path)
    shift
    install_local_path="$1"
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

#################      INSTALL Model-Mesh Serving      #################
# Pull model serving configs if local path not given
info "Getting Model-Mesh Serving configs"
if [[ -n $install_local_path ]]; then
  if [[ -f $install_local_path ]] && [[ $install_local_path =~ \.t?gz$ ]]; then
    tar -xf "$install_local_path"
    cd "$(basename "$(basename $install_local_path .tgz)" .tar.gz)"
  elif [[ -d $install_local_path ]]; then
    cd "$install_local_path"
  else
    die "Could not find provided path to Model-Mesh Serving install configs: $install_local_path"
  fi
else
  # Example version: wml-serving-0.3.0_165 forms URL https://na.artifactory.swg-devops.com/artifactory/wcp-ai-foundation-team-generic-virtual/model-serving/wml-serving-0.3.0_165.tgz
  echo "Pulling Model-Mesh Serving: ${base_artifactory_path}/${artifactory_model_serve_path}/${model_serve_version}.tgz"
  if [[ -z $artifactory_user ]] || [[ -z $artifactory_apikey ]]; then
    die "To pull Model-Mesh Serving tarfile, need to set artifactory user and api key."
  fi

  curl -sSLf -u "${artifactory_user}:${artifactory_apikey}" -o "${model_serve_version}.tgz" "${base_artifactory_path}/${artifactory_model_serve_path}/${model_serve_version}.tgz"
  tar -xf "${model_serve_version}.tgz"
  rm "${model_serve_version}.tgz"
  cd ${model_serve_version}
  info "Successfully pulled Model-Mesh Serving configs version: ${model_serve_version}"
fi

# Ensure the namespace is overridden for all the resources
cd default
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

# Clean up controller deployments and serviceaccounts with old names, regardless of --delete option
kubectl delete deployment wmlserving-operator model-serving-controller --ignore-not-found=true
kubectl delete serviceaccount wmlserving-operator model-serving-controller --ignore-not-found=true

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

if ! kubectl get secret ibm-entitlement-key >/dev/null 2>&1; then
  # NOTE: users will have to make sure that the ibm-entitlement-key has the correct secret for pulling image
  if [[ $redsonja_images == "true" ]]; then
    if [[ -z $redsonja_apikey ]]; then
      die "Must set redsonja user and api key to create ibm-entitlement-key."
    fi

    kubectl create secret docker-registry ibm-entitlement-key \
      --docker-server=us.icr.io \
      --docker-username="iamapikey" \
      --docker-password="$redsonja_apikey"
  else
    if [[ -z $artifactory_user ]] || [[ -z $artifactory_apikey ]]; then
      die "Must set Artifactory user and api key to create ibm-entitlement-key."
    fi

    kubectl create secret docker-registry ibm-entitlement-key \
      --docker-server=wcp-ai-foundation-team-docker-virtual.artifactory.swg-devops.com \
      --docker-username="$artifactory_user" \
      --docker-password="$artifactory_apikey"
  fi
else
  info "ibm-entitlement-key secret found"
fi

if [[ $redsonja_images == "true" ]]; then
  info "Replacing all aritfactory image registries with redsonja"
  replace_artifactory_registry "manager/kustomization.yaml"
  replace_artifactory_registry "runtimes/kustomization.yaml"
  replace_artifactory_registry "default/config-defaults.yaml"
fi

info "Creating storage-config secret if it does not exist"
kubectl create -f default/storage-secret.yaml 2>/dev/null || :
kubectl get secret storage-config

info "Installing Model-Mesh Serving CRDs, RBACs, and controller"
kustomize build default | kubectl apply -f -

info "Waiting for Model-Mesh Serving controller pod to be up..."
wait_for_pods_ready "-l control-plane=wmlserving-controller"

info "Installing Model-Mesh Serving built-in runtimes"
kustomize build runtimes --load-restrictor LoadRestrictionsNone | kubectl apply -f -

success "Successfully installed Model-Mesh Serving!"
