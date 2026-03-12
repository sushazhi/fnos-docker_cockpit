package handler

import (
	"dockpit/pkg/docker"
	"dockpit/pkg/response"

	"github.com/docker/docker/api/types/network"
	"github.com/gin-gonic/gin"
)

type NetworkHandler struct {
	service *docker.NetworkService
}

func NewNetworkHandler() *NetworkHandler {
	return &NetworkHandler{
		service: docker.NewNetworkService(),
	}
}

func (h *NetworkHandler) List(c *gin.Context) {
	ctx, cancel := docker.Context()
	defer cancel()

	networks, err := h.service.List(ctx)
	if err != nil {
		response.DockerError(c, "获取网络列表失败", err.Error())
		return
	}

	type NetworkWithContainers struct {
		network.Inspect
		ContainerCount int `json:"ContainerCount"`
	}

	result := make([]NetworkWithContainers, len(networks))
	for i, net := range networks {
		containerCount := 0
		if len(net.Containers) > 0 {
			containerCount = len(net.Containers)
		} else {
			info, err := h.service.Inspect(ctx, net.ID)
			if err == nil && info.Containers != nil {
				containerCount = len(info.Containers)
			}
		}
		result[i] = NetworkWithContainers{
			Inspect:        net,
			ContainerCount: containerCount,
		}
	}

	response.Success(c, gin.H{"networks": result})
}

func (h *NetworkHandler) Get(c *gin.Context) {
	ctx, cancel := docker.Context()
	defer cancel()

	id := c.Param("id")
	info, err := h.service.Inspect(ctx, id)
	if err != nil {
		response.DockerError(c, "获取网络信息失败", err.Error())
		return
	}

	response.Success(c, gin.H{"info": info})
}

type NetworkCreateRequest struct {
	Name       string            `json:"name" binding:"required"`
	Driver     string            `json:"driver"`
	Subnet     string            `json:"subnet"`
	Gateway    string            `json:"gateway"`
	Labels     map[string]string `json:"labels"`
	Internal   bool              `json:"internal"`
	Attachable bool              `json:"attachable"`
}

func (h *NetworkHandler) Create(c *gin.Context) {
	ctx, cancel := docker.Context()
	defer cancel()

	var req NetworkCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	opts := network.CreateOptions{
		Driver:     req.Driver,
		Internal:   req.Internal,
		Attachable: req.Attachable,
		Labels:     req.Labels,
	}

	if req.Subnet != "" || req.Gateway != "" {
		opts.IPAM = &network.IPAM{
			Config: []network.IPAMConfig{{
				Subnet:  req.Subnet,
				Gateway: req.Gateway,
			}},
		}
	}

	result, err := h.service.Create(ctx, req.Name, opts)
	if err != nil {
		response.DockerError(c, "创建网络失败", err.Error())
		return
	}

	addAuditLog(c, "network_create", map[string]interface{}{"name": req.Name})
	response.Success(c, gin.H{"success": true, "id": result.ID})
}

func (h *NetworkHandler) Remove(c *gin.Context) {
	ctx, cancel := docker.Context()
	defer cancel()

	id := c.Param("id")
	if err := h.service.Remove(ctx, id); err != nil {
		response.DockerError(c, "删除网络失败", err.Error())
		return
	}

	addAuditLog(c, "network_remove", map[string]interface{}{"id": id})
	response.Success(c, gin.H{"success": true})
}

type NetworkConnectRequest struct {
	ContainerID string `json:"containerId" binding:"required"`
	IPAddress   string `json:"ipAddress"`
}

func (h *NetworkHandler) Connect(c *gin.Context) {
	ctx, cancel := docker.Context()
	defer cancel()

	id := c.Param("id")
	var req NetworkConnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	var config *network.EndpointSettings
	if req.IPAddress != "" {
		config = &network.EndpointSettings{
			IPAMConfig: &network.EndpointIPAMConfig{
				IPv4Address: req.IPAddress,
			},
		}
	}

	if err := h.service.Connect(ctx, id, req.ContainerID, config); err != nil {
		response.DockerError(c, "连接网络失败", err.Error())
		return
	}

	addAuditLog(c, "network_connect", map[string]interface{}{
		"network":   id,
		"container": req.ContainerID,
	})
	response.Success(c, gin.H{"success": true})
}

type NetworkDisconnectRequest struct {
	ContainerID string `json:"containerId" binding:"required"`
	Force       bool   `json:"force"`
}

func (h *NetworkHandler) Disconnect(c *gin.Context) {
	ctx, cancel := docker.Context()
	defer cancel()

	id := c.Param("id")
	var req NetworkDisconnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := h.service.Disconnect(ctx, id, req.ContainerID, req.Force); err != nil {
		response.DockerError(c, "断开网络失败", err.Error())
		return
	}

	addAuditLog(c, "network_disconnect", map[string]interface{}{
		"network":   id,
		"container": req.ContainerID,
	})
	response.Success(c, gin.H{"success": true})
}

func (h *NetworkHandler) Prune(c *gin.Context) {
	ctx, cancel := docker.ContextWithTimeout(120000 * 1000000)
	defer cancel()

	report, err := h.service.Prune(ctx)
	if err != nil {
		response.DockerError(c, "清理网络失败", err.Error())
		return
	}

	addAuditLog(c, "network_prune", nil)
	response.Success(c, gin.H{
		"success":        true,
		"networksDeleted": report.NetworksDeleted,
	})
}
