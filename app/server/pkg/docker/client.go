package docker

import (
	"context"
	"dockpit/internal/config"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

var cli *client.Client

type Info = system.Info

func Init() error {
	var err error
	cli, err = client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	return err
}

func GetClient() *client.Client {
	return cli
}

func IsInitialized() bool {
	return cli != nil
}

func Ping(ctx context.Context) error {
	_, err := cli.Ping(ctx)
	return err
}

type ContainerService struct{}

func NewContainerService() *ContainerService {
	return &ContainerService{}
}

func (s *ContainerService) List(ctx context.Context, all bool) ([]types.Container, error) {
	return cli.ContainerList(ctx, container.ListOptions{All: all})
}

func (s *ContainerService) Inspect(ctx context.Context, id string) (types.ContainerJSON, error) {
	return cli.ContainerInspect(ctx, id)
}

func (s *ContainerService) Start(ctx context.Context, id string) error {
	return cli.ContainerStart(ctx, id, container.StartOptions{})
}

func (s *ContainerService) Stop(ctx context.Context, id string, timeout *int) error {
	return cli.ContainerStop(ctx, id, container.StopOptions{Timeout: timeout})
}

func (s *ContainerService) Restart(ctx context.Context, id string, timeout *int) error {
	return cli.ContainerRestart(ctx, id, container.StopOptions{Timeout: timeout})
}

func (s *ContainerService) Pause(ctx context.Context, id string) error {
	return cli.ContainerPause(ctx, id)
}

func (s *ContainerService) Unpause(ctx context.Context, id string) error {
	return cli.ContainerUnpause(ctx, id)
}

func (s *ContainerService) Remove(ctx context.Context, id string, force, removeVolumes bool) error {
	return cli.ContainerRemove(ctx, id, container.RemoveOptions{
		Force:         force,
		RemoveVolumes: removeVolumes,
	})
}

func (s *ContainerService) Logs(ctx context.Context, id string, opts container.LogsOptions) ([]byte, error) {
	reader, err := cli.ContainerLogs(ctx, id, opts)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// 读取所有日志数据
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	// Docker 容器日志是二进制格式，每行前面有 8 字节的头部
	// 格式: [stream_type:1][padding:3][size:4][payload:size]
	// stream_type: 0=stdin, 1=stdout, 2=stderr
	// 需要解析并提取实际的日志内容
	var result strings.Builder
	for i := 0; i < len(data); {
		// 确保有足够的字节读取头部
		if i+8 > len(data) {
			break
		}

		// 读取 payload 大小（大端序）
		size := int(data[i+4])<<24 | int(data[i+5])<<16 | int(data[i+6])<<8 | int(data[i+7])

		// 检查 size 是否有效
		if size <= 0 || size > 1024*1024 { // 最大 1MB 每行
			// 可能是原始文本格式，直接返回
			return data, nil
		}

		// 确保有足够的字节读取 payload
		if i+8+size > len(data) {
			break
		}

		// 提取 payload
		payload := data[i+8 : i+8+size]
		result.Write(payload)

		// 移动到下一个记录
		i += 8 + size
	}

	if result.Len() == 0 {
		// 如果没有解析出内容，可能是原始格式，直接返回
		return data, nil
	}

	return []byte(result.String()), nil
}

func (s *ContainerService) Stats(ctx context.Context, id string) (container.StatsResponseReader, error) {
	return cli.ContainerStats(ctx, id, false)
}

func (s *ContainerService) Create(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, name string) (container.CreateResponse, error) {
	return cli.ContainerCreate(ctx, config, hostConfig, networkingConfig, nil, name)
}

func (s *ContainerService) Rename(ctx context.Context, id, newName string) error {
	return cli.ContainerRename(ctx, id, newName)
}

func (s *ContainerService) Update(ctx context.Context, id string, updateConfig container.UpdateConfig) error {
	_, err := cli.ContainerUpdate(ctx, id, updateConfig)
	return err
}

func (s *ContainerService) Commit(ctx context.Context, id string, options container.CommitOptions) (types.IDResponse, error) {
	return cli.ContainerCommit(ctx, id, options)
}

func (s *ContainerService) ExecCreate(ctx context.Context, id string, opts container.ExecOptions) (types.IDResponse, error) {
	return cli.ContainerExecCreate(ctx, id, opts)
}

func (s *ContainerService) ExecAttach(ctx context.Context, execID string, opts container.ExecAttachOptions) (types.HijackedResponse, error) {
	return cli.ContainerExecAttach(ctx, execID, opts)
}

func (s *ContainerService) ExecResize(ctx context.Context, execID string, opts container.ResizeOptions) error {
	return cli.ContainerExecResize(ctx, execID, opts)
}

type ImageService struct{}

func NewImageService() *ImageService {
	return &ImageService{}
}

func (s *ImageService) List(ctx context.Context, opts image.ListOptions) ([]image.Summary, error) {
	return cli.ImageList(ctx, opts)
}

func (s *ImageService) Inspect(ctx context.Context, id string) (types.ImageInspect, error) {
	inspect, _, err := cli.ImageInspectWithRaw(ctx, id)
	return inspect, err
}

func (s *ImageService) Pull(ctx context.Context, ref string, opts image.PullOptions) ([]byte, error) {
	if cli == nil {
		return nil, fmt.Errorf("Docker client not initialized")
	}
	reader, err := cli.ImagePull(ctx, ref, opts)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// 读取所有输出直到完成
	output, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read pull output: %w", err)
	}
	return output, nil
}

