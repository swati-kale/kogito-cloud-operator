// Copyright 2019 Red Hat, Inc. and/or its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package steps

import (
	"github.com/cucumber/godog"
	"github.com/kiegroup/kogito-cloud-operator/test/smoke/framework"
)

// RegisterCliSteps register all CLI steps existing
func registerKogitoJobsServiceSteps(s *godog.Suite, data *Data) {
	s.Step(`^Deploy Kogito Jobs Service with (\d+) replicas$`, data.deployKogitoJobsServiceWithReplicas)
	s.Step(`^Deploy Kogito Jobs Service with persistence and (\d+) replicas$`, data.deployKogitoJobsServiceWithPersistenceAndReplicas)
	s.Step(`^Kogito Jobs Service has (\d+) pods running within (\d+) minutes$`, data.kogitoJobsServiceHasPodsRunningWithinMinutes)
	s.Step(`^Scale Kogito Jobs Service to (\d+) pods within (\d+) minutes$`, data.scaleKogitoJobsServiceToPodsWithinMinutes)
}

func (data *Data) deployKogitoJobsServiceWithReplicas(replicas int) error {
	return framework.DeployKogitoJobsService(data.Namespace, replicas, false)
}

func (data *Data) deployKogitoJobsServiceWithPersistenceAndReplicas(replicas int) error {
	return framework.DeployKogitoJobsService(data.Namespace, replicas, true)
}

func (data *Data) kogitoJobsServiceHasPodsRunningWithinMinutes(pods, timeoutInMin int) error {
	return framework.WaitForKogitoJobsService(data.Namespace, pods, timeoutInMin)
}

func (data *Data) scaleKogitoJobsServiceToPodsWithinMinutes(nbPods, timeoutInMin int) error {
	err := framework.SetKogitoJobsServiceReplicas(data.Namespace, nbPods)
	if err != nil {
		return err
	}
	return framework.WaitForKogitoJobsService(data.Namespace, nbPods, timeoutInMin)
}
