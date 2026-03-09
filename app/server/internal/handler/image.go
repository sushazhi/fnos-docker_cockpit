package handler

import (
	"bufio"
	"dockpit/pkg/docker"
	"dockpit/pkg/response"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/gin-gonic/gin"
)

type ImageHandler struct {
	service       *docker.ImageService
	containerSrv  *docker.ContainerService
}

func NewImageHandler() *ImageHandler {
	return &ImageHandler{
		service:      docker.NewImageService(),
		containerSrv: docker.NewContainerService(),
	}
}

func (h *ImageHandler) List(c *gin.Context) {
	ctx, cancel := docker.Context()
	defer cancel()

	// 获取镜像列表
	images, err := h.service.List(ctx, image.ListOptions{})
	if err != nil {
		response.DockerError(c, "获取镜像列表失败", err.Error())
		return
	}

	// 获取所有容器（包括停止的），用于计算镜像使用数
	containers, err := h.containerSrv.List(ctx, true)
	if err != nil {
		// 如果获取容器失败，仍然返回镜像列表，只是不计算使用数
		log.Printf("[镜像列表] 获取容器列表失败: %v", err)
		response.Success(c, gin.H{"images": images})
		return
	}

	// 构建镜像ID到容器数量的映射
	imageContainerCount := make(map[string]int64)
	for _, container := range containers {
		if container.ImageID != "" {
			imageContainerCount[container.ImageID]++
		}
	}

	// 更新镜像的 Containers 字段
	for i := range images {
		if count, ok := imageContainerCount[images[i].ID]; ok {
			images[i].Containers = count
		} else {
			images[i].Containers = 0
		}
	}

	response.Success(c, gin.H{"images": images})
}

func (h *ImageHandler) Get(c *gin.Context) {
	ctx, cancel := docker.Context()
	defer cancel()

	id := c.Param("id")
	info, err := h.service.Inspect(ctx, id)
	if err != nil {
		response.DockerError(c, "获取镜像信息失败", err.Error())
		return
	}

	response.Success(c, gin.H{"info": info})
}

type ImagePullRequest struct {
	Image    string `json:"image" binding:"required"`
	Registry string `json:"registry"` // 镜像加速源地址
}

// normalizeImageName 清理和规范化镜像名称
func normalizeImageName(name string) string {
	// 移除前后空白
	name = strings.TrimSpace(name)
	// 移除开头的斜杠
	name = strings.TrimPrefix(name, "/")
	// 移除结尾的斜杠
	name = strings.TrimSuffix(name, "/")
	// 替换连续的斜杠为单个斜杠
	for strings.Contains(name, "//") {
		name = strings.ReplaceAll(name, "//", "/")
	}
	// 转换为小写（Docker 镜像名称必须是小写）
	name = strings.ToLower(name)
	return name
}

// normalizeRegistry 清理镜像加速源地址，移除协议前缀
func normalizeRegistry(registry string) string {
	// 移除反引号和其他特殊字符
	registry = strings.ReplaceAll(registry, "`", "")
	registry = strings.ReplaceAll(registry, "'", "")
	registry = strings.ReplaceAll(registry, "\"", "")
	// 移除前后空白
	registry = strings.TrimSpace(registry)
	// 移除 http:// 或 https:// 前缀（必须在替换连续斜杠之前）
	registry = strings.TrimPrefix(registry, "http://")
	registry = strings.TrimPrefix(registry, "https://")
	// 移除开头的斜杠
	registry = strings.TrimPrefix(registry, "/")
	// 移除结尾的斜杠
	registry = strings.TrimSuffix(registry, "/")
	// 转换为小写
	registry = strings.ToLower(registry)
	return registry
}

