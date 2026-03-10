package handler

import (
	"bufio"
	"bytes"
	"context"
	"dockpit/internal/config"
	"dockpit/pkg/docker"
	"dockpit/pkg/response"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/gin-gonic/gin"
)

type ComposeHandler struct{}

func NewComposeHandler() *ComposeHandler {
	return &ComposeHandler{}
}

// isValidComposeName 验证 compose 项目名称是否合法
func isValidComposeName(name string) bool {
	if name == "" || len(name) > 128 {
		return false
	}
	// 只允许字母、数字、-、_
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

func (h *ComposeHandler) getComposeDir() string {
	return filepath.Join(config.Get().DataDir, "compose")
}

// isValidName 验证名称是否合法，防止命令注入
func isValidName(name string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, name)
	return matched && len(name) > 0 && len(name) <= 255
}

// sanitizePath 验证路径是否在允许的基础目录内，防止路径遍历攻击
func sanitizePath(baseDir, userPath string) (string, error) {
	// 清理路径，移除 . 和 .. 等相对路径组件
	cleaned := filepath.Clean(filepath.Join(baseDir, userPath))

	// 确保清理后的路径在基础目录内
	if !strings.HasPrefix(cleaned, baseDir) {
		return "", errors.New("path traversal detected: access denied")
	}

	return cleaned, nil
}

// getAllowedBaseDirs 返回允许访问的基础目录列表
func getAllowedBaseDirs() []string {
	dataDir := config.Get().DataDir
	return []string{
		dataDir,
		filepath.Join(dataDir, "compose"),
		// 移除 /tmp，避免符号链接攻击风险
	}
}

// validatePath 验证用户提供的路径是否在允许的目录内
func validatePath(userPath string) (string, error) {
	cleaned := filepath.Clean(userPath)

	resolved, err := filepath.EvalSymlinks(cleaned)
	if err != nil {
		return "", err
	}

	allowedDirs := getAllowedBaseDirs()
	for _, baseDir := range allowedDirs {
		resolvedBase, err := filepath.EvalSymlinks(baseDir)
		if err != nil {
			continue
		}
		if strings.HasPrefix(resolved, resolvedBase) {
			return cleaned, nil
		}
	}

	return "", errors.New("access denied: path is outside allowed directories")
}

func (h *ComposeHandler) ensureDir() {
	os.MkdirAll(h.getComposeDir(), 0755)
}

func (h *ComposeHandler) List(c *gin.Context) {
	h.ensureDir()
	dir := h.getComposeDir()
	log.Printf("[Compose] List called, dir: %s", dir)

	type ComposeItem struct {
		Name     string    `json:"name"`
		File     string    `json:"file"`
		Path     string    `json:"path"`
		Status   string    `json:"status"`
		Modified time.Time `json:"modified"`
		Size     int64     `json:"size"`
		Services int       `json:"services"`
	}

	list := make([]ComposeItem, 0)
	projectMap := make(map[string]bool)

	// 首先使用 docker compose ls 获取所有运行中的项目
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	cmd := exec.CommandContext(ctx, "docker", "compose", "ls", "--format", "json")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err == nil {
		log.Printf("[Compose] docker compose ls output: %s", out.String())
		var projects []struct {
			Name    string `json:"Name"`
			Status  string `json:"Status"`
			Config  string `json:"ConfigFiles"`
		}
		if err := json.Unmarshal(out.Bytes(), &projects); err == nil {
			log.Printf("[Compose] Found %d running projects from docker compose ls", len(projects))
			for _, p := range projects {
				log.Printf("[Compose] Project: %s, Config: %s", p.Name, p.Config)
				configFiles := strings.Split(p.Config, ",")
				if len(configFiles) > 0 {
					configPath := strings.TrimSpace(configFiles[0])
					if configPath != "" {
						info, err := os.Stat(configPath)
						if err != nil {
							log.Printf("[Compose] Failed to stat config file %s: %v", configPath, err)
							continue
						}
						projectMap[p.Name] = true
						
						// 获取服务数量
						services := h.getServiceCount(configPath)
						
						list = append(list, ComposeItem{
							Name:     p.Name,
							File:     filepath.Base(configPath),
							Path:     configPath,
							Status:   "running",
							Modified: info.ModTime(),
							Size:     info.Size(),
							Services: services,
						})
					}
				}
			}
		} else {
			log.Printf("[Compose] Failed to parse docker compose ls output: %v", err)
		}
	} else {
		log.Printf("[Compose] docker compose ls failed: %v", err)
	}
	cancel()

	// 扫描本地 compose 目录
	entries, err := os.ReadDir(dir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if !strings.HasSuffix(name, ".yml") && !strings.HasSuffix(name, ".yaml") {
				continue
			}

			projectName := strings.TrimSuffix(name, filepath.Ext(name))
			// 如果已经在运行列表中，跳过
			if projectMap[projectName] {
				continue
			}

			info, _ := entry.Info()
			path := filepath.Join(dir, name)
			
			// 获取服务数量
			services := h.getServiceCount(path)

			list = append(list, ComposeItem{
				Name:     projectName,
				File:     name,
				Path:     path,
				Status:   "stopped",
				Modified: info.ModTime(),
				Size:     info.Size(),
				Services: services,
			})
		}
	}

	log.Printf("[Compose] Returning %d projects", len(list))
	response.Success(c, gin.H{"projects": list})
}

