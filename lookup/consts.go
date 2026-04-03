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

type DeviceType string

const (
	ResourceAttributePrefix                = "resource.aws.com/"
	DeviceGroup1Id                         = ResourceAttributePrefix + "devicegroup1_id"
	DeviceGroup4Id                         = ResourceAttributePrefix + "devicegroup4_id"
	DeviceGroup8Id                         = ResourceAttributePrefix + "devicegroup8_id"
	DeviceGroup16Id                        = ResourceAttributePrefix + "devicegroup16_id"
	DeviceTypeEFA               DeviceType = "efa"
	DeviceTypeNeuron            DeviceType = "neuron"
	NeuronInstanceTypeSysFsPath            = "/sys/devices/virtual/dmi/id/product_name"
)