func (h *ImageHandler) Pull(c *gin.Context) {
	ctx, cancel := docker.ContextWithTimeout(10 * 60 * 1000 * 1000000)
	defer cancel()

	var req ImagePullRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	// 规范化镜像名称
	imageName := normalizeImageName(req.Image)

	// 如果配置了镜像加速源,修改镜像名称
	if req.Registry != "" {
		// 解析镜像名称，分离仓库地址、镜像名和标签
		// 镜像名称格式: [registry/][namespace/]name[:tag]
		registry := normalizeRegistry(req.Registry)

		// 移除默认仓库前缀
		imageName = strings.TrimPrefix(imageName, "docker.io/")
		imageName = strings.TrimPrefix(imageName, "library/")

		// 检查镜像名称是否已经包含了其他仓库地址
		// 如果包含斜杠，说明可能有命名空间或自定义仓库
		parts := strings.Split(imageName, "/")
		if len(parts) > 1 && (strings.Contains(parts[0], ".") || strings.Contains(parts[0], ":")) {
			// 第一部分包含点或冒号，说明是自定义仓库地址，需要替换
			imageName = strings.Join(parts[1:], "/")
		}

		// 添加镜像加速源前缀
		imageName = registry + "/" + imageName
	}
	output, err := h.service.Pull(ctx, imageName, image.PullOptions{})
	if err != nil {
		response.DockerError(c, "拉取镜像失败", err.Error())
		return
	}

	outputStr := string(output)
	log.Printf("[镜像拉取] 拉取输出: %s", outputStr)

	// 检查输出中是否包含错误信息
	if strings.Contains(outputStr, "error") || strings.Contains(outputStr, "Error") {
		response.DockerError(c, "拉取镜像失败", outputStr)
		return
	}

	addAuditLog(c, "image_pull", map[string]interface{}{"image": req.Image, "registry": req.Registry})
	response.Success(c, gin.H{"success": true, "output": outputStr})
}

// PullStream 使用 SSE 实时推送拉取进度
func (h *ImageHandler) PullStream(c *gin.Context) {
	ctx, cancel := docker.ContextWithTimeout(10 * 60 * 1000 * 1000000)
	defer cancel()

	// 从 URL 参数获取镜像名称和加速源
	imageNameParam := c.Query("image")
	registry := c.Query("registry")

	if imageNameParam == "" {
		response.BadRequest(c, "镜像名称不能为空")
		return
	}

	// 规范化镜像名称
	imageName := normalizeImageName(imageNameParam)

	// 如果配置了镜像加速源,修改镜像名称
	if registry != "" {
		registry = normalizeRegistry(registry)
		imageName = registry + "/" + imageName
	}

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// 获取拉取 reader
	reader, err := h.service.PullReader(ctx, imageName, image.PullOptions{})
	if err != nil {
		c.SSEvent("error", gin.H{"message": err.Error()})
		return
	}
	defer reader.Close()

	// 使用 scanner 逐行读取进度
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			c.SSEvent("progress", gin.H{"data": line})
			c.Writer.Flush()
		}
	}

	if err := scanner.Err(); err != nil {
		c.SSEvent("error", gin.H{"message": err.Error()})
		return
	}

	c.SSEvent("complete", gin.H{"message": "拉取完成"})
}

type ImagePushRequest struct {
	Image string `json:"image" binding:"required"`
}

func (h *ImageHandler) Push(c *gin.Context) {
	ctx, cancel := docker.ContextWithTimeout(10 * 60 * 1000 * 1000000)
	defer cancel()

	var req ImagePushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	output, err := h.service.Push(ctx, req.Image, image.PushOptions{})
	if err != nil {
		response.DockerError(c, "推送镜像失败", err.Error())
		return
	}

	addAuditLog(c, "image_push", map[string]interface{}{"image": req.Image})
	response.Success(c, gin.H{"success": true, "output": string(output)})
}

type ImageRemoveRequest struct {
	Force bool `json:"force"`
}

func (h *ImageHandler) Remove(c *gin.Context) {
	ctx, cancel := docker.ContextWithTimeout(2 * 60 * 1000 * 1000000)
	defer cancel()

	id := c.Param("id")
	var req ImageRemoveRequest
	c.ShouldBindJSON(&req)

	_, err := h.service.Remove(ctx, id, image.RemoveOptions{Force: req.Force})
	if err != nil {
		response.DockerError(c, "删除镜像失败", err.Error())
		return
	}

	addAuditLog(c, "image_remove", map[string]interface{}{"id": id})
	response.Success(c, gin.H{"success": true})
}

type ImageTagRequest struct {
	Source string `json:"source" binding:"required"`
	Target string `json:"target" binding:"required"`
}

func (h *ImageHandler) Tag(c *gin.Context) {
	ctx, cancel := docker.Context()
	defer cancel()

	var req ImageTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := h.service.Tag(ctx, req.Source, req.Target); err != nil {
		response.DockerError(c, "标记镜像失败", err.Error())
		return
	}

	response.Success(c, gin.H{"success": true})
}

