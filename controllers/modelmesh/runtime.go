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
package modelmesh

import (
	"fmt"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	api "wmlserving.ai.ibm.com/controller/api/v1"
)

const (
	modelsDirVolume string  = "models-dir"
	socketVolume    string  = "domain-socket"
	ModelsDir       string  = "/models"
	ModelDirScale   float64 = 1.5
)

//Sets the model mesh grace period to match the deployment grace period
func (m *Deployment) syncGracePeriod(deployment *appsv1.Deployment) error {
	if deployment.Spec.Template.Spec.TerminationGracePeriodSeconds != nil {
		gracePeriodS := deployment.Spec.Template.Spec.TerminationGracePeriodSeconds
		gracePeriodMs := *gracePeriodS * int64(1000)
		gracePeriodMsStr := strconv.FormatInt(gracePeriodMs, 10)
		err := setEnvironmentVar("mm", "SHUTDOWN_TIMEOUT_MS", gracePeriodMsStr, deployment)
		return err
	}

	return nil
}

func (m *Deployment) addVolumesToDeployment(deployment *appsv1.Deployment) error {
	rt := m.Owner
	modelsDirSize := calculateModelDirSize(rt)

	volumes := []corev1.Volume{
		{
			Name: modelsDirVolume,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{Medium: "", SizeLimit: modelsDirSize},
			},
		},
	}

	if hasUnixSockets, _, _ := unixDomainSockets(rt); hasUnixSockets {
		volumes = append(volumes, corev1.Volume{
			Name: socketVolume,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	}

	deployment.Spec.Template.Spec.Volumes = volumes

	return nil
}

func (m *Deployment) addStorageConfigVolume(deployment *appsv1.Deployment) error {
	rt := m.Owner
	// need to mount storage volume for built-in adapters and the scenarios where StorageHelper is not disabled/specified.
	if rt.Spec.BuiltInAdapter != nil || useStorageHelper(rt) {
		storageVolume := corev1.Volume{
			Name: ConfigStorageMount,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: StorageSecretName,
				},
			},
		}

		deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, storageVolume)
	}

	return nil
}

// calculate emptyDir Size
func calculateModelDirSize(rt *api.ServingRuntime) *resource.Quantity {

	memorySize := resource.MustParse("0")

	for _, cspec := range rt.Spec.Containers {
		memorySize.Add(cspec.Resources.Limits[corev1.ResourceMemory])
	}

	return resource.NewQuantity(int64(float64(memorySize.Value())*ModelDirScale), resource.BinarySI)
}

//Adds the provided runtime to the deployment
func (m *Deployment) addRuntimeToDeployment(deployment *appsv1.Deployment) error {
	rt := m.Owner

	// first prepare the common variables needed for both adapter and other containers
	lifecycle := &corev1.Lifecycle{
		PreStop: &corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/prestop",
				Port: intstr.FromInt(8090),
			},
		},
	}
	securityContext := &corev1.SecurityContext{
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{"ALL"},
		},
	}

	volumeMounts := []corev1.VolumeMount{
		{
			Name:      modelsDirVolume,
			MountPath: ModelsDir,
		},
	}

	// Now add the containers specified in serving runtime spec
	for _, cspec := range rt.Spec.Containers {
		coreEnv := make([]corev1.EnvVar, len(cspec.Env))
		for i, cenv := range cspec.Env {
			coreEnv[i].Name = cenv.Name
			coreEnv[i].Value = cenv.Value
			coreEnv[i].ValueFrom = cenv.ValueFrom
		}

		//translate our container spec to corev1 container spec
		corecspec := corev1.Container{
			Args:            cspec.Args,
			Command:         cspec.Command,
			Env:             coreEnv,
			Image:           cspec.Image,
			Name:            cspec.Name,
			Resources:       cspec.Resources,
			ImagePullPolicy: cspec.ImagePullPolicy,
			WorkingDir:      cspec.WorkingDir,
			LivenessProbe:   cspec.LivenessProbe,
			VolumeMounts:    volumeMounts,
			Lifecycle:       lifecycle,
			SecurityContext: securityContext,
		}

		err := addDomainSocketMount(rt, &corecspec)
		if err != nil {
			return err
		}

		if found, i, _ := findContainer(cspec.Name, deployment); found {
			deployment.Spec.Template.Spec.Containers[i] = corecspec
		} else {
			deployment.Spec.Template.Spec.Containers = append(deployment.Spec.Template.Spec.Containers, corecspec)
		}
	}

	if rt.Spec.BuiltInAdapter != nil {
		// BuiltInAdapter is specified, so prepare adapter container
		// Validation is already happened in reconcile logic, so just append "-adapter" to runtimeName for adapterName
		runtimeName := string(rt.Spec.BuiltInAdapter.ServerType)
		runtimeAdapterName := runtimeName + "-adapter"

		builtInAdapterContainer := corev1.Container{
			Command:         []string{"/opt/app/" + runtimeAdapterName},
			Image:           m.PullerImage,
			Name:            runtimeAdapterName,
			Lifecycle:       lifecycle,
			SecurityContext: securityContext,
		}

		// the puller and adapter containers are the same image and are given the
		// same resources
		builtInAdapterContainer.Resources = *m.PullerResources

		builtInAdapterContainer.VolumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      ConfigStorageMount,
			MountPath: PullerConfigPath,
			ReadOnly:  true,
		})

		builtInAdapterContainer.Env = []corev1.EnvVar{
			{
				Name:  "ADAPTER_PORT",
				Value: "8085",
			},
			{
				Name:  "RUNTIME_PORT",
				Value: strconv.Itoa(rt.Spec.BuiltInAdapter.RuntimeManagementPort),
			},
			{
				Name: "CONTAINER_MEM_REQ_BYTES",
				ValueFrom: &corev1.EnvVarSource{
					ResourceFieldRef: &corev1.ResourceFieldSelector{
						ContainerName: runtimeName,
						Resource:      "requests.memory",
					},
				},
			},
			{
				Name:  "MEM_BUFFER_BYTES",
				Value: strconv.Itoa(rt.Spec.BuiltInAdapter.MemBufferBytes),
			},
			{
				Name:  "LOADTIME_TIMEOUT",
				Value: strconv.Itoa(rt.Spec.BuiltInAdapter.ModelLoadingTimeoutMillis),
			},
			{
				Name:  "USE_EMBEDDED_PULLER",
				Value: "true",
			},
		}

		deployment.Spec.Template.Spec.Containers = append(deployment.Spec.Template.Spec.Containers, builtInAdapterContainer)
	}

	return nil
}

