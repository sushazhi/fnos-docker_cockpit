<template>
  <div class="page">
    <div class="header">
      <span class="header-title">{{ t('home.title') }}</span>
      <button class="header-action" @click="refresh">
        <svg
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <polyline points="23 4 23 10 17 10" />
          <polyline points="1 20 1 14 7 14" />
          <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15" />
        </svg>
      </button>
    </div>

    <div v-if="loading" class="loading">
      <div class="spinner"></div>
    </div>

    <template v-else>
      <div class="section-header">
        <div class="section-icon overview">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="3" y="3" width="7" height="7" />
            <rect x="14" y="3" width="7" height="7" />
            <rect x="14" y="14" width="7" height="7" />
            <rect x="3" y="14" width="7" height="7" />
          </svg>
        </div>
        <div class="section-info">
          <div class="section-title">{{ t('home.overview') }}</div>
        </div>
      </div>
      <div class="stats-card">
        <div class="stats-grid">
          <div class="stat-item">
            <div class="stat-value">{{ stats.containers }}</div>
            <div class="stat-label">{{ t('home.containers') }}</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">{{ stats.images }}</div>
            <div class="stat-label">{{ t('home.images') }}</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">{{ stats.networks }}</div>
            <div class="stat-label">{{ t('home.networks') }}</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">{{ stats.volumes }}</div>
            <div class="stat-label">{{ t('home.volumes') }}</div>
          </div>
        </div>

        <!-- 总CPU和内存占用 -->
        <div class="resource-usage" v-if="runningContainers.length > 0">
          <div class="usage-item">
            <div class="usage-header">
              <span class="usage-label">{{ t('home.cpu') }}</span>
              <span class="usage-value">{{ totalCpuUsage.toFixed(1) }}%</span>
            </div>
            <div class="usage-bar">
              <div
                class="usage-bar-fill"
                :style="{ width: Math.min(totalCpuUsage, 100) + '%' }"
              ></div>
            </div>
          </div>
          <div class="usage-item">
            <div class="usage-header">
              <span class="usage-label">{{ t('home.mem') }}</span>
              <span class="usage-value">
                {{ formatBytes(totalMemoryBytes.used) }} / {{ formatBytes(totalMemoryBytes.limit) }}
              </span>
            </div>
            <div class="usage-bar">
              <div
                class="usage-bar-fill memory"
                :style="{ width: Math.min(totalMemoryUsage, 100) + '%' }"
              ></div>
            </div>
          </div>
        </div>
      </div>

      <div class="section-header">
        <div class="section-icon docker">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z" />
            <polyline points="22,6 12,13 2,6" />
          </svg>
        </div>
        <div class="section-info">
          <div class="section-title">{{ t('home.dockerInfo') }}</div>
        </div>
      </div>
      <div class="info-card">
        <div class="info-item" v-if="dockerInfo">
          <span class="info-label">{{ t('home.version') }}</span>
          <span class="info-value">{{ dockerInfo.ServerVersion || '-' }}</span>
        </div>
        <div class="info-item" v-if="dockerInfo">
          <span class="info-label">{{ t('home.running') }}</span>
          <span class="info-value">{{ dockerInfo.ContainersRunning || 0 }}</span>
        </div>
        <div class="info-item" v-if="dockerInfo">
          <span class="info-label">{{ t('home.os') }}</span>
          <span class="info-value">{{ dockerInfo.OperatingSystem || '-' }}</span>
        </div>
        <div class="info-item" v-if="dockerInfo">
          <span class="info-label">{{ t('home.kernel') }}</span>
          <span class="info-value">{{ dockerInfo.KernelVersion || '-' }}</span>
        </div>
        <div class="empty-state" v-if="!dockerInfo">
          <div class="empty-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10" />
              <line x1="12" y1="8" x2="12" y2="12" />
              <line x1="12" y1="16" x2="12.01" y2="16" />
            </svg>
          </div>
          <div class="empty-text">{{ t('home.dockerNotRunning') }}</div>
        </div>
      </div>

      <div class="section-header">
        <div class="section-icon running">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polygon points="5 3 19 12 5 21 5 3" />
          </svg>
        </div>
        <div class="section-info">
          <div class="section-title">{{ t('home.runningContainers') }}</div>
        </div>
      </div>
      <div class="list-card">
        <div v-if="runningContainers.length > 0">
          <div
            v-for="c in runningContainers"
            :key="c.Id"
            class="list-item"
            @click="$router.push(`/container/${c.Id}`)"
          >
            <div class="item-icon container-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="2" y="7" width="20" height="14" rx="2" ry="2" />
                <path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16" />
              </svg>
            </div>
            <div class="item-content">
              <div class="item-title">{{ getContainerName(c) }}</div>
              <div class="item-subtitle">{{ c.Image }}</div>
              <div class="item-stats" v-if="containerStats[c.Id]">
                <span class="stat-tag cpu">
                  {{ t('home.cpu') }}: {{ containerStats[c.Id].CPUPerc || '-' }}
                </span>
                <span class="stat-tag mem">
                  {{ t('home.mem') }}: {{ formatMemory(containerStats[c.Id].MemUsage) }}
                </span>
                <span
                  class="stat-tag net"
                  v-if="containerStats[c.Id].NetIORate && containerStats[c.Id].NetIORate !== '-'"
                >
                  {{ containerStats[c.Id].NetIORate }}
                </span>
                <!-- 端口映射显示 -->
                <div class="item-ports" v-if="getUniquePorts(c).length > 0">
                  <div class="port-item" v-for="port in getUniquePorts(c)" :key="port.PrivatePort">
                    <span class="port-text" @click="openPortInNewTab(port.PublicPort)">
                      {{ port.PublicPort }}:{{ port.PrivatePort }}
                    </span>
                  </div>
                </div>
              </div>
            </div>
            <span class="badge badge-success">{{ t('containers.state.running') }}</span>
          </div>
          <div class="more-link" @click="$router.push('/containers')">
            {{ t('common.viewAll') }} {{ runningContainers.length }}
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="9 18 15 12 9 6" />
            </svg>
          </div>
        </div>
        <div v-else class="empty-state">
          <div class="empty-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="2" y="7" width="20" height="14" rx="2" ry="2" />
              <path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16" />
            </svg>
          </div>
          <div class="empty-text">{{ t('common.noData') }}</div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '../services/api'

