package handler

import (
    "bytes"
    "context"
    "errors"
    "fmt"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "sort"
    "strings"
    "time"

    "dockpit/internal/config"
    "dockpit/pkg/docker"
    "dockpit/pkg/response"

    "github.com/docker/docker/api/types"
    dockcontainer "github.com/docker/docker/api/types/container"
    "github.com/docker/docker/api/types/filters"
    "github.com/docker/docker/pkg/stdcopy"
    "github.com/gin-gonic/gin"
)

const (
    hybridComposeProjectLabel     = "com.docker.compose.project"
    hybridComposeServiceLabel     = "com.docker.compose.service"
    hybridComposeConfigFilesLabel = "com.docker.compose.project.config_files"
)

type ComposeHybridHandler struct{}

func NewComposeHybridHandler() *ComposeHybridHandler {
    return &ComposeHybridHandler{}
}

func (h *ComposeHybridHandler) getComposeDir() string {
    return filepath.Join(config.Get().DataDir, "compose")
}

func (h *ComposeHybridHandler) ensureDir() {
    _ = os.MkdirAll(h.getComposeDir(), 0755)
}

type hybridProjectSummary struct {
    name       string
    path       string
    running    bool
    services   map[string]struct{}
    lastCreate int64
}

func (h *ComposeHybridHandler) List(c *gin.Context) {
    h.ensureDir()

    type ComposeItem struct {
        Name     string    `json:"name"`
        File     string    `json:"file"`
        Path     string    `json:"path"`
        Status   string    `json:"status"`
        Modified time.Time `json:"modified"`
        Size     int64     `json:"size"`
        Services int       `json:"services"`
    }

    projects := make(map[string]*hybridProjectSummary)

    runtimeContainers, err := h.listRuntimeContainers("", true)
    if err == nil {
        for _, ctr := range runtimeContainers {
            project := ctr.Labels[hybridComposeProjectLabel]
            if project == "" {
                continue
            }

            summary := projects[project]
            if summary == nil {
                summary = &hybridProjectSummary{
                    name:     project,
                    services: map[string]struct{}{},
                }
                projects[project] = summary
            }

            if strings.EqualFold(ctr.State, "running") {
                summary.running = true
            }
            if ctr.Created > summary.lastCreate {
                summary.lastCreate = ctr.Created
            }
            if summary.path == "" {
                summary.path = firstConfigPath(ctr.Labels[hybridComposeConfigFilesLabel])
            }

            service := ctr.Labels[hybridComposeServiceLabel]
            if service == "" {
                service = normalizedHybridContainerName(ctr)
            }
            if service != "" {
                summary.services[service] = struct{}{}
            }
        }
    }

    entries, err := os.ReadDir(h.getComposeDir())
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
            if _, exists := projects[projectName]; exists {
                continue
            }

            info, _ := entry.Info()
            projects[projectName] = &hybridProjectSummary{
                name:       projectName,
                path:       filepath.Join(h.getComposeDir(), name),
                running:    false,
                services:   map[string]struct{}{},
                lastCreate: info.ModTime().Unix(),
            }
        }
    }

    list := make([]ComposeItem, 0, len(projects))
    for _, p := range projects {
        status := "stopped"
        if p.running {
            status = "running"
        }

        modified := time.Time{}
        if p.lastCreate > 0 {
            modified = time.Unix(p.lastCreate, 0)
        }

        list = append(list, ComposeItem{
            Name:     p.name,
            File:     filepath.Base(p.path),
            Path:     p.path,
            Status:   status,
            Modified: modified,
            Size:     0,
            Services: len(p.services),
        })
    }

    sort.Slice(list, func(i, j int) bool {
        return list[i].Name < list[j].Name
    })

    response.Success(c, gin.H{"projects": list})
}

