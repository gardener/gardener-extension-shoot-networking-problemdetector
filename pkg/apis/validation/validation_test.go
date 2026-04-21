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

var _ = Describe("ValidateIndependentProbes", func() {
	// validTCPProbe returns a minimal valid TCP probe.
	validTCPProbe := func() config.IndependentProbe {
		return config.IndependentProbe{
			JobID:    "test-probe",
			Host:     "example.com",
			Protocol: config.ProbeProtocolTCP,
			Port:     443,
		}
	}

	It("accepts a valid TCP probe", func() {
		Expect(validation.ValidateIndependentProbes([]config.IndependentProbe{validTCPProbe()})).To(Succeed())
	})

	It("accepts a valid HTTPS probe", func() {
		probe := validTCPProbe()
		probe.Protocol = config.ProbeProtocolHTTPS
		Expect(validation.ValidateIndependentProbes([]config.IndependentProbe{probe})).To(Succeed())
	})

	It("accepts a valid Ping probe", func() {
		probe := config.IndependentProbe{
			JobID:     "ping-probe",
			IPAddress: "192.0.2.10",
			Protocol:  config.ProbeProtocolPing,
		}
		Expect(validation.ValidateIndependentProbes([]config.IndependentProbe{probe})).To(Succeed())
	})

	Describe("jobID validation", func() {
		It("rejects an empty jobID", func() {
			probe := validTCPProbe()
			probe.JobID = ""
			Expect(validation.ValidateIndependentProbes([]config.IndependentProbe{probe})).To(MatchError(ContainSubstring("empty jobID")))
		})

		DescribeTable("rejects jobIDs with surrounding whitespace",
			func(jobID string) {
				probe := validTCPProbe()
				probe.JobID = jobID
				Expect(validation.ValidateIndependentProbes([]config.IndependentProbe{probe})).
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
			Expect(validation.ValidateIndependentProbes([]config.IndependentProbe{p1, p2})).To(MatchError(ContainSubstring("duplicate")))
		})
	})

	Describe("host validation", func() {
		DescribeTable("rejects IP addresses as host",
			func(ip string) {
				probe := validTCPProbe()
				probe.Host = ip
				Expect(validation.ValidateIndependentProbes([]config.IndependentProbe{probe})).
					To(MatchError(ContainSubstring("must be a hostname, not an IP address")))
			},
			Entry("IPv4", "192.0.2.10"),
			Entry("IPv6", "::1"),
			Entry("IPv4-mapped IPv6", "::ffff:192.0.2.10"),
		)

		DescribeTable("rejects invalid hostnames",
			func(host string) {
				probe := validTCPProbe()
				probe.Host = host
				Expect(validation.ValidateIndependentProbes([]config.IndependentProbe{probe})).
					To(MatchError(ContainSubstring("must be a valid hostname")))
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
				Expect(validation.ValidateIndependentProbes([]config.IndependentProbe{probe})).To(Succeed())
			},
			Entry("simple domain", "example.com"),
			Entry("subdomain", "api.example.com"),
			Entry("single label", "localhost"),
			Entry("with digits", "node1.example.com"),
		)

		It("rejects a probe with neither host nor ipAddress", func() {
			probe := validTCPProbe()
			probe.Host = ""
			Expect(validation.ValidateIndependentProbes([]config.IndependentProbe{probe})).
				To(MatchError(ContainSubstring("must have host or ipAddress")))
		})
	})

	Describe("ipAddress validation", func() {
		It("rejects an invalid ipAddress", func() {
			probe := validTCPProbe()
			probe.Host = ""
			probe.IPAddress = "not-an-ip"
			Expect(validation.ValidateIndependentProbes([]config.IndependentProbe{probe})).
				To(MatchError(ContainSubstring("invalid ipAddress")))
		})

		It("accepts a valid IPv4 address", func() {
			probe := validTCPProbe()
			probe.Host = ""
			probe.IPAddress = "192.0.2.10"
			Expect(validation.ValidateIndependentProbes([]config.IndependentProbe{probe})).To(Succeed())
		})
	})

	Describe("port validation", func() {
		DescribeTable("rejects out-of-range ports",
			func(port int) {
				probe := validTCPProbe()
				probe.Port = port
				Expect(validation.ValidateIndependentProbes([]config.IndependentProbe{probe})).
					To(MatchError(ContainSubstring("invalid port")))
			},
			Entry("zero", 0),
			Entry("negative", -1),
			Entry("too large", 65536),
		)
	})

	Describe("period validation", func() {
		It("rejects a period shorter than 60s", func() {
			probe := validTCPProbe()
			probe.Period = &metav1.Duration{Duration: 30 * time.Second}
			Expect(validation.ValidateIndependentProbes([]config.IndependentProbe{probe})).
				To(MatchError(ContainSubstring("must be at least 60s")))
		})

		It("accepts a period of exactly 60s", func() {
			probe := validTCPProbe()
			probe.Period = &metav1.Duration{Duration: 60 * time.Second}
			Expect(validation.ValidateIndependentProbes([]config.IndependentProbe{probe})).To(Succeed())
		})

		It("accepts a nil period", func() {
			probe := validTCPProbe()
			probe.Period = nil
			Expect(validation.ValidateIndependentProbes([]config.IndependentProbe{probe})).To(Succeed())
		})
	})

	Describe("protocol validation", func() {
		It("rejects an unsupported protocol", func() {
			probe := validTCPProbe()
			probe.Protocol = "UDP"
			Expect(validation.ValidateIndependentProbes([]config.IndependentProbe{probe})).
				To(MatchError(ContainSubstring("unsupported protocol")))
		})

		It("rejects HTTPS probe without host", func() {
			probe := validTCPProbe()
			probe.Protocol = config.ProbeProtocolHTTPS
			probe.Host = ""
			probe.IPAddress = "192.0.2.10"
			Expect(validation.ValidateIndependentProbes([]config.IndependentProbe{probe})).
				To(MatchError(ContainSubstring("requires host")))
		})

		It("rejects Ping probe without ipAddress", func() {
			probe := validTCPProbe()
			probe.Protocol = config.ProbeProtocolPing
			// Host is set but Ping requires ipAddress specifically
			Expect(validation.ValidateIndependentProbes([]config.IndependentProbe{probe})).
				To(MatchError(ContainSubstring("requires ipAddress")))
		})
	})
})
