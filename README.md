# gpuutiliz: basic periodic logging of GPU utilisation

Log information about GPU utilisation and GPU memory utilisation per-device and per-process.

Data gathered via the [NVIDIA Management Library (NVML)](https://developer.nvidia.com/nvidia-management-library-nvml).

## Usage

```
$ ./gpuutiliz -h
Usage of ./gpuutiliz:
  -devlog string
    	File path of GPU device utilisation log (default gpu-dev-util(-$SLURM_JOB_ID(-$SLURM_ARRAY_TASK_ID)).log
  -frequency int
    	Sampling frequency (seconds) (default 5)
  -proclog string
    	File path of GPU process utilisation log (default gpu-proc-util(-$SLURM_JOB_ID(-$SLURM_ARRAY_TASK_ID)).log
```

NB must be run on a machine with the NVIDIA device driver installed, otherwise will fail with a segmentation fault.

To run in a Slurm batch job, you could run as a background process like so:

```
#!/bin/bash
#SBATCH --time=00:30:00
#SBATCH --gres=gpu:2

/path/to/gpuutiliz &

./someCudaProgram.exe
```

NB here data will only be collected for the GPUs allocated to the Slurm job (via a device cgroup).


## Example output

```
$ cat gpu-dev-util.log | column -t
timestamp             dev_uuid                                  dev_util_pc  dev_mem_used_mb  dev_mem_used_pc  dev_mem_io_pc  dev_temp_deg_c  dev_power_w
2023-09-11T09:59:45Z  GPU-04d1014e-5cc0-4443-db48-f0b9e38e693e  0            866              1                0              37              59
2023-09-11T09:59:45Z  GPU-dcc5f6e3-2de3-8adc-67df-bdd813ab6428  0            866              1                0              30              57
2023-09-11T09:59:45Z  GPU-b7f25043-705d-8287-9789-74ebf2164ae7  97           18462            22               19             58              315
2023-09-11T09:59:45Z  GPU-314c6802-4912-0008-05c7-f1c442181a09  100          18462            22               18             57              299
2023-09-11T09:59:50Z  GPU-04d1014e-5cc0-4443-db48-f0b9e38e693e  0            866              1                0              37              59
2023-09-11T09:59:50Z  GPU-dcc5f6e3-2de3-8adc-67df-bdd813ab6428  0            866              1                0              30              57
2023-09-11T09:59:50Z  GPU-b7f25043-705d-8287-9789-74ebf2164ae7  100          18462            22               18             57              312
2023-09-11T09:59:50Z  GPU-314c6802-4912-0008-05c7-f1c442181a09  100          18462            22               18             57              317

$ cat gpu-proc-util.log | column -t
timestamp             dev_uuid                                  pid    proc_name  proc_mem_mb
2023-09-11T09:59:45Z  GPU-b7f25043-705d-8287-9789-74ebf2164ae7  15848  python     17592
2023-09-11T09:59:45Z  GPU-314c6802-4912-0008-05c7-f1c442181a09  16017  python     17592
2023-09-11T09:59:50Z  GPU-b7f25043-705d-8287-9789-74ebf2164ae7  15848  python     17592
2023-09-11T09:59:50Z  GPU-314c6802-4912-0008-05c7-f1c442181a09  16017  python     17592
```

## Building

```
git clone $THISREPO
cd gpuutiliz
go build
```

## Caveats

  * Assuming Slurm GPU job tasks are all single-node
  * May not produce accurate results for MIG slices (not tested yet)
