module github.com/giantswarm/conditions-handler

go 1.14

require (
	github.com/giantswarm/conditions v0.3.0
	github.com/giantswarm/microerror v0.3.0
	github.com/giantswarm/micrologger v0.5.0
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	sigs.k8s.io/cluster-api v0.3.10
	sigs.k8s.io/controller-runtime v0.10.2
	sigs.k8s.io/yaml v1.2.0
)

replace sigs.k8s.io/cluster-api v0.3.10 => github.com/giantswarm/cluster-api v0.3.10-gs
