// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	healthcheckconfigv1alpha1 "github.com/gardener/gardener/extensions/pkg/controller/healthcheck/config/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Configuration contains information about the network problem detector configuration.
type Configuration struct {
	metav1.TypeMeta `json:",inline"`

	// NetworkProblemDetector contains the configuration for the network problem detector
	// +optional
	NetworkProblemDetector *NetworkProblemDetector `json:"networkProblemDetector,omitempty"`

	// HealthCheckConfig is the config for the health check controller.
	// +optional
	HealthCheckConfig *healthcheckconfigv1alpha1.HealthCheckConfig `json:"healthCheckConfig,omitempty"`
}

// NetworkProblemDetector contains the configuration for the network problem detector.
type NetworkProblemDetector struct {
	// DefaultPeriod is the default period for jobs running in the agent.
	// +optional
	DefaultPeriod *metav1.Duration `json:"defaultPeriod,omitempty"`

	// PSPDisabled is a flag to disable pod security policy.
	// +optional
	PSPDisabled *bool `json:"pspDisabled,omitempty"`

	// PingEnabled is a flag if ICMP ping checks should be performed.
	// +optional
	PingEnabled *bool `json:"pingEnabled,omitempty"`
}
