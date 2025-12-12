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

	appsv1 "k8s.io/api/apps/v1"
	kbatch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	geofrontv1alpha1 "github.com/dayaliuzzo/Smooth-Operator/api/v1alpha1"
	"github.com/dayaliuzzo/Smooth-Operator/internal/controllers/common"
)

const ownerKey = ".metadata.controller"

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
	// TODO(user): your logic here

	return ctrl.Result{}, nil
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
		For(&geofrontv1alpha1.Eva{}).
		Owns(&appsv1.Deployment{}).
		Owns(&kbatch.Job{}).
		Owns(&corev1.Service{}).
		Named("eva").
		Complete(r)
}
