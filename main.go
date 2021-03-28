package main

import (
	"os"

	"github.com/tomoasleep/k8s-avahi/controller"
	"github.com/tomoasleep/k8s-avahi/mdns"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
)

func main() {
	scheme := runtime.NewScheme()
	log := ctrl.Log.WithName("setup")

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     ":8080",
		Port:                   9443,
		LeaderElection:         false,
		LeaderElectionID:       "tomoasleep.k8s-avahi",
		HealthProbeBindAddress: ":9090",
	})
	if err != nil {
		log.Error(err, "unable to start manager")
		os.Exit(1)
	}

	cli, err := mdns.NewClient(mdns.WithSystemBus())
	if err != nil {
		log.Error(err, "unable to start clien")
		os.Exit(1)
	}

	c := &controller.Controller{
		Client:     mgr.GetClient(),
		Log:        ctrl.Log.WithName("controller").WithName("k8s-avahi"),
		MdnsClient: cli,
	}
	defer c.Close()

	err = c.SetupManager(mgr)
	if err != nil {
		log.Error(err, "unable to setup manager")
		os.Exit(1)
	}

	err = mgr.AddHealthzCheck("ping", healthz.Ping)
	if err != nil {
		log.Error(err, "unable to add healthz check")
		os.Exit(1)
	}

	err = mgr.AddReadyzCheck("ping", healthz.Ping)
	if err != nil {
		log.Error(err, "unable to add readyz check")
		os.Exit(1)
	}

	err = mgr.Start(ctrl.SetupSignalHandler())
	if err != nil {
		log.Error(err, "unable to start controller")
		os.Exit(1)
	}
}
