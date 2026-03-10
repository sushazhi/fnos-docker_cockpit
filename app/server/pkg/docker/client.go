package docker

import (
	"context"
	"dockpit/internal/config"
	"fmt"
	"io"
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

func (s *ContainerService) ExecCreate(ctx context.Context, id string, opts types.ExecConfig) (types.IDResponse, error) {
	return cli.ContainerExecCreate(ctx, id, opts)
}

func (s *ContainerService) ExecAttach(ctx context.Context, execID string, opts types.ExecStartCheck) (types.HijackedResponse, error) {
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

func (s *ImageService) Search(ctx context.Context, term string, opts types.ImageSearchOptions) ([]registry.SearchResult, error) {
	return cli.ImageSearch(ctx, term, opts)
}

func (s *ImageService) Prune(ctx context.Context) (types.ImagesPruneReport, error) {
	return cli.ImagesPrune(ctx, filters.NewArgs())
}

func (s *ImageService) History(ctx context.Context, id string) ([]image.HistoryResponseItem, error) {
	return cli.ImageHistory(ctx, id)
}

func (s *ImageService) CheckUpdate(ctx context.Context, localImage, remoteImage string) (bool, string, error) {
	inspect, _, err := cli.ImageInspectWithRaw(ctx, localImage)
	if err != nil {
		return false, "", err
	}

	if len(inspect.RepoDigests) == 0 {
		return false, "", nil
	}

	localDigest := inspect.RepoDigests[0]
	parts := strings.SplitN(localDigest, "@", 2)
	if len(parts) != 2 {
		return false, "", nil
	}
	repo := parts[0]
	localDigestHash := parts[1]

	pullReader, err := cli.ImagePull(ctx, remoteImage, image.PullOptions{})
	if err != nil {
		return false, "", err
	}
	defer pullReader.Close()

	io.ReadAll(pullReader)

	updatedInspect, _, err := cli.ImageInspectWithRaw(ctx, localImage)
	if err != nil {
		return false, "", err
	}

	if len(updatedInspect.RepoDigests) == 0 {
		return false, "", nil
	}

	updatedDigest := ""
	for _, d := range updatedInspect.RepoDigests {
		if strings.HasPrefix(d, repo+"@") {
			updatedDigest = d
			break
		}
	}

	if updatedDigest == "" {
		return false, "", nil
	}

	updatedDigestHash := strings.SplitN(updatedDigest, "@", 2)[1]
	hasUpdate := localDigestHash != updatedDigestHash

	return hasUpdate, updatedDigestHash, nil
}

type NetworkService struct{}

func NewNetworkService() *NetworkService {
	return &NetworkService{}
}

func (s *NetworkService) List(ctx context.Context) ([]types.NetworkResource, error) {
	return cli.NetworkList(ctx, network.ListOptions{})
}

func (s *NetworkService) Inspect(ctx context.Context, id string) (types.NetworkResource, error) {
	return cli.NetworkInspect(ctx, id, network.InspectOptions{})
}

func (s *NetworkService) Create(ctx context.Context, name string, opts types.NetworkCreate) (types.NetworkCreateResponse, error) {
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

func (s *NetworkService) Prune(ctx context.Context) (types.NetworksPruneReport, error) {
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

func (s *VolumeService) Prune(ctx context.Context) (types.VolumesPruneReport, error) {
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
