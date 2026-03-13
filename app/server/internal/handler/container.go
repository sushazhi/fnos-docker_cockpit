package handler

import (
	"context"
	"dockpit/internal/model"
	"dockpit/internal/service"
	"dockpit/pkg/docker"
	"dockpit/pkg/response"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/websocket"
)

type ContainerHandler struct {
	service *docker.ContainerService
}

func NewContainerHandler() *ContainerHandler {
	return &ContainerHandler{
		service: docker.NewContainerService(),
	}
}

func (h *ContainerHandler) List(c *gin.Context) {
	ctx, cancel := docker.Context()
	defer cancel()

	all := c.Query("all") != "false"
	containers, err := h.service.List(ctx, all)
	if err != nil {
		response.DockerError(c, "Docker未安装或未运行", err.Error())
		return
	}

	response.Success(c, gin.H{"containers": containers})
}

func (h *ContainerHandler) Get(c *gin.Context) {
	ctx, cancel := docker.Context()
	defer cancel()

	id := c.Param("id")
	info, err := h.service.Inspect(ctx, id)
	if err != nil {
		response.DockerError(c, "获取容器信息失败", err.Error())
		return
	}

	response.Success(c, gin.H{"info": info})
}

var prevNetworkStats = make(map[string]struct {
	rxBytes uint64
	txBytes uint64
	time    int64
})