// PullReader 返回拉取 reader，用于流式读取进度
func (s *ImageService) PullReader(ctx context.Context, ref string, opts image.PullOptions) (io.ReadCloser, error) {
	if cli == nil {
		return nil, fmt.Errorf("Docker client not initialized")
	}
	return cli.ImagePull(ctx, ref, opts)
}

func (s *ImageService) Push(ctx context.Context, ref string, opts image.PushOptions) ([]byte, error) {
	reader, err := cli.ImagePush(ctx, ref, opts)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	
	buf := make([]byte, 1024*1024)
	n, _ := reader.Read(buf)
	return buf[:n], nil
}

func (s *ImageService) Remove(ctx context.Context, id string, opts image.RemoveOptions) ([]image.DeleteResponse, error) {
	return cli.ImageRemove(ctx, id, opts)
}

func (s *ImageService) Tag(ctx context.Context, source, target string) error {
	return cli.ImageTag(ctx, source, target)
}

func (s *ImageService) RemoveTag(ctx context.Context, tag string) error {
	_, err := cli.ImageRemove(ctx, tag, image.RemoveOptions{})
	return err
}

func (s *ImageService) Search(ctx context.Context, term string, opts registry.SearchOptions) ([]registry.SearchResult, error) {
	return cli.ImageSearch(ctx, term, opts)
}

func (s *ImageService) Prune(ctx context.Context) (image.PruneReport, error) {
	return cli.ImagesPrune(ctx, filters.NewArgs())
}

func (s *ImageService) History(ctx context.Context, id string) ([]image.HistoryResponseItem, error) {
	return cli.ImageHistory(ctx, id)
}

