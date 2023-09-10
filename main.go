package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

type devSample struct {
	uuid      string
	utilPC    uint32
	memIOPC   uint32
	memUsedMB uint64
	memUsedPC uint64
	tempDegC  uint32
	powerW    uint32
}

func genLogFileName(kind string) string {
	if kind != "dev" && kind != "proc" {
		panic("log type must be 'dev' or 'kind'")
	}

	path := fmt.Sprintf("gpu-%s-util", kind)
	if jid := os.Getenv("SLURM_JOB_ID"); jid != "" {
		path += "-" + jid
		if atid := os.Getenv("SLURM_ARRAY_TASK_ID"); atid != "" {
			path += "-" + atid
		}
	}
	path += ".log"
	return path
}

func (s devSample) String() string {
	return fmt.Sprintf("%v %d %d %d %d %d %d",
		s.uuid, s.utilPC, s.memUsedMB, s.memUsedPC, s.memIOPC, s.tempDegC, s.powerW)
}

func NewDevSample(d nvml.Device) devSample {
	var s devSample
	uuid, ret := d.GetUUID()
	if ret != nvml.SUCCESS {
		log.Fatalf("Unable to get uuid of device: %v", nvml.ErrorString(ret))
	}
	s.uuid = uuid

	util, ret := nvml.DeviceGetUtilizationRates(d)
	if ret != nvml.SUCCESS {
		log.Fatalf("Unable to get utilisation info for device %s: %v", s.uuid, nvml.ErrorString(ret))
	}
	s.utilPC = util.Gpu
	s.memIOPC = util.Memory

	devMemInfo, ret := nvml.DeviceGetMemoryInfo(d)
	if ret != nvml.SUCCESS {
		log.Fatalf("Unable to get mem info for device %s: %v", s.uuid, nvml.ErrorString(ret))
	}
	s.memUsedMB = devMemInfo.Used / (1024 * 1024)
	s.memUsedPC = devMemInfo.Used * 100 / devMemInfo.Total

	s.tempDegC, ret = nvml.DeviceGetTemperature(d, nvml.TEMPERATURE_GPU)
	if ret != nvml.SUCCESS {
		log.Fatalf("Unable to get temperature info for device %s: %v", s.uuid, nvml.ErrorString(ret))
	}
	devPowerInfo, ret := nvml.DeviceGetPowerUsage(d)
	if ret != nvml.SUCCESS {
		log.Fatalf("Unable to get power info for device %s: %v", s.uuid, nvml.ErrorString(ret))
	}
	s.powerW = devPowerInfo / 1000.
	return s
}

func timestep(devLog io.Writer, procLog io.Writer) {
	timeStr := time.Now().UTC().Format("2006-01-02T15:04:05Z ")

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

		s := NewDevSample(device)
		fmt.Fprintf(devLog, "%s %s\n", timeStr, s)

		procs, ret := nvml.DeviceGetComputeRunningProcesses(device)
		if ret != nvml.SUCCESS {
			log.Fatalf("Unable to get process info for GPU at index %d: %v", i, nvml.ErrorString(ret))
		}
		for _, proc := range procs {
			procName, ret := nvml.SystemGetProcessName(int(proc.Pid))
			if ret != nvml.SUCCESS {
				log.Fatalf("Unable to get process name for pid %d: %v", proc.Pid, nvml.ErrorString(ret))
			}
			procMemMB := proc.UsedGpuMemory / (1024 * 1024)
			fmt.Fprintf(procLog, "%s %s %d %s %d\n", timeStr, uuid, proc.Pid, procName, procMemMB)
		}
	}
}

func main() {
	// Arg parsing
	devLogPath := flag.String("devlog", "", "File path of GPU device utilisation log (default gpu-dev-util(-$SLURM_JOB_ID(-$SLURM_ARRAY_TASK_ID)).log")
	procLogPath := flag.String("proclog", "", "File path of GPU process utilisation log (default gpu-proc-util(-$SLURM_JOB_ID(-$SLURM_ARRAY_TASK_ID)).log")
	dtInt := flag.Int("frequency", 5, "Sampling frequency (seconds)")
	flag.Parse()

	dt := time.Duration(*dtInt) * time.Second
	if *devLogPath == "" {
		*devLogPath = genLogFileName("dev")
	}
	if *procLogPath == "" {
		*procLogPath = genLogFileName("proc")
	}

	// Create and open log files
	devLog, err := os.OpenFile(*devLogPath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("GPU device log file %s not created: %v\n", *devLogPath, err.Error())
	}
	defer devLog.Close()
	devLog.WriteString("timestamp dev_uuid dev_util_pc dev_mem_used_mb dev_mem_used_pc dev_mem_io_pc dev_temp_deg_c dev_power_w\n")

	procLog, err := os.OpenFile(*procLogPath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("GPU process log file %s not created: %v\n", *procLogPath, err.Error())
	}
	defer procLog.Close()
	procLog.WriteString("timestamp dev_uuid pid proc_name proc_mem_mb\n")

	// NVML initialisation and deferred shutdown
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
		time.Sleep(dt)
	}
}
