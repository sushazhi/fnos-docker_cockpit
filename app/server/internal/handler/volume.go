package handler

import (
	"dockpit/pkg/docker"
	"dockpit/pkg/response"

	"github.com/docker/docker/api/types/volume"
	"github.com/gin-gonic/gin"
)

type VolumeHandler struct {
	service *docker.VolumeService
}

func NewVolumeHandler() *VolumeHandler {
	return &VolumeHandler{
		service: docker.NewVolumeService(),
	}
}

func (h *VolumeHandler) List(c *gin.Context) {
	ctx, cancel := docker.Context()
	defer cancel()

	result, err := h.service.List(ctx)
	if err != nil {
		response.DockerError(c, "获取存储卷列表失败", err.Error())
		return
	}

	// 获取所有容器（包括停止的）
	containerService := docker.NewContainerService()
	containers, err := containerService.List(ctx, true)
	if err != nil {
		// 如果获取容器失败，仍然返回存储卷列表
		response.Success(c, gin.H{"volumes": result.Volumes})
		return
	}

	// 构建存储卷使用情况映射
	volumeUsage := make(map[string][]string) // volumeName -> []containerNames
	for _, cont := range containers {
		for _, mount := range cont.Mounts {
			if mount.Type == "volume" && mount.Name != "" {
				containerName := cont.Names[0]
				if len(containerName) > 0 && containerName[0] == '/' {
					containerName = containerName[1:]
				}
				volumeUsage[mount.Name] = append(volumeUsage[mount.Name], containerName)
			}
		}
	}

	// 为每个存储卷添加实际使用信息
	type VolumeWithUsage struct {
		*volume.Volume
		Containers []string `json:"containers,omitempty"`
		Used       bool     `json:"used"`
	}

	volumesWithUsage := make([]VolumeWithUsage, 0, len(result.Volumes))
	for _, vol := range result.Volumes {
		containers, used := volumeUsage[vol.Name]
		volumesWithUsage = append(volumesWithUsage, VolumeWithUsage{
			Volume:     vol,
			Containers: containers,
			Used:       used,
		})
	}

	response.Success(c, gin.H{"volumes": volumesWithUsage})
}

func (h *VolumeHandler) Get(c *gin.Context) {
	ctx, cancel := docker.Context()
	defer cancel()

	name := c.Param("name")
	info, err := h.service.Inspect(ctx, name)
	if err != nil {
		response.DockerError(c, "获取存储卷信息失败", err.Error())
		return
	}

	response.Success(c, gin.H{"info": info})
}

type VolumeCreateRequest struct {
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	DriverOpts map[string]string `json:"driverOpts"`
	Labels     map[string]string `json:"labels"`
}

func (h *VolumeHandler) Create(c *gin.Context) {
	ctx, cancel := docker.Context()
	defer cancel()

	var req VolumeCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	result, err := h.service.Create(ctx, volume.CreateOptions{
		Name:       req.Name,
		Driver:     req.Driver,
		DriverOpts: req.DriverOpts,
		Labels:     req.Labels,
	})
	if err != nil {
		response.DockerError(c, "创建存储卷失败", err.Error())
		return
	}

	addAuditLog(c, "volume_create", map[string]interface{}{"name": req.Name})
	response.Success(c, gin.H{"success": true, "name": result.Name})
}

type VolumeRemoveRequest struct {
	Force bool `json:"force"`
}

func (h *VolumeHandler) Remove(c *gin.Context) {
	ctx, cancel := docker.Context()
	defer cancel()

	name := c.Param("name")
	var req VolumeRemoveRequest
	c.ShouldBindJSON(&req)

	if err := h.service.Remove(ctx, name, req.Force); err != nil {
		response.DockerError(c, "删除存储卷失败", err.Error())
		return
	}

	addAuditLog(c, "volume_remove", map[string]interface{}{"name": name})
	response.Success(c, gin.H{"success": true})
}

func (h *VolumeHandler) Prune(c *gin.Context) {
	ctx, cancel := docker.ContextWithTimeout(120000 * 1000000)
	defer cancel()

	report, err := h.service.Prune(ctx)
	if err != nil {
		response.DockerError(c, "清理存储卷失败", err.Error())
		return
	}

	addAuditLog(c, "volume_prune", nil)
	response.Success(c, gin.H{
		"success":        true,
		"volumesDeleted": report.VolumesDeleted,
		"spaceReclaimed": report.SpaceReclaimed,
	})
}
