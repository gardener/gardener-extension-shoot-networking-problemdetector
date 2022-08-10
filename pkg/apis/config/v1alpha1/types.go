// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	healthcheckconfigv1alpha1 "github.com/gardener/gardener/extensions/pkg/apis/config/v1alpha1"

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

	// K8sExporter configures the K8s exporter for updating node conditions and creating events.
	// +optional
	K8sExporter *K8sExporter `json:"k8sExporter,omitempty"`
}

// K8sExporter contains information for the K8s exporter which patches the node conditions periodically if enabled.
type K8sExporter struct {
	// Enabled if true, the K8s exporter is active
	Enabled bool `json:"enabled"`
	// HeartbeatPeriod defines the update frequency of the node conditions.
	// +optional
	HeartbeatPeriod *metav1.Duration `json:"heartbeatPeriod,omitempty"`
}
