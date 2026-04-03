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
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEFADeviceGroupIDs(t *testing.T) {
	tests := []struct {
		name          string
		deviceID      string
		mockInstance  string
		mockError     error
		expectError   bool
		errorContains string
	}{
		{
			name:          "invalid BDF format",
			deviceID:      "invalid-bdf",
			expectError:   true,
			errorContains: "invalid BDF format",
		},
		{
			name:         "valid BDF",
			deviceID:     "0000:c9:00.0",
			mockInstance: "trn2.48xlarge",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalOsReadFile := osReadFile
			originalGetDeviceGroupIDs := getDeviceGroupIDs
			defer func() {
				osReadFile = originalOsReadFile
				getDeviceGroupIDs = originalGetDeviceGroupIDs
			}()

			osReadFile = func(path string) ([]byte, error) {
				if tt.mockError != nil {
					return nil, tt.mockError
				}
				return []byte(tt.mockInstance), nil
			}
			getDeviceGroupIDs = func(deviceID string, deviceType DeviceType) (map[string]string, error) {
				return map[string]string{"mock": "result"}, nil
			}

			result, err := GetEFADeviceGroupIDs(tt.deviceID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, result)
			} else {
				assert.Equal(t, map[string]string{"mock": "result"}, result)
			}
		})
	}
}

func TestGetNeuronDeviceGroupIDs(t *testing.T) {
	tests := []struct {
		name          string
		deviceID      string
		expectError   bool
		errorContains string
	}{
		{
			name:          "invalid device ID - non-numeric",
			deviceID:      "abc",
			expectError:   true,
			errorContains: "invalid neuron device ID",
		},
		{
			name:          "invalid device ID - empty",
			deviceID:      "",
			expectError:   true,
			errorContains: "invalid neuron device ID",
		},
		{
			name:          "invalid device ID - negative",
			deviceID:      "-1",
			expectError:   true,
			errorContains: "invalid neuron device ID",
		},
		{
			name:        "valid device ID",
			deviceID:    "0",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalGetDeviceGroupIDs := getDeviceGroupIDs
			defer func() {
				getDeviceGroupIDs = originalGetDeviceGroupIDs
			}()

			getDeviceGroupIDs = func(deviceID string, deviceType DeviceType) (map[string]string, error) {
				return map[string]string{"mock": "result"}, nil
			}

			result, err := GetNeuronDeviceGroupIDs(tt.deviceID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, result)
			} else {
				assert.Equal(t, map[string]string{"mock": "result"}, result)
			}
		})
	}
}

func TestGetDeviceGroupIDs(t *testing.T) {
	// Setup mock data
	originalDeviceTypeMappings := deviceTypeMappings
	originalNeuronToEfa := neuronToEfa
	defer func() {
		deviceTypeMappings = originalDeviceTypeMappings
		neuronToEfa = originalNeuronToEfa
	}()

	neuronToEfa = map[string]string{
		"0":  "0000:c9:00.0",
		"1":  "0000:b4:00.0",
		"2":  "0000:b3:00.0",
		"3":  "0000:ca:00.0",
		"4":  "0000:6c:00.0",
		"5":  "0000:57:00.0",
		"6":  "0000:56:00.0",
		"7":  "0000:6d:00.0",
		"8":  "0000:98:00.0",
		"9":  "0000:83:00.0",
		"10": "0000:82:00.0",
		"11": "0000:99:00.0",
		"12": "0000:f5:00.0",
		"13": "0000:e0:00.0",
		"14": "0000:df:00.0",
		"15": "0000:f6:00.0",
	}

	deviceTypeMappings = map[DeviceType]map[string]string{
		DeviceTypeEFA: {
			"0000:c9:00.0": "0",
			"0000:b4:00.0": "1",
			"0000:b3:00.0": "2",
			"0000:ca:00.0": "3",
			"0000:6c:00.0": "4",
			"0000:57:00.0": "5",
			"0000:56:00.0": "6",
			"0000:6d:00.0": "7",
			"0000:98:00.0": "8",
			"0000:83:00.0": "9",
			"0000:82:00.0": "10",
			"0000:99:00.0": "11",
			"0000:f5:00.0": "12",
			"0000:e0:00.0": "13",
			"0000:df:00.0": "14",
			"0000:f6:00.0": "15",
		},
		DeviceTypeNeuron: neuronToEfa,
	}

	tests := []struct {
		name         string
		deviceID     string
		deviceType   DeviceType
		instanceType string
	}{
		{
			name:       "successful neuron device lookup",
			deviceID:   "0",
			deviceType: DeviceTypeNeuron,
		},
		{
			name:       "successful EFA device lookup",
			deviceID:   "0000:c9:00.0",
			deviceType: DeviceTypeEFA,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getDeviceGroupIDsImpl(tt.deviceID, tt.deviceType)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Contains(t, result, DeviceGroup1Id)
			assert.Contains(t, result, DeviceGroup4Id)
			assert.Contains(t, result, DeviceGroup8Id)
			assert.Contains(t, result, DeviceGroup16Id)
			assert.Equal(t, "0000:c9:00.0", result[DeviceGroup1Id])
			assert.Equal(t, "13713e551cd37c19", result[DeviceGroup4Id])
			assert.Equal(t, "3593812b8066a078", result[DeviceGroup8Id])
			assert.Equal(t, "821d14afaf11327d", result[DeviceGroup16Id])

		})
	}
}

