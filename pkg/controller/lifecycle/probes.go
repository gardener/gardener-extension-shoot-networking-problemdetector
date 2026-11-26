// SPDX-FileCopyrightText: Contributors to the Gardener project
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
// configurations of the agent. jobIDPrefix is prepended to each probe's JobID when constructing
// the agent job name; callers must use a distinct prefix per probe source (operator vs. shoot) so
// that equal JobIDs from different sources never collide in the agent's flat job namespace.
// Job IDs take the form "indep-{proto}-n-{jobIDPrefix}{jobID}" / "indep-{proto}-p-{jobIDPrefix}{jobID}".
func addAdditionalProbeJobs(agentConfig *nwpdconfig.AgentConfig, probes []config.AdditionalProbe, jobIDPrefix string) error {
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
			JobID: fmt.Sprintf("indep-%s-n-%s%s", protoPrefix, jobIDPrefix, probe.JobID),
			Args:  hostArgs,
		}
		podJob := nwpdconfig.Job{
			JobID: fmt.Sprintf("indep-%s-p-%s%s", protoPrefix, jobIDPrefix, probe.JobID),
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

// checkCrossSourceJobIDCollisions returns an error if any shoot probe would produce the same
// agent job ID as an operator probe after source prefixes are applied. Because operator probes
// use an empty prefix and shoot probes use "s-", an operator probe with jobID "s-foo" and a
// shoot probe with jobID "foo" both resolve to "indep-{proto}-{n|p}-s-foo".
func checkCrossSourceJobIDCollisions(operatorProbes, shootProbes []config.AdditionalProbe) error {
	if len(operatorProbes) == 0 || len(shootProbes) == 0 {
		return nil
	}
	opIDs := make(map[string]struct{}, len(operatorProbes)*2)
	for _, p := range operatorProbes {
		proto := strings.ToLower(string(p.Protocol))
		opIDs[fmt.Sprintf("indep-%s-n-%s", proto, p.JobID)] = struct{}{}
		opIDs[fmt.Sprintf("indep-%s-p-%s", proto, p.JobID)] = struct{}{}
	}
	for _, p := range shootProbes {
		proto := strings.ToLower(string(p.Protocol))
		for _, id := range []string{
			fmt.Sprintf("indep-%s-n-s-%s", proto, p.JobID),
			fmt.Sprintf("indep-%s-p-s-%s", proto, p.JobID),
		} {
			if _, exists := opIDs[id]; exists {
				return fmt.Errorf("shoot probe %q collides with an operator probe: agent job ID %q would be duplicated (operator probe jobIDs must not start with %q)", p.JobID, id, "s-")
			}
		}
	}
	return nil
}

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
