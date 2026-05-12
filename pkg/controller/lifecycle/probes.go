// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package lifecycle

import (
	"errors"
	"fmt"
	"strings"

	nwpdconfig "github.com/gardener/network-problem-detector/pkg/common/config"
	"k8s.io/apimachinery/pkg/runtime"
	sigsjson "sigs.k8s.io/json"

	"github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/apis/config"
	configv1alpha1 "github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/apis/validation"
)

// decodedShootConfig holds the decoded per-shoot configuration.
type decodedShootConfig struct {
	IcmpEnabled      *bool
	AdditionalProbes []config.AdditionalProbe
}

// decodeShootProviderConfig decodes per-shoot configuration from an Extension's ProviderConfig.
// Returns a zero-value struct if ProviderConfig is nil or empty.
func decodeShootProviderConfig(rawExt *runtime.RawExtension) (decodedShootConfig, error) {
	if rawExt == nil || len(rawExt.Raw) == 0 {
		return decodedShootConfig{}, nil
	}
	var cfg configv1alpha1.ShootProviderConfig
	strictErrs, err := sigsjson.UnmarshalStrict(rawExt.Raw, &cfg)
	if err != nil || len(strictErrs) > 0 {
		return decodedShootConfig{}, fmt.Errorf("failed to decode shoot provider config: %w", errors.Join(append(strictErrs, err)...))
	}
	configv1alpha1.SetDefaults_ShootProviderConfig(&cfg)
	probes := make([]config.AdditionalProbe, len(cfg.AdditionalProbes))
	for i, p := range cfg.AdditionalProbes {
		probes[i] = config.AdditionalProbe{
			JobID:    p.JobID,
			Protocol: config.ProbeProtocol(p.Protocol),
			Host:     p.Host,
			Port:     p.Port,
			Period:   p.Period,
		}
	}
	return decodedShootConfig{
		IcmpEnabled:      cfg.IcmpEnabled,
		AdditionalProbes: probes,
	}, nil
}

// addAdditionalProbeJobs appends jobs for additional probes to both HostNetwork and PodNetwork
// configurations of the agent. Jobs are added with prefixed IDs: "indep-n-<jobID>" for host network
// and "indep-p-<jobID>" for pod network.
func addAdditionalProbeJobs(agentConfig *nwpdconfig.AgentConfig, probes []config.AdditionalProbe) error {
	if len(probes) == 0 {
		return nil
	}

	if err := validation.ValidateAdditionalProbes(probes); err != nil {
		return err
	}

	for _, probe := range probes {
		hostArgs, podArgs, err := buildProbeArgs(probe)
		if err != nil {
			return err
		}

		protoPrefix := strings.ToLower(string(probe.Protocol))
		hostJob := nwpdconfig.Job{
			JobID: fmt.Sprintf("indep-%s-n-%s", protoPrefix, probe.JobID),
			Args:  hostArgs,
		}
		podJob := nwpdconfig.Job{
			JobID: fmt.Sprintf("indep-%s-p-%s", protoPrefix, probe.JobID),
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
func buildProbeArgs(probe config.AdditionalProbe) (hostArgs, podArgs []string, err error) {
	var args []string
	switch probe.Protocol {
	case config.ProbeProtocolTCP:
		args = []string{"checkTCPPort", "--endpoints", fmt.Sprintf("%s:%s:%d", probe.Host, probe.Host, probe.Port)}
	case config.ProbeProtocolHTTPS:
		args = []string{"checkHTTPSGet", "--endpoints", fmt.Sprintf("%s:%d", probe.Host, probe.Port)}
	case config.ProbeProtocolICMP:
		// pingHost --hosts <host>:<host>
		args = []string{"pingHost", "--hosts", fmt.Sprintf("%s:%s", probe.Host, probe.Host)}
	default:
		return nil, nil, fmt.Errorf("unsupported probe protocol %q for jobID %q", probe.Protocol, probe.JobID)
	}

	if probe.Period != nil {
		args = append(args, "--period", probe.Period.Duration.String())
	} else {
		args = append(args, "--period", "60s")
	}

	return args, args, nil
}