func (h *ComposeHandler) getServiceCount(configPath string) int {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", configPath, "ps", "--format", "json")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err == nil {
		lines := strings.Split(strings.TrimSpace(out.String()), "\n")
		count := 0
		for _, line := range lines {
			if line != "" {
				count++
			}
		}
		return count
	}
	return 0
}

func (h *ComposeHandler) Get(c *gin.Context) {
	h.ensureDir()
	name := c.Param("name")

	// 安全验证：验证项目名称
	if !isValidComposeName(name) {
		response.BadRequest(c, "无效的项目名称")
		return
	}

	var path string
	var content []byte
	var err error

	// 首先尝试从运行中的项目获取配置路径
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	cmd := exec.CommandContext(ctx, "docker", "compose", "ls", "--format", "json")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err == nil {
		var projects []struct {
			Name   string `json:"Name"`
			Config string `json:"ConfigFiles"`
		}
		if err := json.Unmarshal(out.Bytes(), &projects); err == nil {
			for _, p := range projects {
				if p.Name == name {
					configFiles := strings.Split(p.Config, ",")
					if len(configFiles) > 0 {
						path = strings.TrimSpace(configFiles[0])
						break
					}
				}
			}
		}
	}
	cancel()

	// 如果没有找到运行中的项目，尝试从本地目录查找
	if path == "" {
		path = filepath.Join(h.getComposeDir(), name+".yml")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			path = filepath.Join(h.getComposeDir(), name+".yaml")
		}
	}

	// 读取文件内容
	content, err = os.ReadFile(path)
	if err != nil {
		response.NotFound(c, "Compose file not found")
		return
	}

	type Service struct {
		Name        string `json:"name"`
		Image       string `json:"image"`
		State       string `json:"state"`
		ContainerId string `json:"containerId"`
	}

	var services []Service

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	cmd = exec.CommandContext(ctx, "docker", "compose", "-f", path, "ps", "--format", "json")
	out.Reset()
	cmd.Stdout = &out
	if err := cmd.Run(); err == nil {
		lines := strings.Split(strings.TrimSpace(out.String()), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			var psOutput struct {
				Name    string `json:"Name"`
				Image   string `json:"Image"`
				State   string `json:"State"`
				ID      string `json:"ID"`
				Service string `json:"Service"`
			}
			if err := json.Unmarshal([]byte(line), &psOutput); err == nil {
				services = append(services, Service{
					Name:        psOutput.Service,
					Image:       psOutput.Image,
					State:       psOutput.State,
					ContainerId: psOutput.ID,
				})
			}
		}
	}
	cancel()

	response.Success(c, gin.H{
		"content":  string(content),
		"path":     path,
		"services": services,
	})
}

type ComposeSaveRequest struct {
	Name    string `json:"name" binding:"required"`
	Content string `json:"content" binding:"required"`
}

func (h *ComposeHandler) Save(c *gin.Context) {
	h.ensureDir()

	var req ComposeSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid parameters")
		return
	}

	// 验证name参数，防止命令注入
	if !isValidName(req.Name) {
		response.BadRequest(c, "Invalid name: only letters, numbers, hyphens and underscores are allowed")
		return
	}

	path := filepath.Join(h.getComposeDir(), req.Name+".yml")
	if err := os.WriteFile(path, []byte(req.Content), 0600); err != nil {
		response.InternalError(c, "Failed to save file")
		return
	}

	addAuditLog(c, "compose_save", map[string]interface{}{"name": req.Name})
	response.Success(c, gin.H{"success": true, "path": path})
}

