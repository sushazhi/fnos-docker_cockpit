package model

import "time"

type ContainerInfo struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Image   string            `json:"image"`
	Status  string            `json:"status"`
	State   string            `json:"state"`
	Ports   string            `json:"ports"`
	Created time.Time         `json:"created"`
	Labels  map[string]string `json:"labels,omitempty"`
}

type ImageInfo struct {
	ID         string    `json:"id"`
	Repository string    `json:"repository"`
	Tag        string    `json:"tag"`
	Size       int64     `json:"size"`
	Created    time.Time `json:"created"`
}

type NetworkInfo struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Scope      string            `json:"scope"`
	Subnet     string            `json:"subnet,omitempty"`
	Gateway    string            `json:"gateway,omitempty"`
	Containers map[string]ContainerEndpoint `json:"containers,omitempty"`
}

type ContainerEndpoint struct {
	Name        string `json:"name"`
	EndpointID  string `json:"endpointId"`
	MacAddress  string `json:"macAddress"`
	IPv4Address string `json:"ipv4Address"`
	IPv6Address string `json:"ipv6Address"`
}

type VolumeInfo struct {
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Mountpoint string            `json:"mountpoint"`
	CreatedAt  time.Time         `json:"createdAt"`
	Labels     map[string]string `json:"labels,omitempty"`
	Size       int64             `json:"size,omitempty"`
}

type ComposeInfo struct {
	Name     string    `json:"name"`
	File     string    `json:"file"`
	Path     string    `json:"path"`
	Status   string    `json:"status"`
	Modified time.Time `json:"modified"`
	Size     int64     `json:"size"`
}

type AuditLog struct {
	ID        string                 `json:"id"`
	Action    string                 `json:"action"`
	Timestamp time.Time              `json:"timestamp"`
	ClientIP  string                 `json:"clientIp"`
	UserAgent string                 `json:"userAgent"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

type Session struct {
	Token     string    `json:"token"`
	CSRFToken string    `json:"csrfToken"`
	CreatedAt time.Time `json:"createdAt"`
	LastAccess time.Time `json:"lastAccess"`
}

type LoginRequest struct {
	Password string `json:"password" binding:"required,min=8,max=128"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required,min=8,max=128"`
	NewPassword     string `json:"newPassword" binding:"required,min=8,max=128"`
}

type ContainerCreateRequest struct {
	Name       string            `json:"name,omitempty" binding:"max=255"`
	Image      string            `json:"image" binding:"required,max=500"`
	Hostname   string            `json:"hostname,omitempty" binding:"max=255"`
	Env        []string          `json:"env,omitempty" binding:"max=100,dive,max=500"`
	Ports      []string          `json:"ports,omitempty" binding:"max=100,dive,max=100"`
	Volumes    []string          `json:"volumes,omitempty" binding:"max=100,dive,max=500"`
	Network    string            `json:"network,omitempty" binding:"max=255"`
	Restart    string            `json:"restart,omitempty" binding:"max=50"`
	Labels     map[string]string `json:"labels,omitempty" binding:"max=100"`
	WorkDir    string            `json:"workdir,omitempty" binding:"max=500"`
	User       string            `json:"user,omitempty" binding:"max=255"`
	Privileged bool              `json:"privileged,omitempty"`
	CapAdd     []string          `json:"capAdd,omitempty" binding:"max=50,dive,max=100"`
	CapDrop    []string          `json:"capDrop,omitempty" binding:"max=50,dive,max=100"`
	Cmd        []string          `json:"cmd,omitempty" binding:"max=100,dive,max=1000"`
}

type ContainerUpdateRequest struct {
	Memory        string `json:"memory,omitempty" binding:"max=50"`
	CPUs          string `json:"cpus,omitempty" binding:"max=50"`
	RestartPolicy string `json:"restartPolicy,omitempty" binding:"max=50"`
}

type ContainerCommitRequest struct {
	Repository string `json:"repository,omitempty" binding:"max=255"`
	Author     string `json:"author,omitempty" binding:"max=255"`
	Message    string `json:"message,omitempty" binding:"max=1000"`
}

type ContainerLogsQuery struct {
	Tail       string `form:"tail" binding:"max=10"`
	Since      string `form:"since" binding:"max=50"`
	Timestamps bool   `form:"timestamps"`
}

type ContainerExecRequest struct {
	Cmd []string `json:"cmd" binding:"required,max=100,dive,max=1000"`
}

type ContainerRenameRequest struct {
	Name string `json:"name" binding:"required,max=255"`
}

type ContainerRemoveRequest struct {
	Force   bool `json:"force"`
	Volumes bool `json:"volumes"`
}

type ContainerStopRequest struct {
	Timeout int `json:"timeout" binding:"max=3600"`
}

// Network 相关请求
type NetworkCreateRequest struct {
	Name       string            `json:"name" binding:"required,max=255"`
	Driver     string            `json:"driver,omitempty" binding:"max=50"`
	Subnet     string            `json:"subnet,omitempty" binding:"max=50"`
	Gateway    string            `json:"gateway,omitempty" binding:"max=50"`
	Labels     map[string]string `json:"labels,omitempty" binding:"max=100"`
}

type NetworkConnectRequest struct {
	ContainerID string `json:"containerId" binding:"required,max=255"`
}

type NetworkDisconnectRequest struct {
	ContainerID string `json:"containerId" binding:"required,max=255"`
}

// Volume 相关请求
type VolumeCreateRequest struct {
	Name       string            `json:"name" binding:"required,max=255"`
	Driver     string            `json:"driver,omitempty" binding:"max=100"`
	Labels     map[string]string `json:"labels,omitempty" binding:"max=100"`
}

// Image 相关请求
type ImageTagRequest struct {
	Source string `json:"source" binding:"required,max=500"`
	Target string `json:"target" binding:"required,max=500"`
}

// Compose 相关请求
type ComposeSaveRequest struct {
	Name    string `json:"name" binding:"required,max=255"`
	Content string `json:"content" binding:"required,max=1048576"` // 最大 1MB
}

// File 相关请求
type FileWriteRequest struct {
	Path    string `json:"path" binding:"required,max=1000"`
	Content string `json:"content" binding:"max=10485760"` // 最大 10MB
}

type FileMkdirRequest struct {
	Path string `json:"path" binding:"required,max=1000"`
}

type FileRenameRequest struct {
	OldPath string `json:"oldPath" binding:"required,max=1000"`
	NewPath string `json:"newPath" binding:"required,max=1000"`
}
