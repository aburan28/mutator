package main

import (
	"flag"
	"log"
	namespace "mutator/pkg/controllers"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var (
	scheme   = runtime.NewScheme()
	entryLog = ctrl.Log.WithName("entrypoint")
)

func main() {
	var debug bool
	var metricsAddr string
	var metricsPort int
	var enableLeaderElection bool
	var probeAddr string

	entryLog.Info("starting manager")

	flag.BoolVar(&debug, "debug", false, "enable debug logging")
	flag.IntVar(&metricsPort, "metrics-port", 8080, "port to expose metrics on")
	flag.BoolVar(&enableLeaderElection, "leader-election", false, "enable leader election")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "address to bind health probe")

	opts := zap.Options{
		Development: true,
	}

	opts.BindFlags(flag.CommandLine)

	flag.Parse()
	config := ctrl.GetConfigOrDie()

	//mgr, err := manager.New(config, manager.Options{})

	mgr, err := ctrl.NewManager(config, ctrl.Options{
		Scheme: scheme,
		Metrics: server.Options{
			BindAddress: metricsAddr,
		},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
	})
	if err != nil {
		log.Fatal(err)
	}
	r := &namespace.NamespaceReconciler{
		Scheme:   mgr.GetScheme(),
		Client:   mgr.GetClient(),
		Log:      ctrl.Log.WithName("controllers").WithName("namespace"),
		Recorder: mgr.GetEventRecorderFor("mutator"),
	}
	err = r.SetupWithManager(mgr)
	if err != nil {
		log.Fatal(err)
	}

}
