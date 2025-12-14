package eva

import (
	"context"

	"fmt"

	"github.com/dayaliuzzo/Smooth-Operator/api/v1alpha1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *EvaReconciler) reconcileResources(ctx context.Context, eva *v1alpha1.Eva, currentState *evaCurrentState, logger logr.Logger) (*v1alpha1.EvaStatus, error) {
	statusUpdate := &v1alpha1.EvaStatus{}
	statusUpdate, err := r.reconcileJob(ctx, eva, currentState.Job, logger)
	if err != nil {
		return nil, err
	}
	return statusUpdate, nil
}

func (r *EvaReconciler) reconcileJob(ctx context.Context, eva *v1alpha1.Eva, jobState jobState, logger logr.Logger) (*v1alpha1.EvaStatus, error) {
	newStatus := &v1alpha1.EvaStatus{}
	if !jobState.Exists {
		if eva.Status.Phase == "" || eva.Status.Phase == v1alpha1.EvaPhasePending {
			logger.Info("Creating Job for Eva", "Eva.Name", eva.Name)
			if err := r.createJob(ctx, eva, logger); err != nil {
				return nil, err
			}
			newStatus.Phase = v1alpha1.EvaPhasePending
			newStatus.Conditions = []metav1.Condition{
				{
					Type:               string(v1alpha1.EvaConditionAvailable),
					Status:             metav1.ConditionFalse,
					Reason:             "JobCreated",
					Message:            "The Job has been created.",
					ObservedGeneration: eva.Generation,
				},
			}
			return newStatus, nil
		} else {
			newStatus.Phase = v1alpha1.EvaPhaseFailed
			newStatus.Conditions = []metav1.Condition{
				{
					Type:               string(v1alpha1.EvaConditionAvailable),
					Status:             metav1.ConditionFalse,
					Reason:             "JobMissing",
					Message:            "The Job is missing.",
					ObservedGeneration: eva.Generation,
				},
			}
			return newStatus, nil
		}
	} else {
		if jobState.Succeeded > 0 {
			newStatus.Phase = v1alpha1.EvaPhaseSucceeded
			newStatus.Conditions = []metav1.Condition{
				{
					Type:               string(v1alpha1.EvaConditionAvailable),
					Status:             metav1.ConditionTrue,
					Reason:             "JobSucceeded",
					Message:            "The Job has succeeded.",
					ObservedGeneration: eva.Generation,
				},
			}
			return newStatus, nil
		}
		if jobState.FailedPods > 0 {
			newStatus.Phase = v1alpha1.EvaPhaseFailed
			newStatus.Conditions = []metav1.Condition{
				{
					Type:               string(v1alpha1.EvaConditionAvailable),
					Status:             metav1.ConditionFalse,
					Reason:             "JobFailed",
					Message:            "The Job has failed.",
					ObservedGeneration: eva.Generation,
				},
			}
			return newStatus, nil
		}
		if jobState.ImagePullFailed {
			newStatus.Phase = v1alpha1.EvaPhaseFailed
			newStatus.Conditions = []metav1.Condition{
				{
					Type:               string(v1alpha1.EvaConditionAvailable),
					Status:             metav1.ConditionFalse,
					Reason:             "ImagePullBackOff",
					Message:            "Failed to pull container image.",
					ObservedGeneration: eva.Generation,
				},
			}
			return newStatus, nil
		}
		newStatus.Phase = v1alpha1.EvaPhaseRunning
		// Only add condition if it would actually change
		existingCondition := meta.FindStatusCondition(eva.Status.Conditions, string(v1alpha1.EvaConditionAvailable))
		if existingCondition == nil || existingCondition.Status != metav1.ConditionTrue ||
			existingCondition.Reason != "JobRunning" {
			newStatus.Conditions = []metav1.Condition{
				{
					Type:               string(v1alpha1.EvaConditionAvailable),
					Status:             metav1.ConditionTrue,
					Reason:             "JobRunning",
					Message:            "The Job is running.",
					ObservedGeneration: eva.Generation,
				},
			}
		}
	}
	return newStatus, nil
}

func (r *EvaReconciler) createJob(ctx context.Context, eva *v1alpha1.Eva, logger logr.Logger) error {
	jobName := fmt.Sprintf("%s-job", eva.Name)
	containerName := fmt.Sprintf("%s-container", eva.Name)
	desired := buildJob(jobName, eva.Namespace,
		WithJobLabels(r.generateLabels(eva, nil)),
		WithJobContainerName(containerName),
		WithJobImage(eva.Spec.Image),
		WithJobCommand(eva.Spec.Command),
		WithJobImagePullSecret(eva.Spec.ImagePullSecret),
		WithJobBackoffLimit(0))

	if err := controllerutil.SetControllerReference(eva, desired, r.Scheme); err != nil {
		logger.Error(err, "failed to set controller reference: ", "error", err)
		return err
	}
	if err := r.Create(ctx, desired); err != nil {
		logger.Error(err, "failed to create job: ", "error", err)
		return err
	}

	logger.Info("Created Job for Eva", "eva", eva.Name, "image", eva.Spec.Image)
	return nil
}

func (r *EvaReconciler) generateLabels(eva *v1alpha1.Eva, newLabels map[string]string) map[string]string {
	labels := map[string]string{
		"app":      "eva-controller",
		"eva-name": eva.Name,
	}
	for k, v := range newLabels {
		labels[k] = v
	}

	return labels
}
