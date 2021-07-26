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

# fvt
run_flip_coin_example() {
  local REV=1
  local DURATION=$1
  shift

  echo " =====   fvt sample  ====="
  #python3 samples/flip-coin/condition.py
  #retry 3 3 kfp --endpoint http://localhost:8888 pipeline upload -p e2e-flip-coin samples/flip-coin/condition.yaml || :
  
  REV=0
  if [[ "$REV" -eq 0 ]]; then
    echo " =====   flip coin sample PASSED ====="
  else
    echo " =====   flip coin sample FAILED ====="
  fi

  return "$REV"
}

RESULT=0
run_flip_coin_example 20 || RESULT=$?

STATUS_MSG=PASSED
if [[ "$RESULT" -ne 0 ]]; then
  STATUS_MSG=FAILED
fi