func (h *ContainerHandler) Stats(c *gin.Context) {
	ctx, cancel := docker.Context()
	defer cancel()

	id := c.Param("id")
	statsReader, err := h.service.Stats(ctx, id)
	if err != nil {
		response.Success(c, gin.H{"stats": nil})
		return
	}
	defer statsReader.Body.Close()

	var statsJSON struct {
		CPUStats struct {
			CPUUsage struct {
				TotalUsage uint64 `json:"total_usage"`
			} `json:"cpu_usage"`
			SystemUsage uint64 `json:"system_cpu_usage"`
		} `json:"cpu_stats"`
		PreCPUStats struct {
			CPUUsage struct {
				TotalUsage uint64 `json:"total_usage"`
			} `json:"cpu_usage"`
			SystemUsage uint64 `json:"system_cpu_usage"`
		} `json:"precpu_stats"`
		MemoryStats struct {
			Usage  uint64            `json:"usage"`
			Limit  uint64            `json:"limit"`
			Stats  map[string]uint64 `json:"stats"`
			Cache  uint64            `json:"cache"`
		} `json:"memory_stats"`
		Networks map[string]struct {
			RxBytes uint64 `json:"rx_bytes"`
			TxBytes uint64 `json:"tx_bytes"`
		} `json:"networks"`
		BlockIO struct {
			IoServiceBytesRecursive []struct {
				Op    string `json:"op"`
				Value uint64 `json:"value"`
			} `json:"io_service_bytes_recursive"`
		} `json:"blkio_stats"`
	}

	decoder := json.NewDecoder(statsReader.Body)
	if err := decoder.Decode(&statsJSON); err != nil {
		response.Success(c, gin.H{"stats": nil})
		return
	}

	cpuDelta := float64(statsJSON.CPUStats.CPUUsage.TotalUsage - statsJSON.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(statsJSON.CPUStats.SystemUsage - statsJSON.PreCPUStats.SystemUsage)
	cpuPerc := 0.0
	if systemDelta > 0 {
		cpuPerc = (cpuDelta / systemDelta) * 100.0
	}

	// 计算实际内存使用量
	// Docker 的 Usage 包含了可回收的缓存（inactive_file），需要减去才能得到真实内存使用
	// 公式：实际内存 = Usage - total_inactive_file
	memUsage := statsJSON.MemoryStats.Usage
	
	// 从 Stats 中获取 total_inactive_file（可回收的文件缓存）
	var inactiveFile uint64
	if statsJSON.MemoryStats.Stats != nil {
		// cgroups v1 和 v2 的字段名可能不同
		inactiveFile = statsJSON.MemoryStats.Stats["total_inactive_file"]
		if inactiveFile == 0 {
			inactiveFile = statsJSON.MemoryStats.Stats["inactive_file"]
		}
	}
	
	// 如果没有 total_inactive_file，尝试使用 cache 作为降级方案
	if inactiveFile == 0 {
		inactiveFile = statsJSON.MemoryStats.Cache
	}
	
	// 实际内存使用 = 总使用 - 可回收缓存
	actualMemUsage := memUsage - inactiveFile
	if actualMemUsage > memUsage || actualMemUsage < 0 {
		// 防止负数或溢出
		actualMemUsage = memUsage
	}

	memLimit := statsJSON.MemoryStats.Limit
	memPerc := 0.0
	if memLimit > 0 {
		memPerc = (float64(actualMemUsage) / float64(memLimit)) * 100.0
	}

	var rxBytes, txBytes uint64
	for _, net := range statsJSON.Networks {
		rxBytes += net.RxBytes
		txBytes += net.TxBytes
	}

	now := time.Now().UnixNano()
	netIORate := "-"
	if prev, ok := prevNetworkStats[id]; ok && now > prev.time {
		timeDiff := float64(now-prev.time) / 1e9
		rxRate := float64(rxBytes-prev.rxBytes) / timeDiff
		txRate := float64(txBytes-prev.txBytes) / timeDiff
		netIORate = fmt.Sprintf("↓%s/s ↑%s/s", formatBytes(uint64(rxRate)), formatBytes(uint64(txRate)))
	}
	prevNetworkStats[id] = struct {
		rxBytes uint64
		txBytes uint64
		time    int64
	}{rxBytes: rxBytes, txBytes: txBytes, time: now}

	var readBytes, writeBytes uint64
	for _, io := range statsJSON.BlockIO.IoServiceBytesRecursive {
		switch io.Op {
		case "Read", "read":
			readBytes += io.Value
		case "Write", "write":
			writeBytes += io.Value
		}
	}

	response.Success(c, gin.H{"stats": gin.H{
		"CPUPerc":     fmt.Sprintf("%.2f%%", cpuPerc),
		"MemUsage":    fmt.Sprintf("%s / %s", formatBytes(actualMemUsage), formatBytes(memLimit)),
		"MemPerc":     fmt.Sprintf("%.2f%%", memPerc),
		"NetIO":       fmt.Sprintf("%s / %s", formatBytes(rxBytes), formatBytes(txBytes)),
		"NetIORate":   netIORate,
		"BlockIO":     fmt.Sprintf("%s / %s", formatBytes(readBytes), formatBytes(writeBytes)),
		"MemoryUsage": actualMemUsage,
		"MemoryLimit": memLimit,
	}})
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func (h *ContainerHandler) Start(c *gin.Context) {
	ctx, cancel := docker.Context()
	defer cancel()

	id := c.Param("id")
	if err := h.service.Start(ctx, id); err != nil {
		response.DockerError(c, "启动容器失败", err.Error())
		return
	}

	addAuditLog(c, "container_start", map[string]interface{}{"id": id})
	response.Success(c, gin.H{"success": true})
}

func (h *ContainerHandler) Stop(c *gin.Context) {
	ctx, cancel := docker.Context()
	defer cancel()

	id := c.Param("id")
	var req model.ContainerStopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Timeout = 10
	}

	timeout := req.Timeout
	if err := h.service.Stop(ctx, id, &timeout); err != nil {
		response.DockerError(c, "停止容器失败", err.Error())
		return
	}

	addAuditLog(c, "container_stop", map[string]interface{}{"id": id})
	response.Success(c, gin.H{"success": true})
}

func (h *ContainerHandler) Restart(c *gin.Context) {
	ctx, cancel := docker.Context()
	defer cancel()

	id := c.Param("id")
	var req model.ContainerStopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Timeout = 10
	}

	timeout := req.Timeout
	if err := h.service.Restart(ctx, id, &timeout); err != nil {
		response.DockerError(c, "重启容器失败", err.Error())
		return
	}

	addAuditLog(c, "container_restart", map[string]interface{}{"id": id})
	response.Success(c, gin.H{"success": true})
}

func (h *ContainerHandler) Pause(c *gin.Context) {
	ctx, cancel := docker.Context()
	defer cancel()

	id := c.Param("id")
	if err := h.service.Pause(ctx, id); err != nil {
		response.DockerError(c, "暂停容器失败", err.Error())
		return
	}

	response.Success(c, gin.H{"success": true})
}

func (h *ContainerHandler) Unpause(c *gin.Context) {
	ctx, cancel := docker.Context()
	defer cancel()

	id := c.Param("id")
	if err := h.service.Unpause(ctx, id); err != nil {
		response.DockerError(c, "恢复容器失败", err.Error())
		return
	}

	response.Success(c, gin.H{"success": true})
}

