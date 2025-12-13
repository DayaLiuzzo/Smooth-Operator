package eva

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	kbatch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type (
	JobOption        func(*kbatch.Job)
	ServiceOption    func(*corev1.Service)
	DeploymentOption func(*appsv1.Deployment)
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

func buildJob(name, namespace string, opts ...JobOption) *kbatch.Job {
	job := &kbatch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    make(map[string]string),
		},
		Spec: kbatch.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: make(map[string]string),
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers:    []corev1.Container{{}},
				},
			},
		},
	}

	for _, opt := range opts {
		opt(job)
	}

	return job
}

func buildService(name, namespace string, opts ...ServiceOption) *corev1.Service {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    make(map[string]string),
		},
		Spec: corev1.ServiceSpec{
			Selector: make(map[string]string),
		},
	}

	for _, opt := range opts {
		opt(service)
	}

	return service
}

func buildDeployment(name, namespace string, opts ...DeploymentOption) *appsv1.Deployment {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    make(map[string]string),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: make(map[string]string),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: make(map[string]string),
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{}},
				},
			},
		},
	}

	for _, opt := range opts {
		opt(deployment)
	}

	return deployment
}

// WithJobImage sets the container image
func WithJobImage(image string) JobOption {
	return func(job *kbatch.Job) {
		if len(job.Spec.Template.Spec.Containers) > 0 {
			job.Spec.Template.Spec.Containers[0].Image = image
		}
	}
}

// WithJobCommand sets the container command
func WithJobCommand(command []string) JobOption {
	return func(job *kbatch.Job) {
		if len(job.Spec.Template.Spec.Containers) > 0 && len(command) > 0 {
			job.Spec.Template.Spec.Containers[0].Command = command
		}
	}
}

// WithJobContainerName sets the container name
func WithJobContainerName(name string) JobOption {
	return func(job *kbatch.Job) {
		if len(job.Spec.Template.Spec.Containers) > 0 {
			job.Spec.Template.Spec.Containers[0].Name = name
		}
	}
}

// WithJobLabels sets metadata labels on both the Job and Pod template
func WithJobLabels(labels map[string]string) JobOption {
	return func(job *kbatch.Job) {
		if job.Labels == nil {
			job.Labels = make(map[string]string)
		}
		if job.Spec.Template.Labels == nil {
			job.Spec.Template.Labels = make(map[string]string)
		}

		// Set labels on the Job itself
		for k, v := range labels {
			job.Labels[k] = v
		}
		// Set labels on the Pod template
		for k, v := range labels {
			job.Spec.Template.Labels[k] = v
		}
	}
}

// WithJobImagePullSecret adds an image pull secret
func WithJobImagePullSecret(secretName string) JobOption {
	return func(job *kbatch.Job) {
		if secretName != "" {
			job.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
				{Name: secretName},
			}
		}
	}
}

// WithJobBackoffLimit sets the backoff limit for failed job attempts
func WithJobBackoffLimit(limit int32) JobOption {
	return func(job *kbatch.Job) {
		job.Spec.BackoffLimit = &limit
	}
}

// WithJobTTLSecondsAfterFinished sets TTL for cleanup after completion
func WithJobTTLSecondsAfterFinished(seconds int32) JobOption {
	return func(job *kbatch.Job) {
		job.Spec.TTLSecondsAfterFinished = &seconds
	}
}

// === Service Options ===

// WithServicePort sets the service port
func WithServicePort(port int32) ServiceOption {
	return func(service *corev1.Service) {
		service.Spec.Ports = []corev1.ServicePort{
			{Port: port},
		}
	}
}

// WithServiceLabels adds labels to the service
func WithServiceLabels(labels map[string]string) ServiceOption {
	return func(service *corev1.Service) {
		if service.Labels == nil {
			service.Labels = make(map[string]string)
		}
		for k, v := range labels {
			service.Labels[k] = v
		}
	}
}

// WithServiceSelector sets the service selector
func WithServiceSelector(selector map[string]string) ServiceOption {
	return func(service *corev1.Service) {
		service.Spec.Selector = selector
	}
}

// WithServiceType sets the service type (ClusterIP, NodePort, LoadBalancer, ExternalName)
func WithServiceType(serviceType corev1.ServiceType) ServiceOption {
	return func(service *corev1.Service) {
		service.Spec.Type = serviceType
	}
}

// === Deployment Options ===

// WithDeploymentReplicas sets the number of replicas
func WithDeploymentReplicas(replicas int32) DeploymentOption {
	return func(deployment *appsv1.Deployment) {
		deployment.Spec.Replicas = &replicas
	}
}

// WithDeploymentImage sets the container image
func WithDeploymentImage(image string) DeploymentOption {
	return func(deployment *appsv1.Deployment) {
		deployment.Spec.Template.Spec.Containers[0].Image = image
	}
}

// WithDeploymentContainerName sets the container name
func WithDeploymentContainerName(name string) DeploymentOption {
	return func(deployment *appsv1.Deployment) {
		deployment.Spec.Template.Spec.Containers[0].Name = name
	}
}

// WithDeploymentCommand sets the container command
func WithDeploymentCommand(command []string) DeploymentOption {
	return func(deployment *appsv1.Deployment) {
		if len(command) > 0 {
			deployment.Spec.Template.Spec.Containers[0].Command = command
		}
	}
}

// WithDeploymentSelector sets labels used for pod selection (MatchLabels) - these also get added to pod template
func WithDeploymentSelector(labels map[string]string) DeploymentOption {
	return func(deployment *appsv1.Deployment) {
		if deployment.Spec.Selector.MatchLabels == nil {
			deployment.Spec.Selector.MatchLabels = make(map[string]string)
		}
		if deployment.Spec.Template.Labels == nil {
			deployment.Spec.Template.Labels = make(map[string]string)
		}

		// Selector labels MUST also be on pods for matching to work
		for k, v := range labels {
			deployment.Spec.Selector.MatchLabels[k] = v
			deployment.Spec.Template.Labels[k] = v
		}
	}
}

// WithDeploymentLabels sets metadata labels on deployment and pod template (not used for selection)
func WithDeploymentLabels(labels map[string]string) DeploymentOption {
	return func(deployment *appsv1.Deployment) {
		if deployment.Labels == nil {
			deployment.Labels = make(map[string]string)
		}
		if deployment.Spec.Template.Labels == nil {
			deployment.Spec.Template.Labels = make(map[string]string)
		}

		for k, v := range labels {
			deployment.Labels[k] = v
			deployment.Spec.Template.Labels[k] = v
		}
	}
}

// WithDeploymentImagePullSecret adds a single image pull secret
func WithDeploymentImagePullSecret(secretName string) DeploymentOption {
	return func(deployment *appsv1.Deployment) {
		if secretName != "" {
			deployment.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
				{Name: secretName},
			}
		}
	}
}

// WithDeploymentPort adds a single container port
func WithDeploymentPort(port int32, protocol corev1.Protocol) DeploymentOption {
	return func(deployment *appsv1.Deployment) {
		deployment.Spec.Template.Spec.Containers[0].Ports = []corev1.ContainerPort{
			{
				ContainerPort: port,
				Protocol:      protocol,
			},
		}
	}
}
