package common

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func IndexFieldByOwner(
	mgr ctrl.Manager,
	obj client.Object,
	ownerKey string,
	ownerKind string,
) error {
	return mgr.GetFieldIndexer().IndexField(
		context.Background(),
		obj,
		ownerKey,
		func(rawObj client.Object) []string {
			owner := metav1.GetControllerOf(rawObj)
			if owner == nil {
				return nil
			}
			if owner.Kind != ownerKind {
				return nil
			}
			return []string{string(owner.UID)}
		},
	)
}

func SetupOwnerIndexes(mgr ctrl.Manager, ownerKind string, indexes map[client.Object]string) error {
	for obj, ownerKey := range indexes {
		if err := IndexFieldByOwner(mgr, obj, ownerKey, ownerKind); err != nil {
			return err
		}
	}
	return nil
}
