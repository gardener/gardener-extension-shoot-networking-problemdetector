# Register Shoot Networking Filter Extension in Shoot Clusters

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