// CheckSearchAvailable 检查Docker Hub搜索API是否可用
func (h *ImageHandler) CheckSearchAvailable(c *gin.Context) {
	client := &http.Client{Timeout: 5 * 1000000000} // 5秒超时

	// 尝试访问Docker Hub搜索API
	apiURL := "https://registry.hub.docker.com/v2/repositories/search/?query=test&page_size=1"
	resp, err := client.Get(apiURL)
	if err != nil {
		log.Printf("[镜像搜索检测] Docker Hub不可访问: %v", err)
		response.Success(c, gin.H{"available": false})
		return
	}
	defer resp.Body.Close()

	available := resp.StatusCode == 200
	log.Printf("[镜像搜索检测] Docker Hub状态: %d, 可用: %v", resp.StatusCode, available)
	response.Success(c, gin.H{"available": available})
}

func (h *ImageHandler) Search(c *gin.Context) {
	term := c.Query("term")
	if term == "" {
		response.BadRequest(c, "请输入搜索关键词")
		return
	}

	limit := parseInt(c.Query("limit"), 25)

	// 只使用官方Docker Hub API
	client := &http.Client{Timeout: 10 * 1000000000} // 10秒超时
	apiURL := "https://registry.hub.docker.com/v2/repositories/search/?query=" + url.QueryEscape(term) + "&page_size=" + strconv.Itoa(limit)

	resp, err := client.Get(apiURL)
	if err != nil {
		log.Printf("[镜像搜索] 连接失败: %v", err)
		response.DockerError(c, "无法连接到Docker Hub", "请检查网络连接或使用代理")
		return
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[镜像搜索] HTTP错误 %d: %s", resp.StatusCode, string(body))
		response.DockerError(c, "搜索镜像失败", fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)))
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[镜像搜索] 读取响应失败: %v", err)
		response.DockerError(c, "读取搜索结果失败", err.Error())
		return
	}

	// 尝试多种响应格式解析
	var results []map[string]interface{}

	// 格式1: Docker Hub API格式 {"results": [...]}
	var dockerHubResp struct {
		Results []struct {
			Name        string `json:"repo_name"`
			Description string `json:"short_description"`
			Stars       int    `json:"star_count"`
			Official    bool   `json:"is_official"`
		} `json:"results"`
	}

	if err := json.Unmarshal(body, &dockerHubResp); err == nil && len(dockerHubResp.Results) > 0 {
		for _, r := range dockerHubResp.Results {
			results = append(results, map[string]interface{}{
				"name":        r.Name,
				"description": r.Description,
				"star_count":  r.Stars,
				"is_official": r.Official,
			})
		}
	} else {
		// 格式2: 直接数组格式 [...]
		var directArray []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Stars       int    `json:"star_count"`
			Official    bool   `json:"is_official"`
		}

		if err := json.Unmarshal(body, &directArray); err == nil && len(directArray) > 0 {
			for _, r := range directArray {
				results = append(results, map[string]interface{}{
					"name":        r.Name,
					"description": r.Description,
					"star_count":  r.Stars,
					"is_official": r.Official,
				})
			}
		} else {
			// 格式3: 其他可能的格式，尝试通用解析
			var genericData interface{}
			if err := json.Unmarshal(body, &genericData); err != nil {
				log.Printf("[镜像搜索] JSON解析失败: %v", err)
				response.DockerError(c, "解析搜索结果失败", err.Error())
				return
			}

			log.Printf("[镜像搜索] 不支持的响应格式: %T", genericData)
			response.DockerError(c, "不支持的响应格式", "镜像源返回的数据格式无法识别")
			return
		}
	}

	response.Success(c, gin.H{"results": results})
}

func (h *ImageHandler) Prune(c *gin.Context) {
	ctx, cancel := docker.ContextWithTimeout(2 * 60 * 1000 * 1000000)
	defer cancel()

	report, err := h.service.Prune(ctx)
	if err != nil {
		response.DockerError(c, "清理镜像失败", err.Error())
		return
	}

	addAuditLog(c, "image_prune", nil)
	response.Success(c, gin.H{
		"success":        true,
		"imagesDeleted":  report.ImagesDeleted,
		"spaceReclaimed": report.SpaceReclaimed,
	})
}

