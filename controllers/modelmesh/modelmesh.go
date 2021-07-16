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
	"context"
	"fmt"
	"strconv"

	"github.com/go-logr/logr"
	mf "github.com/manifestival/manifestival"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
	api "wmlserving.ai.ibm.com/controller/api/v1"
	"wmlserving.ai.ibm.com/controller/controllers/config"
)

const logYaml = false

const ModelMeshEtcdPrefix = "mm"
const ModelMeshVModelOwner = "wcp" // "watson core predictor"

//Models a deployment
type Deployment struct {
	ServiceName        string
	Name               string
	Namespace          string
	Owner              *api.ServingRuntime
	Log                logr.Logger
	Metrics            bool
	PrometheusPort     uint16
	PrometheusScheme   string
	ModelMeshImage     string
	ModelMeshResources *corev1.ResourceRequirements
	// internal fields used when templating
	ModelMeshLimitCPU       string
	ModelMeshRequestsCPU    string
	ModelMeshLimitMemory    string
	ModelMeshRequestsMemory string
	// end internal fields
	PullerImage         string
	PullerImageCommand  []string
	PullerResources     *corev1.ResourceRequirements
	Replicas            uint16
	Port                uint16
	TLSSecretName       string
	TLSClientAuth       string
	EtcdSecretName      string
	ServiceAccountName  string
	AnnotationConfigMap *corev1.ConfigMap
	EnableAccessLogging bool
	Client              client.Client
}

func (m *Deployment) Apply(ctx context.Context) error {
	clientParam := m.Client

	m.Log.Info("Applying model mesh deployment", "pods", m.Replicas)

	// set internal fields before rendering from the template
	m.ModelMeshLimitCPU = m.ModelMeshResources.Limits.Cpu().String()
	m.ModelMeshLimitMemory = m.ModelMeshResources.Limits.Memory().String()
	m.ModelMeshRequestsCPU = m.ModelMeshResources.Requests.Cpu().String()
	m.ModelMeshRequestsMemory = m.ModelMeshResources.Requests.Memory().String()

	manifest, err := config.Manifest(clientParam, "config/internal/base/deployment.yaml.tmpl", m)
	if err != nil {
		return fmt.Errorf("Error loading model mesh deployment yaml: %w", err)
	}

	if len(manifest.Resources()) != 1 {
		// manifestival.ManifestFrom will hide yaml parsing errors and not include those resources. This check ensures we parsed the proper number of resources.
		return fmt.Errorf("Unexpected number of resources (%d) found in the deployment template. This is likely due to bad or missing config which caused a hidden yaml parsing error.", len(manifest.Resources()))
	}

	configMapErr := m.setConfigMap()
	if configMapErr != nil {
		return configMapErr
	}

	manifest, err = manifest.Transform(
		mf.InjectOwner(m.Owner),
		mf.InjectNamespace(m.Namespace),
		func(resource *unstructured.Unstructured) error {
			var deployment = &appsv1.Deployment{}
			if tErr := scheme.Scheme.Convert(resource, deployment, nil); tErr != nil {
				return tErr
			}

			if tErr := m.transform(deployment,
				m.addVolumesToDeployment,
				m.addStorageConfigVolume,
				m.addMMDomainSocketMount,
				m.addPassThroughPodFieldsToDeployment,
				m.addRuntimeToDeployment,
				m.syncGracePeriod,
				m.addMMEnvVars,
				m.addModelTypeConstraints,
				m.configureMMDeploymentForEtcdSecret,
				m.configureMMDeploymentForTLSSecret,
				m.configureRuntimeAnnotations,
			); tErr != nil {
				return tErr
			}

			return scheme.Scheme.Convert(deployment, resource, nil)
		},
	)
	if err != nil {
		return fmt.Errorf("Error transforming: %w", err)
	}

	if useStorageHelper(m.Owner) {
		manifest, err = manifest.Transform(
			addPullerTransform(m.Owner, m.PullerImage, m.PullerImageCommand, m.PullerResources),
		)
		if err != nil {
			return fmt.Errorf("Error transforming: %w", err)
		}
	}

	if logYaml {
		b, _ := yaml.Marshal(manifest.Resources())
		m.Log.Info(string(b))
	}

	err = manifest.Apply()
	if err != nil {
		return err
	}

	return nil
}

func (m *Deployment) Delete(ctx context.Context, client client.Client) error {
	m.Log.Info("Deleting model mesh deployment ", "m", m)
	return config.Delete(client, m.Owner, "config/internal/base/deployment.yaml.tmpl", m)
}

