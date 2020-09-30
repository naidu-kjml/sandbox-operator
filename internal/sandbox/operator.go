package sandbox

import (
	"context"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	ctrl "sigs.k8s.io/controller-runtime"

	devopsv1 "gitlab.com/stackvista/devops/devopserator/apis/devops/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	devopscontroller "gitlab.com/stackvista/devops/devopserator/controllers/devops"
	"gitlab.com/stackvista/devops/devopserator/internal/logr"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(devopsv1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

type OperatorConfig struct {
	MetricsAddr          string
	EnableLeaderElection bool
}

func StartOperator(ctx context.Context, config *OperatorConfig) error {
	logger := logr.Ctx(ctx)
	ctrl.SetLogger(logger)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: config.MetricsAddr,
		Port:               9443,
		LeaderElection:     config.EnableLeaderElection,
		LeaderElectionID:   "6221cfa4.devopserator.stackstate.com",
		Namespace:          "",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&devopscontroller.SandboxReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Sandbox"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Sandbox")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		return err
	}

	return nil
}