func (h *ImageHandler) History(c *gin.Context) {
	ctx, cancel := docker.Context()
	defer cancel()

	id := c.Param("id")
	history, err := h.service.History(ctx, id)
	if err != nil {
		response.DockerError(c, "获取镜像历史失败", err.Error())
		return
	}

	response.Success(c, gin.H{"history": history})
}

func (h *ImageHandler) CheckUpdate(c *gin.Context) {
	ctx, cancel := docker.ContextWithTimeout(3 * 60 * 1000 * 1000000)
	defer cancel()

	id := c.Param("id")

	inspect, err := h.service.Inspect(ctx, id)
	if err != nil {
		response.DockerError(c, "获取镜像信息失败", err.Error())
		return
	}

	localImage := ""
	if len(inspect.RepoTags) > 0 && inspect.RepoTags[0] != "<none>:<none>" {
		localImage = inspect.RepoTags[0]
	} else if len(inspect.RepoDigests) > 0 {
		parts := strings.SplitN(inspect.RepoDigests[0], "@", 2)
		if len(parts) > 0 {
			localImage = parts[0] + ":latest"
		}
	}

	if localImage == "" {
		response.Success(c, gin.H{"hasUpdate": false, "message": "无法确定镜像名称"})
		return
	}

	hasUpdate, newDigest, err := h.service.CheckUpdate(ctx, localImage, localImage)
	if err != nil {
		response.Success(c, gin.H{
			"hasUpdate":  false,
			"error":      err.Error(),
			"localImage": localImage,
		})
		return
	}

	response.Success(c, gin.H{
		"hasUpdate":  hasUpdate,
		"newDigest":  newDigest,
		"localImage": localImage,
	})
}

func (h *ImageHandler) Update(c *gin.Context) {
	ctx, cancel := docker.ContextWithTimeout(10 * 60 * 1000 * 1000000)
	defer cancel()

	id := c.Param("id")

	inspect, err := h.service.Inspect(ctx, id)
	if err != nil {
		response.DockerError(c, "获取镜像信息失败", err.Error())
		return
	}

	imageName := ""
	if len(inspect.RepoTags) > 0 && inspect.RepoTags[0] != "<none>:<none>" {
		imageName = inspect.RepoTags[0]
	} else if len(inspect.RepoDigests) > 0 {
		parts := strings.SplitN(inspect.RepoDigests[0], "@", 2)
		if len(parts) > 0 {
			imageName = parts[0] + ":latest"
		}
	}

	if imageName == "" {
		response.BadRequest(c, "无法确定镜像名称")
		return
	}

	output, err := h.service.Pull(ctx, imageName, image.PullOptions{})
	if err != nil {
		response.DockerError(c, "更新镜像失败", err.Error())
		return
	}

	addAuditLog(c, "image_update", map[string]interface{}{"image": imageName})
	response.Success(c, gin.H{"success": true, "output": string(output), "image": imageName})
}

type ImageBuildRequest struct {
	Tag        string            `json:"tag"`
	Dockerfile string            `json:"dockerfile"`
	Path       string            `json:"path"`
	BuildArgs  map[string]string `json:"buildArgs"`
	NoCache    bool              `json:"noCache"`
	Pull       bool              `json:"pull"`
}

func (h *ImageHandler) Build(c *gin.Context) {
	ctx, cancel := docker.ContextWithTimeout(10 * 60 * 1000 * 1000000)
	defer cancel()

	var req ImageBuildRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	buildArgs := make(map[string]*string)
	for k, v := range req.BuildArgs {
		buildArgs[k] = &v
	}

	cli := docker.GetClient()
	buildResp, err := cli.ImageBuild(ctx, nil, types.ImageBuildOptions{
		Tags:       []string{req.Tag},
		Dockerfile: req.Dockerfile,
		BuildArgs:  buildArgs,
		NoCache:    req.NoCache,
		PullParent: req.Pull,
	})
	if err != nil {
		response.DockerError(c, "构建镜像失败", err.Error())
		return
	}
	defer buildResp.Body.Close()

	addAuditLog(c, "image_build", map[string]interface{}{"tag": req.Tag})
	response.Success(c, gin.H{"success": true})
}
