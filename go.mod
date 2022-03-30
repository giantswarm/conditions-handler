module github.com/giantswarm/conditions-handler

go 1.14

require (
	github.com/giantswarm/conditions v0.3.0
	github.com/giantswarm/microerror v0.3.0
	github.com/giantswarm/micrologger v0.5.0
	k8s.io/api v0.23.5
	k8s.io/apimachinery v0.23.5
	sigs.k8s.io/cluster-api v0.3.10
	sigs.k8s.io/controller-runtime v0.11.2
	sigs.k8s.io/structured-merge-diff/v3 v3.0.1 // indirect
	sigs.k8s.io/yaml v1.3.0
)

replace sigs.k8s.io/cluster-api v0.3.10 => github.com/giantswarm/cluster-api v0.3.10-gs
