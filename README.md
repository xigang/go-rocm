# Go Bindings for the ROCm SMI Library

[![Go Report Card](https://goreportcard.com/badge/github.com/xigang/go-rocm)](https://goreportcard.com/report/github.com/xigang/go-rocm)
[![GoDoc](https://godoc.org/github.com/xigang/go-rocm?status.svg)](https://godoc.org/github.com/xigang/go-rocm)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A lightweight Go wrapper for AMD's ROCm System Management Interface (SMI) library. Monitor and control AMD GPUs with ease.

## Overview

This library provides Go bindings for the ROCm SMI library, allowing you to:
- Monitor GPU metrics (temperature, power, memory, utilization)
- Control GPU settings (fan speed, clock frequencies)
- Access device information (name, PCIe config, driver version)

## Requirements

- ROCm 5.0 or later
- Go 1.21 or later
- GCC/Clang compiler

## Quick Start

1. **Install ROCm**
   ```bash
   # Verify ROCm installation
   ls /opt/rocm/include/rocm_smi/rocm_smi.h
   ls /opt/rocm/lib/librocm_smi64.so
   ```

2. **Install Package**
   ```bash
   go get github.com/xigang/go-rocm
   ```

3. **Basic Usage**
   ```go
   package main

   import (
       "log"
       "github.com/xigang/go-rocm/pkg/rocm"
   )

   func main() {
       // Initialize ROCm SMI
       if err := rocm.Initialize(); err != nil {
           log.Fatal(err)
       }
       defer rocm.Shutdown()

       // Get GPU count
       count, err := rocm.DeviceCount()
       if err != nil {
           log.Fatal(err)
       }

       // Print GPU information
       for i := 0; i < count; i++ {
           device := rocm.NewDevice(uint32(i))
           name, _ := device.GetDeviceName()
           temp, _ := device.GetTemperature(0, rocm.TempCurrent)
           power, _ := device.GetPowerUsage()
           util, _ := device.GetUtilization()

           log.Printf("GPU %d: %s", i, name)
           log.Printf("  Temperature: %d°C", temp)
           log.Printf("  Power: %d W", power/1000000)
           log.Printf("  Utilization: %d%%", util)
       }
   }
   ```

## Building
```sh
make build
```
