// Copyright 2021 IBM Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package controllers

import (
	"testing"

	servingv1 "wmlserving.ai.ibm.com/controller/api/v1"
	mmeshapi "wmlserving.ai.ibm.com/controller/generated/mmesh"

	"github.com/stretchr/testify/assert"
)

func Test_DecodeModelState(t *testing.T) {

	testData := map[*mmeshapi.ModelStatusInfo][]interface{}{
		{
			Status: mmeshapi.ModelStatusInfo_LOADING_FAILED,
			Errors: []string{"There are no running instances that meet the label requirements of type mt:SomeType: [mt:SomeType]"},
		}: {
			servingv1.Loading, servingv1.RuntimeUnhealthy, "Waiting for supporting runtime Pod to become available",
		},
		{
			Status: mmeshapi.ModelStatusInfo_LOADING_FAILED,
			Errors: []string{"There are no running instances that meet the label requirements of type rt:SomeRuntime: [rt:SomeRuntime]"},
		}: {
			servingv1.Loading, servingv1.RuntimeUnhealthy, "Waiting for supporting runtime Pod to become available",
		},
		{
			Status: mmeshapi.ModelStatusInfo_LOADING_FAILED,
			Errors: []string{"There are no running instances that meet the label requirements of type rt:SomeRuntime: [_no_runtime]"},
		}: {
			servingv1.FailedToLoad, servingv1.RuntimeNotRecognized, "Specified runtime name not recognized",
		},
		{
			Status: mmeshapi.ModelStatusInfo_LOADING_FAILED,
			Errors: []string{"There are no running instances that meet the label requirements of type mt:SomeType: [_no_runtime]"},
		}: {
			servingv1.FailedToLoad, servingv1.NoSupportingRuntime, "No ServingRuntime supports specified model type",
		},
		{
			Status: mmeshapi.ModelStatusInfo_LOADING_FAILED,
			Errors: []string{"Random loading failure message", "Some other error message"},
		}: {
			servingv1.FailedToLoad, servingv1.ModelLoadFailed, "Random loading failure message",
		},
		{
			Status: mmeshapi.ModelStatusInfo_LOADED,
		}: {
			servingv1.Loaded, servingv1.FailureReason(""), "",
		},
	}

	for input, expected := range testData {
		st, reason, msg := decodeModelState(input)
		assert.Equal(t, expected, []interface{}{st, reason, msg})
	}
}
