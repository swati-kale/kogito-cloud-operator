// Copyright 2020 Red Hat, Inc. and/or its affiliates
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

package framework

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"

	"io/ioutil"
)

const (
	logFolder = "logs"
	logSuffix = ".log"
)

var (
	monitoredNamespaces = make(map[string]monitoredNamespace)
)

// StartPodLogCollector monitors a namespace and stores logs of all pods running in the namespace
func StartPodLogCollector(namespace string) error {
	if isNamespaceMonitored(namespace) {
		return errors.New("namespace is already monitored")
	}

	if err := createLogFolder(namespace); err != nil {
		return fmt.Errorf("Error while creating log folder: %v", err)
	}

	monitoredNamespace := monitoredNamespace{
		pods:           make(map[string]monitoredPod),
		stopMonitoring: make(chan bool),
	}
	monitoredNamespaces[namespace] = monitoredNamespace

	scanningPeriod := time.NewTicker(5 * time.Second)
	defer scanningPeriod.Stop()
	for {
		select {
		case <-monitoredNamespace.stopMonitoring:
			return nil
		case <-scanningPeriod.C:
			if pods, err := GetPods(namespace); err != nil {
				GetLogger(namespace).Errorf("Error while getting pods in namespace '%s': %v", namespace, err)
			} else {
				for _, pod := range pods.Items {
					if !isPodMonitored(namespace, pod.Name) && IsPodRunning(&pod) {
						initMonitoredPod(namespace, pod.Name)
						for _, container := range pod.Spec.Containers {
							initMonitoredContainer(namespace, pod.Name, container.Name)
							go storeContainerLogWithFollow(namespace, pod.Name, container.Name)
						}
					}
				}
			}
		}
	}
}

func isNamespaceMonitored(namespace string) bool {
	_, exists := monitoredNamespaces[namespace]
	return exists
}

func createLogFolder(namespace string) error {
	return os.MkdirAll(logFolder+"/"+namespace, os.ModePerm)
}

func isPodMonitored(namespace, podName string) bool {
	_, exists := monitoredNamespaces[namespace].pods[podName]
	return exists
}

func initMonitoredPod(namespace, podName string) {
	monitoredPod := monitoredPod{
		containers: make(map[string]monitoredContainer),
	}
	monitoredNamespaces[namespace].pods[podName] = monitoredPod
}

func initMonitoredContainer(namespace, podName, containerName string) {
	monitoredContainer := monitoredContainer{loggingFinished: false}
	monitoredNamespaces[namespace].pods[podName].containers[containerName] = monitoredContainer
}

func storeContainerLogWithFollow(namespace, podName, containerName string) {
	log, err := getContainerLogWithFollow(namespace, podName, containerName)
	if err != nil {
		GetLogger(namespace).Errorf("Error while retrieving log of pod '%s': %v", podName, err)
	}

	if isContainerLoggingFinished(namespace, podName, containerName) {
		GetLogger(namespace).Debugf("Logging of container '%s' of pod '%s' already finished, retrieved log will be ignored.", containerName, podName)
	} else {
		markContainerLoggingAsFinished(namespace, podName, containerName)
		if err := writeLogIntoTheFile(namespace, podName, containerName, log); err != nil {
			GetLogger(namespace).Errorf("Error while writing log into the file: %v", err)
		}
	}
}

// Log is returned once container is terminated
func getContainerLogWithFollow(namespace, podName, containerName string) (string, error) {
	return kubernetes.PodC(kubeClient).GetLogsWithFollow(namespace, podName, containerName)
}

func isContainerLoggingFinished(namespace, podName, containerName string) bool {
	monitoredContainer := monitoredNamespaces[namespace].pods[podName].containers[containerName]
	return monitoredContainer.loggingFinished
}

func markContainerLoggingAsFinished(namespace, podName, containerName string) {
	monitoredContainer := monitoredNamespaces[namespace].pods[podName].containers[containerName]
	monitoredContainer.loggingFinished = true
}

func writeLogIntoTheFile(namespace, podName, containerName, log string) error {
	return ioutil.WriteFile(logFolder+"/"+namespace+"/"+podName+"-"+containerName+logSuffix, []byte(log), 0644)
}

// StopPodLogCollector waits until all logs are stored on disc
func StopPodLogCollector(namespace string) error {
	if !isNamespaceMonitored(namespace) {
		return errors.New("namespace is not monitored")
	}
	stopNamespaceMonitoring(namespace)
	storeUnfinishedContainersLog(namespace)
	return nil
}

func stopNamespaceMonitoring(namespace string) {
	monitoredNamespaces[namespace].stopMonitoring <- true
	close(monitoredNamespaces[namespace].stopMonitoring)
}

// Write log of all containers of pods in namespace which didn't store their log yet
func storeUnfinishedContainersLog(namespace string) {
	for podName, pod := range monitoredNamespaces[namespace].pods {
		for containerName, container := range pod.containers {
			if !container.loggingFinished {
				storeContainerLog(namespace, podName, containerName)
			}
		}
	}
}

// Write container log into filesystem
func storeContainerLog(namespace string, podName, containerName string) {
	log, err := getContainerLog(namespace, podName, containerName)
	if err != nil {
		GetLogger(namespace).Errorf("Error while retrieving log of container '%s' in pod '%s': %v", containerName, podName, err)
	}

	if err := writeLogIntoTheFile(namespace, podName, containerName, log); err != nil {
		GetLogger(namespace).Errorf("Error while writing log into the file: %v", err)
	}
}

func getContainerLog(namespace, podName, containerName string) (string, error) {
	return kubernetes.PodC(kubeClient).GetLogs(namespace, podName, containerName)
}

type monitoredNamespace struct {
	pods           map[string]monitoredPod
	stopMonitoring chan bool
}

type monitoredPod struct {
	containers map[string]monitoredContainer
}

type monitoredContainer struct {
	loggingFinished bool
}
