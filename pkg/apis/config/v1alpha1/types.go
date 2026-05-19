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

// ProbeProtocol defines the protocol for an additional probe.
type ProbeProtocol string

const (
	// ProbeProtocolTCP uses a TCP connection check.
	ProbeProtocolTCP ProbeProtocol = "TCP"
	// ProbeProtocolHTTPS uses an HTTPS GET request (TLS without certificate verification).
	ProbeProtocolHTTPS ProbeProtocol = "HTTPS"
	// ProbeProtocolICMP uses an ICMP ping check.
	ProbeProtocolICMP ProbeProtocol = "ICMP"
)

// AdditionalProbe defines a single network probe that is logically decoupled from the Shoot/Seed cluster topology.
type AdditionalProbe struct {
	// JobID is the unique identifier for this probe job.
	JobID string `json:"jobID"`
	// Protocol is the probe protocol: TCP, HTTPS, or ICMP.
	Protocol ProbeProtocol `json:"protocol"`
	// Host is the target hostname or IP address.
	// Required for all protocols. For TCP and HTTPS, used as the connection target.
	// For ICMP, used as the ping target.
	Host string `json:"host"`
	// Port is the target port (1-65535). Not used for ICMP probes.
	Port int `json:"port"`
	// Period optionally overrides the default check period for this probe.
	// +optional
	Period *metav1.Duration `json:"period,omitempty"`
}

// NetworkProblemDetector contains the configuration for the network problem detector.
type NetworkProblemDetector struct {
	// DefaultPeriod optionally overrides the default period for jobs running in the agent.
	// +optional
	DefaultPeriod *metav1.Duration `json:"defaultPeriod,omitempty"`

	// MaxPeerNodes optionally overrides the MaxPeerNodes in the agent config (maximum number of is the default period for jobs running in the agent.
	// +optional
	MaxPeerNodes *int `json:"maxPeerNodes,omitempty"`

	// IcmpEnabled is a flag if ICMP ping checks should be performed.
	// +optional
	IcmpEnabled *bool `json:"icmpEnabled,omitempty"`

	// K8sExporter configures the K8s exporter for updating node conditions and creating events.
	// +optional
	K8sExporter *K8sExporter `json:"k8sExporter,omitempty"`

	// AdditionalProbes defines additional probes that run independently of the Shoot/Seed cluster topology,
	// enabling infrastructure-level network diagnostics.
	// +optional
	AdditionalProbes []AdditionalProbe `json:"additionalProbes,omitempty"`
}

// ShootProviderConfig is the per-shoot configuration stored in Extension.spec.providerConfig.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ShootProviderConfig struct {
	metav1.TypeMeta `json:",inline"`
	// IcmpEnabled enables ICMP ping checks for this shoot.
	// +optional
	IcmpEnabled *bool `json:"icmpEnabled,omitempty"`
	// AdditionalProbes defines additional probe jobs for this shoot.
	// +optional
	AdditionalProbes []AdditionalProbe `json:"additionalProbes,omitempty"`
}

// K8sExporter contains information for the K8s exporter which patches the node conditions periodically if enabled.
type K8sExporter struct {
	// Enabled if true, the K8s exporter is active
	Enabled bool `json:"enabled"`
	// HeartbeatPeriod defines the update frequency of the node conditions.
	// +optional
	HeartbeatPeriod *metav1.Duration `json:"heartbeatPeriod,omitempty"`
	// MinFailingPeerNodeShare if > 0, reports node conditions `ClusterNetworkProblems` or `HostNetworkProblems` for node checks only if minimum share of destination peer nodes are failing. Valid range: [0.0,1.0]
	// +optional
	MinFailingPeerNodeShare *float64 `json:"minFailingPeerNodeShare,omitempty"`
}
