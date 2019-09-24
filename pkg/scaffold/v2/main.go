/*
Copyright 2019 The Kubernetes Authors.

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

package v2

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/pkg/scaffold/input"
	"sigs.k8s.io/kubebuilder/pkg/scaffold/util"
	"sigs.k8s.io/kubebuilder/pkg/scaffold/v1/resource"
	"sigs.k8s.io/kubebuilder/pkg/scaffold/v2/internal"

)

const (
	apiPkgImportScaffoldMarker    = "// +kubebuilder:scaffold:imports"
	apiSchemeScaffoldMarker       = "// +kubebuilder:scaffold:scheme"
	reconcilerSetupScaffoldMarker = "// +kubebuilder:scaffold:builder"
)

var _ input.File = &Main{}

// Main scaffolds a main.go to run Controllers
type Main struct {
	input.Input
}

// GetInput implements input.File
func (m *Main) GetInput() (input.Input, error) {
	if m.Path == "" {
		m.Path = filepath.Join("main.go")
	}
	m.TemplateBody = mainTemplate
	return m.Input, nil
}

// Update updates main.go with code fragments required to wire a new
// resource/controller.
func (m *Main) Update(opts *MainUpdateOptions) error {
	path := "main.go"

	resPkg, _ := util.GetResourceInfo(opts.Resource, opts.Project.Repo, opts.Project.Domain)

	// generate all the code fragments
	apiImportCodeFragment := fmt.Sprintf(`%s%s "%s/%s"
`, opts.Resource.Group, opts.Resource.Version, resPkg, opts.Resource.Version)
	ctrlImportCodeFragment := fmt.Sprintf(`"%s/controllers"
`, opts.Project.Repo)
	addschemeCodeFragment := fmt.Sprintf(`_ = %s%s.AddToScheme(scheme)
`, opts.Resource.Group, opts.Resource.Version)
	reconcilerSetupCodeFragment := fmt.Sprintf(`if err = (&controllers.%sReconciler{
	 	Client: mgr.GetClient(),
        Log: ctrl.Log.WithName("controllers").WithName("%s"),
	}).SetupWithManager(mgr); err != nil {
	 	setupLog.Error(err, "unable to create controller", "controller", "%s")
	 	os.Exit(1)
    }
`, opts.Resource.Kind, opts.Resource.Kind, opts.Resource.Kind)
	webhookSetupCodeFragment := fmt.Sprintf(`if err = (&%s%s.%s{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "%s")
		os.Exit(1)
	}
`, opts.Resource.Group, opts.Resource.Version, opts.Resource.Kind, opts.Resource.Kind)

	if opts.WireResource {
		err := internal.InsertStringsInFile(path,
			map[string][]string{
				apiPkgImportScaffoldMarker: {apiImportCodeFragment},
				apiSchemeScaffoldMarker:    {addschemeCodeFragment},
			})
		if err != nil {
			return err
		}
	}

	if opts.WireController {
		return internal.InsertStringsInFile(path,
			map[string][]string{
				apiPkgImportScaffoldMarker:    {apiImportCodeFragment, ctrlImportCodeFragment},
				apiSchemeScaffoldMarker:       {addschemeCodeFragment},
				reconcilerSetupScaffoldMarker: {reconcilerSetupCodeFragment},
			})
	}

	if opts.WireWebhook {
		return internal.InsertStringsInFile(path,
			map[string][]string{
				apiPkgImportScaffoldMarker:    {apiImportCodeFragment, ctrlImportCodeFragment},
				apiSchemeScaffoldMarker:       {addschemeCodeFragment},
				reconcilerSetupScaffoldMarker: {webhookSetupCodeFragment},
			})
	}

	return nil
}

// MainUpdateOptions contains info required for wiring an API/Controller in
// main.go.
type MainUpdateOptions struct {
	// Project contains info about the project
	Project *input.ProjectFile

	// Resource is the resource being added
	Resource *resource.Resource

	// Flags to indicate if resource/controller is being scaffolded or not
	WireResource   bool
	WireController bool
	WireWebhook    bool
}

var mainTemplate = fmt.Sprintf(`{{ .Boilerplate }}

package main

import (
	"flag"
    "os"

	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/log/zap"
    "k8s.io/apimachinery/pkg/runtime"
	kubemetrics "sigs.k8s.io/kubebuilder/pkg/kube-metrics"
    "sigs.k8s.io/kubebuilder/pkg/k8sutil"
	"sigs.k8s.io/kubebuilder/pkg/metrics"

	%s
)

var (
	scheme = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

// Change below variables to serve metrics on different host or port.
const (
	metricsHost               = "0.0.0.0"
	metricsPort         int32 = 8080
	operatorMetricsPort int32 = 8443
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	%s
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,

    ctrl.SetLogger(zap.Logger(true))

    // Get a config to talk to the apiserver
    cfg, err := config.GetConfig()
    if err != nil {
       setupLog.Error(err, "unable to get config manager")
       os.Exit(1)
    }

    // Create a new Cmd to provide shared dependencies and start components
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		Port:               9443, 
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	%s

    // Metrics 
	if err = serveCRMetrics(cfg); err != nil {
		log.Info("Could not generate and serve custom resource metrics", "error", err.Error())
	}

	// Add to the below struct any other metrics ports you want to expose.
	servicePorts := []v1.ServicePort{
		{Port: metricsPort, Name: metrics.OperatorPortName, Protocol: v1.ProtocolTCP, TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: metricsPort}},
		{Port: operatorMetricsPort, Name: metrics.CRPortName, Protocol: v1.ProtocolTCP, TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: operatorMetricsPort}},
	}
	// Create Service object to expose the metrics port(s).
	service, err := metrics.CreateMetricsService(ctx, cfg, servicePorts)
	if err != nil {
		log.Info("Could not create metrics Service", "error", err.Error())
	}

	// CreateServiceMonitors will automatically create the prometheus-operator ServiceMonitor resources
	// necessary to configure Prometheus to scrape metrics from this operator project.
	services := []*v1.Service{service}
	_, err = metrics.CreateServiceMonitors(cfg, namespace, services)
	if err != nil {
		log.Info("Could not create ServiceMonitor object", "error", err.Error())
		// If this operator is deployed to a cluster without the prometheus-operator running, it will return
		// ErrServiceMonitorNotPresent, which can be used to safely skip ServiceMonitor creation.
		if err == metrics.ErrServiceMonitorNotPresent {
			log.Info("Install prometheus-operator in your cluster to create ServiceMonitor objects", "error", err.Error())
		}
	}

    //Start the manager 
	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

// serveCRMetrics gets the Operator/CustomResource GVKs and generates metrics based on those types.
// It serves those metrics on "http://metricsHost:operatorMetricsPort".
func serveCRMetrics(cfg *rest.Config) error {
	// Below function returns filtered operator/CustomResource specific GVKs.
	// For more control override the below GVK list with your own custom logic.
	filteredGVK, err := k8sutil.GetGVKsFromAddToScheme(apis.AddToScheme)
	if err != nil {
		return err
	}
	// Get the namespace the operator is currently deployed in.
	operatorNs, err := k8sutil.GetOperatorNamespace()
	if err != nil {
		return err
	}
	// To generate metrics in other namespaces, add the values below.
	ns := []string{operatorNs}
	// Generate and serve custom resource specific metrics.
	err = kubemetrics.GenerateAndServeCRMetrics(cfg, ns, filteredGVK, metricsHost, operatorMetricsPort)
	if err != nil {
		return err
	}
	return nil
}
`, apiPkgImportScaffoldMarker, apiSchemeScaffoldMarker, reconcilerSetupScaffoldMarker)
