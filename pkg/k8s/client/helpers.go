/*
(c) Copyright IBM Corp. 2024
(c) Copyright Instana Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"github.com/instana/instana-agent-operator/pkg/result"
	k8sClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func wasRetrieved(_ k8sClient.Object) result.Result[bool] {
	return result.OfSuccess(true)
}

func ifNotFound(err error) (bool, error) {
	return false, k8sClient.IgnoreNotFound(err)
}

func doNotExist(res result.Result[bool]) bool {
	return res.IsSuccess() && !res.ToOptional().Get()
}