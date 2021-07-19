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
/*
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

package controllers

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"

	"wmlserving.ai.ibm.com/controller/controllers/modelmesh"

	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"wmlserving.ai.ibm.com/controller/pkg/mmesh"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// ServiceReconciler reconciles a ServingRuntime object
type ServiceReconciler struct {
	client.Client
	Log                  logr.Logger
	Scheme               *runtime.Scheme
	ConfigProvider       *ConfigProvider
	ConfigMapName        types.NamespacedName
	ControllerDeployment types.NamespacedName

	ModelMeshService *mmesh.MMService
	ModelEventStream *mmesh.ModelMeshEventStream
}

func (r *ServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.V(1).Info("Service reconciler called")
	var cfg *Config
	var changed bool
	if req.NamespacedName == r.ConfigMapName || !r.ModelEventStream.IsWatching() {
		cfg = r.ConfigProvider.GetConfig()
		tlsConfig, err := r.tlsConfigFromSecret(ctx, cfg.TLS.SecretName)
		if err != nil {
			return RequeueResult, err
		}
		var metricsPort uint16 = 0
		if cfg.Metrics.Enabled {
			metricsPort = cfg.Metrics.Port
		}
		changed = r.ModelMeshService.UpdateConfig(
			cfg.InferenceServiceName, cfg.InferenceServicePort,
			cfg.ModelMeshEndpoint, cfg.TLS.SecretName, tlsConfig, cfg.HeadlessService, metricsPort)
	}

	if changed || req.NamespacedName != r.ConfigMapName {
		d := &appsv1.Deployment{}
		err := r.Client.Get(ctx, r.ControllerDeployment, d)
		if err != nil {
			if k8serr.IsNotFound(err) {
				// No need to delete because the Deployment is the owner
				return ctrl.Result{}, nil
			}
			return RequeueResult, err
		}
		err2, requeue := r.applyService(ctx, d)
		if err2 != nil || requeue {
			//TODO probably shorter requeue time (immediate?) for service recreate case
			return RequeueResult, err2
		}
	}

	if cfg != nil {
		err := r.ModelEventStream.UpdateWatchedService(ctx, cfg.GetEtcdSecretName(), r.ModelMeshService.Name)
		if err != nil {
			return RequeueResult, err
		}
	}
	return ctrl.Result{}, nil
}

func (r *ServiceReconciler) tlsConfigFromSecret(ctx context.Context, secretName string) (*tls.Config, error) {
	if secretName == "" {
		return nil, nil
	}
	tlsSecret := corev1.Secret{}
	err := r.Client.Get(ctx, client.ObjectKey{
		Name:      secretName,
		Namespace: r.ControllerDeployment.Namespace}, &tlsSecret)
	if err != nil {
		r.Log.Error(err, "Unable to access TLS secret", "secretName", secretName)
		return nil, fmt.Errorf("unable to access TLS secret '%s': %v", secretName, err)
	}
	cert, ok2 := tlsSecret.Data[modelmesh.TLSSecretCertKey]
	key, ok := tlsSecret.Data[modelmesh.TLSSecretKeyKey]
	if !ok || !ok2 {
		r.Log.Error(err, "TLS secret missing required keys", "secretName", secretName)
		return nil, fmt.Errorf("TLS secret '%s' missing %s and/or %s",
			secretName, modelmesh.TLSSecretCertKey, modelmesh.TLSSecretKeyKey)
	}
	certificate, err := tls.X509KeyPair(cert, key)
	if err != nil {
		r.Log.Error(err, "Could not load client key pair")
		return nil, fmt.Errorf("could not load client key pair: %v", err)
	}
	certPool, _ := x509.SystemCertPool() // this returns a copy
	if certPool == nil {
		certPool = x509.NewCertPool()
	}
	if ok := certPool.AppendCertsFromPEM(cert); !ok {
		return nil, errors.New("failed to append ca certs")
	}
	return &tls.Config{Certificates: []tls.Certificate{certificate}, RootCAs: certPool}, nil
}

func (r *ServiceReconciler) applyService(ctx context.Context, d *appsv1.Deployment) (error, bool) {
	s := &corev1.Service{}
	serviceName := r.ModelMeshService.Name
	exists := true
	err := r.Get(ctx, types.NamespacedName{Name: serviceName, Namespace: r.ControllerDeployment.Namespace}, s)
	if err != nil && k8serr.IsNotFound(err) {
		exists = false
		s.Name = serviceName
		s.Namespace = r.ControllerDeployment.Namespace
	} else if err != nil {
		return err, false
	}

	headless := r.ModelMeshService.Headless
	if exists && (s.Spec.ClusterIP == "None") != headless {
		r.Log.Info("Deleting/recreating Service because headlessness changed",
			"service", serviceName, "headless", headless)
		// Have to recreate since ClusterIP field is immutable
		if err = r.Delete(ctx, s); err != nil {
			return err, false
		}
		// This will trigger immediate re-reconcilation
		return nil, true
	}

	commonLabelValue := "modelmesh-controller"
	s.Labels = map[string]string{
		"modelmesh-service":            serviceName,
		"app.kubernetes.io/instance":   commonLabelValue,
		"app.kubernetes.io/managed-by": commonLabelValue,
		"app.kubernetes.io/name":       commonLabelValue,
	}
	s.Spec.Selector = map[string]string{"modelmesh-service": serviceName}
	s.Spec.Ports = []corev1.ServicePort{
		{
			Name:       "grpc",
			Port:       int32(r.ModelMeshService.Port),
			TargetPort: intstr.FromString("grpc"),
		},
	}

	if r.ModelMeshService.MetricsPort > 0 {
		s.Spec.Ports = append(s.Spec.Ports, corev1.ServicePort{
			Name:       "prometheus",
			Port:       int32(r.ModelMeshService.MetricsPort),
			TargetPort: intstr.FromString("prometheus"),
		})
	}

	err = controllerutil.SetControllerReference(d, s, r.Scheme)
	if err != nil {
		r.Log.Error(err, "Could not set owner reference")
	}

	if !exists {
		if headless {
			s.Spec.ClusterIP = "None"
		}
		r.ModelMeshService.Disconnect()
		err = r.Create(ctx, s)
		if err != nil {
			r.Log.Error(err, "Could not create service")
		}
	} else {
		err = r.ModelMeshService.Connect()
		if err != nil {
			r.Log.Error(err, "Error establishing model-mesh gRPC conn")
		}
		err2 := r.Update(ctx, s)
		if err2 != nil {
			r.Log.Error(err, "Could not update service")
			if err == nil {
				err = err2
			}
		}
	}

	return err, false
}

func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	filter := func(meta metav1.Object) bool {
		return meta.GetName() == r.ControllerDeployment.Name &&
			meta.GetNamespace() == r.ControllerDeployment.Namespace
	}
	err := ctrl.NewControllerManagedBy(mgr).
		Named("ServiceReconciler").
		For(&appsv1.Deployment{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool { return filter(event.Object) },
			UpdateFunc: func(event event.UpdateEvent) bool { return filter(event.ObjectNew) },
			DeleteFunc: func(event event.DeleteEvent) bool { return false },
		})).
		Owns(&corev1.Service{}).
		Watches(&source.Kind{Type: &corev1.ConfigMap{}},
			ConfigWatchHandler(r.ConfigMapName, func() []reconcile.Request {
				return []reconcile.Request{{NamespacedName: r.ConfigMapName}}
			}, r.ConfigProvider, &r.Client)).
		Complete(r)

	return err
}