func (s *ImageService) CheckUpdate(ctx context.Context, localImage, remoteImage string, localDigests []string) (bool, string, string, error) {
	
	// 直接使用传入的 RepoDigests，而不是重新 inspect
	if len(localDigests) == 0 {
		return false, "", "本地镜像没有RepoDigests信息，无法检查更新（可能是本地构建或特殊导入的镜像）", nil
	}

	localDigest := localDigests[0]
	parts := strings.SplitN(localDigest, "@", 2)
	if len(parts) != 2 {
		return false, "", "无法解析本地镜像的digest信息", nil
	}
	localDigestHash := parts[1]

	// 使用 go-containerregistry 库查询远程镜像的 digest

	// 解析镜像名称和标签
	imageName := remoteImage
	tag := "latest"
	if strings.Contains(remoteImage, ":") {
		parts := strings.SplitN(remoteImage, ":", 2)
		imageName = parts[0]
		tag = parts[1]
	}

	// 对于镜像加速源，必须去掉加速源前缀，查询Docker Hub（原始源）
	// 因为：
	// 1. 加速源通常不支持Registry API，只是pull代理
	// 2. 本地镜像的RepoDigests记录的是Docker Hub的digest
	// 3. 必须查询原始源才能准确比较
	if strings.Contains(imageName, "/") {
		parts := strings.Split(imageName, "/")
		// 如果第一部分看起来像域名（包含 . 或 :），说明是加速源
		if strings.Contains(parts[0], ".") || strings.Contains(parts[0], ":") {
			// 去掉加速源前缀，使用原始镜像名称查询Docker Hub
			if len(parts) > 1 {
				imageName = strings.Join(parts[1:], "/")
			}
		}
	}

	// 智能选择查询策略：
	// 1. 如果配置了镜像加速源，直接查询加速源（因为Docker Hub在中国可能无法访问）
	// 2. 如果未配置加速源，查询Docker Hub
	var remoteDigestHash string
	var err error

	// 检查是否配置了镜像加速源
	if strings.Contains(remoteImage, "/") {
		parts := strings.Split(remoteImage, "/")
		if strings.Contains(parts[0], ".") || strings.Contains(parts[0], ":") {
			// 配置了镜像加速源，直接查询加速源
			registryHost := parts[0]
			// 去掉可能存在的tag
			imagePathWithPossibleTag := strings.Join(parts[1:], "/")
			imagePath := imagePathWithPossibleTag
			if strings.Contains(imagePathWithPossibleTag, ":") {
				imagePathParts := strings.SplitN(imagePathWithPossibleTag, ":", 2)
				imagePath = imagePathParts[0]
			}

			remoteDigestHash, err = s.queryWithGoContainerRegistry(fmt.Sprintf("%s/%s", registryHost, imagePath), tag)
			if err != nil {
				return false, "", "无法检查更新（网络限制），建议手动更新镜像", nil
			}
		} else {
			// 未配置加速源，查询Docker Hub
			remoteDigestHash, err = s.queryWithGoContainerRegistry(imageName, tag)
			if err != nil {
				return false, "", "无法检查更新（网络限制），建议手动更新镜像", nil
			}
		}
	} else {
		// 未配置加速源，查询Docker Hub
		remoteDigestHash, err = s.queryWithGoContainerRegistry(imageName, tag)
		if err != nil {
			return false, "", "无法检查更新（网络限制），建议手动更新镜像", nil
		}
	}


	// 比较本地和远程的 digest
	hasUpdate := localDigestHash != remoteDigestHash


	var reason string
	if hasUpdate {
		reason = "发现新版本可用"
	} else {
		reason = "当前已是最新版本"
	}

	return hasUpdate, remoteDigestHash, reason, nil
}

// queryWithGoContainerRegistry 使用 go-containerregistry 库查询镜像 digest
func (s *ImageService) queryWithGoContainerRegistry(imageName, tag string) (string, error) {
	// 构建镜像引用
	ref, err := name.ParseReference(fmt.Sprintf("%s:%s", imageName, tag))
	if err != nil {
		return "", fmt.Errorf("解析镜像名称失败: %v", err)
	}
	
	
	// 获取远程镜像的描述信息
	desc, err := remote.Get(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return "", fmt.Errorf("获取远程镜像失败: %v", err)
	}
	
	// 获取 digest
	digest := desc.Digest.String()
	if digest == "" {
		return "", fmt.Errorf("未获取到 digest")
	}
	
	// 提取 digest hash
	if strings.Contains(digest, ":") {
		return strings.SplitN(digest, ":", 2)[1], nil
	}
	return digest, nil
}

