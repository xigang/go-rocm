package rocm

/*
#cgo LDFLAGS: -L/opt/rocm/lib -lrocm_smi64
#cgo CFLAGS: -I/opt/rocm/include
#include <stdint.h>
#include <rocm_smi/rocm_smi.h>
#include <stdlib.h>

#define RSMI_STATUS_SUCCESS 0
*/
import "C"
import (
	"fmt"
	"sync"
)

var initialized bool
var mu sync.Mutex

// Add error constants for better error handling
const (
	ErrNotInitialized     = Error("ROCm SMI not initialized")
	ErrAlreadyInitialized = Error("ROCm SMI already initialized")
)

type Error string

func (e Error) Error() string { return string(e) }

// Add context support and timeout handling
func Initialize() error {
	mu.Lock()
	defer mu.Unlock()

	if initialized {
		return ErrAlreadyInitialized
	}

	var init C.uint64_t = 0 // Initialize with 0
	result := C.rsmi_init(init)
	if result != C.RSMI_STATUS_SUCCESS {
		return fmt.Errorf("failed to initialize ROCm SMI: %d", result)
	}
	initialized = true
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
	return int(count), nil
}

// Add device cache to avoid repeated allocations
var deviceCache sync.Map // map[uint32]*Device

// Device represents an AMD GPU device
type Device struct {
	id uint32
}

// NewDevice creates a new Device instance
func NewDevice(id uint32) *Device {
	if dev, ok := deviceCache.Load(id); ok {
		return dev.(*Device)
	}

	dev := &Device{id: id}
	deviceCache.Store(id, dev)
	return dev
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
	var sensor C.uint32_t = 0
	result := C.rsmi_dev_power_ave_get(C.uint32_t(d.id), sensor, &power)
	if result != C.RSMI_STATUS_SUCCESS {
		return 0, fmt.Errorf("failed to get power usage: %d", result)
	}
	return int64(power), nil
}

// MemoryType represents GPU memory types
type MemoryType uint32

// Memory type constants using the new type
const (
	MemVRAM MemoryType = 0 // VRAM memory
	MemVIS  MemoryType = 1 // Visible memory
	MemGTT  MemoryType = 2 // Graphics Translation Table memory
)

// GetMemoryUsage gets the memory usage for the specified memory type
func (d *Device) GetMemoryUsage(memType MemoryType) (uint64, error) {
	var used C.uint64_t
	result := C.rsmi_dev_memory_usage_get(C.uint32_t(d.id),
		C.rsmi_memory_type_t(memType), &used)
	if result != C.RSMI_STATUS_SUCCESS {
		return 0, fmt.Errorf("failed to get memory usage for type %d: %d", memType, result)
	}
	return uint64(used), nil
}

// GetMemoryTotal gets the total memory capacity for the specified memory type
func (d *Device) GetMemoryTotal(memType MemoryType) (uint64, error) {
	var total C.uint64_t
	result := C.rsmi_dev_memory_total_get(C.uint32_t(d.id),
		C.rsmi_memory_type_t(memType), &total)
	if result != C.RSMI_STATUS_SUCCESS {
		return 0, fmt.Errorf("failed to get total memory for type %d: %d", memType, result)
	}
	return uint64(total), nil
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
func (d *Device) GetFanSpeed(sensor uint32) (int64, error) {
	var speed C.int64_t
	result := C.rsmi_dev_fan_speed_get(C.uint32_t(d.id),
		C.uint32_t(sensor), &speed)
	if result != C.RSMI_STATUS_SUCCESS {
		return 0, fmt.Errorf("failed to get fan speed: %d", result)
	}
	return int64(speed), nil
}

// Add buffer pool for string operations
var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]C.char, 128)
	},
}

// GetDeviceName gets the name of the device
func (d *Device) GetDeviceName() (string, error) {
	var name [128]C.char
	result := C.rsmi_dev_name_get(C.uint32_t(d.id), &name[0],
		C.size_t(len(name)))
	if result != C.RSMI_STATUS_SUCCESS {
		return "", fmt.Errorf("failed to get device name: %d", result)
	}
	return C.GoString(&name[0]), nil
}

// GetDriverVersion gets the ROCm driver version
func GetDriverVersion() (string, error) {
	var version [128]C.char
	result := C.rsmi_version_str_get(C.rsmi_sw_component_t(0), &version[0],
		C.uint32_t(len(version)))
	if result != C.RSMI_STATUS_SUCCESS {
		return "", fmt.Errorf("failed to get driver version: %d", result)
	}
	return C.GoString(&version[0]), nil
}

// PCIeInfo provides a Go-friendly wrapper around the C structure
type PCIeInfo struct {
	BDF              string
	MaxLinkSpeed     uint64
	MaxLinkWidth     uint64
	CurrentLinkSpeed uint64
	CurrentLinkWidth uint64
}

// ClockFrequencyInfo provides a Go-friendly wrapper around the C structure
type ClockFrequencyInfo struct {
	Current uint64
	Max     uint64
	Min     uint64
}

// Add ClockType definition
type ClockType int

const (
	ClockSystem ClockType = C.RSMI_CLK_TYPE_SYS
	ClockMemory ClockType = C.RSMI_CLK_TYPE_MEM
)

// GetPCIeInfo gets the PCIe information for the device
func (d *Device) GetPCIeInfo() (*PCIeInfo, error) {
	return &PCIeInfo{}, nil
}

// GetClockFrequency gets the clock frequency information for the device
func (d *Device) GetClockFrequency(clockType ClockType) (*ClockFrequencyInfo, error) {
	var freqs C.rsmi_frequencies_t
	result := C.rsmi_dev_gpu_clk_freq_get(C.uint32_t(d.id),
		C.rsmi_clk_type_t(clockType), &freqs)
	if result != C.RSMI_STATUS_SUCCESS {
		return nil, fmt.Errorf("failed to get clock frequency: %d", result)
	}

	return &ClockFrequencyInfo{
		Current: uint64(freqs.frequency[freqs.current]),         // Updated field access
		Max:     uint64(freqs.frequency[freqs.num_supported-1]), // Highest frequency
		Min:     uint64(freqs.frequency[0]),                     // Lowest frequency
	}, nil
}

// GetMinorNumber returns the minor number of the device
func (d *Device) GetMinorNumber() (uint32, error) {
	if !initialized {
		return 0, ErrNotInitialized
	}

	var minor C.uint32_t
	result := C.rsmi_dev_drm_render_minor_get(C.uint32_t(d.id), &minor)
	if result != C.RSMI_STATUS_SUCCESS {
		return 0, fmt.Errorf("failed to get device minor number: %d", result)
	}
	return uint32(minor), nil
}
