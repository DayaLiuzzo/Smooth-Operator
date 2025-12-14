package eva

import (
	"context"

	"github.com/dayaliuzzo/Smooth-Operator/api/v1alpha1"
	"github.com/go-logr/logr"
	kbatch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

	// Check for image pull errors in Pods
	jobState.ImagePullFailed, err = r.checkPodImagePullErrors(ctx, job, logger)
	if err != nil {
		logger.Error(err, "Failed to check pod image pull status")
	}

	return jobState, nil
}

// checkPodImagePullErrors checks if any pods owned by the job have image pull errors
func (r *EvaReconciler) checkPodImagePullErrors(ctx context.Context, job *kbatch.Job, logger logr.Logger) (bool, error) {
	podList := &corev1.PodList{}
	if err := r.List(ctx, podList, client.InNamespace(job.Namespace), client.MatchingLabels{"job-name": job.Name}); err != nil {
		return false, err
	}

	for _, pod := range podList.Items {
		// Check container statuses for image pull errors
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.State.Waiting != nil {
				reason := containerStatus.State.Waiting.Reason
				if reason == "ImagePullBackOff" || reason == "ErrImagePull" {
					logger.Info("Detected image pull failure", "pod", pod.Name, "container", containerStatus.Name, "reason", reason)
					return true, nil
				}
			}
		}
		// Also check init container statuses
		for _, containerStatus := range pod.Status.InitContainerStatuses {
			if containerStatus.State.Waiting != nil {
				reason := containerStatus.State.Waiting.Reason
				if reason == "ImagePullBackOff" || reason == "ErrImagePull" {
					logger.Info("Detected image pull failure in init container", "pod", pod.Name, "container", containerStatus.Name, "reason", reason)
					return true, nil
				}
			}
		}
	}

	return false, nil
}
