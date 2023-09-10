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

## Building

```
git clone $THISREPO
cd gpuutiliz
go build
```

## Caveats

Memory usage doesn't match the values shown by `nvidia-smi`,
possibly because the figures presented by this tool might include memory reserved for NVIDIA system management.