// queryRegistryWithAuth 查询需要认证的 Registry API
func (s *ImageService) queryRegistryWithAuth(registryHost, imagePath, tag string) (string, error) {
	// 步骤1: 获取认证token
	// 从 Www-Authenticate 头解析认证URL
	authURL := fmt.Sprintf("https://%s/openapi/v1/auth/token/%s/repository:%s:pull", registryHost, registryHost, imagePath)
	
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(authURL)
	if err != nil {
		return "", fmt.Errorf("获取认证token失败: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("认证失败: HTTP %d", resp.StatusCode)
	}
	
	// 解析token
	var tokenResp struct {
		Token string `json:"token"`
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取token响应失败: %v", err)
	}
	
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("解析token失败: %v", err)
	}
	
	if tokenResp.Token == "" {
		return "", fmt.Errorf("未获取到token")
	}
	
	// 步骤2: 使用token查询manifest
	manifestURL := fmt.Sprintf("https://%s/v2/%s/manifests/%s", registryHost, imagePath, tag)
	
	req, err := http.NewRequest("HEAD", manifestURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+tokenResp.Token)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	
	resp, err = client.Do(req)
	if err != nil {
		return "", fmt.Errorf("查询manifest失败: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("manifest查询失败: HTTP %d", resp.StatusCode)
	}
	
	// 获取 digest
	digest := resp.Header.Get("Docker-Content-Digest")
	if digest == "" {
		return "", fmt.Errorf("未获取到digest")
	}
	
	// 提取 digest hash
	if strings.Contains(digest, ":") {
		return strings.SplitN(digest, ":", 2)[1], nil
	}
	return digest, nil
}

type NetworkService struct{}

func NewNetworkService() *NetworkService {
	return &NetworkService{}
}

func (s *NetworkService) List(ctx context.Context) ([]network.Inspect, error) {
	return cli.NetworkList(ctx, network.ListOptions{})
}

func (s *NetworkService) Inspect(ctx context.Context, id string) (network.Inspect, error) {
	return cli.NetworkInspect(ctx, id, network.InspectOptions{})
}

func (s *NetworkService) Create(ctx context.Context, name string, opts network.CreateOptions) (network.CreateResponse, error) {
	return cli.NetworkCreate(ctx, name, opts)
}

func (s *NetworkService) Remove(ctx context.Context, id string) error {
	return cli.NetworkRemove(ctx, id)
}

func (s *NetworkService) Connect(ctx context.Context, networkID, containerID string, config *network.EndpointSettings) error {
	return cli.NetworkConnect(ctx, networkID, containerID, config)
}

func (s *NetworkService) Disconnect(ctx context.Context, networkID, containerID string, force bool) error {
	return cli.NetworkDisconnect(ctx, networkID, containerID, force)
}

func (s *NetworkService) Prune(ctx context.Context) (network.PruneReport, error) {
	return cli.NetworksPrune(ctx, filters.NewArgs())
}

type VolumeService struct{}

func NewVolumeService() *VolumeService {
	return &VolumeService{}
}

func (s *VolumeService) List(ctx context.Context) (volume.ListResponse, error) {
	return cli.VolumeList(ctx, volume.ListOptions{})
}

func (s *VolumeService) Inspect(ctx context.Context, name string) (volume.Volume, error) {
	return cli.VolumeInspect(ctx, name)
}

func (s *VolumeService) Create(ctx context.Context, opts volume.CreateOptions) (volume.Volume, error) {
	return cli.VolumeCreate(ctx, opts)
}

func (s *VolumeService) Remove(ctx context.Context, name string, force bool) error {
	return cli.VolumeRemove(ctx, name, force)
}

func (s *VolumeService) Prune(ctx context.Context) (volume.PruneReport, error) {
	return cli.VolumesPrune(ctx, filters.NewArgs())
}

type SystemService struct{}

func NewSystemService() *SystemService {
	return &SystemService{}
}

func (s *SystemService) Info(ctx context.Context) (Info, error) {
	return cli.Info(ctx)
}

func (s *SystemService) Version(ctx context.Context) (types.Version, error) {
	return cli.ServerVersion(ctx)
}

func (s *SystemService) DiskUsage(ctx context.Context) (types.DiskUsage, error) {
	return cli.DiskUsage(ctx, types.DiskUsageOptions{})
}

func (s *SystemService) Prune(ctx context.Context, all, volumes bool) (*types.BuildCachePruneReport, error) {
	opts := types.BuildCachePruneOptions{}
	if all {
		opts.All = true
	}
	return cli.BuildCachePrune(ctx, opts)
}

func (s *SystemService) Events(ctx context.Context, opts events.ListOptions) (<-chan events.Message, <-chan error) {
	return cli.Events(ctx, opts)
}

func Context() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), config.Get().Docker.Timeout)
}

func ContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// queryDockerHubAPI 查询 Docker Hub API 获取镜像 digest
func (s *ImageService) queryDockerHubAPI(apiURL string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	
	// 解析 JSON 响应
	var tagInfo struct {
		Images []struct {
			Digest string `json:"digest"`
		} `json:"images"`
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}
	
	if err := json.Unmarshal(body, &tagInfo); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}
	
	// 获取第一个镜像的 digest
	if len(tagInfo.Images) == 0 {
		return "", fmt.Errorf("未找到镜像信息")
	}
	
	digest := tagInfo.Images[0].Digest
	if digest == "" {
		return "", fmt.Errorf("未获取到 digest")
	}
	
	// 提取 digest hash
	if strings.Contains(digest, ":") {
		return strings.SplitN(digest, ":", 2)[1], nil
	}
	return digest, nil
}

// checkUpdateByPull 通过 docker pull 方式检查更新（会下载镜像）
