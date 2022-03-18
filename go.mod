module github.com/giantswarm/conditions-handler

go 1.14

require (
	github.com/giantswarm/conditions v0.3.0
	github.com/giantswarm/microerror v0.4.0
	github.com/giantswarm/micrologger v1.0.0
	k8s.io/api v0.18.19
	k8s.io/apimachinery v0.18.19
	sigs.k8s.io/cluster-api v0.3.10
	sigs.k8s.io/controller-runtime v0.6.4
	sigs.k8s.io/yaml v1.2.0
)

replace sigs.k8s.io/cluster-api v0.3.10 => github.com/giantswarm/cluster-api v0.3.10-gs
