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

// ValidateIndependentProbes checks that all probes are valid and that JobIDs are unique.
func ValidateIndependentProbes(probes []config.IndependentProbe) error {
	seen := make(map[string]struct{}, len(probes))
	for _, probe := range probes {
		if probe.JobID == "" {
			return fmt.Errorf("independent probe has empty jobID")
		}
		if probe.JobID != strings.TrimSpace(probe.JobID) {
			return fmt.Errorf("independent probe jobID %q must not have leading or trailing whitespace", probe.JobID)
		}
		if _, exists := seen[probe.JobID]; exists {
			return fmt.Errorf("duplicate independent probe jobID %q", probe.JobID)
		}
		seen[probe.JobID] = struct{}{}

		if strings.TrimSpace(probe.Host) == "" && strings.TrimSpace(probe.IPAddress) == "" {
			return fmt.Errorf("independent probe %q must have host or ipAddress", probe.JobID)
		}
		if probe.Host != "" {
			if net.ParseIP(probe.Host) != nil {
				return fmt.Errorf("independent probe %q has invalid host %q: must be a hostname, not an IP address (use ipAddress field instead)", probe.JobID, probe.Host)
			}
			if len(k8svalidation.IsDNS1123Subdomain(probe.Host)) > 0 {
				return fmt.Errorf("independent probe %q has invalid host %q: must be a valid hostname", probe.JobID, probe.Host)
			}
		}
		if probe.IPAddress != "" && net.ParseIP(probe.IPAddress) == nil {
			return fmt.Errorf("independent probe %q has invalid ipAddress %q: must be a valid IP address", probe.JobID, probe.IPAddress)
		}
		switch probe.Protocol {
		case config.ProbeProtocolTCP:
			if probe.Port < 1 || probe.Port > 65535 {
				return fmt.Errorf("independent probe %q has invalid port %d: must be in range [1, 65535]", probe.JobID, probe.Port)
			}
		case config.ProbeProtocolHTTPS:
			if probe.Port < 1 || probe.Port > 65535 {
				return fmt.Errorf("independent probe %q has invalid port %d: must be in range [1, 65535]", probe.JobID, probe.Port)
			}
			if strings.TrimSpace(probe.Host) == "" {
				return fmt.Errorf("independent probe %q with protocol HTTPS requires host", probe.JobID)
			}
		case config.ProbeProtocolPing:
			if strings.TrimSpace(probe.IPAddress) == "" {
				return fmt.Errorf("independent probe %q with protocol Ping requires ipAddress", probe.JobID)
			}
		default:
			return fmt.Errorf("independent probe %q has unsupported protocol %q: must be TCP, HTTPS, or Ping", probe.JobID, probe.Protocol)
		}
		if probe.Period != nil && probe.Period.Duration < 60*time.Second {
			return fmt.Errorf("independent probe %q has invalid period %s: must be at least 60s", probe.JobID, probe.Period.Duration)
		}
	}
	return nil
}
