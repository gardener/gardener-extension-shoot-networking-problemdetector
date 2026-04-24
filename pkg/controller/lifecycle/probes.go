// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package lifecycle

import (
	"encoding/json"
	"fmt"

	nwpdconfig "github.com/gardener/network-problem-detector/pkg/common/config"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/apis/config"
	configv1alpha1 "github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/apis/validation"
)

// decodedShootConfig holds the decoded per-shoot configuration.
type decodedShootConfig struct {
	PingEnabled       *bool
	IndependentProbes []config.IndependentProbe
}

// decodeShootProviderConfig decodes per-shoot configuration from an Extension's ProviderConfig.
// Returns a zero-value struct if ProviderConfig is nil or empty.
func decodeShootProviderConfig(rawExt *runtime.RawExtension) (decodedShootConfig, error) {
	if rawExt == nil || len(rawExt.Raw) == 0 {
		return decodedShootConfig{}, nil
	}
	var cfg configv1alpha1.ShootProviderConfig
	if err := json.Unmarshal(rawExt.Raw, &cfg); err != nil {
		return decodedShootConfig{}, fmt.Errorf("failed to unmarshal shoot provider config: %w", err)
	}
	probes := make([]config.IndependentProbe, len(cfg.IndependentProbes))
	for i, p := range cfg.IndependentProbes {
		probes[i] = config.IndependentProbe{
			JobID:     p.JobID,
			Protocol:  config.ProbeProtocol(p.Protocol),
			Host:      p.Host,
			IPAddress: p.IPAddress,
			Port:      p.Port,
			Period:    p.Period,
		}
	}
	return decodedShootConfig{
		PingEnabled:       cfg.PingEnabled,
		IndependentProbes: probes,
	}, nil
}

// addIndependentProbeJobs appends jobs for independent probes to both HostNetwork and PodNetwork
// configurations of the agent. Jobs are added with prefixed IDs: "indep-h-<jobID>" for host network
// and "indep-p-<jobID>" for pod network.
func addIndependentProbeJobs(agentConfig *nwpdconfig.AgentConfig, probes []config.IndependentProbe) error {
	if len(probes) == 0 {
		return nil
	}

	if err := validation.ValidateIndependentProbes(probes); err != nil {
		return err
	}

	for _, probe := range probes {
		hostArgs, podArgs, err := buildProbeArgs(probe)
		if err != nil {
			return err
		}

		hostJob := nwpdconfig.Job{
			JobID: "indep-h-" + probe.JobID,
			Args:  hostArgs,
		}
		podJob := nwpdconfig.Job{
			JobID: "indep-p-" + probe.JobID,
			Args:  podArgs,
		}

		if agentConfig.HostNetwork != nil {
			agentConfig.HostNetwork.Jobs = append(agentConfig.HostNetwork.Jobs, hostJob)
		}
		if agentConfig.PodNetwork != nil {
			agentConfig.PodNetwork.Jobs = append(agentConfig.PodNetwork.Jobs, podJob)
		}
	}
	return nil
}

// buildProbeArgs returns the job args for the host-network and pod-network jobs for a given probe.
// Both are identical since the probe target is external and independent of the network context.
func buildProbeArgs(probe config.IndependentProbe) (hostArgs, podArgs []string, err error) {
	var args []string
	switch probe.Protocol {
	case config.ProbeProtocolTCP:
		// checkTCPPort --endpoints <hostname>:<ip>:<port>
		// IPAddress overrides the IP to connect to while Host remains the label.
		// Falls back to Host when IPAddress is not set.
		// Falls back to IPAddress as label when Host is not set.
		ip := probe.Host
		if probe.IPAddress != "" {
			ip = probe.IPAddress
		}
		hostname := probe.Host
		if hostname == "" {
			hostname = probe.IPAddress
		}
		endpoint := fmt.Sprintf("%s:%s:%d", hostname, ip, probe.Port)
		args = []string{"checkTCPPort", "--endpoints", endpoint}
	case config.ProbeProtocolHTTPS:
		// checkHTTPSGet --endpoints <hostname>[:<port>]
		// Uses InsecureSkipVerify: true for non-kube-apiserver endpoints.
		endpoint := fmt.Sprintf("%s:%d", probe.Host, probe.Port)
		args = []string{"checkHTTPSGet", "--endpoints", endpoint}
	case config.ProbeProtocolPing:
		// pingHost --hosts <hostname>:<ip>
		// Host is the label; IPAddress is the target (required for ping).
		hostname := probe.Host
		if hostname == "" {
			hostname = probe.IPAddress
		}
		args = []string{"pingHost", "--hosts", fmt.Sprintf("%s:%s", hostname, probe.IPAddress)}
	default:
		return nil, nil, fmt.Errorf("unsupported probe protocol %q for jobID %q", probe.Protocol, probe.JobID)
	}

	if probe.Period != nil {
		args = append(args, "--period", probe.Period.Duration.String())
	}

	return args, append([]string(nil), args...), nil
}
