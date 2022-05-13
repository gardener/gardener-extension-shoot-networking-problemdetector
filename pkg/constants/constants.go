// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package constants

import "path/filepath"

const (
	// ExtensionType is the name of the extension type.
	ExtensionType = "shoot-networking-problemdetector"
	// ServiceName is the name of the service.
	ServiceName = ExtensionType

	// ApplicationName is the name for resource describing the components deployed by the extension controller.
	ApplicationName = "network-problem-detector"

	// AgentImageName image name for network problem detector agent
	AgentImageName = "network-problem-detector-agent"
	// ControllerImageName image name for network problem detector controller
	ControllerImageName = "network-problem-detector-controller"

	extensionServiceName = "extension-" + ServiceName
	// NamespaceKubeSystem kube-system namespace
	NamespaceKubeSystem = "kube-system"
	// ManagedResourceNamesControllerSeed is the name used to describe the managed seed resources for the controller.
	ManagedResourceNamesControllerSeed = extensionServiceName + "-controller-seed"
	// ManagedResourceNamesControllerShoot is the name used to describe the managed shoot resources for the controller.
	ManagedResourceNamesControllerShoot = extensionServiceName + "-controller-shoot"
	// ManagedResourceNamesAgentShoot is the name used to describe the managed shoot resources for the agents.
	ManagedResourceNamesAgentShoot = extensionServiceName + "-agent-shoot"

	// ShootAccessSecretName is the name of the shoot access secret in the seed.
	ShootAccessSecretName = extensionServiceName
	// ShootAccessServiceAccountName is the name of the service account used for accessing the shoot.
	ShootAccessServiceAccountName = ShootAccessSecretName

	// NetworkProblemDetectorControllerChartNameSeed is the chart name for nwpd controller resources in the seed.
	NetworkProblemDetectorControllerChartNameSeed = "shoot-network-problem-detector-controller-seed"
	// NetworkProblemDetectorControllerChartNameShoot is the chart name for nwpd controller resources in the shoot.
	NetworkProblemDetectorControllerChartNameShoot = "shoot-network-problem-detector-controller-shoot"
)

// ChartsPath is the path to the charts
var ChartsPath = filepath.Join("charts", "internal")
