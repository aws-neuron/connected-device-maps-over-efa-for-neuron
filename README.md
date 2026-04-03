## Connected Device Maps over EFA for Neuron

A Go library that provides mappings between AWS Neuron devices and their corresponding EFA (Elastic Fabric Adapter) devices. These mappings enable aligned device allocation for Neuron + EFA workloads on supported EC2 instance types.

### Background

When allocating Neuron and EFA devices to a pod, the devices must be aligned — each Neuron device must be paired with its corresponding EFA device for optimal performance. Standard PCI root-based matching does not produce optimal allocations for Trainium instances.

This library provides pre-computed mappings from Neuron device IDs to EFA Bus:Device:Function (BDF) identifiers, enabling correct aligned allocation.

### Installation

```bash
go get github.com/aws-neuron/connected-device-maps-over-efa-for-neuron
```

### Usage

```go
import (
    "github.com/aws-neuron/connected-device-maps-over-efa-for-neuron/lookup"
)
```

Get the device group IDs for an EFA device:
```go
deviceGroupIDs, err := lookup.GetEFADeviceGroupIDs(deviceID)
```

Get the device group IDs for a Neuron device:
```go
deviceGroupIDs, err := lookup.GetNeuronDeviceGroupIDs(deviceID)
```

### Supported Instance Types

See `lookup/config/efa-enabled-instances.json` for the full list of supported EFA-enabled Neuron instance types.

### Adding Support for New Instance Types

To add support for a new EFA-enabled Neuron instance type, update `lookup/config/efa-enabled-instances.json`.

### Adding Support for New Instance Families

Mappings from Neuron device IDs to EFA BDFs are stored as JSON files in the `lookup/neuron` directory. These mappings are constructed from the Neuron collectives topology definitions and the Neuron driver device mappings.

`lookup/config/instance-family-mappings.json` maps instance families to their corresponding mapping files. Update this file when adding a new instance family.

## Security

See [CONTRIBUTING](CONTRIBUTING.md#security-issue-notifications) for more information.

## License

This project is licensed under the Apache-2.0 License. See the [LICENSE](LICENSE) file.
