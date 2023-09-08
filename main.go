package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

type logWriter struct {
}

func (writer logWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(time.Now().UTC().Format("2006-01-02T15:04:05Z ") + string(bytes))
}

func timestep(devLog log.Logger, procLog log.Logger) {
	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		log.Fatalf("Unable to get device count: %v", nvml.ErrorString(ret))
	}

	for i := 0; i < count; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			log.Fatalf("Unable to get device at index %d: %v", i, nvml.ErrorString(ret))
		}

		uuid, ret := device.GetUUID()
		if ret != nvml.SUCCESS {
			log.Fatalf("Unable to get uuid of device at index %d: %v", i, nvml.ErrorString(ret))
		}

		util, ret := nvml.DeviceGetUtilizationRates(device)
		if ret != nvml.SUCCESS {
			log.Fatalf("Unable to get utilisation info for device at index %d: %v", i, nvml.ErrorString(ret))
		}
		devUtilPercent := util.Gpu
		devMemIOPercent := util.Memory

		devMemInfo, ret := nvml.DeviceGetMemoryInfo(device)
		if ret != nvml.SUCCESS {
			log.Fatalf("Unable to get mem info for device at index %d: %v", i, nvml.ErrorString(ret))
		}

		devMemUsedMB := devMemInfo.Used / (1024 * 1024)
		devMemUsedPercent := devMemInfo.Used * 100 / devMemInfo.Total

		devTempDegC, ret := nvml.DeviceGetTemperature(device, nvml.TEMPERATURE_GPU)
		if ret != nvml.SUCCESS {
			log.Fatalf("Unable to get temperature info for device at index %d: %v", i, nvml.ErrorString(ret))
		}
		devPowerInfo, ret := nvml.DeviceGetPowerUsage(device)
		if ret != nvml.SUCCESS {
			log.Fatalf("Unable to get power info for device at index %d: %v", i, nvml.ErrorString(ret))
		}
		devPowerW := devPowerInfo / 1000.

		devLog.Printf("%v %d %d %d %d %d %d\n",
			uuid, devUtilPercent, devMemUsedMB, devMemUsedPercent, devMemIOPercent, devTempDegC, devPowerW)
	}
}

func main() {
	var dt time.Duration = 5

	// Logging setup
        f, err := os.OpenFile("dev.log", os.O_RDWR|os.O_CREATE, 0666)
        if err != nil {
           fmt.Println("GPU device log file not created", err.Error())
        }
        defer f.Close()
	f.WriteString("time date dev_uuid dev_util_pc dev_mem_used_mb dev_mem_used_pc dev_mem_io_pc dev_temp_deg_c dev_power_w\n")
        var devLog log.Logger
        //devLog.SetOutput(new(logWriter))
        devLog.SetFlags(log.Ldate|log.Ltime)
        devLog.SetOutput(f)

        f2, err := os.OpenFile("proc.log", os.O_RDWR|os.O_CREATE, 0666)
        if err != nil {
           fmt.Println("GPU process log file not created", err.Error())
        }
        defer f2.Close()
        var procLog log.Logger
        //procLog.SetOutput(new(logWriter))
        procLog.SetFlags(log.Ldate|log.Ltime)
        procLog.SetOutput(f)


	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		log.Fatalf("Unable to initialize NVML: %v", nvml.ErrorString(ret))
	}
	defer func() {
		ret := nvml.Shutdown()
		if ret != nvml.SUCCESS {
			log.Fatalf("Unable to shutdown NVML: %v", nvml.ErrorString(ret))
		}
	}()

	for {
		timestep(devLog, procLog)
		time.Sleep(dt * time.Second)
	}

}
