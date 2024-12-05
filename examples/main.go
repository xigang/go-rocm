package main

import (
	"github.com/xigang/go-rocm/pkg/rocm"
	"k8s.io/klog/v2"
)

func main() {
	if err := rocm.Initialize(); err != nil {
		klog.Fatal("Failed to initialize ROCm SMI:", err)
	}
	defer rocm.Shutdown()

	count, err := rocm.DeviceCount()
	if err != nil {
		klog.Fatal("Failed to get device count:", err)
	}

	for i := 0; i < count; i++ {
		device := rocm.NewDevice(uint32(i))
		klog.Infof("Device %d: %+v", i, device)

		name, err := device.GetDeviceName()
		if err != nil {
			klog.Errorf("Failed to get device name for device %d: %v", i, err)
		}

		temp, err := device.GetTemperature(0, rocm.TempCurrent)
		if err != nil {
			klog.Errorf("Failed to get temperature for device %d: %v", i, err)
		}

		power, err := device.GetPowerUsage()
		if err != nil {
			klog.Errorf("Failed to get power for device %d: %v", i, err)
		}

		util, err := device.GetUtilization()
		if err != nil {
			klog.Errorf("Failed to get utilization for device %d: %v", i, err)
		}

		memUsed, err := device.GetMemoryUsage(rocm.MemVRAM)
		if err != nil {
			klog.Errorf("Failed to get memory info for device %d: %v", i, err)
		}

		memTotal, err := device.GetMemoryTotal(rocm.MemVRAM)
		if err != nil {
			klog.Errorf("Failed to get memory total for device %d: %v", i, err)
		}

		klog.Infof("Device %d (%s):\n", i, name)
		klog.Infof("  Temperature: %d°C\n", temp)
		klog.Infof("  Power: %d W\n", power/1000000)
		klog.Infof("  Utilization: %d%%\n", util)
		klog.Infof("  Memory: %d/%d\n", memUsed, memTotal)
	}
}
