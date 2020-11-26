package internal

const (
	// lastDeployedReleaseVersion is the version annotation put into Cluster CR to
	// define which Giant Swarm release version was last successfully deployed
	// during cluster creation or upgrade. Versions are defined as semver version
	// without the "v" prefix, e.g. 14.1.0, which means that cluster was created
	// with or upgraded to Giant Swarm release v14.1.0.
	LastDeployedReleaseVersion = "release.giantswarm.io/last-deployed-version"

	// UpgradingToNodePools is set to True during the first cluster upgrade to node pools release.
	UpgradingToNodePools = "release.giantswarm.io/upgrading-to-node-pools"
)