func addDomainSocketMount(rt *api.ServingRuntime, c *corev1.Container) error {
	var requiresDomainSocketMounting bool
	var domainSocketMountPoint string
	endpoints := []*string{
		rt.Spec.GrpcDataEndpoint,
		//		rt.Spec.HTTPDataEndpoint,
		rt.Spec.GrpcMultiModelManagementEndpoint,
	}
	for _, endpointStr := range endpoints {
		if endpointStr != nil {
			e, err := ParseEndpoint(*endpointStr)
			if err != nil {
				return err
			}
			if udsE, ok := e.(UnixEndpoint); ok {
				requiresDomainSocketMounting = true
				_mountPoint := udsE.ParentPath
				if domainSocketMountPoint != "" && domainSocketMountPoint != _mountPoint {
					return fmt.Errorf("Only one unix domain socket path is allowed. Found %v", endpoints)
				}

				domainSocketMountPoint = _mountPoint
			}
		}
	}
	if requiresDomainSocketMounting {
		c.VolumeMounts = append(c.VolumeMounts, corev1.VolumeMount{
			Name:      socketVolume,
			MountPath: domainSocketMountPoint,
		})
	}

	return nil
}

func (m *Deployment) addPassThroughPodFieldsToDeployment(deployment *appsv1.Deployment) error {
	rt := m.Owner
	// these fields map directly to pod spec fields
	deployment.Spec.Template.Spec.NodeSelector = rt.Spec.NodeSelector
	deployment.Spec.Template.Spec.Tolerations = rt.Spec.Tolerations
	archNodeSelector := corev1.NodeSelectorTerm{
		MatchExpressions: []corev1.NodeSelectorRequirement{
			{
				Key:      "kubernetes.io/arch",
				Operator: corev1.NodeSelectorOpIn,
				Values:   []string{"amd64"},
			},
		},
	}
	deployment.Spec.Template.Spec.Affinity = rt.Spec.Affinity
	if rt.Spec.Affinity == nil {
		deployment.Spec.Template.Spec.Affinity = &corev1.Affinity{
			NodeAffinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{
						archNodeSelector,
					},
				},
			},
		}
	} else if rt.Spec.Affinity.NodeAffinity == nil {
		deployment.Spec.Template.Spec.Affinity.NodeAffinity = &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					archNodeSelector,
				},
			},
		}
	} else if rt.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		deployment.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{
				archNodeSelector,
			},
		}
	} else {
		nodeSelectors := rt.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
		deployment.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = append(nodeSelectors, archNodeSelector)
	}

	return nil
}

func (m *Deployment) configureRuntimeAnnotations(deployment *appsv1.Deployment) error {

	// default annotations
	// annotations data has single string with key=value values line separated
	annotationsMap := map[string]string{
		"productID":     "7320f6c142574f48a46f2a8e82736ded",
		"productMetric": "FREE",
		"productName":   "IBM Watson Machine Learning Core",
	}

	if deployment.ObjectMeta.Annotations == nil {
		deployment.ObjectMeta.Annotations = make(map[string]string)
	}

	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = make(map[string]string)
	}

	for key, value := range annotationsMap {
		// set annotations for deployment
		deployment.ObjectMeta.Annotations[key] = value
		// set annotations for pods created by deployment
		deployment.Spec.Template.Annotations[key] = value
	}

	// annotations from configmap
	if m.AnnotationConfigMap == nil {
		return nil
	}

	m.Log.Info("Found product ConfigMap for ServingRuntime, setting annotations", "configMapName", m.AnnotationConfigMap.Name)

	cfg, ok := m.AnnotationConfigMap.Data["annotations"]
	if !ok {
		return fmt.Errorf("ConfigMap must contain a key named annotations")
	}

	properties := strings.Split(cfg, "\n")
	for _, propStr := range properties {
		prop := strings.Split(propStr, "=")
		if len(prop) == 2 {
			trimmedKey := strings.TrimSpace(strings.Trim(prop[0], "'\""))
			trimmedVal := strings.TrimSpace(strings.Trim(prop[1], "'\""))
			annotationsMap[trimmedKey] = trimmedVal
		}
	}
	// set default value
	if _, exists := annotationsMap["conversionRatio"]; !exists {
		annotationsMap["conversionRatio"] = "1:1"
	}

	// assign the values retrieved from configmap to runtime deployment
	for key, value := range annotationsMap {
		// set annotations for deployment
		deployment.ObjectMeta.Annotations[key] = value
		// set annotations for pods created by deployment
		deployment.Spec.Template.Annotations[key] = value
	}

	return nil
}
