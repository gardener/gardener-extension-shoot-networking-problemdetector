// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"fmt"
	"net"
	"strings"
	"time"

	k8svalidation "k8s.io/apimachinery/pkg/util/validation"

	"github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/apis/config"
)

// ValidateAdditionalProbes checks that all probes are valid and that JobIDs are unique.
func ValidateAdditionalProbes(probes []config.AdditionalProbe) error {
	seen := make(map[string]struct{}, len(probes))
	for _, probe := range probes {
		if probe.JobID == "" {
			return fmt.Errorf("additional probe has empty jobID")
		}
		if probe.JobID != strings.TrimSpace(probe.JobID) {
			return fmt.Errorf("additional probe jobID %q must not have leading or trailing whitespace", probe.JobID)
		}
		if _, exists := seen[probe.JobID]; exists {
			return fmt.Errorf("duplicate additional probe jobID %q", probe.JobID)
		}
		seen[probe.JobID] = struct{}{}

		if strings.TrimSpace(probe.Host) == "" {
			return fmt.Errorf("additional probe %q must have host", probe.JobID)
		}
		// host can be either a valid hostname or a valid IP address
		if net.ParseIP(probe.Host) == nil && len(k8svalidation.IsDNS1123Subdomain(probe.Host)) > 0 {
			return fmt.Errorf("additional probe %q has invalid host %q: must be a valid hostname or IP address", probe.JobID, probe.Host)
		}

		switch probe.Protocol {
		case config.ProbeProtocolTCP:
			if probe.Port < 1 || probe.Port > 65535 {
				return fmt.Errorf("additional probe %q has invalid port %d: must be in range [1, 65535]", probe.JobID, probe.Port)
			}
		case config.ProbeProtocolHTTPS:
			if probe.Port < 1 || probe.Port > 65535 {
				return fmt.Errorf("additional probe %q has invalid port %d: must be in range [1, 65535]", probe.JobID, probe.Port)
			}
		case config.ProbeProtocolICMP:
			// host is already validated above
		default:
			return fmt.Errorf("additional probe %q has unsupported protocol %q: must be TCP, HTTPS, or ICMP", probe.JobID, probe.Protocol)
		}
		if probe.Period != nil && probe.Period.Duration < 10*time.Second {
			return fmt.Errorf("additional probe %q has invalid period %s: must be at least 10s", probe.JobID, probe.Period.Duration)
		}
	}
	return nil
}
