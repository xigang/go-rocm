package rocm

/*
#cgo LDFLAGS: -L/opt/rocm/lib -lrocm_smi64
#cgo CFLAGS: -I/opt/rocm/include
#include <rocm_smi/rocm_smi.h>
#include <stdlib.h>
*/

import "C"
import (
	"fmt"
	"sync"
)

var initialized bool
var mu sync.Mutex

// Initialize initializes the ROCm SMI Library
func Initialize() error {
	mu.Lock()
	defer mu.Unlock()

	if !initialized {
		result := C.rsmi_init()
		if result != C.RSMI_STATUS_SUCCESS {
			return fmt.Errorf("failed to initialize ROCm SMI: %d", result)
		}
		initialized = true
	}
	return nil
}

// Shutdown shuts down the ROCm SMI Library
func Shutdown() error {
	mu.Lock()
	defer mu.Unlock()

	if initialized {
		result := C.rsmi_shut_down()
		if result != C.RSMI_STATUS_SUCCESS {
			return fmt.Errorf("failed to shutdown ROCm SMI: %d", result)
		}
		initialized = false
	}
	return nil
}

// DeviceCount returns the number of GPU devices
func DeviceCount() (int, error) {
	var count C.uint32_t
	result := C.rsmi_num_monitor_devices(&count)
	if result != C.RSMI_STATUS_SUCCESS {
		return 0, fmt.Errorf("failed to get device count: %d", result)
	}
	return uint32(count), nil
}

// Device represents an AMD GPU device
type Device struct {
	id uint32
}

// NewDevice creates a new Device instance
func NewDevice(id uint32) *Device {
	return &Device{id: id}
}

// Temperature represents GPU temperature types
type Temperature int

const (
	TempCurrent Temperature = C.RSMI_TEMP_CURRENT
	TempMax     Temperature = C.RSMI_TEMP_MAX
	TempMin     Temperature = C.RSMI_TEMP_MIN
)

// GetTemperature returns the temperature of a GPU device
func (d *Device) GetTemperature(sensor uint32, metric Temperature) (int64, error) {
	var temp C.int64_t
	result := C.rsmi_dev_temp_metric_get(C.uint32_t(d.id), C.uint32_t(sensor),
		C.rsmi_temperature_metric_t(metric), &temp)
	if result != C.RSMI_STATUS_SUCCESS {
		return 0, fmt.Errorf("failed to get temperature: %d", result)
	}
	return int64(temp), nil
}

// GetPowerUsage gets the power usage in microwatts
func (d *Device) GetPowerUsage() (int64, error) {
	var power C.uint64_t
	result := C.rsmi_dev_power_ave_get(C.uint32_t(d.id), &power)
	if result != C.RSMI_STATUS_SUCCESS {
		return 0, fmt.Errorf("failed to get power usage: %d", result)
	}
	return int64(power), nil
}

// GetMemoryUsage gets the memory usage for the specified memory type
func (d *Device) GetMemoryUsage(memType uint32) (uint64, error) {
	var used C.uint64_t
	result := C.rsmi_dev_memory_usage_get(C.uint32_t(d.id), C.uint32_t(memType), &used)
	if result != C.RSMI_STATUS_SUCCESS {
		return 0, fmt.Errorf("failed to get memory usage: %d", result)
	}
	return uint64(used), nil
}

// GetUtilization gets the GPU utilization percentage
func (d *Device) GetUtilization() (uint32, error) {
	var util C.uint32_t
	result := C.rsmi_dev_busy_percent_get(C.uint32_t(d.id), &util)
	if result != C.RSMI_STATUS_SUCCESS {
		return 0, fmt.Errorf("failed to get utilization: %d", result)
	}
	return uint32(util), nil
}

// GetFanSpeed gets the fan speed percentage
func (d *Device) GetFanSpeed(sensor uint32) (int32, error) {
	var speed C.int32_t
	result := C.rsmi_dev_fan_speed_get(C.uint32_t(d.id), C.uint32_t(sensor), &speed)
	if result != C.RSMI_STATUS_SUCCESS {
		return 0, fmt.Errorf("failed to get fan speed: %d", result)
	}
	return int32(speed), nil
}

// GetDeviceName gets the name of the device
func (d *Device) GetDeviceName() (string, error) {
	var name [128]C.char
	result := C.rsmi_dev_name_get(C.uint32_t(d.id), &name[0], C.uint32_t(len(name)))
	if result != C.RSMI_STATUS_SUCCESS {
		return "", fmt.Errorf("failed to get device name: %d", result)
	}
	return C.GoString(&name[0]), nil
}

// GetDriverVersion gets the ROCm driver version
func GetDriverVersion() (string, error) {
	var version [128]C.char
	result := C.rsmi_version_str_get(&version[0], C.uint32_t(len(version)))
	if result != C.RSMI_STATUS_SUCCESS {
		return "", fmt.Errorf("failed to get driver version: %d", result)
	}
	return C.GoString(&version[0]), nil
}

// GetPCIeInfo gets PCIe information for the device
type PCIeInfo struct {
	BDF              string
	MaxLinkSpeed     uint64
	MaxLinkWidth     uint64
	CurrentLinkSpeed uint64
	CurrentLinkWidth uint64
}

func (d *Device) GetPCIeInfo() (*PCIeInfo, error) {
	var bdf [32]C.char
	var maxSpeed, maxWidth, curSpeed, curWidth C.uint64_t

	result := C.rsmi_dev_pci_id_get(C.uint32_t(d.id), &bdf[0], C.uint32_t(len(bdf)))
	if result != C.RSMI_STATUS_SUCCESS {
		return nil, fmt.Errorf("failed to get PCI BDF: %d", result)
	}

	result = C.rsmi_dev_pci_bandwidth_get(C.uint32_t(d.id), &maxSpeed, &maxWidth,
		&curSpeed, &curWidth)
	if result != C.RSMI_STATUS_SUCCESS {
		return nil, fmt.Errorf("failed to get PCIe bandwidth info: %d", result)
	}

	return &PCIeInfo{
		BDF:              C.GoString(&bdf[0]),
		MaxLinkSpeed:     uint64(maxSpeed),
		MaxLinkWidth:     uint64(maxWidth),
		CurrentLinkSpeed: uint64(curSpeed),
		CurrentLinkWidth: uint64(curWidth),
	}, nil
}

// GetClockFrequency gets the clock frequency for the specified type
type ClockType int

const (
	ClockSystem ClockType = C.RSMI_CLK_TYPE_SYS
	ClockMemory ClockType = C.RSMI_CLK_TYPE_MEM
)

// GetClockFrequency gets the clock frequency for the specified type
func (d *Device) GetClockFrequency(clockType ClockType) (uint64, error) {
	var freq C.uint64_t
	result := C.rsmi_dev_gpu_clk_freq_get(C.uint32_t(d.id),
		C.rsmi_clk_type_t(clockType), &freq)
	if result != C.RSMI_STATUS_SUCCESS {
		return 0, fmt.Errorf("failed to get clock frequency: %d", result)
	}
	return uint64(freq), nil
}
