/*
Copyright 2024.

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
// Code generated by applyconfiguration-gen. DO NOT EDIT.

package applyconfiguration

import (
	v1alpha1 "github.com/openshift/kueue-operator/pkg/apis/kueueoperator/v1alpha1"
	internal "github.com/openshift/kueue-operator/pkg/generated/applyconfiguration/internal"
	kueueoperatorv1alpha1 "github.com/openshift/kueue-operator/pkg/generated/applyconfiguration/kueueoperator/v1alpha1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	testing "k8s.io/client-go/testing"
)

// ForKind returns an apply configuration type for the given GroupVersionKind, or nil if no
// apply configuration type exists for the given GroupVersionKind.
func ForKind(kind schema.GroupVersionKind) interface{} {
	switch kind {
	// Group=operator.openshift.io, Version=v1alpha1
	case v1alpha1.SchemeGroupVersion.WithKind("ByWorkload"):
		return &kueueoperatorv1alpha1.ByWorkloadApplyConfiguration{}
	case v1alpha1.SchemeGroupVersion.WithKind("ExternalFramework"):
		return &kueueoperatorv1alpha1.ExternalFrameworkApplyConfiguration{}
	case v1alpha1.SchemeGroupVersion.WithKind("GangScheduling"):
		return &kueueoperatorv1alpha1.GangSchedulingApplyConfiguration{}
	case v1alpha1.SchemeGroupVersion.WithKind("Integrations"):
		return &kueueoperatorv1alpha1.IntegrationsApplyConfiguration{}
	case v1alpha1.SchemeGroupVersion.WithKind("Kueue"):
		return &kueueoperatorv1alpha1.KueueApplyConfiguration{}
	case v1alpha1.SchemeGroupVersion.WithKind("KueueConfiguration"):
		return &kueueoperatorv1alpha1.KueueConfigurationApplyConfiguration{}
	case v1alpha1.SchemeGroupVersion.WithKind("KueueOperandSpec"):
		return &kueueoperatorv1alpha1.KueueOperandSpecApplyConfiguration{}
	case v1alpha1.SchemeGroupVersion.WithKind("KueueStatus"):
		return &kueueoperatorv1alpha1.KueueStatusApplyConfiguration{}
	case v1alpha1.SchemeGroupVersion.WithKind("LabelKeys"):
		return &kueueoperatorv1alpha1.LabelKeysApplyConfiguration{}
	case v1alpha1.SchemeGroupVersion.WithKind("Preemption"):
		return &kueueoperatorv1alpha1.PreemptionApplyConfiguration{}
	case v1alpha1.SchemeGroupVersion.WithKind("WorkloadManagement"):
		return &kueueoperatorv1alpha1.WorkloadManagementApplyConfiguration{}

	}
	return nil
}

func NewTypeConverter(scheme *runtime.Scheme) *testing.TypeConverter {
	return &testing.TypeConverter{Scheme: scheme, TypeResolver: internal.Parser()}
}
