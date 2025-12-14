package eva

import (
	"context"

	"github.com/dayaliuzzo/Smooth-Operator/api/v1alpha1"
	"github.com/go-logr/logr"
	kbatch "k8s.io/api/batch/v1"
)

// getCurrentState observes the current state of all resources managed by the Eva controller
func (r *EvaReconciler) getCurrentState(ctx context.Context, eva *v1alpha1.Eva, logger logr.Logger) (evaCurrentState, error) {
	var err error
	currentState := evaCurrentState{}

	currentState.Job, err = r.getJobState(ctx, eva, logger)
	if err != nil {
		return currentState, err
	}
	return currentState, nil
}

// getJobState observes the current state of the Job owned by this Eva
func (r *EvaReconciler) getJobState(ctx context.Context, eva *v1alpha1.Eva, logger logr.Logger) (jobState, error) {
	var err error
	var job *kbatch.Job
	jobState := jobState{}
	job, err = GetOwnedJob(ctx, r.Client, eva, ownerKey)
	if err != nil {
		return jobState, err
	}
	if job == nil {
		jobState.Exists = false
		return jobState, nil
	}

	jobState.Exists = true
	jobState.Succeeded = job.Status.Succeeded
	jobState.Active = job.Status.Active
	jobState.FailedPods = job.Status.Failed

	return jobState, nil
}
