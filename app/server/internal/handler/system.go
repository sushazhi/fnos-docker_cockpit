package handler

import (
	"bufio"
	"dockpit/pkg/docker"
	"dockpit/pkg/response"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/system"
	"github.com/gin-gonic/gin"
	"runtime/debug"
)

type SystemHandler struct {
	service *docker.SystemService
}

func NewSystemHandler() *SystemHandler {
	return &SystemHandler{
		service: docker.NewSystemService(),
	}
}

// getHostMemory 获取宿主机物理内存信息
// 注意：如果程序运行在 Docker 容器中，/proc/meminfo 返回的是容器内存限制
// 所以优先使用 Docker Info 返回的宿主机内存
func getHostMemory(dockerMemTotal int64) map[string]interface{} {
	result := map[string]interface{}{
		"total":     int64(0),
		"available": int64(0),
		"used":      int64(0),
	}

	// 优先使用 Docker 返回的宿主机内存（更可靠）
	if dockerMemTotal > 0 {
		result["total"] = dockerMemTotal
		// Docker 不直接返回 available，尝试从 /proc/meminfo 获取
		if runtime.GOOS == "linux" {
			if file, err := os.Open("/proc/meminfo"); err == nil {
				defer file.Close()
				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					line := scanner.Text()
					fields := strings.Fields(line)
					if len(fields) < 2 {
						continue
					}
					key := strings.TrimSuffix(fields[0], ":")
					if key == "MemAvailable" {
						value, _ := strconv.ParseInt(fields[1], 10, 64)
						// /proc/meminfo 中的值以 kB 为单位，转换为字节
						value *= 1024
						// 如果 available 大于 total，说明读取的是容器内存，不可用
						if value < dockerMemTotal {
							result["available"] = value
							result["used"] = dockerMemTotal - value
						}
						break
					}
				}
			}
		}
		return result
	}

	// 降级：直接读取 /proc/meminfo（可能在容器中不准确）
	if runtime.GOOS != "linux" {
		return result
	}

	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return result
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var total, available int64
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		key := strings.TrimSuffix(fields[0], ":")
		value, _ := strconv.ParseInt(fields[1], 10, 64)
		// /proc/meminfo 中的值以 kB 为单位，转换为字节
		value *= 1024

		switch key {
		case "MemTotal":
			total = value
		case "MemAvailable":
			available = value
		}
	}

	result["total"] = total
	result["available"] = available
	result["used"] = total - available
	return result
}

func (h *SystemHandler) Info(c *gin.Context) {
	ctx, cancel := docker.Context()
	defer cancel()

	info, err := h.service.Info(ctx)
	if err != nil {
		response.DockerError(c, "获取Docker信息失败", err.Error())
		return
	}

	// 同时获取版本信息（包含 API 版本）
	version, err := h.service.Version(ctx)
	if err != nil {
		response.DockerError(c, "获取Docker版本失败", err.Error())
		return
	}

	// 将 API 版本等信息合并到 info 中
	infoWithVersion := struct {
		system.Info
		APIVersion    string `json:"ApiVersion"`
		MinAPIVersion string `json:"MinAPIVersion,omitempty"`
	}{
		Info:          info,
		APIVersion:    version.APIVersion,
		MinAPIVersion: version.MinAPIVersion,
	}

	// 获取宿主机物理内存信息，传入 Docker 返回的内存总量
	hostMemory := getHostMemory(info.MemTotal)

	response.Success(c, gin.H{
		"info":       infoWithVersion,
		"hostMemory": hostMemory,
	})
}

func (h *SystemHandler) Version(c *gin.Context) {
	ctx, cancel := docker.Context()
	defer cancel()

	version, err := h.service.Version(ctx)
	if err != nil {
		response.DockerError(c, "获取Docker版本失败", err.Error())
		return
	}

	response.Success(c, gin.H{"version": version})
}

func (h *SystemHandler) DiskUsage(c *gin.Context) {
	ctx, cancel := docker.Context()
	defer cancel()

	usage, err := h.service.DiskUsage(ctx)
	if err != nil {
		response.DockerError(c, "获取磁盘使用情况失败", err.Error())
		return
	}

	var imagesSize, containersSize, volumesSize, buildCacheSize int64
	for _, img := range usage.Images {
		imagesSize += img.Size
	}
	for _, container := range usage.Containers {
		containersSize += int64(container.SizeRw)
	}
	for _, vol := range usage.Volumes {
		volumesSize += vol.UsageData.Size
	}
	for _, bc := range usage.BuildCache {
		buildCacheSize += int64(bc.Size)
	}

	response.Success(c, gin.H{
		"usage":          usage,
		"ImagesSize":     imagesSize,
		"ContainersSize": containersSize,
		"VolumesSize":    volumesSize,
		"BuildCacheSize": buildCacheSize,
		"TotalSize":      imagesSize + containersSize + volumesSize + buildCacheSize,
	})
}

type SystemPruneRequest struct {
	All     bool `json:"all"`
	Volumes bool `json:"volumes"`
}

func (h *SystemHandler) Prune(c *gin.Context) {
	ctx, cancel := docker.ContextWithTimeout(5 * 60 * 1000 * 1000000)
	defer cancel()

	var req SystemPruneRequest
	c.ShouldBindJSON(&req)

	report, err := h.service.Prune(ctx, req.All, req.Volumes)
	if err != nil {
		response.DockerError(c, "清理系统失败", err.Error())
		return
	}

	spaceReclaimed := uint64(0)
	if report != nil {
		spaceReclaimed = report.SpaceReclaimed
	}

	addAuditLog(c, "system_prune", map[string]interface{}{
		"all":     req.All,
		"volumes": req.Volumes,
	})
	response.Success(c, gin.H{
		"success":        true,
		"spaceReclaimed": spaceReclaimed,
	})
}

func (h *SystemHandler) Check(c *gin.Context) {
	ctx, cancel := docker.Context()
	defer cancel()

	err := docker.Ping(ctx)
	if err != nil {
		response.Success(c, gin.H{
			"available": false,
			"error":     err.Error(),
		})
		return
	}

	response.Success(c, gin.H{"available": true})
}

func (h *SystemHandler) Events(c *gin.Context) {
	ctx := c.Request.Context()

	eventsChan, errChan := h.service.Events(ctx, events.ListOptions{})

	for {
		select {
		case event := <-eventsChan:
			c.SSEvent("event", event)
			c.Writer.Flush()
		case err := <-errChan:
			if err != nil {
				c.SSEvent("error", err.Error())
			}
			return
		case <-ctx.Done():
			return
		}
	}
}

// getDockerSDKVersion 从构建信息中获取 Docker SDK 版本
func getDockerSDKVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}

	// 查找 docker/docker 依赖
	for _, dep := range info.Deps {
		if dep.Path == "github.com/docker/docker" {
			return dep.Version
		}
	}

	return "unknown"
}

// AppInfo 返回应用信息，包括 Docker SDK 版本
// 版本号通过 ldflags 在构建时注入: -ldflags "-X dockpit/internal/handler.appVersion=1.0.0"
var appVersion = "1.0.0" // 默认值，构建时会被覆盖

func (h *SystemHandler) AppInfo(c *gin.Context) {
	response.Success(c, gin.H{
		"version":          appVersion,
		"dockerSDKVersion": getDockerSDKVersion(),
		"goVersion":        runtime.Version(),
		"platform":         runtime.GOOS + "/" + runtime.GOARCH,
	})
}
