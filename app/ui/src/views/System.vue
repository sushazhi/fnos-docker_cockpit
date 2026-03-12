<template>
  <div class="page">
    <div class="header">
      <button class="header-back" @click="$router.back()">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
          <polyline points="15 18 9 12 15 6" />
        </svg>
      </button>
      <span class="header-title">{{ t('system.title') }}</span>
    </div>

    <div v-if="loading" class="loading">
      <div class="spinner"></div>
    </div>

    <template v-else>
      <div class="section-header">
        <div class="section-icon docker">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z" />
            <polyline points="22,6 12,13 2,6" />
          </svg>
        </div>
        <div class="section-info">
          <div class="section-title">Docker</div>
        </div>
      </div>
      <div class="info-card">
        <div class="info-item">
          <span class="info-label">{{ t('system.version') }}</span>
          <span class="info-value">{{ info.ServerVersion || '-' }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">{{ t('system.apiVersion') }}</span>
          <span class="info-value">{{ info.ApiVersion || info.APIVersion || '-' }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">{{ t('system.os') }}</span>
          <span class="info-value">{{ info.OperatingSystem || '-' }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">{{ t('system.kernel') }}</span>
          <span class="info-value">{{ info.KernelVersion || '-' }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">{{ t('system.arch') }}</span>
          <span class="info-value">{{ info.Architecture || '-' }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">{{ t('system.cpus') }}</span>
          <span class="info-value">{{ info.NCPU || '-' }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">{{ t('system.memory') }}</span>
          <span class="info-value">{{ formatMemory(systemMemory) }}</span>
        </div>
      </div>

      <div class="section-header">
        <div class="section-icon containers">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="2" y="7" width="20" height="14" rx="2" ry="2" />
            <path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16" />
          </svg>
        </div>
        <div class="section-info">
          <div class="section-title">{{ t('system.containers') }}</div>
        </div>
      </div>
      <div class="stats-card">
        <div class="stats-row">
          <div class="stat-item">
            <div class="stat-value">{{ info.Containers || 0 }}</div>
            <div class="stat-label">{{ t('system.total') }}</div>
          </div>
          <div class="stat-item">
            <div class="stat-value running">{{ info.ContainersRunning || 0 }}</div>
            <div class="stat-label">{{ t('system.running') }}</div>
          </div>
          <div class="stat-item">
            <div class="stat-value stopped">{{ info.ContainersStopped || 0 }}</div>
            <div class="stat-label">{{ t('system.stopped') }}</div>
          </div>
          <div class="stat-item">
            <div class="stat-value paused">{{ info.ContainersPaused || 0 }}</div>
            <div class="stat-label">{{ t('system.paused') }}</div>
          </div>
        </div>
      </div>

      <div class="section-header">
        <div class="section-icon images">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
            <circle cx="8.5" cy="8.5" r="1.5" />
            <polyline points="21 15 16 10 5 21" />
          </svg>
        </div>
        <div class="section-info">
          <div class="section-title">{{ t('system.images') }}</div>
        </div>
      </div>
      <div class="info-card">
        <div class="info-item">
          <span class="info-label">{{ t('system.count') }}</span>
          <span class="info-value">{{ info.Images || 0 }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">{{ t('system.diskUsage') }}</span>
          <span class="info-value">{{ formatMemory(disk.ImagesSize) }}</span>
        </div>
      </div>

      <div class="section-header">
        <div class="section-icon disk">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <ellipse cx="12" cy="5" rx="9" ry="3" />
            <path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3" />
            <path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5" />
          </svg>
        </div>
        <div class="section-info">
          <div class="section-title">{{ t('system.diskUsage') }}</div>
        </div>
      </div>
      <div class="info-card">
        <div class="info-item">
          <span class="info-label">{{ t('system.images') }}</span>
          <span class="info-value">{{ formatMemory(disk.ImagesSize) }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">{{ t('system.containers') }}</span>
          <span class="info-value">{{ formatMemory(disk.ContainersSize) }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">{{ t('system.volumes') }}</span>
          <span class="info-value">{{ formatMemory(disk.VolumesSize) }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">{{ t('system.buildCache') }}</span>
          <span class="info-value">{{ formatMemory(disk.BuildCacheSize) }}</span>
        </div>
        <div class="info-item total">
          <span class="info-label">{{ t('system.total') }}</span>
          <span class="info-value">{{ formatMemory(disk.TotalSize) }}</span>
        </div>
      </div>

      <div class="section-header">
        <div class="section-icon storage">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z" />
          </svg>
        </div>
        <div class="section-info">
          <div class="section-title">{{ t('system.storageDriver') }}</div>
        </div>
      </div>
      <div class="info-card">
        <div class="info-item">
          <span class="info-label">{{ t('system.driver') }}</span>
          <span class="info-value">{{ info.Driver || '-' }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">{{ t('system.dockerRoot') }}</span>
          <span class="info-value">{{ info.DockerRootDir || '-' }}</span>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '../services/api'

const { t } = useI18n()
const loading = ref(true)
const info = ref({})
const hostMemory = ref({ total: 0, available: 0, used: 0 })
const disk = ref({
  ImagesSize: 0,
  ContainersSize: 0,
  VolumesSize: 0,
  BuildCacheSize: 0,
  TotalSize: 0
})

// 获取系统总内存 - 优先使用宿主机物理内存
const systemMemory = computed(() => {
  // 优先使用从 /proc/meminfo 读取的宿主机物理内存
  if (hostMemory.value && hostMemory.value.total > 0) {
    return hostMemory.value.total
  }
  // 降级使用 Docker 返回的内存信息
  return (
    info.value.MemTotal || info.value.MemoryTotal || info.value.mem_total || info.value.memory || 0
  )
})

function formatMemory(bytes) {
  if (bytes === null || bytes === undefined) return '-'
  if (bytes === 0) return '0 B'
  const gb = bytes / 1024 / 1024 / 1024
  if (gb < 1) {
    const mb = bytes / 1024 / 1024
    if (mb < 1) {
      const kb = bytes / 1024
      return kb.toFixed(1) + ' KB'
    }
    return mb.toFixed(1) + ' MB'
  }
  return gb.toFixed(2) + ' GB'
}

async function load() {
  loading.value = true
  try {
    const [infoRes, diskRes] = await Promise.all([
      api.get('/api/system/info'),
      api.get('/api/system/df').catch(() => null)
    ])
    info.value = infoRes.info || {}
    // 获取宿主机物理内存信息
    hostMemory.value = infoRes.hostMemory || { total: 0, available: 0, used: 0 }
    // 调试：打印内存相关字段
    console.log('System info:', info.value)
    console.log('Host memory:', hostMemory.value)
    if (diskRes) {
      disk.value = {
        ImagesSize: diskRes.ImagesSize || 0,
        ContainersSize: diskRes.ContainersSize || 0,
        VolumesSize: diskRes.VolumesSize || 0,
        BuildCacheSize: diskRes.BuildCacheSize || 0,
        TotalSize: diskRes.TotalSize || 0
      }
    }
  } catch (e) {
    console.error('Failed to load system info:', e)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  load()
})
</script>

<style scoped>
.section-header {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  gap: 12px;
}

.section-icon {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 10px;
  flex-shrink: 0;
}

.section-icon svg {
  width: 18px;
  height: 18px;
}

.section-icon.docker {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
  color: white;
}

.section-icon.containers {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.section-icon.images {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  color: white;
}

.section-icon.disk {
  background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
  color: white;
}

.section-icon.storage {
  background: linear-gradient(135deg, #fa709a 0%, #fee140 100%);
  color: white;
}

.section-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-color);
}

.info-card {
  background: var(--card-bg);
  margin: 0 12px 12px;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: var(--shadow-sm);
}

[data-theme='dark'] .info-card {
  box-shadow: none;
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 16px;
}

.info-item + .info-item {
  border-top: 1px solid var(--border-color);
}

.info-item.total {
  background: var(--hover-bg);
}

.info-label {
  font-size: 15px;
  color: var(--text-color);
}

.info-value {
  font-size: 14px;
  color: var(--text-secondary);
  text-align: right;
  max-width: 60%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.stats-card {
  background: var(--card-bg);
  margin: 0 12px 12px;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: var(--shadow-sm);
}

[data-theme='dark'] .stats-card {
  box-shadow: none;
}

.stats-row {
  display: flex;
  padding: 16px 8px;
}

.stat-item {
  flex: 1;
  text-align: center;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-color);
}

.stat-value.running {
  color: #00c853;
}

.stat-value.stopped {
  color: #fa2a2d;
}

.stat-value.paused {
  color: #ff9800;
}

.stat-label {
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 4px;
}
</style>
