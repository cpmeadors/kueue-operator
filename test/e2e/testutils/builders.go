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

package testutils

import (
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

type TestResourceBuilder struct {
	namespace      string
	queueName      string
	containerImage string
}

func NewTestResourceBuilder(namespace, queueName string) *TestResourceBuilder {
	return &TestResourceBuilder{
		namespace:      namespace,
		queueName:      queueName,
		containerImage: GetContainerImageForWorkloads(),
	}
}

func (b *TestResourceBuilder) NewPod() *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "test-pod-",
			Namespace:    b.namespace,
			Annotations: map[string]string{
				"kueue.x-k8s.io/queue-name": b.queueName,
			},
		},
		Spec: corev1.PodSpec{
			SecurityContext: &corev1.PodSecurityContext{
				RunAsNonRoot: ptr.To(true),
				SeccompProfile: &corev1.SeccompProfile{
					Type: corev1.SeccompProfileTypeRuntimeDefault,
				},
			},
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:    "test-container",
					Image:   b.containerImage,
					Command: []string{"sh", "-c", "echo Hello Kueue; sleep 3600"},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("200m"),
							corev1.ResourceMemory: resource.MustParse("512Mi"),
						},
					},
					SecurityContext: &corev1.SecurityContext{
						AllowPrivilegeEscalation: ptr.To(false),
						Capabilities: &corev1.Capabilities{
							Drop: []corev1.Capability{"ALL"},
						},
					},
				},
			},
		},
	}
}

func (b *TestResourceBuilder) NewJob() *batchv1.Job {
	labels := map[string]string{}
	if b.namespace == "kueue-managed-test" {
		labels["kueue.x-k8s.io/queue-name"] = b.queueName
	}
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "test-job-",
			Namespace:    b.namespace,
			Labels:       labels,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: *b.NewPod().Spec.DeepCopy(),
			},
		},
	}
}

func (b *TestResourceBuilder) NewStatefulSet() *appsv1.StatefulSet {
	var replicas int32 = 1
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "test-ss-",
			Namespace:    b.namespace,
			Annotations: map[string]string{
				"kueue.x-k8s.io/queue-name": b.queueName,
			},
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: ptr.To(replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test-statefulset"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "test-statefulset"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "test-container",
							Image:   b.containerImage,
							Command: []string{"sh", "-c", "sleep 3600"},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("200m"),
									corev1.ResourceMemory: resource.MustParse("512Mi"),
								},
							},
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: ptr.To(false),
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{"ALL"},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (b *TestResourceBuilder) NewDeployment() *appsv1.Deployment {
	var replicas int32 = 1
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "test-deploy-",
			Namespace:    b.namespace,
			Annotations: map[string]string{
				"kueue.x-k8s.io/queue-name": b.queueName,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test-deployment"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "test-deployment"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "test-container",
							Image:   b.containerImage,
							Command: []string{"sh", "-c", "sleep 3600"},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("200m"),
									corev1.ResourceMemory: resource.MustParse("512Mi"),
								},
							},
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: ptr.To(false),
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{"ALL"},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (b *TestResourceBuilder) NewPodWithoutQueue() *corev1.Pod {
	pod := b.NewPod()
	delete(pod.Annotations, "kueue.x-k8s.io/queue-name")
	return pod
}

func (b *TestResourceBuilder) NewJobWithoutQueue() *batchv1.Job {
	job := b.NewJob()
	if job.Labels != nil {
		delete(job.Labels, "kueue.x-k8s.io/queue-name")
	}
	return job
}

func (b *TestResourceBuilder) NewDeploymentWithoutQueue() *appsv1.Deployment {
	deploy := b.NewDeployment()
	delete(deploy.Annotations, "kueue.x-k8s.io/queue-name")
	return deploy
}

func (b *TestResourceBuilder) NewStatefulSetWithoutQueue() *appsv1.StatefulSet {
	ss := b.NewStatefulSet()
	delete(ss.Annotations, "kueue.x-k8s.io/queue-name")
	return ss
}