const { t } = useI18n()
const loading = ref(true)
const stats = ref({ containers: 0, images: 0, networks: 0, volumes: 0 })
const dockerInfo = ref(null)
const hostMemory = ref({ total: 0, used: 0, available: 0 })
const runningContainers = ref([])
const containerStats = ref({})
let statsInterval = null
let isVisible = true

function startPolling() {
  if (statsInterval) return
  statsInterval = setInterval(loadContainerStats, 5000)
}

function stopPolling() {
  if (statsInterval) {
    clearInterval(statsInterval)
    statsInterval = null
  }
}

function handleVisibilityChange() {
  isVisible = !document.hidden
  if (isVisible) {
    startPolling()
    loadContainerStats()
  } else {
    stopPolling()
  }
}

// 计算总CPU占用
const totalCpuUsage = computed(() => {
  if (Object.keys(containerStats.value).length === 0) return 0

  let totalCpu = 0
  for (const id in containerStats.value) {
    const stats = containerStats.value[id]
    // 使用后端返回的 CPUPerc 字段
    if (stats.CPUPerc) {
      const cpuPercent = parseFloat(stats.CPUPerc)
      if (!isNaN(cpuPercent)) {
        totalCpu += cpuPercent
      }
    }
  }

  return totalCpu
})

// 计算总内存占用 - 运行容器内存总和 / 系统总内存
const totalMemoryUsage = computed(() => {
  const totalMem = hostMemory.value.total || dockerInfo.value?.MemTotal || 0
  if (totalMem === 0) return 0

  // 累加所有运行容器的内存使用
  let containerUsed = 0
  for (const id in containerStats.value) {
    const stats = containerStats.value[id]
    if (stats.MemoryUsage !== undefined) {
      containerUsed += stats.MemoryUsage || 0
    }
  }

  return (containerUsed / totalMem) * 100
})

