// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	healthcheckconfigv1alpha1 "github.com/gardener/gardener/extensions/pkg/apis/config/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Configuration contains information about the network problem detector configuration.
type Configuration struct {
	metav1.TypeMeta

	// NetworkProblemDetector contains the configuration for the network problem detector
	NetworkProblemDetector *NetworkProblemDetector

	// HealthCheckConfig is the config for the health check controller.
	HealthCheckConfig *healthcheckconfigv1alpha1.HealthCheckConfig
}

// ProbeProtocol defines the protocol for an independent probe.
type ProbeProtocol string

const (
	// ProbeProtocolTCP uses a TCP connection check.
	ProbeProtocolTCP ProbeProtocol = "TCP"
	// ProbeProtocolHTTPS uses an HTTPS GET request (TLS without certificate verification).
	ProbeProtocolHTTPS ProbeProtocol = "HTTPS"
	// ProbeProtocolPing uses an ICMP ping check. Requires IPAddress to be set.
	ProbeProtocolPing ProbeProtocol = "Ping"
)

// IndependentProbe defines a single network probe that is logically decoupled from the Shoot/Seed cluster topology.
type IndependentProbe struct {
	// JobID is the unique identifier for this probe job.
	JobID string
	// Protocol is the probe protocol: TCP or HTTPS.
	Protocol ProbeProtocol
	// Host is the target hostname used for labeling and HTTPS checks.
	// Optional for TCP probes when IPAddress is set; required for HTTPS probes.
	Host string
	// IPAddress optionally overrides the IP address used for the TCP connection.
	// When set, the TCP check connects to this IP while Host is still used as the endpoint label.
	// Has no effect for HTTPS probes.
	IPAddress string
	// Port is the target port (1-65535).
	Port int
	// Period optionally overrides the default check period for this probe.
	Period *metav1.Duration
}

// NetworkProblemDetector contains the configuration for the network problem detector.
type NetworkProblemDetector struct {
	// DefaultPeriod optionally overrides the default period for jobs running in the agent.
	DefaultPeriod *metav1.Duration

	// MaxPeerNodes optionally overrides the MaxPeerNodes in the agent config (maximum number of is the default period for jobs running in the agent.
	MaxPeerNodes *int

	// PingEnabled is a flag if ICMP ping checks should be performed.
	PingEnabled *bool

	// K8sExporter configures the K8s exporter for updating node conditions and creating events.
	K8sExporter *K8sExporter

	// IndependentProbes defines probes that run independently of the Shoot/Seed cluster topology,
	// enabling infrastructure-level network diagnostics.
	IndependentProbes []IndependentProbe
}

// K8sExporter contains information for the K8s exporter which patches the node conditions periodically if enabled.
type K8sExporter struct {
	// Enabled if true, the K8s exporter is active
	Enabled bool
	// HeartbeatPeriod defines the update frequency of the node conditions.
	HeartbeatPeriod *metav1.Duration
	// MinFailingPeerNodeShare if > 0, reports node conditions `ClusterNetworkProblems` or `HostNetworkProblems` for node checks only if minimum share of destination peer nodes are failing. Valid range: [0.0,1.0]
	MinFailingPeerNodeShare *float64
}
