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
	"crypto/sha256"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	klog "k8s.io/klog/v2"
)

var (
	osReadFile        = os.ReadFile
	getDeviceGroupIDs = getDeviceGroupIDsImpl
	getInstanceType   = getInstanceTypeImpl
)

func GetEFADeviceGroupIDs(deviceID string) (map[string]string, error) {
	// Validate BDF format (e.g., "0000:c9:00.0")
	if !isValidBDF(deviceID) {
		return nil, fmt.Errorf("invalid BDF format: %s", deviceID)
	}

	return getDeviceGroupIDs(deviceID, DeviceTypeEFA)
}

func GetNeuronDeviceGroupIDs(deviceID string) (map[string]string, error) {
	// Validate device ID is numeric
	if !isValidNeuronDeviceID(deviceID) {
		return nil, fmt.Errorf("invalid neuron device ID: %s", deviceID)
	}
	return getDeviceGroupIDs(deviceID, DeviceTypeNeuron)
}

func isValidBDF(bdf string) bool {
	// Match format: 0000:c9:00.0
	matched, _ := regexp.MatchString(`^[0-9a-fA-F]{4}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}\.[0-9a-fA-F]$`, bdf)
	return matched
}

func isValidNeuronDeviceID(deviceID string) bool {
	intID, err := strconv.Atoi(deviceID)
	return err == nil && deviceID != "" && intID >= 0
}

func getDeviceGroupIDsImpl(deviceID string, deviceType DeviceType) (map[string]string, error) {
	if initError != nil {
		return nil, fmt.Errorf("initialization failed: %v", initError)
	}

	mapping := deviceTypeMappings[deviceType]
	if mapping == nil {
		return nil, fmt.Errorf("unsupported device type: %s", deviceType)
	}

	otherDeviceId, exists := mapping[deviceID]
	if !exists {
		klog.Warningf("Device ID %s not found in mapping, likely not an EFA-enabled instance", deviceID)
		return map[string]string{}, nil
	}

	var neuronDeviceID string
	var efaBDF string
	if deviceType == DeviceTypeEFA {
		neuronDeviceID = otherDeviceId
		efaBDF = deviceID
	} else {
		neuronDeviceID = deviceID
		efaBDF = otherDeviceId
	}

	devicegroup_ids := map[string]string{
		// This allows efa + neuron aligned allocation for 1 neuron device.
		// Note: for allocation on trn1, only 8 pods can request aligned allocation with group1
		// this is because only 8 efa devices are on trn1, and neuron to efa devices are mapped 2:1
		DeviceGroup1Id: efaBDF,
		// Returns a hash, ensuring neuron devices in the same topology group of 4 get the same group ID.
		// Their corresponding EFA devices will also get the same group ID.
		DeviceGroup4Id: getDeviceGroupHash(neuronDeviceID, 4),
		// Returns a hash, ensuring neuron devices in the same topology group of 8 get the same group ID.
		// Their corresponding EFA devices will also get the same group ID.
		DeviceGroup8Id: getDeviceGroupHash(neuronDeviceID, 8),
		// Returns a hash, ensuring devices in the same topology group of 16 (i.e. all neuron devices)
		// get the same group ID. Their corresponding EFA devices will also get the same group ID.
		DeviceGroup16Id: getDeviceGroupHash(neuronDeviceID, 16),
	}

	klog.InfoS("Retrieved device group IDs", "deviceType", deviceType, "deviceId", deviceID, "deviceGroupIds", devicegroup_ids)

	return devicegroup_ids, nil
}

func getInstanceTypeImpl() (string, error) {
	instanceType, err := osReadFile(NeuronInstanceTypeSysFsPath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(instanceType)), nil
}

func getInstanceFamily(instanceType string) string {
	parts := strings.Split(instanceType, ".")
	family := parts[0]

	return instanceFamilyMappings[family]
}

func getDeviceGroupHash(neuronDeviceIDStr string, groupSize int) string {
	// get all EFA BDFs for devices with same device / groupSize value
	neuronDeviceID, _ := strconv.Atoi(neuronDeviceIDStr)
	groupIndex := neuronDeviceID / groupSize

	var efaBDFs []string

	// Get all devices in the same topology group
	for i := groupIndex * groupSize; i < (groupIndex+1)*groupSize; i++ {
		deviceStr := strconv.Itoa(i)
		if efaBDF, exists := neuronToEfa[deviceStr]; exists {
			efaBDFs = append(efaBDFs, efaBDF)
		}
	}
	// Construct hash
	if len(efaBDFs) > 0 {
		return constructDeviceGroupHash(efaBDFs)
	}
	return ""
}

func constructDeviceGroupHash(efaBDFs []string) string {
	// Remove duplicate BDFs (relevant in the case of trn1)
	seen := make(map[string]bool)
	var unique []string

	for _, bdf := range efaBDFs {
		if !seen[bdf] {
			seen[bdf] = true
			unique = append(unique, bdf)
		}
	}

	// Construct hash by joining sorted BDFs
	sort.Strings(unique)
	data := strings.Join(unique, ",")

	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)[:16] // truncate to 16 chars
}
