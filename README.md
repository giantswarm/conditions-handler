[![CircleCI](https://circleci.com/gh/giantswarm/conditionshandler.svg?style=shield)](https://circleci.com/gh/giantswarm/conditionshandler)

# conditionshandler

`giantswarm/conditionshandler` module contains implementations of provider-independent
condition handlers that can be plugged into multiple operators, or even used from
other non-operator tools, like CLI apps.

Other Cluster API conditions projects:
- https://github.com/giantswarm/conditions

To read more about conditions, check out these articles:
- Cluster API: [Conditions - Cluster status at glance](https://github.com/kubernetes-sigs/cluster-api/blob/master/docs/proposals/20200506-conditions.md)
- Kubernetes API conventions, section about conditions: [Typical Status Properties - Conditions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties)
