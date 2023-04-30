package utils

const (
	ResourceName = "gpu-share.com/gpu-mem"
	CountName    = "gpu-share.com/gpu-count"

	EnvNVGPU              = "NVIDIA_VISIBLE_DEVICES"
	EnvResourceIndex      = "GPU_SHARE_COM_GPU_MEM_IDX"
	EnvResourceByPod      = "GPU_SHARE_COM_GPU_MEM_POD"
	EnvResourceByDev      = "GPU_SHARE_COM_GPU_MEM_DEV"
	EnvAssignedFlag       = "GPU_SHARE_COM_GPU_MEM_ASSIGNED"
	EnvResourceAssumeTime = "GPU_SHARE_COM_GPU_MEM_ASSUME_TIME"
)
