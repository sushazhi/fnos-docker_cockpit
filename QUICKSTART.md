# Dockpit 快速开始指南

## 开发环境设置

### 前置要求

- Go 1.21+
- Node.js 18+
- Docker
- Make (可选)

### 1. 克隆项目

```bash
git clone <repository-url>
cd fnos-docker_cockpit
```

### 2. 后端开发

```bash
cd app/server

# 安装依赖
go mod download

# 运行开发服务器
go run main.go

# 或使用 Make
make dev-backend
```

后端将在 `http://localhost:8807` 启动

### 3. 前端开发

```bash
cd app/ui

# 安装依赖
npm install

# 运行开发服务器
npm run dev

# 或使用 Make
make dev-frontend
```

前端将在 `http://localhost:3000` 启动，API 请求会自动代理到后端

## 使用新功能

### 批量操作容器

```typescript
import { useContainerStore } from '@/stores/container'

const containerStore = useContainerStore()

// 批量启动容器
await containerStore.batchOperation('start', ['id1', 'id2', 'id3'])

// 批量停止容器
await containerStore.batchOperation('stop', ['id1', 'id2'], {
  timeout: 30
})

// 批量删除容器
await containerStore.batchOperation('remove', ['id1', 'id2'], {
  force: true
})
```

### 使用虚拟滚动

```vue
<template>
  <VirtualList 
    :items="containers" 
    :item-height="70" 
    key-field="Id"
  >
    <template #item="{ item }">
      <div class="container-item">
        {{ item.Names[0] }}
      </div>
    </template>
  </VirtualList>
</template>

<script setup>
import VirtualList from '@/components/VirtualList.vue'
import { ref } from 'vue'

const containers = ref([])
</script>
```

### 使用防抖和节流

```typescript
import { debounce, throttle } from '@/utils/debounce'

// 搜索输入防抖
const handleSearch = debounce((query: string) => {
  console.log('Searching:', query)
}, 300)

// 滚动事件节流
const handleScroll = throttle(() => {
  console.log('Scrolling...')
}, 100)
```

### 使用骨架屏

```vue
<template>
  <SkeletonLoader v-if="loading" :count="5" />
  <div v-else>
    <!-- 实际内容 -->
  </div>
</template>

<script setup>
import SkeletonLoader from '@/components/SkeletonLoader.vue'
import { ref } from 'vue'

const loading = ref(true)
</script>
```

## 代码质量检查

### 运行 Linter

```bash
# 后端
cd app/server
golangci-lint run

# 前端
cd app/ui
npm run lint
```

### 格式化代码

```bash
# 前端
cd app/ui
npm run format
```

## 构建项目

### 使用 Make

```bash
# 构建前后端
make build

# 清理构建产物
make clean
```

### 手动构建

```bash
# 构建前端
cd app/ui
npm run build

# 构建后端
cd app/server
go build -o dockpit
```

## 依赖管理

### 检查依赖更新

```bash
# 使用 Make
make deps-check

# 或手动运行
bash scripts/check-deps.sh  # Linux/Mac
pwsh scripts/check-deps.ps1 # Windows
```

### 更新依赖

```bash
# 使用 Make
make deps-update

# 或手动更新
cd app/ui && npm update
cd app/server && go get -u ./... && go mod tidy
```

## 测试

```bash
# 运行所有测试
make test

# 仅后端测试
cd app/server
go test ./...

# 仅前端测试
cd app/ui
npm test
```

## 常见问题

### Q: Docker 连接失败

A: 确保 Docker 守护进程正在运行：
```bash
# Linux
sudo systemctl start docker

# Windows/Mac
# 启动 Docker Desktop
```

### Q: 前端代理不工作

A: 检查 `vite.config.ts` 中的代理配置，确保后端端口正确

### Q: Go 模块下载慢

A: 配置 Go 代理：
```bash
go env -w GOPROXY=https://goproxy.cn,direct
```

### Q: npm 安装慢

A: 使用国内镜像：
```bash
npm config set registry https://registry.npmmirror.com
```

## 开发工作流

1. 创建功能分支
2. 开发新功能
3. 运行 linter: `make lint`
4. 运行测试: `make test`
5. 提交代码
6. 创建 Pull Request

## 生产部署

### 构建 FPK 包

```bash
# Windows
./build.ps1 -Version 1.0.0

# Linux/Mac
./build.sh 1.0.0
```

### 手动部署

1. 构建前端: `cd app/ui && npm run build`
2. 构建后端: `cd app/server && go build`
3. 复制 `app/ui/dist` 到服务器的 `ui` 目录
4. 运行后端: `./dockpit`

## 更多信息

- 改进文档: [IMPROVEMENTS.md](./IMPROVEMENTS.md)
- 项目说明: [README.md](./README.md)
- API 文档: (待添加)
