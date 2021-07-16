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

package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"regexp"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	servingaiibmcomv1 "wmlserving.ai.ibm.com/controller/api/v1"
	servingv1 "wmlserving.ai.ibm.com/controller/api/v1"
	"wmlserving.ai.ibm.com/controller/controllers"
	"wmlserving.ai.ibm.com/controller/controllers/modelmesh"
	"wmlserving.ai.ibm.com/controller/pkg/mmesh"
	// +kubebuilder:scaffold:imports
)

var (
	scheme              = runtime.NewScheme()
	setupLog            = ctrl.Log.WithName("setup")
	ControllerNamespace string
)

const (
	ControllerNamespaceEnvVar      = "NAMESPACE"
	DefaultControllerNamespace     = "model-serving"
	KubeNamespaceFile              = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	ControllerPodNameEnvVar        = "POD_NAME"
	ControllerDeploymentNameEnvVar = "CONTROLLER_DEPLOYMENT"
	DefaultControllerName          = "wmlserving-controller"
	UserConfigMapName              = "model-serving-config"
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	err := servingv1.AddToScheme(scheme)
	if err != nil {
		log.Fatalf("cannot add model serving v1 scheme, %v", err)
	}
	_ = batchv1.AddToScheme(scheme)
	_ = servingaiibmcomv1.AddToScheme(scheme)
	_ = servingv1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	// ----- mmesh related envar setup -----
	controllerNamespace := os.Getenv(ControllerNamespaceEnvVar)
	if controllerNamespace == "" {
		bytes, err := ioutil.ReadFile(KubeNamespaceFile)
		if err != nil {
			//TODO check kube context and retrieve namespace from there
			setupLog.Info("Error reading Kube-mounted namespace file, reverting to default namespace",
				"file", KubeNamespaceFile, "err", err, "default", DefaultControllerNamespace)
			controllerNamespace = DefaultControllerNamespace
		} else {
			controllerNamespace = string(bytes)
		}
	}
	ControllerNamespace = controllerNamespace

	controllerDeploymentName := os.Getenv(ControllerDeploymentNameEnvVar)
	if controllerDeploymentName == "" {
		podName := os.Getenv(ControllerPodNameEnvVar)
		if podName != "" {
			if matches := regexp.MustCompile("(.*)-.*-.*").FindStringSubmatch(podName); len(matches) == 2 {
				deployment := matches[1]
				setupLog.Info("Use controller deployment from POD_NAME", "Deployment", deployment)
				controllerDeploymentName = deployment
			}
		}
		if controllerDeploymentName == "" {
			setupLog.Info("Skip empty Controller deployment from Env Var, use default",
				"name", DefaultControllerName)
			controllerDeploymentName = DefaultControllerName
		}
	}

	// TODO: use the manager client instead. This will require restructuring the dependency
	// relationship with the manager so that this code runs after mgr.Start()
	cfg := config.GetConfigOrDie()
	cl, err := client.New(cfg, client.Options{})
	if err != nil {
		setupLog.Error(err, "unable to create an api server client")
		os.Exit(1)
	}

	cp, err := controllers.NewConfigProvider(context.Background(), cl, types.NamespacedName{Name: UserConfigMapName, Namespace: ControllerNamespace})
	if err != nil {
		setupLog.Error(err, "Error loading user config from configmap", "ConfigMapName", UserConfigMapName)
		os.Exit(1)
	}
	conf := cp.GetConfig()

	setupLog.Info("Using puller", "image", conf.StorageHelperImage.TaggedImage())
	setupLog.Info("Using model-mesh", "image", conf.ModelMeshImage.TaggedImage())
	setupLog.Info("Using inference service", "name", conf.InferenceServiceName, "port", conf.InferenceServicePort)

	// mmesh service kubedns or hostname
	mmeshEndpoint := conf.ModelMeshEndpoint

	setupLog.Info("MMesh Configuration", "serviceName", conf.InferenceServiceName, "port", conf.InferenceServicePort,
		"mmeshEndpoint", mmeshEndpoint)

	//TODO: this should be moved out of package globals
	modelmesh.StorageSecretName = conf.StorageSecretName

	// ----- end of mmesh related envar setup -----

	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                  scheme,
		MetricsBindAddress:      metricsAddr,
		Namespace:               ControllerNamespace,
		Port:                    9443,
		LeaderElection:          enableLeaderElection,
		LeaderElectionID:        "wmlserving-controller-leader-lock",
		LeaderElectionNamespace: ControllerNamespace,
		HealthProbeBindAddress:  probeAddr,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	_, err = mmesh.InitGrpcResolver(ControllerNamespace, mgr)
	if err != nil {
		setupLog.Error(err, "Failed to Initialize Grpc Resolver, exit")
		os.Exit(1)
	}

	mmService := mmesh.NewMMService()

	modelEventStream, err := mmesh.NewModelEventStream(ctrl.Log.WithName("ModelMeshEventStream"),
		mgr.GetClient(), ControllerNamespace)
	if err != nil {
		setupLog.Error(err, "Failed to Initialize Model Event Stream, exit")
		os.Exit(1)
	}

	if err = (&controllers.ServiceReconciler{
		Client:               mgr.GetClient(),
		Log:                  ctrl.Log.WithName("controllers").WithName("Service"),
		Scheme:               mgr.GetScheme(),
		ControllerDeployment: types.NamespacedName{Namespace: ControllerNamespace, Name: controllerDeploymentName},
		ModelMeshService:     mmService,
		ModelEventStream:     modelEventStream,
		ConfigProvider:       cp,
		ConfigMapName:        types.NamespacedName{Namespace: ControllerNamespace, Name: UserConfigMapName},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Service")
		os.Exit(1)
	}

	if err = (&controllers.PredictorReconciler{
		Client:    mgr.GetClient(),
		Log:       ctrl.Log.WithName("controllers").WithName("Predictor"),
		MMService: mmService,
	}).SetupWithManager(mgr, modelEventStream); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Predictor")
		os.Exit(1)
	}

	if err = (&controllers.ServingRuntimeReconciler{
		Client:              mgr.GetClient(),
		Log:                 ctrl.Log.WithName("controllers").WithName("ServingRuntime"),
		Scheme:              mgr.GetScheme(),
		ConfigProvider:      cp,
		ConfigKey:           types.NamespacedName{Namespace: ControllerNamespace, Name: UserConfigMapName},
		DeploymentNamespace: ControllerNamespace,
		DeploymentName:      controllerDeploymentName,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ServingRuntime")
		os.Exit(1)
	}

	// +kubebuilder:scaffold:builder

	// Add Healthz Endpoint
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}

	// Add Readyz Endpoint
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
