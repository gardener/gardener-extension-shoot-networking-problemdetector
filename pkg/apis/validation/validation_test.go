// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validation_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/apis/config"
	"github.com/gardener/gardener-extension-shoot-networking-problemdetector/pkg/apis/validation"
)

var _ = Describe("ValidateAdditionalProbes", func() {
	// validTCPProbe returns a minimal valid TCP probe.
	validTCPProbe := func() config.AdditionalProbe {
		return config.AdditionalProbe{
			JobID:    "test-probe",
			Host:     "example.com",
			Protocol: config.ProbeProtocolTCP,
			Port:     443,
		}
	}

	It("accepts a valid TCP probe", func() {
		Expect(validation.ValidateAdditionalProbes([]config.AdditionalProbe{validTCPProbe()})).To(Succeed())
	})

	It("accepts a valid HTTPS probe", func() {
		probe := validTCPProbe()
		probe.Protocol = config.ProbeProtocolHTTPS
		Expect(validation.ValidateAdditionalProbes([]config.AdditionalProbe{probe})).To(Succeed())
	})

	It("accepts a valid ICMP probe with IP address", func() {
		probe := config.AdditionalProbe{
			JobID:    "icmp-probe",
			Host:     "192.0.2.10",
			Protocol: config.ProbeProtocolICMP,
		}
		Expect(validation.ValidateAdditionalProbes([]config.AdditionalProbe{probe})).To(Succeed())
	})

	It("accepts a valid ICMP probe with hostname", func() {
		probe := config.AdditionalProbe{
			JobID:    "icmp-probe",
			Host:     "gateway.example.com",
			Protocol: config.ProbeProtocolICMP,
		}
		Expect(validation.ValidateAdditionalProbes([]config.AdditionalProbe{probe})).To(Succeed())
	})

	Describe("jobID validation", func() {
		It("rejects an empty jobID", func() {
			probe := validTCPProbe()
			probe.JobID = ""
			Expect(validation.ValidateAdditionalProbes([]config.AdditionalProbe{probe})).To(MatchError(ContainSubstring("empty jobID")))
		})

		DescribeTable("rejects jobIDs with surrounding whitespace",
			func(jobID string) {
				probe := validTCPProbe()
				probe.JobID = jobID
				Expect(validation.ValidateAdditionalProbes([]config.AdditionalProbe{probe})).
					To(MatchError(ContainSubstring("leading or trailing whitespace")))
			},
			Entry("whitespace only", "   "),
			Entry("leading space", " myid"),
			Entry("trailing space", "myid "),
			Entry("both", " myid "),
		)

		It("rejects duplicate jobIDs", func() {
			p1 := validTCPProbe()
			p2 := validTCPProbe()
			Expect(validation.ValidateAdditionalProbes([]config.AdditionalProbe{p1, p2})).To(MatchError(ContainSubstring("duplicate")))
		})
	})

	Describe("host validation", func() {
		It("rejects a probe with empty host", func() {
			probe := validTCPProbe()
			probe.Host = ""
			Expect(validation.ValidateAdditionalProbes([]config.AdditionalProbe{probe})).
				To(MatchError(ContainSubstring("must have host")))
		})

		It("accepts an IP address as host", func() {
			probe := validTCPProbe()
			probe.Host = "192.0.2.10"
			Expect(validation.ValidateAdditionalProbes([]config.AdditionalProbe{probe})).To(Succeed())
		})

		It("accepts an IPv6 address as host", func() {
			probe := validTCPProbe()
			probe.Host = "::1"
			Expect(validation.ValidateAdditionalProbes([]config.AdditionalProbe{probe})).To(Succeed())
		})

		DescribeTable("rejects invalid host values",
			func(host string) {
				probe := validTCPProbe()
				probe.Host = host
				Expect(validation.ValidateAdditionalProbes([]config.AdditionalProbe{probe})).
					To(MatchError(ContainSubstring("must be a valid hostname or IP address")))
			},
			Entry("contains spaces", "not valid"),
			Entry("contains underscore", "not_valid.example.com"),
			Entry("starts with hyphen", "-invalid.com"),
			Entry("ends with hyphen", "invalid-.com"),
		)

		DescribeTable("accepts valid hostnames",
			func(host string) {
				probe := validTCPProbe()
				probe.Host = host
				Expect(validation.ValidateAdditionalProbes([]config.AdditionalProbe{probe})).To(Succeed())
			},
			Entry("simple domain", "example.com"),
			Entry("subdomain", "api.example.com"),
			Entry("single label", "localhost"),
			Entry("with digits", "node1.example.com"),
		)
	})

	Describe("port validation", func() {
		DescribeTable("rejects out-of-range ports",
			func(port int) {
				probe := validTCPProbe()
				probe.Port = port
				Expect(validation.ValidateAdditionalProbes([]config.AdditionalProbe{probe})).
					To(MatchError(ContainSubstring("invalid port")))
			},
			Entry("zero", 0),
			Entry("negative", -1),
			Entry("too large", 65536),
		)
	})

	Describe("period validation", func() {
		It("rejects a period shorter than 10s", func() {
			probe := validTCPProbe()
			probe.Period = &metav1.Duration{Duration: 5 * time.Second}
			Expect(validation.ValidateAdditionalProbes([]config.AdditionalProbe{probe})).
				To(MatchError(ContainSubstring("must be at least 10s")))
		})

		It("accepts a period of exactly 10s", func() {
			probe := validTCPProbe()
			probe.Period = &metav1.Duration{Duration: 10 * time.Second}
			Expect(validation.ValidateAdditionalProbes([]config.AdditionalProbe{probe})).To(Succeed())
		})

		It("accepts a nil period", func() {
			probe := validTCPProbe()
			probe.Period = nil
			Expect(validation.ValidateAdditionalProbes([]config.AdditionalProbe{probe})).To(Succeed())
		})
	})

	Describe("protocol validation", func() {
		It("rejects an unsupported protocol", func() {
			probe := validTCPProbe()
			probe.Protocol = "UDP"
			Expect(validation.ValidateAdditionalProbes([]config.AdditionalProbe{probe})).
				To(MatchError(ContainSubstring("unsupported protocol")))
		})

		It("rejects HTTPS probe without host", func() {
			probe := validTCPProbe()
			probe.Protocol = config.ProbeProtocolHTTPS
			probe.Host = ""
			Expect(validation.ValidateAdditionalProbes([]config.AdditionalProbe{probe})).
				To(MatchError(ContainSubstring("must have host")))
		})

		It("rejects ICMP probe without host", func() {
			probe := config.AdditionalProbe{
				JobID:    "icmp-bad",
				Protocol: config.ProbeProtocolICMP,
			}
			Expect(validation.ValidateAdditionalProbes([]config.AdditionalProbe{probe})).
				To(MatchError(ContainSubstring("must have host")))
		})
	})
})
