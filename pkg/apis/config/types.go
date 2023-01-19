// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	healthcheckconfig "github.com/gardener/gardener/extensions/pkg/apis/config"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Configuration contains information about the network problem detector configuration.
type Configuration struct {
	metav1.TypeMeta

	// NetworkProblemDetector contains the configuration for the network problem detector
	NetworkProblemDetector *NetworkProblemDetector

	// HealthCheckConfig is the config for the health check controller.
	HealthCheckConfig *healthcheckconfig.HealthCheckConfig
}

// NetworkProblemDetector contains the configuration for the network problem detector.
type NetworkProblemDetector struct {
	// DefaultPeriod optionally overrides the default period for jobs running in the agent.
	DefaultPeriod *metav1.Duration

	// MaxPeerNodes optionally overrides the MaxPeerNodes in the agent config (maximum number of is the default period for jobs running in the agent.
	MaxPeerNodes *int

	// PSPDisabled is a flag to disable pod security policy.
	PSPDisabled *bool

	// PingEnabled is a flag if ICMP ping checks should be performed.
	PingEnabled *bool

	// K8sExporter configures the K8s exporter for updating node conditions and creating events.
	K8sExporter *K8sExporter
}

// K8sExporter contains information for the K8s exporter which patches the node conditions periodically if enabled.
type K8sExporter struct {
	// Enabled if true, the K8s exporter is active
	Enabled bool
	// HeartbeatPeriod defines the update frequency of the node conditions.
	HeartbeatPeriod *metav1.Duration
}