// 计算总内存使用量（字节）- 运行容器内存总和 / 系统总内存
const totalMemoryBytes = computed(() => {
  const totalMem = hostMemory.value.total || dockerInfo.value?.MemTotal || 0

  // 累加所有运行容器的内存使用
  let containerUsed = 0
  for (const id in containerStats.value) {
    const stats = containerStats.value[id]
    if (stats.MemoryUsage !== undefined) {
      containerUsed += stats.MemoryUsage || 0
    }
  }

  return { used: containerUsed, limit: totalMem }
})

// 格式化字节数为可读格式
function formatBytes(bytes) {
  if (bytes === 0) return '0B'

  const k = 1024
  const sizes = ['B', 'K', 'M', 'G', 'T']
  const i = Math.floor(Math.log(bytes) / Math.log(k))

  const value = bytes / Math.pow(k, i)

  // 根据大小决定小数位数
  if (i < 2) {
    return value.toFixed(0) + sizes[i]
  } else if (value >= 100) {
    return value.toFixed(0) + sizes[i]
  } else if (value >= 10) {
    return value.toFixed(1) + sizes[i]
  } else {
    return value.toFixed(2) + sizes[i]
  }
}

function getContainerName(c) {
  if (c.Names && c.Names.length > 0) {
    return c.Names[0].replace(/^\//, '')
  }
  return c.Id?.substring(0, 12) || 'unknown'
}

// 获取去重后的端口列表（处理 IPv4/IPv6 重复绑定的情况）
function getUniquePorts(c) {
  if (!c.Ports || c.Ports.length === 0) return []

  const seen = new Set()
  return c.Ports.filter(port => {
    if (!port.PublicPort) return false
    const key = `${port.PublicPort}:${port.PrivatePort}`
    if (seen.has(key)) return false
    seen.add(key)
    return true
  })
}

// 格式化内存显示，只显示使用量部分
function formatMemory(memUsage) {
  if (!memUsage) return '-'
  // MemUsage 格式为 "50MiB / 1GiB"，只取第一部分
  const parts = memUsage.split(' / ')
  return parts[0] || memUsage
}

// 在新标签页打开端口
function openPortInNewTab(port) {
  const host = window.location.hostname
  window.open(`http://${host}:${port}`, '_blank')
}

async function loadContainerStats() {
  // 并行加载所有运行中容器的统计信息（不限制数量）
  const promises = runningContainers.value.map(async c => {
    try {
      const data = await api.get(`/api/container/${c.Id}/stats`)
      if (data.stats) {
        containerStats.value[c.Id] = data.stats
      }
    } catch (e) {
      console.error('Failed to load stats for container:', c.Id, e)
    }
  })

  // 等待所有请求完成
  await Promise.all(promises)
}

async function refresh() {
  loading.value = true
  try {
    const [infoRes, containersRes, imagesRes, networksRes, volumesRes] = await Promise.all([
      api.get('/api/system/info').catch(() => null),
      api.get('/api/containers').catch(() => ({ containers: [] })),
      api.get('/api/images').catch(() => ({ images: [] })),
      api.get('/api/networks').catch(() => ({ networks: [] })),
      api.get('/api/volumes').catch(() => ({ volumes: [] }))
    ])

    dockerInfo.value = infoRes?.info || null
    // 获取宿主机内存信息
    if (infoRes?.hostMemory) {
      hostMemory.value = infoRes.hostMemory
    }
    const containers = containersRes.containers || []
    runningContainers.value = containers.filter(c => c.State === 'running')

    stats.value = {
      containers: containers.length,
      images: (imagesRes.images || []).length,
      networks: (networksRes.networks || []).length,
      volumes: (volumesRes.volumes || []).length
    }

    loadContainerStats()
  } catch (e) {
    console.error('加载数据失败:', e)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  refresh()
  startPolling()
  document.addEventListener('visibilitychange', handleVisibilityChange)
})

onUnmounted(() => {
  stopPolling()
  document.removeEventListener('visibilitychange', handleVisibilityChange)
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

.section-icon.overview {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.section-icon.docker {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
  color: white;
}

.section-icon.running {
  background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
  color: white;
}

.section-info {
  flex: 1;
}

.section-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-color);
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

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  padding: 16px 8px;
}

.stat-item {
  text-align: center;
  padding: 8px;
}

.stat-value {
  font-size: 28px;
  font-weight: 600;
  color: var(--text-color);
  letter-spacing: -0.5px;
}

.stat-label {
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 4px;
}

.resource-usage {
  padding: 16px;
  border-top: 1px solid var(--border-color);
}

.usage-item {
  margin-bottom: 12px;
}

.usage-item:last-child {
  margin-bottom: 0;
}

.usage-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}

.usage-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
}