func (h *ComposeHandler) Delete(c *gin.Context) {
	h.ensureDir()
	name := c.Param("name")

	// 验证name参数，防止命令注入
	if !isValidName(name) {
		response.BadRequest(c, "Invalid name: only letters, numbers, hyphens and underscores are allowed")
		return
	}

	// 使用安全路径
	path, err := sanitizePath(h.getComposeDir(), name+".yml")
	if err != nil {
		response.BadRequest(c, "Invalid path")
		return
	}

	// 检查 .yml 文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// 尝试 .yaml 文件
		path, err = sanitizePath(h.getComposeDir(), name+".yaml")
		if err != nil {
			response.BadRequest(c, "Invalid path")
			return
		}
		if _, err := os.Stat(path); os.IsNotExist(err) {
			response.NotFound(c, "Compose file not found")
			return
		}
	}

	os.Remove(path)
	addAuditLog(c, "compose_delete", map[string]interface{}{"name": name})
	response.Success(c, gin.H{"success": true})
}

func (h *ComposeHandler) runCompose(c *gin.Context, name string, args ...string) (string, error) {
	h.ensureDir()

	// 安全验证：验证项目名称
	if !isValidComposeName(name) {
		return "", errors.New("无效的项目名称")
	}

	var path string

	// 首先尝试从运行中的项目获取配置路径
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	cmd := exec.CommandContext(ctx, "docker", "compose", "ls", "--format", "json")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err == nil {
		var projects []struct {
			Name   string `json:"Name"`
			Config string `json:"ConfigFiles"`
		}
		if err := json.Unmarshal(out.Bytes(), &projects); err == nil {
			for _, p := range projects {
				if p.Name == name {
					configFiles := strings.Split(p.Config, ",")
					if len(configFiles) > 0 {
						path = strings.TrimSpace(configFiles[0])
						break
					}
				}
			}
		}
	}
	cancel()

	// 如果没有找到运行中的项目，尝试从本地目录查找（使用安全路径）
	if path == "" {
		var err error
		path, err = sanitizePath(h.getComposeDir(), name+".yml")
		if err != nil {
			return "", err
		}
		// 检查 .yml 文件
		if _, err := os.Stat(path); os.IsNotExist(err) {
			// 尝试 .yaml 文件
			path, err = sanitizePath(h.getComposeDir(), name+".yaml")
			if err != nil {
				return "", err
			}
			if _, err := os.Stat(path); os.IsNotExist(err) {
				return "", os.ErrNotExist
			}
		}
	}

	cmdArgs := []string{"compose", "-f", path}
	cmdArgs = append(cmdArgs, args...)

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cmd = exec.CommandContext(ctx, "docker", cmdArgs...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

type ComposeUpRequest struct {
	Build         bool `json:"build"`
	ForceRecreate bool `json:"forceRecreate"`
}

func (h *ComposeHandler) Up(c *gin.Context) {
	name := c.Param("name")

	var req ComposeUpRequest
	c.ShouldBindJSON(&req)

	args := []string{"up", "-d"}
	if req.Build {
		args = append(args, "--build")
	}
	if req.ForceRecreate {
		args = append(args, "--force-recreate")
	}

	output, err := h.runCompose(c, name, args...)
	if err != nil {
		response.DockerError(c, "Failed to start compose", output)
		return
	}

	addAuditLog(c, "compose_up", map[string]interface{}{"name": name})
	response.Success(c, gin.H{"success": true, "output": output})
}

type ComposeDownRequest struct {
	Volumes       bool `json:"volumes"`
	RemoveOrphans bool `json:"removeOrphans"`
}

func (h *ComposeHandler) Down(c *gin.Context) {
	name := c.Param("name")

	var req ComposeDownRequest
	c.ShouldBindJSON(&req)

	args := []string{"down"}
	if req.Volumes {
		args = append(args, "--volumes")
	}
	if req.RemoveOrphans {
		args = append(args, "--remove-orphans")
	}

	output, err := h.runCompose(c, name, args...)
	if err != nil {
		response.DockerError(c, "Failed to stop compose", output)
		return
	}

	addAuditLog(c, "compose_down", map[string]interface{}{"name": name})
	response.Success(c, gin.H{"success": true, "output": output})
}

func (h *ComposeHandler) Ps(c *gin.Context) {
	name := c.Param("name")

	output, err := h.runCompose(c, name, "ps")
	if err != nil {
		response.DockerError(c, "Failed to get compose status", output)
		return
	}

	response.Success(c, gin.H{"output": output})
}

type ComposeLogsQuery struct {
	Tail   string `form:"tail"`
	Follow bool   `form:"follow"`
}

func (h *ComposeHandler) Logs(c *gin.Context) {
	name := c.Param("name")

	var query ComposeLogsQuery
	c.ShouldBindQuery(&query)

	args := []string{"logs"}
	if query.Tail != "" {
		args = append(args, "--tail", query.Tail)
	}
	if query.Follow {
		args = append(args, "-f")
	}

	output, err := h.runCompose(c, name, args...)
	if err != nil {
		response.DockerError(c, "Failed to get logs", output)
		return
	}

	response.Success(c, gin.H{"logs": output})
}

func (h *ComposeHandler) Pull(c *gin.Context) {
	name := c.Param("name")

	output, err := h.runCompose(c, name, "pull")
	if err != nil {
		response.DockerError(c, "Failed to pull images", output)
		return
	}

	response.Success(c, gin.H{"success": true, "output": output})
}

func (h *ComposeHandler) Build(c *gin.Context) {
	name := c.Param("name")

	output, err := h.runCompose(c, name, "build")
	if err != nil {
		response.DockerError(c, "Failed to build", output)
		return
	}

	response.Success(c, gin.H{"success": true, "output": output})
}

func (h *ComposeHandler) Restart(c *gin.Context) {
	name := c.Param("name")

	output, err := h.runCompose(c, name, "restart")
	if err != nil {
		response.DockerError(c, "Failed to restart", output)
		return
	}

	addAuditLog(c, "compose_restart", map[string]interface{}{"name": name})
	response.Success(c, gin.H{"success": true, "output": output})
}

func (h *ComposeHandler) Stop(c *gin.Context) {
	name := c.Param("name")

	output, err := h.runCompose(c, name, "stop")
	if err != nil {
		response.DockerError(c, "Failed to stop", output)
		return
	}

	addAuditLog(c, "compose_stop", map[string]interface{}{"name": name})
	response.Success(c, gin.H{"success": true, "output": output})
}

func (h *ComposeHandler) Start(c *gin.Context) {
	name := c.Param("name")

	output, err := h.runCompose(c, name, "start")
	if err != nil {
		response.DockerError(c, "Failed to start", output)
		return
	}

	addAuditLog(c, "compose_start", map[string]interface{}{"name": name})
	response.Success(c, gin.H{"success": true, "output": output})
}

type FileHandler struct{}

func NewFileHandler() *FileHandler {
	return &FileHandler{}
}

func (h *FileHandler) List(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		response.BadRequest(c, "Path is required")
		return
	}

	// 验证路径安全性
	validatedPath, err := validatePath(path)
	if err != nil {
		response.Error(c, 403, "ACCESS_DENIED", err.Error())
		return
	}

	entries, err := os.ReadDir(validatedPath)
	if err != nil {
		response.DockerError(c, "Failed to read directory", err.Error())
		return
	}

	type FileInfo struct {
		Name    string    `json:"name"`
		IsDir   bool      `json:"isDir"`
		Size    int64     `json:"size"`
		ModTime time.Time `json:"modTime"`
		Mode    string    `json:"mode"`
	}

	var files []FileInfo
	for _, entry := range entries {
		info, _ := entry.Info()
		files = append(files, FileInfo{
			Name:    entry.Name(),
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
			Mode:    info.Mode().String(),
		})
	}

	response.Success(c, gin.H{"files": files})
}

func (h *FileHandler) Read(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		response.BadRequest(c, "Path is required")
		return
	}

	// 验证路径安全性
	validatedPath, err := validatePath(path)
	if err != nil {
		response.Error(c, 403, "ACCESS_DENIED", err.Error())
		return
	}

	content, err := os.ReadFile(validatedPath)
	if err != nil {
		response.DockerError(c, "Failed to read file", err.Error())
		return
	}

	response.Success(c, gin.H{"content": string(content)})
}