func TestGetDeviceGroupIDsNoMocks(t *testing.T) {
	tests := []struct {
		name          string
		deviceID      string
		instanceType  string
		deviceType    DeviceType
		expectError   bool
		errorContains string
	}{
		{
			name:          "unsupported device type",
			deviceID:      "0",
			deviceType:    "invalid",
			expectError:   true,
			errorContains: "unsupported device type",
		},
		{
			name:        "unsupported efa instance type",
			deviceID:    "0",
			deviceType:  DeviceTypeNeuron,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getDeviceGroupIDsImpl(tt.deviceID, tt.deviceType)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, map[string]string{}, result)
			}
		})
	}
}

func TestGetDeviceGroupIDsImplWithInitError(t *testing.T) {
	// Save original state
	originalInitError := initError
	defer func() {
		initError = originalInitError
	}()

	// Set initError to simulate initialization failure
	initError = fmt.Errorf("mock initialization error")

	tests := []struct {
		name       string
		deviceID   string
		deviceType DeviceType
	}{
		{
			name:       "neuron device with init error",
			deviceID:   "0",
			deviceType: DeviceTypeNeuron,
		},
		{
			name:       "EFA device with init error",
			deviceID:   "0000:c9:00.0",
			deviceType: DeviceTypeEFA,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getDeviceGroupIDsImpl(tt.deviceID, tt.deviceType)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "initialization failed")
			assert.Contains(t, err.Error(), "mock initialization error")
			assert.Nil(t, result)
		})
	}
}