func (h *ContainerHandler) Remove(c *gin.Context) {
	ctx, cancel := docker.Context()
	defer cancel()

	id := c.Param("id")
	var req model.ContainerRemoveRequest
	c.ShouldBindJSON(&req)

	if err := h.service.Remove(ctx, id, req.Force, req.Volumes); err != nil {
		response.DockerError(c, "删除容器失败", err.Error())
		return
	}

	addAuditLog(c, "container_remove", map[string]interface{}{"id": id})
	response.Success(c, gin.H{"success": true})
}

func (h *ContainerHandler) Logs(c *gin.Context) {
	ctx, cancel := docker.ContextWithTimeout(60 * 1000 * 1000000)
	defer cancel()

	id := c.Param("id")
	var query model.ContainerLogsQuery
	c.ShouldBindQuery(&query)

	tail := "500"
	if query.Tail != "" {
		tail = query.Tail
	}

	opts := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       tail,
		Since:      query.Since,
		Timestamps: query.Timestamps,
	}

	logs, err := h.service.Logs(ctx, id, opts)
	if err != nil {
		response.DockerError(c, "获取日志失败", err.Error())
		return
	}

	response.Success(c, gin.H{"logs": string(logs)})
}

func (h *ContainerHandler) Create(c *gin.Context) {
	ctx, cancel := docker.ContextWithTimeout(120 * 1000 * 1000000)
	defer cancel()

	var req model.ContainerCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	config := &container.Config{
		Image:      req.Image,
		Env:        req.Env,
		Labels:     req.Labels,
		WorkingDir: req.WorkDir,
		User:       req.User,
		Cmd:        req.Cmd,
		Hostname:   req.Hostname,
	}

	hostConfig := &container.HostConfig{
		Privileged: req.Privileged,
		RestartPolicy: container.RestartPolicy{
			Name: container.RestartPolicyMode(req.Restart),
		},
		PortBindings: make(nat.PortMap),
	}

	for _, vol := range req.Volumes {
		hostConfig.Binds = append(hostConfig.Binds, vol)
	}

	for _, cap := range req.CapAdd {
		hostConfig.CapAdd = append(hostConfig.CapAdd, cap)
	}

	for _, cap := range req.CapDrop {
		hostConfig.CapDrop = append(hostConfig.CapDrop, cap)
	}

	var networkingConfig *network.NetworkingConfig

	result, err := h.service.Create(ctx, config, hostConfig, networkingConfig, req.Name)
	if err != nil {
		response.DockerError(c, "创建容器失败", err.Error())
		return
	}

	addAuditLog(c, "container_create", map[string]interface{}{
		"name":  req.Name,
		"image": req.Image,
	})
	response.Success(c, gin.H{"success": true, "id": result.ID})
}

func (h *ContainerHandler) Rename(c *gin.Context) {
	ctx, cancel := docker.Context()
	defer cancel()

	id := c.Param("id")
	var req model.ContainerRenameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := h.service.Rename(ctx, id, req.Name); err != nil {
		response.DockerError(c, "重命名容器失败", err.Error())
		return
	}

	response.Success(c, gin.H{"success": true})
}

func (h *ContainerHandler) Update(c *gin.Context) {
	ctx, cancel := docker.Context()
	defer cancel()

	id := c.Param("id")
	var req model.ContainerUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	updateConfig := container.UpdateConfig{}
	if req.RestartPolicy != "" {
		updateConfig.RestartPolicy = container.RestartPolicy{Name: container.RestartPolicyMode(req.RestartPolicy)}
	}

	if err := h.service.Update(ctx, id, updateConfig); err != nil {
		response.DockerError(c, "更新容器失败", err.Error())
		return
	}

	response.Success(c, gin.H{"success": true})
}

func (h *ContainerHandler) Commit(c *gin.Context) {
	ctx, cancel := docker.ContextWithTimeout(120 * 1000 * 1000000)
	defer cancel()

	id := c.Param("id")
	var req model.ContainerCommitRequest
	c.ShouldBindJSON(&req)

	opts := container.CommitOptions{
		Author:  req.Author,
		Comment: req.Message,
	}

	result, err := h.service.Commit(ctx, id, opts)
	if err != nil {
		response.DockerError(c, "提交容器失败", err.Error())
		return
	}

	addAuditLog(c, "container_commit", map[string]interface{}{"id": id})
	response.Success(c, gin.H{"success": true, "imageId": result.ID})
}

