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
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/gin-gonic/gin"
)

type ImageHandler struct {
	service      *docker.ImageService
	containerSrv *docker.ContainerService
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

// isValidImageName 验证镜像名称是否合法
func isValidImageName(name string) bool {
	if name == "" || len(name) > 256 {
		return false
	}
	// 只允许字母、数字、-、_、/、.
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') ||
			c == '-' || c == '_' || c == '/' || c == '.') {
			return false
		}
	}
	// 不能以 . 或 - 开头
	if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "-") {
		return false
	}
	return true
}

// isValidSearchTerm 验证搜索关键词是否合法
func isValidSearchTerm(term string) bool {
	if term == "" || len(term) > 128 {
		return false
	}
	// 只允许字母、数字、-、_、/、.、空格
	for _, c := range term {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '_' ||
			c == '/' || c == '.' || c == ' ') {
			return false
		}
	}
	return true
}

// isValidRegistry 验证加速源地址是否合法
func isValidRegistry(registry string) bool {
	if registry == "" || len(registry) > 256 {
		return false
	}
	// 只允许字母、数字、-、_、.、:
	for _, c := range registry {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') ||
			c == '-' || c == '_' || c == '.' || c == ':') {
			return false
		}
	}
	// 必须包含至少一个点（域名格式）
	if !strings.Contains(registry, ".") {
		return false
	}
	return true
}

