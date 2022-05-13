// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	healthcheckconfig "github.com/gardener/gardener/extensions/pkg/controller/healthcheck/config"

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
	// DefaultPeriod is the default period for jobs running in the agent.
	DefaultPeriod *metav1.Duration

	// PSPDisabled is a flag to disable pod security policy.
	PSPDisabled *bool

	// PingEnabled is a flag if ICMP ping checks should be performed.
	PingEnabled *bool
}