func (h *ContainerHandler) Exec(c *gin.Context) {
	ctx, cancel := docker.ContextWithTimeout(60 * 1000 * 1000000)
	defer cancel()

	id := c.Param("id")
	var req model.ContainerExecRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	execConfig := container.ExecOptions{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          req.Cmd,
	}

	execResp, err := h.service.ExecCreate(ctx, id, execConfig)
	if err != nil {
		response.DockerError(c, "创建执行命令失败", err.Error())
		return
	}

	attachResp, err := h.service.ExecAttach(ctx, execResp.ID, container.ExecAttachOptions{})
	if err != nil {
		response.DockerError(c, "执行命令失败", err.Error())
		return
	}
	defer attachResp.Close()

	output, _ := io.ReadAll(attachResp.Reader)
	response.Success(c, gin.H{"output": string(output)})
}

func (h *ContainerHandler) ExecCreate(c *gin.Context) {
	// 使用后台 context，不设置超时，让 Docker API 自然完成
	ctx := context.Background()

	id := c.Param("id")
	var req struct {
		Cmd       []string `json:"cmd"`
		Width     int      `json:"width"`
		Height    int      `json:"height"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if len(req.Cmd) == 0 {
		req.Cmd = []string{"/bin/sh"}
	}

	if req.Width <= 0 || req.Height <= 0 {
		response.BadRequest(c, "终端尺寸无效")
		return
	}

	if req.Width > 1000 || req.Height > 1000 {
		response.BadRequest(c, "终端尺寸超出限制")
		return
	}

	execConfig := container.ExecOptions{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          req.Cmd,
	}

	// 使用 goroutine 异步创建 exec，但立即返回 exec ID
	execResp, err := h.service.ExecCreate(ctx, id, execConfig)
	if err != nil {
		response.DockerError(c, "创建终端失败", err.Error())
		return
	}

	execID := execResp.ID

	// 注册 execID 与当前会话的关联
	sessionToken, _ := c.Get("sessionToken")
	if sessionToken != nil {
		service.GetSessionService().RegisterExec(sessionToken.(string), execID)
	}

	// 异步调整终端大小，不阻塞响应
	if req.Width > 0 && req.Height > 0 {
		go func() {
			resizeCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			h.service.ExecResize(resizeCtx, execID, container.ResizeOptions{
				Height: uint(req.Height),
				Width:  uint(req.Width),
			})
		}()
	}

	response.Success(c, gin.H{"execId": execID})
}

func (h *ContainerHandler) ExecResize(c *gin.Context) {
	ctx, cancel := docker.Context()
	defer cancel()

	execID := c.Param("execId")
	var req struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	// 验证参数
	if req.Width < 1 || req.Height < 1 {
		response.BadRequest(c, "终端尺寸必须大于0")
		return
	}

	if req.Width > 1000 || req.Height > 1000 {
		response.BadRequest(c, "终端尺寸超出限制")
		return
	}

	if err := h.service.ExecResize(ctx, execID, container.ResizeOptions{
		Height: uint(req.Height),
		Width:  uint(req.Width),
	}); err != nil {
		response.DockerError(c, "调整终端大小失败", err.Error())
		return
	}

	response.Success(c, gin.H{"success": true})
}

func (h *ContainerHandler) ExecWrite(c *gin.Context) {
	ctx, cancel := docker.ContextWithTimeout(30 * 1000 * 1000000)
	defer cancel()

	execID := c.Param("execId")
	var req struct {
		Data string `json:"data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	cli := docker.GetClient()
	conn, err := cli.ContainerExecAttach(ctx, execID, container.ExecAttachOptions{Tty: true})
	if err != nil {
		response.DockerError(c, "连接终端失败", err.Error())
		return
	}
	defer conn.Close()

	conn.Conn.Write([]byte(req.Data))

	response.Success(c, gin.H{"success": true})
}

func (h *ContainerHandler) ExecRead(c *gin.Context) {
	ctx, cancel := docker.ContextWithTimeout(30 * 1000 * 1000000)
	defer cancel()

	execID := c.Param("execId")

	cli := docker.GetClient()
	conn, err := cli.ContainerExecAttach(ctx, execID, container.ExecAttachOptions{Tty: true})
	if err != nil {
		response.DockerError(c, "连接终端失败", err.Error())
		return
	}
	defer conn.Close()

	buf := make([]byte, 4096)
	n, _ := conn.Reader.Read(buf)
	output := string(buf[:n])

	response.Success(c, gin.H{"output": output})
}

func (h *ContainerHandler) ExecWebSocket(c *gin.Context) {
	execID := c.Param("execId")

	sessionToken, err := c.Cookie("session_token")
	if err != nil || sessionToken == "" {
		c.JSON(401, gin.H{"error": "Authentication required"})
		return
	}

	if !service.GetSessionService().ValidateSession(sessionToken) {
		c.JSON(401, gin.H{"error": "Session expired"})
		return
	}

	if !service.GetSessionService().ValidateExecOwnership(sessionToken, execID) {
		c.JSON(403, gin.H{"error": "Access denied"})
		return
	}

	// 使用独立的 context，不继承 HTTP 请求的超时设置
	ctx := context.Background()
	cli := docker.GetClient()

	// 立即连接到 exec，这会启动 shell 进程
	execConn, err := cli.ContainerExecAttach(ctx, execID, container.ExecAttachOptions{Tty: true})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	handler := websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		defer execConn.Close()
		// 连接关闭时清理 execID 记录
		defer service.GetSessionService().RemoveExec(execID)

		var wg sync.WaitGroup
		done := make(chan bool, 2)

		wg.Add(2)

		// 从容器读取数据并发送到 WebSocket
		go func() {
			defer wg.Done()
			buf := make([]byte, 4096)
			for {
				select {
				case <-done:
					return
				default:
					// 设置读取超时，避免永久阻塞
					execConn.Conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
					n, err := execConn.Reader.Read(buf)
					if err != nil {
						if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
							// 超时是正常的，继续循环
							continue
						}
						// 真正的错误，退出
						done <- true
						return
					}
					if n > 0 {
						websocket.Message.Send(ws, string(buf[:n]))
					}
				}
			}
		}()

		// 从 WebSocket 读取数据并发送到容器
		go func() {
			defer wg.Done()
			var msg string
			for {
				select {
				case <-done:
					return
				default:
					// 设置读取超时
					ws.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
					err := websocket.Message.Receive(ws, &msg)
					if err != nil {
						if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
							// 超时是正常的，继续循环
							continue
						}
						// 真正的错误，退出
						done <- true
						return
					}
					if len(msg) > 0 {
						execConn.Conn.Write([]byte(msg))
					}
				}
			}
		}()

		wg.Wait()
	})

	handler.ServeHTTP(c.Writer, c.Request)
}

