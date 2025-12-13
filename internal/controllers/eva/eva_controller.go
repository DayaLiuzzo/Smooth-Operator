/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package eva

import (
	"context"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	kbatch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/dayaliuzzo/Smooth-Operator/api/v1alpha1"
	"github.com/dayaliuzzo/Smooth-Operator/internal/controllers/common"
)

const ownerKey = ".metadata.controller"
const evaFinalizer = "geofront.nerv.com/finalizer"

// EvaReconciler reconciles a Eva object
type EvaReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=geofront.nerv.com,resources=evas,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=geofront.nerv.com,resources=evas/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=geofront.nerv.com,resources=evas/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Eva object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.4/pkg/reconcile
func (r *EvaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)
	logger.Info("Reconciling Eva", "request", req)

	var eva v1alpha1.Eva
	if err := r.Get(ctx, req.NamespacedName, &eva); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if !eva.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.handleDelete(ctx, &eva, logger)
	}
	if !controllerutil.ContainsFinalizer(&eva, evaFinalizer) {
		logger.Info("Adding finalizer", "finalizer", evaFinalizer)
		controllerutil.AddFinalizer(&eva, evaFinalizer)
		if err := r.Update(ctx, &eva); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	currentState, err := r.getCurrentState(ctx, &eva, logger)
	if err != nil {
		return ctrl.Result{}, err
	}
	logger.Info("Current State", "state", currentState)
	statusUpdate, err := r.reconcileResources(ctx, &eva, &currentState, logger)
	if err != nil {
		return ctrl.Result{}, err
	}

	for _, condition := range statusUpdate.Conditions {
		meta.SetStatusCondition(&eva.Status.Conditions, condition)
	}
	eva.Status.Phase = statusUpdate.Phase
	eva.Status.ObservedGeneration = eva.Generation
	if err := r.Status().Update(ctx, &eva); err != nil {
		return ctrl.Result{}, err
	}
	logger.Info("Reconciliation complete", "statusUpdate", statusUpdate)

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

func (r *EvaReconciler) handleDelete(ctx context.Context, eva *v1alpha1.Eva, logger logr.Logger) (ctrl.Result, error) {
	logger.Info("Removing finalizer", "finalizer", evaFinalizer)
	if controllerutil.ContainsFinalizer(eva, evaFinalizer) {
		controllerutil.RemoveFinalizer(eva, evaFinalizer)
		
		if err := r.Update(ctx, eva); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func (r *EvaReconciler) reconcileResources(ctx context.Context, eva *v1alpha1.Eva, currentState *evaCurrentState, logger logr.Logger) (*v1alpha1.EvaStatus, error) {
	statusUpdate := &v1alpha1.EvaStatus{}
	if !currentState.Job.Exists {
		statusUpdate.Phase = v1alpha1.EvaPhasePending
		statusUpdate.Conditions = []metav1.Condition{
			{
				Type:               string(v1alpha1.EvaConditionAvailable),
				Status:             metav1.ConditionFalse,
				Reason:             "JobNotCreated",
				Message:            "The Job has not been created yet.",
				ObservedGeneration: eva.Generation,
			},
		}
	} else {
		statusUpdate.Phase = v1alpha1.EvaPhaseRunning
		statusUpdate.Conditions = []metav1.Condition{
			{
				Type:               string(v1alpha1.EvaConditionAvailable),
				Status:             metav1.ConditionTrue,
				Reason:             "JobRunning",
				Message:            "The Job is running.",
				ObservedGeneration: eva.Generation,
			},
		}
	}
	return statusUpdate, nil
}

func (r *EvaReconciler) getCurrentState(ctx context.Context, eva *v1alpha1.Eva, logger logr.Logger) (evaCurrentState, error) {
	var err error
	currentState := evaCurrentState{}

	currentState.Job, err = r.getJobState(ctx, eva, logger)
	if err != nil {
		return currentState, err
	}
	return currentState, nil
}

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
	return jobState, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EvaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := common.SetupOwnerIndexes(mgr, "Eva", map[client.Object]string{
		&kbatch.Job{}:        ownerKey,
		&corev1.Service{}:    ownerKey,
		&appsv1.Deployment{}: ownerKey,
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Eva{}).
		Owns(&appsv1.Deployment{}).
		Owns(&kbatch.Job{}).
		Owns(&corev1.Service{}).
		Named("eva").
		Complete(r)
}
