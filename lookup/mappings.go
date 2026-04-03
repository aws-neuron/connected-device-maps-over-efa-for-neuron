/*
* Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */
package lookup

import (
	"embed"
	_ "embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"

	"k8s.io/klog/v2"
)

//go:embed config/*.json neuron/*.json
var actualConfigFS embed.FS

var (
	efaEnabledDevice       map[string]bool
	instanceFamilyMappings map[string]string
	// EFA BDF -> NeuronDeviceID
	efaToNeuron = make(map[string]string)
	// NeuronDeviceID -> EFA BDF
	neuronToEfa = make(map[string]string)

	deviceTypeMappings = map[DeviceType]map[string]string{
		DeviceTypeEFA:    efaToNeuron,
		DeviceTypeNeuron: neuronToEfa,
	}

	initMappings = initMappingsImpl
	initError    error
	configFS     fs.FS = actualConfigFS
)

func initMappingsImpl() {
	// Load EFA enabled instances
	efaData, err := fs.ReadFile(configFS, "config/efa-enabled-instances.json")
	if err != nil {
		initError = fmt.Errorf("failed to load EFA enabled instances config: %v", err)
		return
	}
	if err := json.Unmarshal(efaData, &efaEnabledDevice); err != nil {
		initError = fmt.Errorf("failed to parse EFA enabled instances config: %v", err)
		return
	}

	// Load instance family mappings
	familyData, err := fs.ReadFile(configFS, "config/instance-family-mappings.json")
	if err != nil {
		initError = fmt.Errorf("failed to load instance family mappings config: %v", err)
		return
	}
	if err := json.Unmarshal(familyData, &instanceFamilyMappings); err != nil {
		initError = fmt.Errorf("failed to parse instance family mappings config: %v", err)
		return
	}

	// store efa enabled instances in json
	// store instance family to family mapping in json
	instanceType, err := getInstanceType()
	if err != nil {
		initError = fmt.Errorf("failed to get instance type: %v", err)
		return
	}

	if !efaEnabledDevice[instanceType] {
		klog.Warningf("Instance type %s is not EFA enabled", instanceType)
		return // This is not an error, just not supported
	}

	instanceFamily := getInstanceFamily(instanceType)
	data, err := fs.ReadFile(configFS, filepath.Join("neuron", instanceFamily+".json"))
	if err != nil {
		initError = fmt.Errorf("failed to load mappings for instance family %s: %v", instanceFamily, err)
		return
	}

	var mapping map[string]string // Neuron Device ID -> EFA BDF
	if err := json.Unmarshal(data, &mapping); err != nil {
		initError = fmt.Errorf("failed to parse mappings for instance family %s: %v", instanceFamily, err)
		return
	}

	for neuronDeviceID, efaBDF := range mapping {
		// Build efaToNeuron table
		efaToNeuron[efaBDF] = neuronDeviceID

		// Build neuronToEfa table
		neuronToEfa[neuronDeviceID] = efaBDF
	}
}

func init() {
	initMappings()
}