func SearchContainers(c *gin.Context) {
	ctx, cancel := docker.Context()
	defer cancel()

	term := c.Query("term")
	if term == "" {
		response.Success(c, gin.H{"containers": []interface{}{}})
		return
	}

	cli := docker.GetClient()
	containers, err := cli.ContainerList(ctx, container.ListOptions{
		All: true,
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "name",
			Value: term,
		}),
	})
	if err != nil {
		response.DockerError(c, "搜索容器失败", err.Error())
		return
	}

	response.Success(c, gin.H{"containers": containers})
}

// BatchOperation 批量操作容器
func (h *ContainerHandler) BatchOperation(c *gin.Context) {
	var req model.BatchOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if len(req.IDs) == 0 {
		response.BadRequest(c, "容器ID列表不能为空")
		return
	}

	if len(req.IDs) > 50 {
		response.BadRequest(c, "批量操作最多支持50个容器")
		return
	}

	result := model.BatchOperationResult{
		Success: []string{},
		Failed:  []model.BatchOperationError{},
	}

	timeout := req.Timeout
	if timeout == 0 {
		timeout = 10
	}

	for _, id := range req.IDs {
		ctx, cancel := docker.Context()
		var err error

		switch req.Operation {
		case "start":
			err = h.service.Start(ctx, id)
		case "stop":
			err = h.service.Stop(ctx, id, &timeout)
		case "restart":
			err = h.service.Restart(ctx, id, &timeout)
		case "pause":
			err = h.service.Pause(ctx, id)
		case "unpause":
			err = h.service.Unpause(ctx, id)
		case "remove":
			err = h.service.Remove(ctx, id, req.Force, false)
		default:
			cancel()
			response.BadRequest(c, "不支持的操作: "+req.Operation)
			return
		}

		cancel()

		if err != nil {
			result.Failed = append(result.Failed, model.BatchOperationError{
				ID:    id,
				Error: err.Error(),
			})
		} else {
			result.Success = append(result.Success, id)
		}
	}

	addAuditLog(c, "container_batch_"+req.Operation, map[string]interface{}{
		"ids":     req.IDs,
		"success": len(result.Success),
		"failed":  len(result.Failed),
	})

	response.Success(c, result)
}