func (m *Deployment) transform(deployment *appsv1.Deployment, funcs ...func(deployment *appsv1.Deployment) error) error {
	for _, f := range funcs {
		err := f(deployment)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Deployment) addMMDomainSocketMount(deployment *appsv1.Deployment) error {
	var found bool
	var index int
	var c corev1.Container
	if found, index, c = findContainer(ModelMeshContainer, deployment); !found {
		return fmt.Errorf("Could not find the model mesh container %v", ModelMeshContainer)
	}

	if hasUnix, mountPoint, err := mountPoint(m.Owner); err != nil {
		return err
	} else if hasUnix {
		c.VolumeMounts = append(c.VolumeMounts, corev1.VolumeMount{
			Name:      "domain-socket",
			MountPath: mountPoint,
		})
	}

	deployment.Spec.Template.Spec.Containers[index] = c

	return nil
}

func (m *Deployment) addMMEnvVars(deployment *appsv1.Deployment) error {
	rt := m.Owner
	if rt.Spec.GrpcDataEndpoint != nil {
		e, err := ParseEndpoint(*rt.Spec.GrpcDataEndpoint)
		if err != nil {
			return err
		}
		if tcpE, ok := e.(TCPEndpoint); ok {
			err = setEnvironmentVar(ModelMeshContainer, ServeGrpcPortEnvVar, tcpE.Port, deployment)
			if err != nil {
				return err
			}
		} else if udsE, ok := e.(UnixEndpoint); ok {
			err = setEnvironmentVar(ModelMeshContainer, ServeGrpcUdsPathEnvVar, udsE.Path, deployment)
			if err != nil {
				return err
			}
		}
	}

	if useStorageHelper(rt) {
		err := setEnvironmentVar(ModelMeshContainer, GrpcPortEnvVar, strconv.Itoa(PullerPortNumber), deployment)
		if err != nil {
			return err
		}
	} else {
		e, err := ParseEndpoint(*rt.Spec.GrpcMultiModelManagementEndpoint)
		if err != nil {
			return err
		}
		if tcpE, ok := e.(TCPEndpoint); ok {
			err = setEnvironmentVar(ModelMeshContainer, GrpcPortEnvVar, tcpE.Port, deployment)
			if err != nil {
				return err
			}
		} else if udsE, ok := e.(UnixEndpoint); ok {
			err = setEnvironmentVar(ModelMeshContainer, GrpcUdsPathEnvVar, udsE.Path, deployment)
			if err != nil {
				return err
			}
		}
	}

	if m.EnableAccessLogging {
		//See https://github.ibm.com/ai-foundation/model-mesh/blob/e0e8570eb9be9a0f13c9e96c2fe3a6c737c67005/src/main/java/com/ibm/watson/modelmesh/ModelMeshEnvVars.java#L49
		err := setEnvironmentVar(ModelMeshContainer, "MM_LOG_EACH_INVOKE", "true", deployment)
		if err != nil {
			return err
		}
	}

	// See https://github.ibm.com/ai-foundation/model-mesh/blob/develop/src/main/java/com/ibm/watson/modelmesh/ModelMeshEnvVars.java#L31
	err := setEnvironmentVar(ModelMeshContainer, "MM_KVSTORE_PREFIX", ModelMeshEtcdPrefix, deployment)
	if err != nil {
		return err
	}
	// See https://github.ibm.com/ai-foundation/model-mesh/blob/898101124694f9eba10a34168ea7aac3a870f12e/src/main/java/com/ibm/watson/modelmesh/ModelMeshEnvVars.java#L65
	err = setEnvironmentVar(ModelMeshContainer, "MM_DEFAULT_VMODEL_OWNER", ModelMeshVModelOwner, deployment)
	if err != nil {
		return err
	}

	return nil
}

func (m *Deployment) setConfigMap() error {
	// get configmap name from servingRuntime
	rt := m.Owner
	configMap := rt.ObjectMeta.Annotations["productConfig"]
	if configMap == "" {
		return nil
	}

	// read configmap data.annotations
	clientParam := m.Client
	annotationConfigMap := &corev1.ConfigMap{}
	configMapErr := clientParam.Get(context.TODO(), client.ObjectKey{
		Name:      configMap,
		Namespace: m.Namespace}, annotationConfigMap)

	if configMapErr != nil {
		return fmt.Errorf("Unable to access ConfigMap '%s': %w", configMap, configMapErr)
	}

	m.AnnotationConfigMap = annotationConfigMap
	return nil
}