func (h *ImageHandler) Pull(c *gin.Context) {
	ctx, cancel := docker.ContextWithTimeout(10 * 60 * 1000 * 1000000)
	defer cancel()

	var req ImagePullRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	// 规范化镜像名称（原始名称，不带加速源）
	originalImageName := normalizeImageName(req.Image)

	// 如果配置了镜像加速源,修改镜像名称
	imageName := originalImageName
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

	// 如果使用了加速源，给镜像添加一个不带加速源的标签
	if req.Registry != "" {
		// 从拉取的镜像名称中提取镜像ID或标签
		// 尝试使用原始镜像名称（不带加速源）作为新标签
		cleanTag := originalImageName
		// 移除默认仓库前缀
		cleanTag = strings.TrimPrefix(cleanTag, "docker.io/")
		cleanTag = strings.TrimPrefix(cleanTag, "library/")

		// 给拉取的镜像添加纯净标签
		if err := h.service.Tag(ctx, imageName, cleanTag); err != nil {
			log.Printf("[镜像拉取] 添加纯净标签失败 %s -> %s: %v", imageName, cleanTag, err)
		} else {
			log.Printf("[镜像拉取] 成功添加纯净标签: %s", cleanTag)
		}
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
	registryParam := c.Query("registry")

	if imageNameParam == "" {
		response.BadRequest(c, "镜像名称不能为空")
		return
	}

	// 规范化镜像名称（原始名称，不带加速源）
	originalImageName := normalizeImageName(imageNameParam)

	// 如果配置了镜像加速源,修改镜像名称
	imageName := originalImageName
	if registryParam != "" {
		registry := normalizeRegistry(registryParam)
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

	// 如果使用了加速源，给镜像添加一个不带加速源的标签
	if registryParam != "" {
		// 尝试使用原始镜像名称（不带加速源）作为新标签
		cleanTag := originalImageName
		// 移除默认仓库前缀
		cleanTag = strings.TrimPrefix(cleanTag, "docker.io/")
		cleanTag = strings.TrimPrefix(cleanTag, "library/")

		// 给拉取的镜像添加纯净标签
		if err := h.service.Tag(ctx, imageName, cleanTag); err != nil {
			log.Printf("[镜像拉取] 添加纯净标签失败 %s -> %s: %v", imageName, cleanTag, err)
		} else {
			log.Printf("[镜像拉取] 成功添加纯净标签: %s", cleanTag)
		}
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

	// 安全验证：验证搜索关键词
	if !isValidSearchTerm(term) {
		response.BadRequest(c, "无效的搜索关键词")
		return
	}

	registry := c.Query("registry") // 从 URL 参数获取加速源
	limit := parseInt(c.Query("limit"), 25)

	// 限制最大搜索数量
	if limit > 100 {
		limit = 100
	}

	// 如果配置了加速源，使用 Docker SDK 搜索（支持加速源）
	if registry != "" {
		registry = normalizeRegistry(registry)
		// 安全验证：验证加速源地址
		if !isValidRegistry(registry) {
			response.BadRequest(c, "无效的加速源地址")
			return
		}
		// 构建带加速源的搜索词，如: docker.1ms.run/nginx
		searchTerm := registry + "/" + term

		ctx, cancel := docker.ContextWithTimeout(30 * 1000 * 1000000)
		defer cancel()

		results, err := h.service.Search(ctx, searchTerm, types.ImageSearchOptions{Limit: limit})
		if err != nil {
			log.Printf("[镜像搜索] Docker SDK 搜索失败: %v", err)
			response.DockerError(c, "搜索镜像失败", "搜索服务暂时不可用")
			return
		}

		// 转换结果格式
		var searchResults []map[string]interface{}
		for _, r := range results {
			// 返回带加速源前缀的完整名称，方便用户直接拉取
			fullName := registry + "/" + r.Name
			searchResults = append(searchResults, map[string]interface{}{
				"name":        fullName,
				"description": r.Description,
				"star_count":  r.StarCount,
				"is_official": r.IsOfficial,
			})
		}

		response.Success(c, gin.H{"results": searchResults})
		return
	}

	// 没有配置加速源，使用 Docker Hub API
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
		log.Printf("[镜像搜索] HTTP错误 %d", resp.StatusCode)
		response.DockerError(c, "搜索镜像失败", fmt.Sprintf("HTTP %d", resp.StatusCode))
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
	registry := c.Query("registry") // 从 URL 参数获取加速源

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

	// 如果配置了加速源，构建带加速源的远程镜像名
	remoteImage := localImage
	if registry != "" {
		registry = normalizeRegistry(registry)
		// 移除默认仓库前缀
		imageName := strings.TrimPrefix(localImage, "docker.io/")
		imageName = strings.TrimPrefix(imageName, "library/")
		// 检查是否已包含自定义仓库
		parts := strings.Split(imageName, "/")
		if len(parts) > 1 && (strings.Contains(parts[0], ".") || strings.Contains(parts[0], ":")) {
			imageName = strings.Join(parts[1:], "/")
		}
		remoteImage = registry + "/" + imageName
	}

	hasUpdate, newDigest, err := h.service.CheckUpdate(ctx, localImage, remoteImage)
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

	// 从请求体获取加速源参数
	var req struct {
		Registry string `json:"registry"`
	}
	c.ShouldBindJSON(&req)

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

	// 如果配置了加速源，构建带加速源的镜像名
	pullImage := imageName
	if req.Registry != "" {
		registry := normalizeRegistry(req.Registry)
		// 移除默认仓库前缀
		cleanImage := strings.TrimPrefix(imageName, "docker.io/")
		cleanImage = strings.TrimPrefix(cleanImage, "library/")
		// 检查是否已包含自定义仓库
		parts := strings.Split(cleanImage, "/")
		if len(parts) > 1 && (strings.Contains(parts[0], ".") || strings.Contains(parts[0], ":")) {
			cleanImage = strings.Join(parts[1:], "/")
		}
		pullImage = registry + "/" + cleanImage
	}

	output, err := h.service.Pull(ctx, pullImage, image.PullOptions{})
	if err != nil {
		response.DockerError(c, "更新镜像失败", err.Error())
		return
	}

	addAuditLog(c, "image_update", map[string]interface{}{"image": imageName, "registry": req.Registry})
	response.Success(c, gin.H{"success": true, "output": string(output), "image": imageName})
}

// GetTags 获取镜像的所有标签
func (h *ImageHandler) GetTags(c *gin.Context) {
	imageName := c.Query("image")
	if imageName == "" {
		response.BadRequest(c, "请输入镜像名称")
		return
	}

	// 移除可能存在的标签部分
	if idx := strings.LastIndex(imageName, ":"); idx > strings.LastIndex(imageName, "/") {
		imageName = imageName[:idx]
	}

	// 移除加速源前缀（如果有）
	registry := c.Query("registry")
	if registry != "" {
		registry = normalizeRegistry(registry)
		// 如果镜像名包含加速源前缀，移除它
		imageName = strings.TrimPrefix(imageName, registry+"/")
	}

	// 如果镜像名还包含域名前缀（如 docker.1ms.run/nginx），移除它
	// Docker Hub 官方镜像格式: library/nginx 或 nginx
	// 第三方镜像格式: user/image
	if strings.Contains(imageName, "/") {
		parts := strings.Split(imageName, "/")
		// 如果第一部分看起来像域名（包含 . 或 :），移除它
		if len(parts) > 1 && (strings.Contains(parts[0], ".") || strings.Contains(parts[0], ":")) {
			imageName = strings.Join(parts[1:], "/")
		}
	}

	// 安全验证：只允许合法的镜像名字符（字母、数字、-、_、/、.）
	if !isValidImageName(imageName) {
		response.BadRequest(c, "无效的镜像名称")
		return
	}

	// 优先使用加速源获取标签
	if registry != "" {
		// 通过加速源的 Registry API 获取标签
		tags, err := h.getTagsFromRegistry(registry, imageName)
		if err == nil && len(tags) > 0 {
			response.Success(c, gin.H{"tags": tags})
			return
		}
		log.Printf("[镜像标签] 从加速源获取失败: %v，尝试 Docker Hub", err)
	}

	// 尝试 Docker Hub API
	client := &http.Client{Timeout: 10 * 1000000000}
	apiURL := "https://registry.hub.docker.com/v2/repositories/" + url.PathEscape(imageName) + "/tags/?page_size=50"

	resp, err := client.Get(apiURL)
	if err != nil {
		log.Printf("[镜像标签] 获取失败: %v", err)
		response.DockerError(c, "获取镜像标签失败", "网络连接错误")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("[镜像标签] HTTP错误 %d", resp.StatusCode)
		response.DockerError(c, "获取镜像标签失败", fmt.Sprintf("HTTP %d", resp.StatusCode))
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		response.DockerError(c, "读取响应失败", err.Error())
		return
	}

	// 解析响应
	var tagsResp struct {
		Results []struct {
			Name string `json:"name"`
		} `json:"results"`
	}

	if err := json.Unmarshal(body, &tagsResp); err != nil {
		response.DockerError(c, "解析响应失败", err.Error())
		return
	}

	// 提取标签名
	var tags []string
	for _, t := range tagsResp.Results {
		if t.Name != "" {
			tags = append(tags, t.Name)
		}
	}

	response.Success(c, gin.H{"tags": tags})
}

// getTagsFromRegistry 从 Registry 获取镜像标签
func (h *ImageHandler) getTagsFromRegistry(registry, imageName string) ([]string, error) {
	// 构建 Registry API URL
	// 格式: https://registry.example.com/v2/library/nginx/tags/list
	client := &http.Client{Timeout: 10 * 1000000000}

	// 对于官方镜像，需要添加 library/ 前缀
	if !strings.Contains(imageName, "/") {
		imageName = "library/" + imageName
	}

	apiURL := fmt.Sprintf("https://%s/v2/%s/tags/list", registry, imageName)

	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tagsResp struct {
		Tags []string `json:"tags"`
	}

	if err := json.Unmarshal(body, &tagsResp); err != nil {
		return nil, err
	}

	return tagsResp.Tags, nil
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

// EditTagsRequest 编辑标签请求
type EditTagsRequest struct {
	Tags []string `json:"tags" binding:"required"`
}

// EditTags 编辑镜像标签
func (h *ImageHandler) EditTags(c *gin.Context) {
	ctx, cancel := docker.Context()
	defer cancel()

	id := c.Param("id")

	var req EditTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 验证标签格式
	for _, tag := range req.Tags {
		if !isValidImageName(tag) {
			response.BadRequest(c, "无效的标签格式: "+tag)
			return
		}
	}

	// 获取镜像信息
	inspect, err := h.service.Inspect(ctx, id)
	if err != nil {
		response.DockerError(c, "获取镜像信息失败", err.Error())
		return
	}

	// 获取当前镜像ID
	imageID := inspect.ID

	// 检查新标签是否与旧标签相同
	if len(req.Tags) > 0 && len(inspect.RepoTags) > 0 {
		oldTag := inspect.RepoTags[0]
		newTag := req.Tags[0]
		if oldTag == newTag {
			response.Success(c, gin.H{"success": true, "message": "标签未改变"})
			return
		}
	}

	// 删除旧标签（保留第一个作为基础）
	if len(inspect.RepoTags) > 0 {
		for i, tag := range inspect.RepoTags {
			if i == 0 && len(req.Tags) > 0 {
				// 保留第一个标签，后续会重新标记
				continue
			}
			if err := h.service.RemoveTag(ctx, tag); err != nil {
				log.Printf("[编辑标签] 删除旧标签失败 %s: %v", tag, err)
			}
		}
	}

	// 添加新标签
	for _, tag := range req.Tags {
		if err := h.service.Tag(ctx, imageID, tag); err != nil {
			response.DockerError(c, "添加标签失败: "+tag, err.Error())
			return
		}
	}

	addAuditLog(c, "image_edit_tags", map[string]interface{}{"image": id, "tags": req.Tags})
	response.Success(c, gin.H{"success": true, "tags": req.Tags})
}

// DetectUpgrade 检测镜像可升级版本
func (h *ImageHandler) DetectUpgrade(c *gin.Context) {
	ctx, cancel := docker.ContextWithTimeout(30 * 1000 * 1000000)
	defer cancel()

	id := c.Param("id")

	inspect, err := h.service.Inspect(ctx, id)
	if err != nil {
		response.DockerError(c, "获取镜像信息失败", err.Error())
		return
	}

	// 获取当前镜像名称和标签
	imageName := ""
	currentTag := "latest"
	if len(inspect.RepoTags) > 0 && inspect.RepoTags[0] != "<none>:<none>" {
		parts := strings.Split(inspect.RepoTags[0], ":")
		if len(parts) >= 2 {
			imageName = parts[0]
			currentTag = parts[1]
		} else {
			imageName = inspect.RepoTags[0]
		}
	} else if len(inspect.RepoDigests) > 0 {
		parts := strings.SplitN(inspect.RepoDigests[0], "@", 2)
		if len(parts) > 0 {
			imageName = parts[0]
		}
	}

	if imageName == "" {
		response.Success(c, gin.H{"available": false, "message": "无法确定镜像名称"})
		return
	}

	// 验证镜像名称格式
	if !isValidImageName(imageName) {
		response.Success(c, gin.H{"available": false, "message": "无效的镜像名称"})
		return
	}

	// 获取镜像加速源
	registry := c.Query("registry")

	// 获取所有可用标签
	var allTags []string
	if registry != "" {
		registry = normalizeRegistry(registry)
		allTags, err = h.getTagsFromRegistry(registry, imageName)
		if err != nil {
			log.Printf("[检测升级] 从加速源获取标签失败: %v", err)
		}
	}

	// 如果加速源失败或未配置，尝试 Docker Hub
	if len(allTags) == 0 {
		client := &http.Client{Timeout: 10 * 1000000000}
		apiURL := "https://registry.hub.docker.com/v2/repositories/" + url.PathEscape(imageName) + "/tags/?page_size=100"

		resp, err := client.Get(apiURL)
		if err != nil {
			response.Success(c, gin.H{"available": false, "message": "无法连接到镜像仓库"})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			response.Success(c, gin.H{"available": false, "message": "镜像仓库返回错误: " + fmt.Sprintf("HTTP %d", resp.StatusCode)})
			return
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			response.Success(c, gin.H{"available": false, "message": "读取响应失败"})
			return
		}

		var tagsResp struct {
			Results []struct {
				Name string `json:"name"`
			} `json:"results"`
		}

		if err := json.Unmarshal(body, &tagsResp); err != nil {
			response.Success(c, gin.H{"available": false, "message": "解析响应失败"})
			return
		}

		for _, t := range tagsResp.Results {
			if t.Name != "" {
				allTags = append(allTags, t.Name)
			}
		}
	}

	if len(allTags) == 0 {
		response.Success(c, gin.H{"available": false, "message": "未找到可用标签"})
		return
	}

	// 分析版本，找出可升级的版本
	upgradableVersions := analyzeVersions(currentTag, allTags)

	response.Success(c, gin.H{
		"available":          true,
		"currentTag":         currentTag,
		"imageName":          imageName,
		"allTags":            allTags[:min(20, len(allTags))],
		"upgradableVersions": upgradableVersions,
	})
}

// analyzeVersions 分析版本，找出可升级的版本
func analyzeVersions(currentTag string, allTags []string) []string {
	// 如果当前标签不是版本号格式，返回空
	if !isVersionTag(currentTag) {
		return []string{}
	}

	var upgradable []string
	currentVersion := parseVersion(currentTag)

	for _, tag := range allTags {
		if isVersionTag(tag) {
			tagVersion := parseVersion(tag)
			if compareVersions(tagVersion, currentVersion) > 0 {
				upgradable = append(upgradable, tag)
			}
		}
	}

	// 按版本号排序
	sortVersions(upgradable)

	return upgradable[:min(10, len(upgradable))]
}

// isVersionTag 检查标签是否是版本号格式
func isVersionTag(tag string) bool {
	// 匹配常见的版本格式: 1.0, 1.0.0, v1.0, v1.0.0, 1.0-alpine, 1.0.0-alpine3.18 等
	matched, _ := regexp.MatchString(`^v?\d+(\.\d+)*([\-\+].*)?$`, tag)
	return matched
}

// parseVersion 解析版本号为可比较的数组
func parseVersion(tag string) []int {
	// 移除 v 前缀
	tag = strings.TrimPrefix(tag, "v")
	// 移除 - 后面的内容（如 -alpine）
	if idx := strings.Index(tag, "-"); idx > 0 {
		tag = tag[:idx]
	}
	if idx := strings.Index(tag, "+"); idx > 0 {
		tag = tag[:idx]
	}

	parts := strings.Split(tag, ".")
	var version []int
	for _, p := range parts {
		if n, err := strconv.Atoi(p); err == nil {
			version = append(version, n)
		}
	}
	return version
}

// compareVersions 比较两个版本号
// 返回: 1 表示 v1 > v2, -1 表示 v1 < v2, 0 表示相等
func compareVersions(v1, v2 []int) int {
	maxLen := len(v1)
	if len(v2) > maxLen {
		maxLen = len(v2)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(v1) {
			n1 = v1[i]
		}
		if i < len(v2) {
			n2 = v2[i]
		}

		if n1 > n2 {
			return 1
		}
		if n1 < n2 {
			return -1
		}
	}
	return 0
}

// sortVersions 对版本号进行排序（降序）
func sortVersions(versions []string) {
	sort.Slice(versions, func(i, j int) bool {
		vi := parseVersion(versions[i])
		vj := parseVersion(versions[j])
		return compareVersions(vi, vj) > 0
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