type FileWriteRequest struct {
	Path    string `json:"path" binding:"required"`
	Content string `json:"content"`
}

func (h *FileHandler) Write(c *gin.Context) {
	var req FileWriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid parameters")
		return
	}

	// 验证路径安全性
	validatedPath, err := validatePath(req.Path)
	if err != nil {
		response.Error(c, 403, "ACCESS_DENIED", err.Error())
		return
	}

	if err := os.WriteFile(validatedPath, []byte(req.Content), 0600); err != nil {
		response.DockerError(c, "Failed to write file", err.Error())
		return
	}

	addAuditLog(c, "file_write", map[string]interface{}{"path": validatedPath})
	response.Success(c, gin.H{"success": true})
}

func (h *FileHandler) Delete(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		response.BadRequest(c, "请指定路径")
		return
	}

	// 验证路径安全性
	validatedPath, err := validatePath(path)
	if err != nil {
		response.Error(c, 403, "ACCESS_DENIED", err.Error())
		return
	}

	if err := os.RemoveAll(validatedPath); err != nil {
		response.DockerError(c, "Failed to delete", err.Error())
		return
	}

	addAuditLog(c, "file_delete", map[string]interface{}{"path": validatedPath})
	response.Success(c, gin.H{"success": true})
}

type FileMkdirRequest struct {
	Path string `json:"path" binding:"required"`
}