func TestIsValidBDF(t *testing.T) {
	tests := []struct {
		name     string
		bdf      string
		expected bool
	}{
		{"valid BDF lowercase", "0000:c9:00.0", true},
		{"valid BDF uppercase", "0000:C9:00.0", true},
		{"valid BDF mixed case", "0000:c9:00.A", true},
		{"invalid - missing colon", "0000c9:00.0", false},
		{"invalid - missing dot", "0000:c9:000", false},
		{"invalid - wrong format", "000:c9:00.0", false},
		{"invalid - empty string", "", false},
		{"invalid - too long", "00000:c9:00.0", false},
		{"invalid - non-hex chars", "0000:gz:00.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidBDF(tt.bdf)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidNeuronDeviceID(t *testing.T) {
	tests := []struct {
		name     string
		deviceID string
		expected bool
	}{
		{"valid single digit", "0", true},
		{"valid multi digit", "15", true},
		{"invalid - empty string", "", false},
		{"invalid - non-numeric", "abc", false},
		{"invalid - mixed", "1a", false},
		{"invalid - negative", "-1", false},
		{"invalid - float", "1.5", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidNeuronDeviceID(tt.deviceID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetNeuronInstanceType(t *testing.T) {
	tests := []struct {
		name        string
		mockReturn  []byte
		mockError   error
		expected    string
		expectError bool
	}{
		{
			name:       "success",
			mockReturn: []byte("trn2.3xlarge"),
			mockError:  nil,
			expected:   "trn2.3xlarge",
		},
		{
			name:        "file read error",
			mockReturn:  nil,
			mockError:   errors.New("file not found"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalOsReadFile := osReadFile
			defer func() {
				osReadFile = originalOsReadFile
			}()

			osReadFile = func(path string) ([]byte, error) {
				if path == "/sys/devices/virtual/dmi/id/product_name" {
					return tt.mockReturn, tt.mockError
				}
				return nil, errors.New("invalid path")
			}

			result, err := getInstanceType()

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGetInstanceFamily(t *testing.T) {
	tests := []struct {
		name         string
		instanceType string
		expected     string
	}{
		{
			name:         "trn2u maps to trn2",
			instanceType: "trn2u.48xlarge",
			expected:     "trn2",
		},
		{
			name:         "trn1 stays trn1",
			instanceType: "trn1.32xlarge",
			expected:     "trn1",
		},
		{
			name:         "trn2 stays trn2",
			instanceType: "trn2.48xlarge",
			expected:     "trn2",
		},
		{
			name:         "no dot returns whole string",
			instanceType: "trn1",
			expected:     "trn1",
		},
		{
			name:         "empty string",
			instanceType: "",
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getInstanceFamily(tt.instanceType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetDeviceGroupHash(t *testing.T) {
	// Setup mock data
	originalNeuronToEfa := neuronToEfa
	defer func() {
		neuronToEfa = originalNeuronToEfa
	}()

	neuronToEfa = map[string]string{
		"0": "0000:c9:00.0",
		"1": "0000:ca:00.0",
		"2": "0000:cb:00.0",
		"3": "0000:cc:00.0",
		"4": "0000:c9:00.0",
		"5": "0000:ca:00.0",
		"6": "0000:cb:00.0",
		"7": "0000:cc:00.0",
	}

	tests := []struct {
		name           string
		neuronDeviceID string
		groupSize      int
		expectedEmpty  bool
		expectedResult string
	}{
		{
			name:           "group size 4, device 0",
			neuronDeviceID: "0",
			groupSize:      4,
			expectedEmpty:  false,
			expectedResult: "5b68000a3e3d51b1",
		},
		{
			name:           "group size 4, device 1",
			neuronDeviceID: "1",
			groupSize:      4,
			expectedEmpty:  false,
			expectedResult: "5b68000a3e3d51b1",
		},
		{
			name:           "group size 4, device 3",
			neuronDeviceID: "3",
			groupSize:      4,
			expectedEmpty:  false,
			expectedResult: "5b68000a3e3d51b1",
		},
		{
			name:           "group size 8, device 7",
			neuronDeviceID: "7",
			groupSize:      8,
			expectedEmpty:  false,
			expectedResult: "5b68000a3e3d51b1", // same hash due to remove duplicates
		},
		{
			name:           "invalid device ID",
			neuronDeviceID: "20",
			groupSize:      4,
			expectedEmpty:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getDeviceGroupHash(tt.neuronDeviceID, tt.groupSize)

			if tt.expectedEmpty {
				assert.Empty(t, result)
			} else {
				assert.NotEmpty(t, result)
				assert.Len(t, result, 16, "Hash should be 16 characters")
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestConstructDeviceGroupHash(t *testing.T) {
	tests := []struct {
		name     string
		efaBDFs  []string
		expected string
	}{
		{
			name:     "multiple BDFs sorted",
			efaBDFs:  []string{"0000:c9:00.0", "0000:ca:00.0"},
			expected: "5ba52364feae51b7", // First 16 chars of SHA256 hash
		},
		{
			name:     "duplicate BDFs removed",
			efaBDFs:  []string{"0000:c9:00.0", "0000:c9:00.0", "0000:ca:00.0"},
			expected: "5ba52364feae51b7", // Same as above after dedup
		},
		{
			name:     "unsorted BDFs get sorted",
			efaBDFs:  []string{"0000:ca:00.0", "0000:c9:00.0"},
			expected: "5ba52364feae51b7", // Same hash due to sorting
		},
		{
			name:     "empty slice",
			efaBDFs:  []string{},
			expected: "e3b0c44298fc1c14", // Hash of empty string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := constructDeviceGroupHash(tt.efaBDFs)
			assert.Len(t, result, 16, "Hash should be 16 characters")
			assert.Equal(t, tt.expected, result)
		})
	}
}
