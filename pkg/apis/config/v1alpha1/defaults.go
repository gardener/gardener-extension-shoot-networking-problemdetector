// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

// SetDefaults_Configuration sets default values for Configuration objects.
func SetDefaults_Configuration(obj *Configuration) {
	if obj.NetworkProblemDetector != nil {
		setDefaultAdditionalProbePeriods(obj.NetworkProblemDetector.AdditionalProbes)
	}
}

// SetDefaults_ShootProviderConfig sets default values for ShootProviderConfig objects.
func SetDefaults_ShootProviderConfig(obj *ShootProviderConfig) {
	setDefaultAdditionalProbePeriods(obj.AdditionalProbes)
}

func setDefaultAdditionalProbePeriods(probes []AdditionalProbe) {
	for i := range probes {
		if probes[i].Period == nil {
			probes[i].Period = &metav1.Duration{Duration: 60 * time.Second}
		}
	}
}
