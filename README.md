# Dockpit

一个简洁、美观的 Docker 管理工具，专为飞牛 OS 设计。

## 功能特性

### 容器管理
- 查看容器列表和状态
- 启动、停止、重启、暂停容器
- 查看容器日志（支持自动刷新）
- 查看容器统计信息（CPU、内存、网络、磁盘）
- 容器终端（Web Terminal）
- 查看容器详情（端口、环境变量、卷挂载、网络）
- 生成 Docker Run 命令
- 生成 Docker Compose 文件
- 编辑容器重启策略

### Compose 管理
- 查看 Compose 项目列表
- 启动、停止、重启 Compose 项目
- 构建 Compose 服务
- 查看服务日志
- 编辑 YAML 文件
- 删除 Compose 项目

### 镜像管理
- 查看本地镜像列表
- 拉取新镜像（支持镜像加速源）
- 删除镜像
- 更新镜像
- 检查镜像更新（需要 Docker Hub 可访问）
- 从 Docker Hub 搜索镜像

### 网络管理
- 查看网络列表
- 创建网络
- 删除网络

### 存储卷管理
- 查看卷列表
- 创建卷
- 删除卷
- 显示卷使用状态（包括停止的 Compose 项目）
- 显示卷所属的 Compose 项目

### 系统设置
- 镜像加速源配置
- 深色/浅色主题切换
- 密码修改
- 审计日志

### 系统信息
- 查看 Docker 版本
- 查看系统资源使用情况

## 技术栈

### 后端
- Go 1.21+
- Docker Engine API
- Gin Web Framework

### 前端
- Vue 3
- Vite
- Vue I18n（国际化）
- Xterm.js（终端）

## 安装

### 飞牛 OS
1. 下载 `.fpk` 安装包
2. 在飞牛 OS 应用中心安装

### 手动构建

#### 后端
```bash
cd app/server
go build -o dockpit
```

#### 前端
```bash
cd app/ui
npm install
npm run build
```

#### 打包
```bash
# Windows
./build.ps1 -Version 1.0.0

# Linux/Mac
./build.sh
```

## 开发

### 启动后端
```bash
cd app/server
go run main.go
```

### 启动前端
```bash
cd app/ui
npm run dev
```

## 国际化

支持以下语言：
- 简体中文 (zh-CN)
- 英文 (en-US)

## 许可证

MIT License