func (h *ComposeHybridHandler) Get(c *gin.Context) {
    h.ensureDir()
    name := c.Param("name")

    var content string
    path, hasLocal := h.findComposeFile(name)
    if hasLocal {
        fileContent, err := os.ReadFile(path)
        if err == nil {
            content = string(fileContent)
        }
    }

    containers, err := h.listRuntimeContainers(name, true)
    if err != nil && !hasLocal {
        response.DockerError(c, "Failed to query compose project", err.Error())
        return
    }

    if len(containers) == 0 && !hasLocal {
        response.NotFound(c, "Compose project not found")
        return
    }

    // 如果没有本地文件，尝试从容器标签中获取配置文件路径并读取内容
    if path == "" || content == "" {
        for _, ctr := range containers {
            if cfg := firstConfigPath(ctr.Labels[hybridComposeConfigFilesLabel]); cfg != "" {
                if path == "" {
                    path = cfg
                }
                // 尝试读取配置文件内容
                if content == "" {
                    if fileContent, err := os.ReadFile(cfg); err == nil {
                        content = string(fileContent)
                    }
                }
                break
            }
        }
    }

    response.Success(c, gin.H{
        "content":  content,
        "path":     path,
        "services": buildHybridServices(containers),
    })
}

type hybridComposeSaveRequest struct {
    Name    string `json:"name" binding:"required"`
    Content string `json:"content" binding:"required"`
}

