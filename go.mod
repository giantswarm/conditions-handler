module github.com/giantswarm/conditions-handler

go 1.14

require (
	github.com/giantswarm/conditions v0.0.0-20201125143305-c28ae85a4d78
	github.com/giantswarm/microerror v0.2.1
	github.com/giantswarm/micrologger v0.3.4
	k8s.io/api v0.18.9
	k8s.io/apimachinery v0.18.9
	sigs.k8s.io/cluster-api v0.3.10
	sigs.k8s.io/controller-runtime v0.6.4
)

replace sigs.k8s.io/cluster-api v0.3.10 => github.com/giantswarm/cluster-api v0.3.10-gs
