module github.com/tomoasleep/k8s-avahi

go 1.15

require (
	github.com/go-logr/logr v0.3.0
	github.com/godbus/dbus/v5 v5.0.2
	github.com/holoplot/go-avahi v0.0.0-20200423113835-c8b94bb23ec8
	golang.org/x/time v0.0.0-20201208040808-7e3f01d25324 // indirect
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	sigs.k8s.io/controller-runtime v0.8.2
)
