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
	"io"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockEfaFS struct{}

func (m *mockEfaFS) Open(name string) (fs.File, error) {
	return nil, errors.New("mock file not found")
}

func (m *mockEfaFS) ReadFile(name string) ([]byte, error) {
	if name == "config/efa-enabled-instances.json" {
		return nil, errors.New("mock config file read error")
	}
	return nil, errors.New("mock file not found")
}

type mockFamilyFS struct{}

func (m *mockFamilyFS) Open(name string) (fs.File, error) {
	if name == "config/efa-enabled-instances.json" {
		return &mockFile{content: `{"trn2.48xlarge": true}`}, nil
	}
	if name == "config/instance-family-mappings.json" {
		return nil, errors.New("mock family mappings read error")
	}
	return nil, errors.New("mock file not found")
}

type mockFile struct {
	content string
	pos     int
}

func (f *mockFile) Read(p []byte) (n int, err error) {
	if f.pos >= len(f.content) {
		return 0, io.EOF
	}
	n = copy(p, f.content[f.pos:])
	f.pos += n
	return n, nil
}

func (f *mockFile) Close() error               { return nil }
func (f *mockFile) Stat() (fs.FileInfo, error) { return nil, errors.New("not implemented") }

func TestInitMappingsImpl(t *testing.T) {
	tests := []struct {
		name             string
		mockInstanceType string
		expectMappings   bool
	}{
		{
			name:             "successful EFA enabled instance",
			mockInstanceType: "trn2.48xlarge",
			expectMappings:   true,
		},
		{
			name:             "EFA disabled instance",
			mockInstanceType: "inf1.xlarge",
			expectMappings:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original state
			originalGetInstanceType := getInstanceType
			originalInitError := initError
			originalNeuronToEfa := neuronToEfa
			originalEfaToNeuron := efaToNeuron
			defer func() {
				getInstanceType = originalGetInstanceType
				initError = originalInitError
				neuronToEfa = originalNeuronToEfa
				efaToNeuron = originalEfaToNeuron
			}()

			// Reset state
			initError = nil
			neuronToEfa = make(map[string]string)
			efaToNeuron = make(map[string]string)

			// Mock getInstanceType
			getInstanceType = func() (string, error) {
				return tt.mockInstanceType, nil
			}

			// Call function
			initMappingsImpl()

			assert.Nil(t, initError)

			if tt.expectMappings {
				assert.NotEmpty(t, neuronToEfa)
				assert.NotEmpty(t, efaToNeuron)
			} else {
				assert.Empty(t, neuronToEfa)
				assert.Empty(t, efaToNeuron)
			}
		})
	}
}

func TestInitMappingsImplErrorPaths(t *testing.T) {
	tests := []struct {
		name              string
		mockInstanceType  string
		mockInstanceError error
		mockConfigFS      fs.FS
		expectError       bool
		errorContains     string
	}{
		{
			name:              "getInstanceType error",
			mockInstanceError: errors.New("file not found"),
			expectError:       true,
			errorContains:     "failed to get instance type",
		},
		{
			name:             "missing efa enabled config file",
			mockInstanceType: "trn2.48xlarge",
			mockConfigFS:     &mockEfaFS{}, // This will cause config file read errors
			expectError:      true,
			errorContains:    "failed to load EFA enabled instances config",
		},
		{
			name:             "missing instance family mapping config file",
			mockInstanceType: "trn2.48xlarge",
			mockConfigFS:     &mockFamilyFS{}, // This will cause config file read errors
			expectError:      true,
			errorContains:    "failed to load instance family mappings config",
		},
		{
			name:             "missing neuron mapping file",
			mockInstanceType: "trn99.48xlarge", // Instance family that doesn't have a JSON file
			expectError:      true,
			errorContains:    "failed to load mappings for instance family",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original state
			originalGetInstanceType := getInstanceType
			originalInitError := initError
			originalConfigFS := configFS
			originalEfaEnabledDevice := efaEnabledDevice
			originalInstanceFamilyMappings := instanceFamilyMappings
			defer func() {
				getInstanceType = originalGetInstanceType
				initError = originalInitError
				configFS = originalConfigFS
				efaEnabledDevice = originalEfaEnabledDevice
				instanceFamilyMappings = originalInstanceFamilyMappings
			}()

			// Reset state
			initError = nil

			// Set up basic config for non-config-error tests
			if tt.mockConfigFS == nil {
				efaEnabledDevice = map[string]bool{
					"trn2.48xlarge":  true,
					"trn99.48xlarge": true, // Enable but no JSON file exists
				}
				instanceFamilyMappings = map[string]string{
					"trn2":  "trn2",
					"trn99": "trn99",
				}
			} else {
				configFS = tt.mockConfigFS
			}

			getInstanceType = func() (string, error) {
				return tt.mockInstanceType, tt.mockInstanceError
			}

			initMappingsImpl()

			if tt.expectError {
				assert.NotNil(t, initError)
				assert.Contains(t, initError.Error(), tt.errorContains)
			} else {
				assert.Nil(t, initError)
			}
		})
	}
}
