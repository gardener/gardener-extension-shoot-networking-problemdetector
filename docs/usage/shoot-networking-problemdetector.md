# Shoot Networking Problem Detector Extension

## Introduction

Within a shoot cluster, it is possible to enable the network problem detector. It is necessary that the Gardener installation your shoot cluster runs in is equipped with a `shoot-networking-problemdetector` extension. Please ask your Gardener operator if the extension is available in your environment.

## Shoot Feature Gate

In most of the Gardener setups the `shoot-networking-problemdetector` extension is not enabled globally and thus must be configured per shoot cluster. Please adapt the shoot specification by the configuration shown below to activate the extension individually.

```yaml
kind: Shoot
...
spec:
  extensions:
    - type: shoot-networking-problemdetector
...
```

## Opt-out

If the shoot network problem detector is globally enabled by default, it can be disabled per shoot. To disable the service for a shoot, the shoot manifest must explicitly state it.

```yaml
apiVersion: core.gardener.cloud/v1beta1
kind: Shoot
...
spec:
  extensions:
    - type: shoot-networking-problemdetector
      disabled: true
...
```

## Shoot-level Configuration (`providerConfig`)

Per-shoot behaviour can be tuned by adding a `providerConfig` to the extension entry. All fields are optional.

```yaml
apiVersion: core.gardener.cloud/v1beta1
kind: Shoot
...
spec:
  extensions:
    - type: shoot-networking-problemdetector
      providerConfig:
        apiVersion: shoot-networking-problemdetector.extensions.config.gardener.cloud/v1alpha1
        kind: NetworkProblemDetectorConfig
        pingEnabled: false
        independentProbes:
          - jobID: check-registry
            protocol: TCP
            host: registry.example.com
            port: 443
          - jobID: ping-gateway
            protocol: Ping
            ipAddress: 192.0.2.1
```

### `pingEnabled`

| Field | Type | Default |
|---|---|---|
| `pingEnabled` | `bool` | operator default |

Enables or disables ICMP ping checks between nodes. When omitted the value configured by the operator is used.

### `independentProbes`

A list of additional network probes that run independently of the shoot cluster topology. Each probe is added as a job to both the host-network and pod-network agents.

| Field | Type | Required | Description |
|---|---|---|---|
| `jobID` | `string` | yes | Unique identifier for the probe job. Must be unique within the list. |
| `protocol` | `string` | yes | Probe protocol: `TCP`, `HTTPS`, or `Ping`. |
| `host` | `string` | see below | Target hostname, used as the endpoint label. |
| `ipAddress` | `string` | see below | Target IP address. Must be a valid IPv4 or IPv6 address. |
| `port` | `int` | TCP/HTTPS | Target port (1–65535). Not used for `Ping`. |
| `period` | `duration` | no | Override the default check interval (e.g. `30s`, `2m`). |

**Protocol requirements:**

| Protocol | `host` | `ipAddress` | `port` |
|---|---|---|---|
| `TCP` | optional when `ipAddress` is set | optional | required |
| `HTTPS` | required | ignored | required |
| `Ping` | optional (falls back to `ipAddress` as label) | required | ignored |

**Protocol behaviour:**

- **TCP** — opens a TCP connection to the target. The agent job argument is `checkTCPPort --endpoints <host>:<ip>:<port>`. When `host` is omitted, `ipAddress` is used as both the connection target and the label.
- **HTTPS** — performs an HTTPS GET (TLS without certificate verification). The agent job argument is `checkHTTPSGet --endpoints <host>:<port>`.
- **Ping** — sends a single ICMP ping to `ipAddress`. The agent job argument is `pingHost --hosts <host>:<ipAddress>`. When `host` is omitted, `ipAddress` is used as the label.

**Examples:**

```yaml
independentProbes:
  # TCP: hostname resolves at runtime
  - jobID: check-registry
    protocol: TCP
    host: registry.example.com
    port: 443

  # TCP: fixed IP, hostname kept as label
  - jobID: check-registry-fixed-ip
    protocol: TCP
    host: registry.example.com
    ipAddress: 192.0.2.10
    port: 443

  # TCP: IP only, no DNS involved
  - jobID: check-internal-endpoint
    protocol: TCP
    ipAddress: 10.0.0.5
    port: 8080

  # HTTPS: checks TLS connectivity
  - jobID: check-api
    protocol: HTTPS
    host: api.example.com
    port: 443

  # Ping: ICMP reachability with explicit label
  - jobID: ping-gateway
    protocol: Ping
    host: gateway.example.com
    ipAddress: 192.0.2.1

  # Ping: ICMP reachability, IP used as label
  - jobID: ping-ip
    protocol: Ping
    ipAddress: 10.0.0.1

  # Custom period
  - jobID: slow-check
    protocol: TCP
    host: slow.example.com
    port: 9090
    period: 60s
```

## Operator Configuration

Operators configure the extension globally via the `Configuration` resource supplied to the controller at startup (typically through the `--config` flag or a mounted `ConfigMap`).

```yaml
apiVersion: shoot-networking-problemdetector.extensions.config.gardener.cloud/v1alpha1
kind: Configuration
networkProblemDetector:
  defaultPeriod: 30s
  maxPeerNodes: 10
  pingEnabled: true
  k8sExporter:
    enabled: true
    heartbeatPeriod: 1m
    minFailingPeerNodeShare: 0.3
  independentProbes:
    - jobID: check-seed-registry
      protocol: TCP
      host: registry.example.com
      port: 443
```

### `networkProblemDetector`

| Field | Type | Default | Description |
|---|---|---|---|
| `defaultPeriod` | `duration` | agent default | Default interval for all check jobs (e.g. `30s`). |
| `maxPeerNodes` | `int` | agent default | Maximum number of peer nodes each agent checks. |
| `pingEnabled` | `bool` | `false` | Enable ICMP ping checks between nodes. Can be overridden per shoot. |
| `k8sExporter` | object | disabled | Configures node condition reporting (see below). |
| `independentProbes` | list | — | Global independent probes added to every shoot. Same schema as the shoot-level probes above; merged with any shoot-level probes. |

### `networkProblemDetector.k8sExporter`

| Field | Type | Default | Description |
|---|---|---|---|
| `enabled` | `bool` | `false` | Activates the Kubernetes exporter, which patches node conditions and creates events. |
| `heartbeatPeriod` | `duration` | — | How often node conditions are updated. |
| `minFailingPeerNodeShare` | `float` | — | Minimum fraction of failing peer nodes [0.0–1.0] before `ClusterNetworkProblems` or `HostNetworkProblems` node conditions are reported. |