func (h *ComposeHybridHandler) Save(c *gin.Context) {
    h.ensureDir()

    var req hybridComposeSaveRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, "Invalid parameters")
        return
    }

    if !isValidName(req.Name) {
        response.BadRequest(c, "Invalid name: only letters, numbers, hyphens and underscores are allowed")
        return
    }

    if existingPath, exists := h.findComposeFile(req.Name); exists {
        response.BadRequest(c, "Compose project already exists: "+existingPath)
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

func (h *ComposeHybridHandler) Delete(c *gin.Context) {
    h.ensureDir()
    name := c.Param("name")

    if !isValidName(name) {
        response.BadRequest(c, "Invalid name: only letters, numbers, hyphens and underscores are allowed")
        return
    }

    removedLocal := false
    if path, exists := h.findComposeFile(name); exists {
        if err := os.Remove(path); err == nil {
            removedLocal = true
        }
    }

    removedRuntime := false
    containers, err := h.listRuntimeContainers(name, true)
    if err == nil {
        cli := docker.GetClient()
        for _, ctr := range containers {
            if rmErr := cli.ContainerRemove(c.Request.Context(), ctr.ID, dockcontainer.RemoveOptions{Force: true, RemoveVolumes: false}); rmErr == nil {
                removedRuntime = true
            }
        }
    }

    if !removedLocal && !removedRuntime {
        response.NotFound(c, "Compose project not found")
        return
    }

    addAuditLog(c, "compose_delete", map[string]interface{}{"name": name})
    response.Success(c, gin.H{"success": true})
}

func (h *ComposeHybridHandler) Up(c *gin.Context) {
    name := c.Param("name")

    var req struct {
        Build         bool `json:"build"`
        ForceRecreate bool `json:"forceRecreate"`
    }
    _ = c.ShouldBindJSON(&req)

    if path, exists := h.findComposeFile(name); exists {
        args := []string{"up", "-d"}
        if req.Build {
            args = append(args, "--build")
        }
        if req.ForceRecreate {
            args = append(args, "--force-recreate")
        }

        output, err := h.runComposeWithFile(name, path, args...)
        if err != nil {
            response.DockerError(c, "Failed to start compose", output)
            return
        }

        addAuditLog(c, "compose_up", map[string]interface{}{"name": name})
        response.Success(c, gin.H{"success": true, "output": output})
        return
    }

    started, err := h.startByRuntime(name)
    if err != nil {
        response.DockerError(c, "Failed to start compose", err.Error())
        return
    }

    addAuditLog(c, "compose_up", map[string]interface{}{"name": name})
    response.Success(c, gin.H{"success": true, "output": fmt.Sprintf("started %d container(s)", started)})
}

func (h *ComposeHybridHandler) Down(c *gin.Context) {
    name := c.Param("name")

    var req struct {
        Volumes       bool `json:"volumes"`
        RemoveOrphans bool `json:"removeOrphans"`
    }
    _ = c.ShouldBindJSON(&req)

    if path, exists := h.findComposeFile(name); exists {
        args := []string{"down"}
        if req.Volumes {
            args = append(args, "--volumes")
        }
        if req.RemoveOrphans {
            args = append(args, "--remove-orphans")
        }

        output, err := h.runComposeWithFile(name, path, args...)
        if err != nil {
            response.DockerError(c, "Failed to stop compose", output)
            return
        }

        addAuditLog(c, "compose_down", map[string]interface{}{"name": name})
        response.Success(c, gin.H{"success": true, "output": output})
        return
    }

    stopped, err := h.stopByRuntime(name)
    if err != nil {
        response.DockerError(c, "Failed to stop compose", err.Error())
        return
    }

    addAuditLog(c, "compose_down", map[string]interface{}{"name": name})
    response.Success(c, gin.H{"success": true, "output": fmt.Sprintf("stopped %d container(s)", stopped)})
}

func (h *ComposeHybridHandler) Ps(c *gin.Context) {
    name := c.Param("name")

    if path, exists := h.findComposeFile(name); exists {
        output, err := h.runComposeWithFile(name, path, "ps")
        if err == nil {
            response.Success(c, gin.H{"output": output})
            return
        }
    }

    containers, err := h.listRuntimeContainers(name, true)
    if err != nil {
        response.DockerError(c, "Failed to get compose status", err.Error())
        return
    }

    lines := make([]string, 0, len(containers))
    for _, ctr := range containers {
        service := ctr.Labels[hybridComposeServiceLabel]
        if service == "" {
            service = normalizedHybridContainerName(ctr)
        }
        lines = append(lines, fmt.Sprintf("%s\t%s\t%s\t%s", service, strings.ToLower(ctr.State), ctr.Image, shortHybridID(ctr.ID)))
    }

    sort.Strings(lines)
    response.Success(c, gin.H{"output": strings.Join(lines, "\n")})
}

type hybridComposeLogsQuery struct {
    Tail   string `form:"tail"`
    Follow bool   `form:"follow"`
}

func (h *ComposeHybridHandler) Logs(c *gin.Context) {
    name := c.Param("name")

    var query hybridComposeLogsQuery
    _ = c.ShouldBindQuery(&query)

    if path, exists := h.findComposeFile(name); exists {
        args := []string{"logs"}
        if query.Tail != "" {
            args = append(args, "--tail", query.Tail)
        }
        if query.Follow {
            args = append(args, "-f")
        }

        output, err := h.runComposeWithFile(name, path, args...)
        if err == nil {
            response.Success(c, gin.H{"logs": output})
            return
        }
    }

    containers, err := h.listRuntimeContainers(name, true)
    if err != nil {
        response.DockerError(c, "Failed to get logs", err.Error())
        return
    }
    if len(containers) == 0 {
        response.NotFound(c, "Compose project not found")
        return
    }

    tail := query.Tail
    if tail == "" {
        tail = "500"
    }

    sort.Slice(containers, func(i, j int) bool {
        return normalizedHybridContainerName(containers[i]) < normalizedHybridContainerName(containers[j])
    })

    cli := docker.GetClient()
    var merged strings.Builder
    for _, ctr := range containers {
        reader, err := cli.ContainerLogs(c.Request.Context(), ctr.ID, dockcontainer.LogsOptions{
            ShowStdout: true,
            ShowStderr: true,
            Tail:       tail,
            Follow:     false,
            Timestamps: false,
        })
        if err != nil {
            continue
        }

        raw, err := io.ReadAll(reader)
        _ = reader.Close()
        if err != nil {
            continue
        }

        var stdout bytes.Buffer
        var stderr bytes.Buffer
        if _, err := stdcopy.StdCopy(&stdout, &stderr, bytes.NewReader(raw)); err != nil {
            stdout.Reset()
            stdout.Write(raw)
        }

        logs := strings.TrimSpace(stdout.String() + stderr.String())
        if logs == "" {
            continue
        }

        if merged.Len() > 0 {
            merged.WriteString("\n")
        }

        merged.WriteString("[")
        merged.WriteString(normalizedHybridContainerName(ctr))
        merged.WriteString("]\n")
        merged.WriteString(logs)
        merged.WriteString("\n")
    }

    response.Success(c, gin.H{"logs": strings.TrimSpace(merged.String())})
}

func (h *ComposeHybridHandler) Pull(c *gin.Context) {
    name := c.Param("name")
    path, exists := h.findComposeFile(name)
    if !exists {
        response.BadRequest(c, "Pull requires a local compose file")
        return
    }

    output, err := h.runComposeWithFile(name, path, "pull")
    if err != nil {
        response.DockerError(c, "Failed to pull images", output)
        return
    }

    response.Success(c, gin.H{"success": true, "output": output})
}

func (h *ComposeHybridHandler) Build(c *gin.Context) {
    name := c.Param("name")
    path, exists := h.findComposeFile(name)
    if !exists {
        response.BadRequest(c, "Build requires a local compose file")
        return
    }

    output, err := h.runComposeWithFile(name, path, "build")
    if err != nil {
        response.DockerError(c, "Failed to build", output)
        return
    }

    response.Success(c, gin.H{"success": true, "output": output})
}

func (h *ComposeHybridHandler) Restart(c *gin.Context) {
    name := c.Param("name")

    if path, exists := h.findComposeFile(name); exists {
        output, err := h.runComposeWithFile(name, path, "restart")
        if err != nil {
            response.DockerError(c, "Failed to restart", output)
            return
        }

        addAuditLog(c, "compose_restart", map[string]interface{}{"name": name})
        response.Success(c, gin.H{"success": true, "output": output})
        return
    }

    restarted, err := h.restartByRuntime(name)
    if err != nil {
        response.DockerError(c, "Failed to restart", err.Error())
        return
    }

    addAuditLog(c, "compose_restart", map[string]interface{}{"name": name})
    response.Success(c, gin.H{"success": true, "output": fmt.Sprintf("restarted %d container(s)", restarted)})
}

func (h *ComposeHybridHandler) Stop(c *gin.Context) {
    name := c.Param("name")

    if path, exists := h.findComposeFile(name); exists {
        output, err := h.runComposeWithFile(name, path, "stop")
        if err != nil {
            response.DockerError(c, "Failed to stop", output)
            return
        }

        addAuditLog(c, "compose_stop", map[string]interface{}{"name": name})
        response.Success(c, gin.H{"success": true, "output": output})
        return
    }

    stopped, err := h.stopByRuntime(name)
    if err != nil {
        response.DockerError(c, "Failed to stop", err.Error())
        return
    }

    addAuditLog(c, "compose_stop", map[string]interface{}{"name": name})
    response.Success(c, gin.H{"success": true, "output": fmt.Sprintf("stopped %d container(s)", stopped)})
}

func (h *ComposeHybridHandler) Start(c *gin.Context) {
    name := c.Param("name")

    if path, exists := h.findComposeFile(name); exists {
        output, err := h.runComposeWithFile(name, path, "start")
        if err != nil {
            response.DockerError(c, "Failed to start", output)
            return
        }

        addAuditLog(c, "compose_start", map[string]interface{}{"name": name})
        response.Success(c, gin.H{"success": true, "output": output})
        return
    }

    started, err := h.startByRuntime(name)
    if err != nil {
        response.DockerError(c, "Failed to start", err.Error())
        return
    }

    addAuditLog(c, "compose_start", map[string]interface{}{"name": name})
    response.Success(c, gin.H{"success": true, "output": fmt.Sprintf("started %d container(s)", started)})
}

func (h *ComposeHybridHandler) findComposeFile(name string) (string, bool) {
    yml := filepath.Join(h.getComposeDir(), name+".yml")
    if _, err := os.Stat(yml); err == nil {
        return yml, true
    }

    yaml := filepath.Join(h.getComposeDir(), name+".yaml")
    if _, err := os.Stat(yaml); err == nil {
        return yaml, true
    }

    return "", false
}

func (h *ComposeHybridHandler) runComposeWithFile(projectName, path string, args ...string) (string, error) {
    cmdArgs := []string{"compose", "-p", projectName, "-f", path}
    cmdArgs = append(cmdArgs, args...)

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    cmd := exec.CommandContext(ctx, "docker", cmdArgs...)
    output, err := cmd.CombinedOutput()
    return string(output), err
}

func (h *ComposeHybridHandler) listRuntimeContainers(projectName string, all bool) ([]types.Container, error) {
    cli := docker.GetClient()
    if cli == nil {
        return nil, errors.New("docker client is not initialized")
    }

    f := filters.NewArgs(filters.Arg("label", hybridComposeProjectLabel))
    if projectName != "" {
        f = filters.NewArgs(filters.Arg("label", hybridComposeProjectLabel+"="+projectName))
    }

    return cli.ContainerList(context.Background(), dockcontainer.ListOptions{All: all, Filters: f})
}

func (h *ComposeHybridHandler) startByRuntime(projectName string) (int, error) {
    cli := docker.GetClient()
    if cli == nil {
        return 0, errors.New("docker client is not initialized")
    }

    containers, err := h.listRuntimeContainers(projectName, true)
    if err != nil {
        return 0, err
    }
    if len(containers) == 0 {
        return 0, errors.New("compose project not found")
    }

    count := 0
    var errs []string
    for _, ctr := range containers {
        if strings.EqualFold(ctr.State, "running") {
            continue
        }
        if err := cli.ContainerStart(context.Background(), ctr.ID, dockcontainer.StartOptions{}); err != nil {
            errs = append(errs, fmt.Sprintf("%s: %v", shortHybridID(ctr.ID), err))
            continue
        }
        count++
    }

    if len(errs) > 0 {
        return count, errors.New(strings.Join(errs, "; "))
    }
    return count, nil
}

func (h *ComposeHybridHandler) stopByRuntime(projectName string) (int, error) {
    cli := docker.GetClient()
    if cli == nil {
        return 0, errors.New("docker client is not initialized")
    }

    containers, err := h.listRuntimeContainers(projectName, true)
    if err != nil {
        return 0, err
    }
    if len(containers) == 0 {
        return 0, errors.New("compose project not found")
    }

    timeout := 10
    count := 0
    var errs []string
    for _, ctr := range containers {
        if !strings.EqualFold(ctr.State, "running") {
            continue
        }
        if err := cli.ContainerStop(context.Background(), ctr.ID, dockcontainer.StopOptions{Timeout: &timeout}); err != nil {
            errs = append(errs, fmt.Sprintf("%s: %v", shortHybridID(ctr.ID), err))
            continue
        }
        count++
    }

    if len(errs) > 0 {
        return count, errors.New(strings.Join(errs, "; "))
    }
    return count, nil
}

func (h *ComposeHybridHandler) restartByRuntime(projectName string) (int, error) {
    cli := docker.GetClient()
    if cli == nil {
        return 0, errors.New("docker client is not initialized")
    }

    containers, err := h.listRuntimeContainers(projectName, true)
    if err != nil {
        return 0, err
    }
    if len(containers) == 0 {
        return 0, errors.New("compose project not found")
    }

    timeout := 10
    count := 0
    var errs []string
    for _, ctr := range containers {
        if err := cli.ContainerRestart(context.Background(), ctr.ID, dockcontainer.StopOptions{Timeout: &timeout}); err != nil {
            errs = append(errs, fmt.Sprintf("%s: %v", shortHybridID(ctr.ID), err))
            continue
        }
        count++
    }

    if len(errs) > 0 {
        return count, errors.New(strings.Join(errs, "; "))
    }
    return count, nil
}

func buildHybridServices(containers []types.Container) []map[string]string {
    services := make([]map[string]string, 0, len(containers))
    for _, ctr := range containers {
        serviceName := ctr.Labels[hybridComposeServiceLabel]
        if serviceName == "" {
            serviceName = normalizedHybridContainerName(ctr)
        }

        services = append(services, map[string]string{
            "name":        serviceName,
            "image":       ctr.Image,
            "state":       strings.ToLower(ctr.State),
            "containerId": ctr.ID,
        })
    }

    sort.Slice(services, func(i, j int) bool {
        return services[i]["name"] < services[j]["name"]
    })

    return services
}

func normalizedHybridContainerName(ctr types.Container) string {
    if len(ctr.Names) == 0 {
        return shortHybridID(ctr.ID)
    }
    return strings.TrimPrefix(ctr.Names[0], "/")
}

func shortHybridID(id string) string {
    if len(id) <= 12 {
        return id
    }
    return id[:12]
}

func firstConfigPath(raw string) string {
    if raw == "" {
        return ""
    }
    parts := strings.Split(raw, ",")
    if len(parts) == 0 {
        return ""
    }
    return strings.TrimSpace(parts[0])
}