func (h *FileHandler) Mkdir(c *gin.Context) {
	var req FileMkdirRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid parameters")
		return
	}

	// 验证路径安全性
	validatedPath, err := validatePath(req.Path)
	if err != nil {
		response.Error(c, 403, "ACCESS_DENIED", err.Error())
		return
	}

	if err := os.MkdirAll(validatedPath, 0755); err != nil {
		response.DockerError(c, "Failed to create directory", err.Error())
		return
	}

	addAuditLog(c, "file_mkdir", map[string]interface{}{"path": validatedPath})
	response.Success(c, gin.H{"success": true})
}

type FileRenameRequest struct {
	OldPath string `json:"oldPath" binding:"required"`
	NewPath string `json:"newPath" binding:"required"`
}

func (h *FileHandler) Rename(c *gin.Context) {
	var req FileRenameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid parameters")
		return
	}

	// 验证旧路径安全性
	validatedOldPath, err := validatePath(req.OldPath)
	if err != nil {
		response.Error(c, 403, "ACCESS_DENIED", "old path: "+err.Error())
		return
	}

	// 验证新路径安全性
	validatedNewPath, err := validatePath(req.NewPath)
	if err != nil {
		response.Error(c, 403, "ACCESS_DENIED", "new path: "+err.Error())
		return
	}

	if err := os.Rename(validatedOldPath, validatedNewPath); err != nil {
		response.DockerError(c, "重命名失败", err.Error())
		return
	}

	addAuditLog(c, "file_rename", map[string]interface{}{
		"oldPath": validatedOldPath,
		"newPath": validatedNewPath,
	})
	response.Success(c, gin.H{"success": true})
}

func (h *FileHandler) Upload(c *gin.Context) {
	path := c.PostForm("path")
	if path == "" {
		response.BadRequest(c, "Target path is required")
		return
	}

	// 验证路径安全性
	validatedPath, err := validatePath(path)
	if err != nil {
		response.Error(c, 403, "ACCESS_DENIED", err.Error())
		return
	}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "Failed to get uploaded file")
		return
	}
	defer file.Close()

	dst, err := os.Create(validatedPath)
	if err != nil {
		response.DockerError(c, "Failed to create file", err.Error())
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		response.DockerError(c, "Failed to write file", err.Error())
		return
	}

	addAuditLog(c, "file_upload", map[string]interface{}{"path": validatedPath})
	response.Success(c, gin.H{"success": true})
}

func (h *FileHandler) Download(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		response.BadRequest(c, "Path is required")
		return
	}

	// 验证路径安全性
	validatedPath, err := validatePath(path)
	if err != nil {
		response.Error(c, 403, "ACCESS_DENIED", err.Error())
		return
	}

	file, err := os.Open(validatedPath)
	if err != nil {
		response.DockerError(c, "Failed to open file", err.Error())
		return
	}
	defer file.Close()

	stat, _ := file.Stat()
	reader := bufio.NewReader(file)

	c.DataFromReader(200, stat.Size(), "application/octet-stream", reader, map[string]string{
		"Content-Disposition": "attachment; filename=" + stat.Name(),
	})
}

func (h *FileHandler) CopyToContainer(c *gin.Context) {
	var req struct {
		ContainerID string `json:"containerId" binding:"required"`
		SrcPath     string `json:"srcPath" binding:"required"`
		DestPath    string `json:"destPath" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid parameters")
		return
	}

	ctx, cancel := docker.ContextWithTimeout(2 * time.Minute)
	defer cancel()

	cli := docker.GetClient()
	err := cli.CopyToContainer(ctx, req.ContainerID, req.DestPath, nil, container.CopyToContainerOptions{})
	if err != nil {
		response.DockerError(c, "Failed to copy file to container", err.Error())
		return
	}

	addAuditLog(c, "file_copy_to_container", map[string]interface{}{
		"container": req.ContainerID,
		"src":       req.SrcPath,
		"dest":      req.DestPath,
	})
	response.Success(c, gin.H{"success": true})
}
