module github.com/giantswarm/conditions-handler

go 1.14

require (
	github.com/caddyserver/caddy v1.0.3 // indirect
	github.com/docker/docker v0.7.3-0.20190327010347-be7ac8be2ae0 // indirect
	github.com/drone/envsubst v1.0.3-0.20200709223903-efdb65b94e5a // indirect
	github.com/giantswarm/conditions v0.4.0
	github.com/giantswarm/microerror v0.3.0
	github.com/giantswarm/micrologger v0.5.0
	github.com/google/go-github v17.0.0+incompatible // indirect
	github.com/gophercloud/gophercloud v0.1.0 // indirect
	gotest.tools v2.2.0+incompatible // indirect
	k8s.io/api v0.21.3
	k8s.io/apimachinery v0.21.3
	sigs.k8s.io/cluster-api v0.4.2
	sigs.k8s.io/controller-runtime v0.9.6
	sigs.k8s.io/kind v0.7.1-0.20200303021537-981bd80d3802 // indirect
	sigs.k8s.io/structured-merge-diff/v3 v3.0.1 // indirect
	sigs.k8s.io/yaml v1.2.0
)

replace sigs.k8s.io/cluster-api v0.3.10 => github.com/giantswarm/cluster-api v0.3.10-gs
