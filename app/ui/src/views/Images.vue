<template>
  <div class="page">
    <div class="header">
      <button class="header-back" @click="$router.back()">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
          <polyline points="15 18 9 12 15 6" />
        </svg>
      </button>
      <span class="header-title">{{ t('images.title') }}</span>
      <button class="header-action" @click="refresh">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="23 4 23 10 17 10" />
          <polyline points="1 20 1 14 7 14" />
          <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15" />
        </svg>
      </button>
    </div>

    <div v-if="loading" class="loading">
      <div class="spinner"></div>
    </div>

    <div v-else-if="images.length === 0" class="empty-state">
      <div class="empty-icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
          <circle cx="8.5" cy="8.5" r="1.5" />
          <polyline points="21 15 16 10 5 21" />
        </svg>
      </div>
      <div class="empty-text">{{ t('common.noData') }}</div>
    </div>

    <!-- 正在拉取的镜像列表 -->
    <div v-if="pullingImages.length > 0" class="list-card pulling-section">
      <div class="section-header">正在拉取</div>
      <div v-for="item in pullingImages" :key="item.name" class="list-item pulling-item">
        <div class="item-icon pulling-icon">
          <svg
            v-if="item.status === 'pulling'"
            class="spin"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <circle cx="12" cy="12" r="10" />
            <path d="M12 6v6l4 2" />
          </svg>
          <svg
            v-else-if="item.status === 'success'"
            viewBox="0 0 24 24"
            fill="none"
            stroke="#10b981"
            stroke-width="2"
          >
            <polyline points="20 6 9 17 4 12" />
          </svg>
          <svg v-else viewBox="0 0 24 24" fill="none" stroke="#ef4444" stroke-width="2">
            <circle cx="12" cy="12" r="10" />
            <line x1="15" y1="9" x2="9" y2="15" />
            <line x1="9" y1="9" x2="15" y2="15" />
          </svg>
        </div>
        <div class="item-content">
          <div class="item-title">{{ item.name }}</div>
          <div class="item-subtitle">{{ item.progress }}</div>
          <!-- 进度条 -->
          <div
            v-if="item.status === 'pulling' && item.percent !== undefined"
            class="progress-bar-container"
          >
            <div class="progress-bar" :style="{ width: item.percent + '%' }"></div>
            <span class="progress-text">{{ item.percent }}%</span>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="list-card">
      <div v-for="img in images" :key="img.Id" class="list-item" @click="showImageActions(img)">
        <div class="item-icon image-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
            <circle cx="8.5" cy="8.5" r="1.5" />
            <polyline points="21 15 16 10 5 21" />
          </svg>
        </div>
        <div class="item-content">
          <div class="item-title-row">
            <span class="item-title">{{ getImageName(img) }}</span>
            <span v-if="getContainerCount(img) > 0" class="in-use-badge">
              使用中 ({{ getContainerCount(img) }})
            </span>
            <span v-else class="in-use-badge" style="background: #999">未使用</span>
            <span v-if="img.hasUpdate" class="update-badge" @click.stop="updateImage(img)">
              {{ t('images.canUpdate') }}
            </span>
          </div>
          <div class="item-subtitle">
            {{ img.Id?.substring(7, 19) }} · {{ formatSize(img.Size) }} ·
            {{ formatDate(img.Created) }}
          </div>
        </div>
        <div class="item-arrow">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="9 18 15 12 9 6" />
          </svg>
        </div>
      </div>
    </div>

    <!-- 底部操作按钮组 -->
    <div class="fab-container">
      <button class="fab" @click="showPullModal = true" :title="t('images.pull')">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="12" y1="5" x2="12" y2="19" />
          <line x1="5" y1="12" x2="19" y2="12" />
        </svg>
      </button>
    </div>

    <!-- 拉取镜像模态框 -->
    <div v-if="showPullModal" class="dialog-overlay" @click.self="showPullModal = false">
      <div class="dialog">
        <div class="dialog-header">
          <h3 class="dialog-title">{{ t('images.pull') }}</h3>
          <button class="dialog-close" @click="showPullModal = false">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>
        <div class="dialog-body">
          <div class="form-field">
            <label class="form-label">{{ t('images.imageName') }}</label>
            <input
              type="text"
              class="form-input"
              v-model="pullImageName"
              placeholder="nginx"
              @keyup.enter="pullImage"
            />
          </div>

          <div class="form-field">
            <label class="form-label">{{ t('images.tag') }}</label>
            <input
              type="text"
              class="form-input"
              v-model="pullImageTag"
              placeholder="latest"
              @keyup.enter="pullImage"
            />
            <div class="form-hint">{{ t('images.tagHint') }}</div>
          </div>

          <div class="form-field">
            <label class="form-label">{{ t('images.platform') }}</label>
            <select class="form-input" v-model="pullPlatform">
              <option value="">{{ t('images.autoDetect') }}</option>
              <option value="linux/amd64">linux/amd64</option>
              <option value="linux/arm64">linux/arm64</option>
            </select>
            <div class="form-hint">{{ t('images.platformHint') }}</div>
          </div>

          <div v-if="pulling" class="pull-progress">
            <div class="spinner"></div>
            <span>{{ t('images.pulling') }}</span>
          </div>
        </div>
        <div class="dialog-footer">
          <button class="dialog-btn secondary" @click="showPullModal = false">
            {{ t('common.cancel') }}
          </button>
          <button
            class="dialog-btn primary"
            @click="pullImage"
            :disabled="pulling || !pullImageName"
          >
            {{ t('images.pull') }}
          </button>
        </div>
      </div>
    </div>

    <!-- 构建镜像模态框 -->
    <div v-if="showBuildModal" class="dialog-overlay" @click.self="showBuildModal = false">
      <div class="dialog dialog-large">
        <div class="dialog-header">
          <h3 class="dialog-title">{{ t('images.build') }}</h3>
          <button class="dialog-close" @click="showBuildModal = false">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>
        <div class="dialog-body">
          <div class="form-field">
            <label class="form-label">{{ t('images.imageName') }}</label>
            <input
              type="text"
              class="form-input"
              v-model="buildImageName"
              placeholder="my-app:latest"
            />
          </div>

          <div class="form-field">
            <label class="form-label">Dockerfile 路径</label>
            <input
              type="text"
              class="form-input"
              v-model="buildDockerfile"
              placeholder="./Dockerfile"
            />
            <div class="form-hint">Dockerfile 文件的路径,默认为 ./Dockerfile</div>
          </div>

          <div class="form-field">
            <label class="form-label">构建上下文</label>
            <input type="text" class="form-input" v-model="BuildContext" placeholder="." />
            <div class="form-hint">构建上下文目录,默认为当前目录</div>
          </div>

          <div v-if="building" class="pull-progress">
            <div class="spinner"></div>
            <span>构建中...</span>
          </div>
        </div>
        <div class="dialog-footer">
          <button class="dialog-btn secondary" @click="showBuildModal = false">
            {{ t('common.cancel') }}
          </button>
          <button
            class="dialog-btn primary"
            @click="buildImage"
            :disabled="building || !buildImageName"
          >
            {{ t('images.build') }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="showActions" class="action-sheet-overlay" @click.self="showActions = false">
      <div class="action-sheet">
        <div class="action-sheet-handle"></div>
        <div class="action-sheet-content">
          <button class="sheet-btn" @click="createContainer">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="2" y="7" width="20" height="14" rx="2" ry="2" />
              <path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16" />
            </svg>
            {{ t('images.createContainer') }}
          </button>
          <button class="sheet-btn" @click="showEditTagsDialog">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path
                d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"
              />
              <line x1="7" y1="7" x2="7.01" y2="7" />
            </svg>
            编辑标签
          </button>
          <button class="sheet-btn" @click="showDetectUpgradeDialog">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
              <polyline points="7 10 12 15 17 10" />
              <line x1="12" y1="15" x2="12" y2="3" />
            </svg>
            检测升级
          </button>
          <button
            v-if="selectedImage?.hasUpdate"
            class="sheet-btn update"
            @click="updateSelectedImage"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
              <polyline points="7 10 12 15 17 10" />
              <line x1="12" y1="15" x2="12" y2="3" />
            </svg>
            {{ t('images.update') }}
          </button>
          <button class="sheet-btn danger" @click="confirmRemove">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="3 6 5 6 21 6" />
              <path
                d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"
              />
            </svg>
            {{ t('images.remove') }}
          </button>
        </div>
      </div>
    </div>

    <ConfirmModal
      v-if="showConfirm"
      :title="t('images.remove')"
      :message="
        t('common.confirmDelete') + ' ' + (selectedImage ? getImageName(selectedImage) : '') + '?'
      "
      :confirm-text="t('images.remove')"
      danger
      @close="showConfirm = false"
      @confirm="removeImage"
    />

    <!-- 编辑标签对话框 -->
    <div v-if="showEditTagsModal" class="dialog-overlay" @click.self="showEditTagsModal = false">
      <div class="dialog">
        <div class="dialog-header">
          <h3 class="dialog-title">编辑标签</h3>
          <button class="dialog-close" @click="showEditTagsModal = false">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>
        <div class="dialog-body">
          <div class="form-field">
            <label class="form-label">当前标签</label>
            <input type="text" class="form-input" :value="editTagsCurrent" readonly />
          </div>
          <div class="form-field">
            <label class="form-label">新仓库名</label>
            <input type="text" class="form-input" v-model="editTagsRepo" placeholder="nginx" />
          </div>
          <div class="form-field">
            <label class="form-label">新标签</label>
            <input type="text" class="form-input" v-model="editTagsTag" placeholder="latest" />
          </div>
          <div
            class="form-hint"
            style="
              background: var(--bg-secondary);
              padding: 12px;
              border-radius: 8px;
              margin-top: 16px;
            "
          >
            编辑标签会创建新标签并删除旧标签。如果新旧标签相同则不会进行任何操作。
          </div>
        </div>
        <div class="dialog-footer">
          <button class="dialog-btn secondary" @click="showEditTagsModal = false">取消</button>
          <button class="dialog-btn primary" @click="saveEditTags" :disabled="editingTags">
            {{ editingTags ? '保存中...' : '确定' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 检测升级对话框 -->
    <div v-if="showDetectUpgradeModal" class="dialog-overlay" @click.self="closeDetectUpgrade">
      <div class="dialog dialog-large">
        <div class="dialog-header">
          <h3 class="dialog-title">检测升级: {{ detectUpgradeImageName }}</h3>
          <button class="dialog-close" @click="closeDetectUpgrade">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>
        <div class="dialog-body">
          <!-- 进度条 -->
          <div v-if="detectUpgradeLoading" class="upgrade-progress">
            <div class="progress-header">
              <span class="progress-title">{{ detectUpgradeStep }}</span>
              <span class="progress-percent">{{ detectUpgradePercent }}%</span>
            </div>
            <div class="progress-bar-container">
              <div class="progress-bar" :style="{ width: detectUpgradePercent + '%' }"></div>
            </div>
          </div>

          <!-- 检测步骤 -->
          <div class="upgrade-steps">
            <div
              v-for="(step, index) in detectUpgradeSteps"
              :key="index"
              class="upgrade-step"
              :class="{ active: step.active, completed: step.completed }"
            >
              <div class="step-indicator">
                <span v-if="step.completed" class="step-check">✓</span>
                <span v-else class="step-number">{{ index + 1 }}</span>
              </div>
              <div class="step-content">
                <div class="step-title">{{ step.title }}</div>
                <div class="step-desc">{{ step.desc }}</div>
              </div>
              <div class="step-meta">
                <span v-if="step.percent" class="step-percent">{{ step.percent }}%</span>
                <span v-if="step.time" class="step-time">{{ step.time }}</span>
              </div>
            </div>
          </div>

          <!-- 可升级版本列表 -->
          <div
            v-if="!detectUpgradeLoading && upgradableVersions.length > 0"
            class="upgrade-versions"
          >
            <h4 class="versions-title">可升级版本</h4>
            <div class="versions-list">
              <div
                v-for="version in upgradableVersions"
                :key="version"
                class="version-item"
                @click="upgradeToVersion(version)"
              >
                <span class="version-name">{{ version }}</span>
                <button class="version-upgrade-btn">升级</button>
              </div>
            </div>
          </div>

          <!-- 无可用升级 -->
          <div
            v-if="!detectUpgradeLoading && detectUpgradeComplete && upgradableVersions.length === 0"
            class="upgrade-empty"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10" />
              <line x1="12" y1="8" x2="12" y2="12" />
              <line x1="12" y1="16" x2="12.01" y2="16" />
            </svg>
            <span>当前已是最新版本</span>
          </div>
        </div>
        <div class="dialog-footer">
          <button class="dialog-btn secondary" @click="closeDetectUpgrade">关闭</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { inject } from 'vue'
import api from '../services/api'
import ConfirmModal from '../components/ConfirmModal.vue'

const { t } = useI18n()
const router = useRouter()
const showToast = inject('showToast')

const loading = ref(true)
const images = ref([])
const showPullModal = ref(false)
const pullImageName = ref('')
const pullImageTag = ref('latest')
const pullPlatform = ref('')
const pulling = ref(false)
const pullingImages = ref([]) // 正在拉取的镜像列表
const showBuildModal = ref(false)
const buildImageName = ref('')
const buildDockerfile = ref('./Dockerfile')
const buildContext = ref('.')
const building = ref(false)
const showActions = ref(false)
const showConfirm = ref(false)
const selectedImage = ref(null)
const updating = ref(false)

// 编辑标签相关
const showEditTagsModal = ref(false)
const editTagsCurrent = ref('')
const editTagsRepo = ref('')
const editTagsTag = ref('')
const editingTags = ref(false)

// 检测升级相关
const showDetectUpgradeModal = ref(false)
const detectUpgradeImageName = ref('')
const detectUpgradeLoading = ref(false)
const detectUpgradeComplete = ref(false)
const detectUpgradePercent = ref(0)
const detectUpgradeStep = ref('')
const upgradableVersions = ref([])
const detectUpgradeSteps = ref([
  {
    title: '准备中...',
    desc: '开始镜像升级检测...',
    percent: 5,
    time: '',
    active: true,
    completed: false
  },
  {
    title: '获取本地镜像信息...',
    desc: '获取本地镜像信息...',
    percent: 15,
    time: '',
    active: false,
    completed: false
  },
  {
    title: '统计使用此镜像的容器...',
    desc: '统计使用此镜像的容器...',
    percent: 25,
    time: '',
    active: false,
    completed: false
  },
  {
    title: '检查远程镜像版本...',
    desc: '检查远程镜像版本...',
    percent: 40,
    time: '',
    active: false,
    completed: false
  }
])

function getImageName(img) {
  if (img.RepoTags && img.RepoTags.length > 0 && img.RepoTags[0] !== '<none>:<none>') {
    let tag = img.RepoTags[0]
    // 移除加速源前缀
    const registry = localStorage.getItem('docker_registry_mirror') || ''
    if (registry) {
      let cleanRegistry = registry.trim()
      // 移除协议前缀
      cleanRegistry = cleanRegistry.replace(/^https?:\/\//, '')
      // 移除末尾的斜杠
      cleanRegistry = cleanRegistry.replace(/\/$/, '')
      if (cleanRegistry && tag.startsWith(cleanRegistry + '/')) {
        tag = tag.substring(cleanRegistry.length + 1)
      }
    }
    return tag
  }
  if (img.RepoDigests && img.RepoDigests.length > 0) {
    let digest = img.RepoDigests[0].split('@')[0]
    // 移除加速源前缀
    const registry = localStorage.getItem('docker_registry_mirror') || ''
    if (registry) {
      let cleanRegistry = registry.trim()
      // 移除协议前缀
      cleanRegistry = cleanRegistry.replace(/^https?:\/\//, '')
      // 移除末尾的斜杠
      cleanRegistry = cleanRegistry.replace(/\/$/, '')
      if (cleanRegistry && digest.startsWith(cleanRegistry + '/')) {
        digest = digest.substring(cleanRegistry.length + 1)
      }
    }
    return digest
  }
  return img.Id?.substring(7, 19) || 'unknown'
}

// 获取镜像被容器使用的数量
// Docker API 返回 -1 表示未计算，需要特殊处理
function getContainerCount(img) {
  const count = img.Containers ?? img.containers ?? -1
  // -1 表示未计算，返回 -1 让调用方知道状态未知
  return count
}

function formatSize(bytes) {
  if (!bytes) return '-'
  const mb = bytes / 1024 / 1024
  if (mb < 1024) return mb.toFixed(1) + ' MB'
  return (mb / 1024).toFixed(2) + ' GB'
}

function formatDate(timestamp) {
  if (!timestamp) return '-'
  const date = new Date(timestamp * 1000)
  const now = new Date()
  const diff = now - date
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))

  if (days === 0) {
    const hours = Math.floor(diff / (1000 * 60 * 60))
    if (hours === 0) {
      const minutes = Math.floor(diff / (1000 * 60))
      return minutes <= 1
        ? t('images.justNow')
        : `${minutes}${t('images.minutesAgo').replace('{n}', '')}`
    }
    return hours === 1 ? t('images.hourAgo') : `${hours}${t('images.hoursAgo').replace('{n}', '')}`
  } else if (days === 1) {
    return t('images.yesterday')
  } else if (days < 7) {
    return `${days}${t('images.daysAgoText')}`
  } else {
    // 显示完整的日期时间，格式：2026/2/24 15:30
    return (
      date.toLocaleDateString() +
      ' ' +
      date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    )
  }
}

async function refresh() {
  loading.value = true
  try {
    const data = await api.get('/api/images')
    images.value = data.images || []
    checkUpdates()
  } catch (e) {
    console.error('Failed to load images:', e)
  } finally {
    loading.value = false
  }
}

async function checkUpdates() {
  // 获取镜像加速源
  let registry = localStorage.getItem('docker_registry_mirror') || ''
  registry = registry.replace(/[`'"]/g, '').trim()

  // 如果没有配置加速源，先检查 Docker Hub 是否可连接
  if (!registry) {
    try {
      const checkData = await api.get('/api/image/search/check')
      if (!checkData.available) {
        return
      }
    } catch (e) {
      return
    }
  }

  for (const img of images.value) {
    if (img.RepoTags && img.RepoTags[0] && img.RepoTags[0] !== '<none>:<none>') {
      try {
        const url = registry
          ? `/api/image/${img.Id}/check-update?registry=${encodeURIComponent(registry)}`
          : `/api/image/${img.Id}/check-update`
        const result = await api.get(url)
        img.hasUpdate = result.hasUpdate
      } catch (e) {
        img.hasUpdate = false
      }
    }
  }
}

async function pullImage() {
  if (!pullImageName.value) return
  pulling.value = true
  try {
    // 清理和规范化镜像名称
    let imageName = pullImageName.value.trim().toLowerCase()
    let tag = (pullImageTag.value || 'latest').trim().toLowerCase()

    // 移除镜像名称中的非法字符
    imageName = imageName.replace(/^[\/]+/, '') // 移除开头的斜杠
    imageName = imageName.replace(/[\/]+$/, '') // 移除结尾的斜杠
    imageName = imageName.replace(/\/+/g, '/') // 替换连续的斜杠为单个斜杠

    // 如果镜像名称中已经包含标签，提取出来
    if (imageName.includes(':')) {
      const parts = imageName.split(':')
      imageName = parts[0]
      tag = parts[1]
    }

    // 构建完整的镜像名称
    const fullName = `${imageName}:${tag}`

    // 获取并清理镜像加速源
    let registry = localStorage.getItem('docker_registry_mirror') || ''
    registry = registry.replace(/[`'"]/g, '') // 移除反引号、单引号、双引号
    registry = registry.trim()

    const payload = {
      image: fullName,
      platform: pullPlatform.value || undefined,
      registry: registry
    }

    // 先关闭弹窗，把镜像加入正在拉取列表
    showPullModal.value = false

    // 检查是否使用了镜像加速源
    const hasRegistry = !!registry
    const pullingImage = {
      name: fullName, // 显示名称（不包含加速源）
      actualName: hasRegistry ? `${registry}/${fullName}` : fullName, // 实际拉取的完整名称
      status: 'pulling',
      progress: hasRegistry ? '准备拉取 (使用加速源)...' : '准备拉取...',
      usingRegistry: hasRegistry,
      registry: registry
    }
    pullingImages.value.push(pullingImage)

    // 异步拉取镜像
    pullImageAsync(payload, pullingImage)

    // 清空输入
    pullImageName.value = ''
    pullImageTag.value = 'latest'
    pullPlatform.value = ''
  } catch (e) {
    const details = e.Details || e.details || ''
    showToast(t('images.pullFailed') + ': ' + e.message + (details ? ' - ' + details : ''))
  } finally {
    pulling.value = false
  }
}

// 异步拉取镜像（使用 SSE 获取实时进度）
async function pullImageAsync(payload, pullingImage) {
  try {
    pullingImage.progress = '正在连接...'
    pullingImage.percent = 0

    // 使用 SSE 获取实时进度
    const params = new URLSearchParams({
      image: payload.image,
      registry: payload.registry || ''
    })

    const eventSource = new EventSource(`/api/image/pull-stream?${params}`)

    // 监听进度事件
    eventSource.addEventListener('progress', (/* event */) => {
      const data = JSON.parse(event.data)
      // 解析进度信息
      try {
        const progressData = JSON.parse(data.data)
        if (progressData.status) {
          pullingImage.progress = progressData.status
        }
        // 尝试从 progress 字段解析
        if (progressData.progress) {
          pullingImage.progress = progressData.progress
        }
        // 尝试从 progressDetail 解析百分比
        if (progressData.progressDetail) {
          const { current, total } = progressData.progressDetail
          if (current && total) {
            pullingImage.percent = Math.round((current / total) * 100)
          }
        }
      } catch (e) {
        pullingImage.progress = data.data
      }
    })

    // 监听错误事件
    eventSource.addEventListener('error', (/* event */) => {
      const data = JSON.parse(event.data)
      pullingImage.status = 'error'
      pullingImage.progress = '拉取失败'
      showToast(t('images.pullFailed') + ': ' + data.message)
      eventSource.close()
    })

    // 监听完成事件
    eventSource.addEventListener('complete', (/* event */) => {
      pullingImage.status = 'success'
      pullingImage.progress = '拉取完成'
      pullingImage.percent = 100
      showToast(t('images.pullSuccess'))
      refresh() // 刷新镜像列表
      eventSource.close()

      // 3秒后从列表中移除
      setTimeout(() => {
        const index = pullingImages.value.indexOf(pullingImage)
        if (index > -1) {
          pullingImages.value.splice(index, 1)
        }
      }, 3000)
    })

    eventSource.onerror = error => {
      console.error('SSE error:', error)
      pullingImage.status = 'error'
      pullingImage.progress = '连接失败'
      eventSource.close()
    }
  } catch (e) {
    pullingImage.status = 'error'
    pullingImage.progress = '拉取失败'
    const details = e.Details || e.details || ''
    showToast(t('images.pullFailed') + ': ' + e.message + (details ? ' - ' + details : ''))
  }
}

async function buildImage() {
  if (!buildImageName.value) return
  building.value = true
  try {
    const payload = {
      name: buildImageName.value,
      dockerfile: buildDockerfile.value || './Dockerfile',
      context: buildContext.value || '.'
    }

    await api.post('/api/image/build', payload)
    showToast('镜像构建成功')
    showBuildModal.value = false
    buildImageName.value = ''
    buildDockerfile.value = './Dockerfile'
    buildContext.value = '.'
    refresh()
  } catch (e) {
    showToast('构建失败: ' + e.message)
  } finally {
    building.value = false
  }
}

function showImageActions(img) {
  selectedImage.value = img
  showActions.value = true
}

function createContainer() {
  showActions.value = false
  const imageName = getImageName(selectedImage.value)
  router.push(`/containers?image=${encodeURIComponent(imageName)}`)
}

function confirmRemove() {
  showActions.value = false
  showConfirm.value = true
}

async function removeImage() {
  try {
    const imageId = selectedImage.value.Id
    await api.post(`/api/image/${imageId}/remove`, { force: true })
    showToast(t('images.removeSuccess'))
    showConfirm.value = false
    // 清除选中的镜像，避免刷新时出现状态混乱
    selectedImage.value = null
    // 延迟刷新，确保Docker完成删除操作
    setTimeout(refresh, 500)
  } catch (e) {
    showToast(t('images.removeFailed') + ': ' + e.message)
  }
}

async function updateImage(img) {
  if (updating.value) return
  updating.value = true
  try {
    // 获取镜像加速源
    let registry = localStorage.getItem('docker_registry_mirror') || ''
    registry = registry.replace(/[`'"]/g, '').trim()

    await api.post(`/api/image/${img.Id}/update`, { registry: registry || undefined })
    showToast(t('images.updateSuccess'))
    refresh()
  } catch (e) {
    showToast(t('images.updateFailed') + ': ' + e.message)
  } finally {
    updating.value = false
  }
}

async function updateSelectedImage() {
  if (!selectedImage.value || updating.value) return
  showActions.value = false
  updating.value = true
  try {
    // 获取镜像加速源
    let registry = localStorage.getItem('docker_registry_mirror') || ''
    registry = registry.replace(/[`'"]/g, '').trim()

    await api.post(`/api/image/${selectedImage.value.Id}/update`, {
      registry: registry || undefined
    })
    showToast(t('images.updateSuccess'))
    refresh()
  } catch (e) {
    showToast(t('images.updateFailed') + ': ' + e.message)
  } finally {
    updating.value = false
  }
}

// 编辑标签功能
function showEditTagsDialog() {
  if (!selectedImage.value) return
  showActions.value = false

  const imageName = getImageName(selectedImage.value)
  const parts = imageName.split(':')
  editTagsRepo.value = parts[0] || ''
  editTagsTag.value = parts[1] || 'latest'
  editTagsCurrent.value = imageName

  showEditTagsModal.value = true
}

async function saveEditTags() {
  if (!selectedImage.value || !editTagsRepo.value || !editTagsTag.value) return

  editingTags.value = true
  try {
    // 清理仓库名中的加速源前缀
    let cleanRepo = editTagsRepo.value.trim()
    const registry = localStorage.getItem('docker_registry_mirror') || ''
    if (registry) {
      let cleanRegistry = registry.trim()
      // 移除协议前缀
      cleanRegistry = cleanRegistry.replace(/^https?:\/\//, '')
      // 移除末尾的斜杠
      cleanRegistry = cleanRegistry.replace(/\/$/, '')
      // 如果仓库名以加速源开头，移除它
      if (cleanRegistry && cleanRepo.startsWith(cleanRegistry + '/')) {
        cleanRepo = cleanRepo.substring(cleanRegistry.length + 1)
      }
    }

    const newTag = `${cleanRepo}:${editTagsTag.value.trim()}`
    await api.post(`/api/image/${selectedImage.value.Id}/edit-tags`, {
      tags: [newTag]
    })
    showToast('标签编辑成功')
    showEditTagsModal.value = false
    refresh()
  } catch (e) {
    showToast('标签编辑失败: ' + e.message)
  } finally {
    editingTags.value = false
  }
}

// 检测升级功能
function showDetectUpgradeDialog() {
  if (!selectedImage.value) return
  showActions.value = false

  const imageName = getImageName(selectedImage.value)
  detectUpgradeImageName.value = imageName
  showDetectUpgradeModal.value = true

  // 重置状态
  detectUpgradeLoading.value = true
  detectUpgradeComplete.value = false
  detectUpgradePercent.value = 5
  detectUpgradeStep.value = '准备中...'
  upgradableVersions.value = []

  // 重置步骤状态
  detectUpgradeSteps.value.forEach((step, index) => {
    step.active = index === 0
    step.completed = false
    step.time = new Date().toLocaleTimeString()
  })

  // 开始检测
  detectUpgrade()
}

async function detectUpgrade() {
  if (!selectedImage.value) return

  try {
    // 模拟进度更新
    const steps = detectUpgradeSteps.value

    // 步骤1: 准备中
    await delay(500)
    steps[0].completed = true
    steps[0].active = false
    steps[1].active = true
    detectUpgradePercent.value = 15
    detectUpgradeStep.value = steps[1].title

    // 步骤2: 获取本地镜像信息
    await delay(800)
    steps[1].completed = true
    steps[1].active = false
    steps[2].active = true
    detectUpgradePercent.value = 25
    detectUpgradeStep.value = steps[2].title

    // 步骤3: 统计容器
    await delay(600)
    steps[2].completed = true
    steps[2].active = false
    steps[3].active = true
    detectUpgradePercent.value = 40
    detectUpgradeStep.value = steps[3].title

    // 步骤4: 检查远程版本
    let registry = localStorage.getItem('docker_registry_mirror') || ''
    registry = registry.replace(/[`'"]/g, '').trim()

    const url = registry
      ? `/api/image/${selectedImage.value.Id}/detect-upgrade?registry=${encodeURIComponent(registry)}`
      : `/api/image/${selectedImage.value.Id}/detect-upgrade`

    const result = await api.get(url)

    steps[3].completed = true
    steps[3].active = false
    detectUpgradePercent.value = 100
    detectUpgradeStep.value = '检测完成'

    if (result.available && result.upgradableVersions) {
      upgradableVersions.value = result.upgradableVersions
    }

    detectUpgradeComplete.value = true
    detectUpgradeLoading.value = false
  } catch (e) {
    showToast('检测升级失败: ' + e.message)
    detectUpgradeLoading.value = false
    detectUpgradeComplete.value = true
  }
}

function closeDetectUpgrade() {
  showDetectUpgradeModal.value = false
}

async function upgradeToVersion(version) {
  if (!selectedImage.value || updating.value) return

  updating.value = true
  try {
    const imageName = detectUpgradeImageName.value.split(':')[0]
    const fullImage = `${imageName}:${version}`
    const oldImageName = detectUpgradeImageName.value

    // 拉取新版本
    await api.post('/api/image/pull', {
      image: fullImage,
      registry: localStorage.getItem('docker_registry_mirror') || undefined
    })

    // 获取使用该镜像的容器
    const containersData = await api.get('/api/containers')
    const containers = containersData.containers || []
    const usingContainers = containers.filter(
      container => container.Image === oldImageName || container.ImageID === selectedImage.value.Id
    )

    // 重启使用该镜像的容器
    if (usingContainers.length > 0) {
      for (const container of usingContainers) {
        try {
          await api.post(`/api/container/${container.Id}/restart`, { timeout: 10 })
        } catch (e) {
          console.error(`重启容器 ${container.Names[0]} 失败:`, e)
        }
      }
      showToast(`已升级到版本: ${version}，并重启了 ${usingContainers.length} 个容器`)
    } else {
      showToast(`已升级到版本: ${version}`)
    }

    // 询问是否删除旧镜像
    if (confirm('升级完成，是否删除旧版本的镜像？')) {
      try {
        await api.post(`/api/image/${selectedImage.value.Id}/remove`, { force: true })
        showToast('旧镜像已删除')
      } catch (e) {
        showToast('删除旧镜像失败: ' + e.message)
      }
    }

    closeDetectUpgrade()
    refresh()
  } catch (e) {
    showToast('升级失败: ' + e.message)
  } finally {
    updating.value = false
  }
}

function delay(ms) {
  return new Promise(resolve => setTimeout(resolve, ms))
}

onMounted(() => {
  refresh()
})
</script>

<style scoped>
.list-card {
  background: var(--card-bg);
  margin: 12px;
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

.item-icon.image-icon {
  background: rgba(102, 126, 234, 0.1);
  color: #667eea;
}

.item-content {
  flex: 1;
  min-width: 0;
}

.item-title-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.item-title {
  font-size: 15px;
  font-weight: 500;
  color: var(--text-color);
  word-break: break-all;
}

.in-use-badge {
  font-size: 11px;
  font-weight: 500;
  color: #fff;
  background: #007dff;
  padding: 2px 8px;
  border-radius: 10px;
  white-space: nowrap;
  flex-shrink: 0;
  margin-left: 8px;
}

.update-badge {
  font-size: 11px;
  font-weight: 500;
  color: #fff;
  background: #28a745;
  padding: 2px 8px;
  border-radius: 10px;
  cursor: pointer;
  white-space: nowrap;
  flex-shrink: 0;
  margin-left: 8px;
}

.update-badge:hover {
  background: #218838;
}

.item-subtitle {
  font-size: 13px;
  color: var(--text-secondary);
  margin-top: 2px;
  font-family: 'HarmonyOS Sans SC', 'SF Mono', 'Consolas', monospace;
}

.item-arrow {
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-tertiary);
  flex-shrink: 0;
}

.item-arrow svg {
  width: 18px;
  height: 18px;
}

.empty-state {
  padding: 60px 20px;
  text-align: center;
}

.empty-icon {
  width: 56px;
  height: 56px;
  margin: 0 auto 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--hover-bg);
  border-radius: 16px;
}

.empty-icon svg {
  width: 28px;
  height: 28px;
  color: var(--text-tertiary);
}

.empty-text {
  font-size: 14px;
  color: var(--text-secondary);
}

.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 16px;
}

.dialog {
  background: var(--card-bg);
  border-radius: 20px;
  width: 100%;
  max-width: 360px;
  overflow: hidden;
  box-shadow: var(--shadow-lg);
}

.dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px;
  border-bottom: 1px solid var(--border-color);
}

.dialog-title {
  font-size: 17px;
  font-weight: 600;
  color: var(--text-color);
}

.dialog-close {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: none;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  border-radius: 8px;
}

.dialog-close svg {
  width: 20px;
  height: 20px;
}

.dialog-body {
  padding: 20px;
}

.form-field {
  margin-bottom: 16px;
}

.form-field:last-child {
  margin-bottom: 0;
}

.form-label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
  margin-bottom: 8px;
}

.form-input {
  width: 100%;
  padding: 12px 14px;
  background: var(--input-bg);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  color: var(--text-color);
  font-size: 15px;
  font-family: inherit;
}

.form-input:focus {
  outline: none;
  border-color: #007dff;
  box-shadow: 0 0 0 3px rgba(0, 125, 255, 0.12);
}

.form-hint {
  font-size: 12px;
  color: var(--text-tertiary);
  margin-top: 6px;
  line-height: 1.4;
}

.pull-progress {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px;
  background: var(--hover-bg);
  border-radius: 12px;
  font-size: 14px;
  color: var(--text-secondary);
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid var(--border-color);
}

.dialog-btn {
  padding: 10px 20px;
  border: none;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
}

.dialog-btn.primary {
  background: #007dff;
  color: white;
}

.dialog-btn.primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.dialog-btn.secondary {
  background: var(--hover-bg);
  color: var(--text-color);
}

.action-sheet-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: flex-end;
  z-index: 1000;
}

.action-sheet {
  background: var(--card-bg);
  border-radius: 20px 20px 0 0;
  width: 100%;
}

.action-sheet-handle {
  width: 36px;
  height: 4px;
  background: var(--border-color);
  border-radius: 2px;
  margin: 12px auto;
}

.action-sheet-content {
  padding: 0 16px 34px;
}

.sheet-btn {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 16px;
  background: var(--hover-bg);
  border: none;
  border-radius: 12px;
  color: var(--text-color);
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  margin-bottom: 8px;
}

.sheet-btn svg {
  width: 20px;
  height: 20px;
}

.sheet-btn.danger {
  color: #fa2a2d;
}

.sheet-btn.update {
  color: #28a745;
}

/* 底部按钮组 */
.fab-container {
  position: fixed;
  right: 20px;
  bottom: 20px;
  display: flex;
  gap: 12px;
  z-index: 100;
}

.fab {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  padding: 0;
  background: #007dff;
  color: white;
  border: none;
  border-radius: 50%;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  box-shadow: 0 4px 12px rgba(0, 125, 255, 0.3);
  transition: all 0.2s;
}

.fab:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(0, 125, 255, 0.4);
}

.fab:active {
  transform: translateY(0);
}

.fab svg {
  width: 20px;
  height: 20px;
}

/* 进度条样式 */
.progress-bar-container {
  margin-top: 8px;
  height: 8px;
  background: var(--border-color);
  border-radius: 4px;
  overflow: hidden;
  position: relative;
}

.progress-bar {
  height: 100%;
  background: linear-gradient(90deg, #007dff, #00c6ff);
  border-radius: 4px;
  transition: width 0.3s ease;
}

.progress-text {
  position: absolute;
  right: 0;
  top: -18px;
  font-size: 11px;
  color: var(--text-secondary);
}

/* 正在拉取区域样式 */
.pulling-section {
  margin-bottom: 16px;
  background: var(--card-bg);
  border: 1px solid var(--border-color);
}

.section-header {
  padding: 12px 16px;
  font-weight: 600;
  color: var(--text-color);
  border-bottom: 1px solid var(--border-color);
  background: var(--bg-secondary);
}

.pulling-item {
  background: rgba(0, 125, 255, 0.05);
}

.pulling-item .item-icon {
  color: #007dff;
}

.pulling-item .item-icon svg {
  width: 24px;
  height: 24px;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

/* 检测升级对话框样式 */
.dialog-large {
  max-width: 480px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
}

.dialog-large .dialog-body {
  flex: 1;
  overflow-y: auto;
  max-height: calc(80vh - 120px);
}

.upgrade-progress {
  margin-bottom: 20px;
}

.progress-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.progress-title {
  font-size: 14px;
  color: var(--text-color);
}

.progress-percent {
  font-size: 14px;
  font-weight: 600;
  color: #007dff;
}

.upgrade-steps {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 20px;
}

.upgrade-step {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-radius: 10px;
  background: var(--bg-secondary);
  opacity: 0.6;
  transition: all 0.3s;
}

.upgrade-step.active {
  opacity: 1;
  background: rgba(0, 125, 255, 0.1);
}

.upgrade-step.completed {
  opacity: 1;
  background: rgba(40, 167, 69, 0.1);
}

.step-indicator {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: var(--border-color);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  flex-shrink: 0;
}

.upgrade-step.active .step-indicator {
  background: #007dff;
  color: white;
}

.upgrade-step.completed .step-indicator {
  background: #28a745;
  color: white;
}

.step-check {
  font-size: 14px;
}

.step-content {
  flex: 1;
  min-width: 0;
}

.step-title {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-color);
  margin-bottom: 2px;
}

.step-desc {
  font-size: 12px;
  color: var(--text-secondary);
}

.step-meta {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 2px;
}

.step-percent {
  font-size: 12px;
  font-weight: 600;
  color: #007dff;
}

.step-time {
  font-size: 11px;
  color: var(--text-tertiary);
}

.upgrade-versions {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--border-color);
}

.versions-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-color);
  margin-bottom: 12px;
}

.versions-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.version-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px;
  background: var(--bg-secondary);
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
}

.version-item:hover {
  background: var(--hover-bg);
}

.version-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-color);
}

.version-upgrade-btn {
  padding: 6px 16px;
  background: #007dff;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
}

.upgrade-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  color: var(--text-secondary);
  gap: 12px;
}

.upgrade-empty svg {
  width: 48px;
  height: 48px;
  stroke: #28a745;
}
</style>
