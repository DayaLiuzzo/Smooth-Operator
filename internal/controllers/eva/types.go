package eva

import ()

type jobState struct {
	Exists          bool
	Ready           bool
	Active          int32
	Succeeded       int32
	FailedPods      int32
	ImagePullFailed bool
}

type deploymentState struct {
	Exists bool
	Ready  bool
}

type serviceState struct {
	Exists bool
}

type evaCurrentState struct {
	Job        jobState
	Service    serviceState
	Deployment deploymentState
}
