package common

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	kbatch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetOwnedJob(ctx context.Context, c client.Client, owner client.Object, ownerKey string) (*kbatch.Job, error) {
	jobList := &kbatch.JobList{}
	if err := c.List(ctx, jobList,
		client.InNamespace(owner.GetNamespace()),
		client.MatchingFields{ownerKey: string(owner.GetUID())}); err != nil {
		return nil, err
	}
	if len(jobList.Items) == 0 {
		return nil, nil
	}
	return &jobList.Items[0], nil
}

func GetOwnedService(ctx context.Context, c client.Client, owner client.Object, ownerKey string) (*corev1.Service, error) {
	svcList := &corev1.ServiceList{}
	if err := c.List(ctx, svcList,
		client.InNamespace(owner.GetNamespace()),
		client.MatchingFields{ownerKey: string(owner.GetUID())}); err != nil {
		return nil, err
	}
	if len(svcList.Items) == 0 {
		return nil, nil
	}
	return &svcList.Items[0], nil
}

func GetOwnedDeployment(ctx context.Context, c client.Client, owner client.Object, ownerKey string) (*appsv1.Deployment, error) {
	depList := &appsv1.DeploymentList{}
	if err := c.List(ctx, depList,
		client.InNamespace(owner.GetNamespace()),
		client.MatchingFields{ownerKey: string(owner.GetUID())}); err != nil {
		return nil, err
	}
	if len(depList.Items) == 0 {
		return nil, nil
	}
	return &depList.Items[0], nil
}