.usage-value {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-color);
}

.usage-bar {
  height: 6px;
  background: var(--hover-bg);
  border-radius: 3px;
  overflow: hidden;
}

.usage-bar-fill {
  height: 100%;
  background: linear-gradient(90deg, #667eea 0%, #764ba2 100%);
  border-radius: 3px;
  transition: width 0.3s ease;
}

.usage-bar-fill.memory {
  background: linear-gradient(90deg, #43e97b 0%, #38f9d7 100%);
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

.list-card {
  background: var(--card-bg);
  margin: 0 12px 12px;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: var(--shadow-sm);
}

[data-theme='dark'] .list-card {
  box-shadow: none;
}

.list-item {
  display: flex;
  align-items: center;
  padding: 14px 16px;
  cursor: pointer;
  transition: background var(--transition-fast);
}

.list-item:hover {
  background: var(--hover-bg);
}

.list-item:active {
  background: var(--active-bg);
}

.list-item + .list-item {
  border-top: 1px solid var(--border-color);
}

.item-icon {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
  margin-right: 12px;
  flex-shrink: 0;
}

.item-icon svg {
  width: 20px;
  height: 20px;
}

.item-icon.container-icon {
  background: rgba(0, 125, 255, 0.1);
  color: #007dff;
}

.item-content {
  flex: 1;
  min-width: 0;
}

.item-title {
  font-size: 15px;
  font-weight: 500;
  color: var(--text-color);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-subtitle {
  font-size: 13px;
  color: var(--text-secondary);
  margin-top: 2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-stats {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 6px;
}

.stat-tag {
  font-size: 11px;
  padding: 2px 6px;
  border-radius: 4px;
  font-weight: 500;
  font-family: 'HarmonyOS Sans SC', 'SF Mono', 'Consolas', monospace;
}

.stat-tag.cpu {
  background: rgba(0, 125, 255, 0.1);
  color: #007dff;
}

.stat-tag.mem {
  background: rgba(40, 167, 69, 0.1);
  color: #28a745;
}

.stat-tag.net {
  background: rgba(255, 193, 7, 0.1);
  color: #ffc107;
}

/* 端口映射样式 */
.item-ports {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 0;
}

.port-item {
  background: rgba(102, 126, 234, 0.1);
  border: 1px solid rgba(102, 126, 234, 0.2);
  border-radius: 4px;
  padding: 2px 6px;
  font-size: 11px;
  font-weight: 500;
  font-family: 'HarmonyOS Sans SC', 'SF Mono', 'Consolas', monospace;
}

.port-text {
  font-size: 11px;
  font-weight: 500;
  color: #667eea;
  cursor: pointer;
  font-family: 'HarmonyOS Sans SC', 'SF Mono', 'Consolas', monospace;
}

.port-text:hover {
  text-decoration: underline;
}

.more-link {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  padding: 14px;
  color: #007dff;
  font-size: 15px;
  font-weight: 500;
  cursor: pointer;
  border-top: 1px solid var(--border-color);
}

.more-link svg {
  width: 16px;
  height: 16px;
}

.empty-state {
  padding: 40px 20px;
  text-align: center;
}

.empty-icon {
  width: 48px;
  height: 48px;
  margin: 0 auto 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--hover-bg);
  border-radius: 12px;
}

.empty-icon svg {
  width: 24px;
  height: 24px;
  color: var(--text-tertiary);
}

.empty-text {
  font-size: 14px;
  color: var(--text-secondary);
}

@media (max-width: 360px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .stat-value {
    font-size: 24px;
  }
}
</style>
