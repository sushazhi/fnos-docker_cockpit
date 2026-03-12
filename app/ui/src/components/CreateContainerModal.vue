<template>
  <div class="dialog-overlay" @click.self="$emit('close')">
    <div class="dialog dialog-large">
      <div class="dialog-header">
        <h3 class="dialog-title">{{ t('containers.create') }}</h3>
        <button class="dialog-close" @click="$emit('close')">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18" />
            <line x1="6" y1="6" x2="18" y2="18" />
          </svg>
        </button>
      </div>
      <div class="dialog-body">
        <div class="form-section">
          <div class="section-title">{{ t('containers.basicInfo') }}</div>
          <div class="form-field">
            <label class="form-label">{{ t('containers.name') }}</label>
            <input
              type="text"
              class="form-input"
              v-model="form.name"
              :placeholder="t('containers.namePlaceholder')"
            />
          </div>
          <div class="form-field">
            <label class="form-label">{{ t('containers.image') }} *</label>
            <div class="image-input-wrapper">
              <input
                type="text"
                class="form-input"
                v-model="form.image"
                placeholder="nginx:latest"
                @focus="showImageDropdown = true"
                @input="filterImages"
              />
              <div v-if="showImageDropdown && filteredImages.length > 0" class="image-dropdown">
                <div
                  v-for="image in filteredImages"
                  :key="image.RepoTags?.[0] || image.Id"
                  class="image-dropdown-item"
                  @click="selectImage(image)"
                >
                  <span class="image-name">{{ getImageName(image) }}</span>
                  <span class="image-size">{{ formatImageSize(image.Size) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="form-section">
          <div class="section-header">
            <div class="section-title">{{ t('containers.ports') }}</div>
            <button class="add-btn" @click="addPort">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="12" y1="5" x2="12" y2="19" />
                <line x1="5" y1="12" x2="19" y2="12" />
              </svg>
            </button>
          </div>
          <div class="dynamic-list">
            <div
              v-for="(port, index) in form.ports"
              :key="'port-' + index"
              class="dynamic-item triple-item"
            >
              <input
                type="text"
                class="form-input"
                v-model="port.host"
                :placeholder="t('containers.hostPort')"
              />
              <input
                type="text"
                class="form-input"
                v-model="port.container"
                :placeholder="t('containers.containerPort')"
              />
              <select class="form-input" v-model="port.protocol">
                <option value="tcp">TCP</option>
                <option value="udp">UDP</option>
              </select>
              <button class="remove-btn" @click="removePort(index)">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <line x1="18" y1="6" x2="6" y2="18" />
                  <line x1="6" y1="6" x2="18" y2="18" />
                </svg>
              </button>
            </div>
          </div>
        </div>

        <div class="form-section">
          <div class="section-header">
            <div class="section-title">{{ t('containers.volumes') }}</div>
            <button class="add-btn" @click="addVolume">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="12" y1="5" x2="12" y2="19" />
                <line x1="5" y1="12" x2="19" y2="12" />
              </svg>
            </button>
          </div>
          <div class="dynamic-list">
            <div
              v-for="(vol, index) in form.volumes"
              :key="'vol-' + index"
              class="dynamic-item triple-item"
            >
              <input
                type="text"
                class="form-input"
                v-model="vol.host"
                :placeholder="t('containers.hostPath')"
              />
              <input
                type="text"
                class="form-input"
                v-model="vol.container"
                :placeholder="t('containers.containerPath')"
              />
              <select class="form-input" v-model="vol.mode">
                <option value="rw">{{ t('containers.readWrite') }}</option>
                <option value="ro">{{ t('containers.readOnly') }}</option>
              </select>
              <button class="remove-btn" @click="removeVolume(index)">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <line x1="18" y1="6" x2="6" y2="18" />
                  <line x1="6" y1="6" x2="18" y2="18" />
                </svg>
              </button>
            </div>
          </div>
        </div>

        <div class="form-section">
          <div class="section-header">
            <div class="section-title">{{ t('containers.env') }}</div>
            <button class="add-btn" @click="addEnv">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="12" y1="5" x2="12" y2="19" />
                <line x1="5" y1="12" x2="19" y2="12" />
              </svg>
            </button>
          </div>
          <div class="dynamic-list">
            <div
              v-for="(env, index) in form.envs"
              :key="'env-' + index"
              class="dynamic-item env-item"
            >
              <input type="text" class="form-input env-key" v-model="env.key" placeholder="KEY" />
              <span class="env-equal">=</span>
              <input
                type="text"
                class="form-input env-value"
                v-model="env.value"
                placeholder="value"
              />
              <button class="remove-btn" @click="removeEnv(index)" v-if="form.envs.length > 1">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <line x1="18" y1="6" x2="6" y2="18" />
                  <line x1="6" y1="6" x2="18" y2="18" />
                </svg>
              </button>
            </div>
          </div>
        </div>

        <div class="form-section collapsible" :class="{ expanded: showAdvanced }">
          <div class="section-header clickable" @click="showAdvanced = !showAdvanced">
            <div class="section-title">{{ t('containers.advancedOptions') }}</div>
            <svg
              class="expand-icon"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
            >
              <polyline points="6 9 12 15 18 9" />
            </svg>
          </div>
          <div class="advanced-content" v-if="showAdvanced">
            <div class="form-field">
              <label class="form-label">{{ t('containers.cmd') }}</label>
              <input
                type="text"
                class="form-input"
                v-model="form.cmd"
                placeholder="nginx -g 'daemon off;'"
              />
              <div class="form-hint">{{ t('containers.cmdHint') }}</div>
            </div>
            <div class="form-field">
              <label class="form-label">{{ t('containers.entrypoint') }}</label>
              <input
                type="text"
                class="form-input"
                v-model="form.entrypoint"
                placeholder="/docker-entrypoint.sh"
              />
              <div class="form-hint">{{ t('containers.entrypointHint') }}</div>
            </div>
            <div class="form-field">
              <label class="form-label">{{ t('containers.workdir') }}</label>
              <input type="text" class="form-input" v-model="form.workdir" placeholder="/app" />
            </div>
            <div class="form-field">
              <label class="form-label">{{ t('containers.user') }}</label>
              <input
                type="text"
                class="form-input"
                v-model="form.user"
                placeholder="root 或 1000:1000"
              />
            </div>
            <div class="form-field">
              <label class="form-label">{{ t('containers.hostname') }}</label>
              <input
                type="text"
                class="form-input"
                v-model="form.hostname"
                placeholder="my-container"
              />
            </div>
            <div class="form-row">
              <div class="form-field half">
                <label class="form-label">{{ t('containers.network') }}</label>
                <select class="form-input" v-model="form.network">
                  <option value="bridge">bridge</option>
                  <option value="host">host</option>
                  <option value="none">none</option>
                </select>
              </div>
              <div class="form-field half">
                <label class="form-label">{{ t('containers.restart') }}</label>
                <select class="form-input" v-model="form.restart">
                  <option value="no">no</option>
                  <option value="always">always</option>
                  <option value="unless-stopped">unless-stopped</option>
                  <option value="on-failure">on-failure</option>
                </select>
              </div>
            </div>
            <div class="form-field">
              <div class="field-header">
                <label class="form-label">{{ t('containers.memoryLimit') }}</label>
                <button class="auto-set-btn" @click="autoSetMemory" type="button">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="23 4 23 10 17 10" />
                    <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10" />
                  </svg>
                  自动设置
                </button>
              </div>
              <div class="input-group">
                <input
                  type="number"
                  class="form-input"
                  v-model.number="form.memory"
                  placeholder="512"
                  min="0"
                />
                <select class="form-input unit-select" v-model="form.memoryUnit">
                  <option value="m">MB</option>
                  <option value="g">GB</option>
                </select>
              </div>
            </div>
            <div class="form-field">
              <div class="field-header">
                <label class="form-label">{{ t('containers.cpuLimit') }}</label>
                <button class="auto-set-btn" @click="autoSetCPU" type="button">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="23 4 23 10 17 10" />
                    <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10" />
                  </svg>
                  自动设置
                </button>
              </div>
              <input
                type="number"
                class="form-input"
                v-model.number="form.cpuShares"
                placeholder="1024"
                min="0"
                step="1"
              />
              <div class="form-hint">{{ t('containers.cpuLimitHint') }}</div>
            </div>
            <div class="form-field">
              <div class="checkbox-item" @click="form.privileged = !form.privileged">
                <div class="checkbox" :class="{ checked: form.privileged }">
                  <svg
                    v-if="form.privileged"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="3"
                  >
                    <polyline points="20 6 9 17 4 12" />
                  </svg>
                </div>
                <span class="checkbox-label">{{ t('containers.privileged') }}</span>
              </div>
              <div class="form-hint">{{ t('containers.privilegedHint') }}</div>
            </div>
            <div class="form-field">
              <div class="checkbox-item" @click="form.autoRemove = !form.autoRemove">
                <div class="checkbox" :class="{ checked: form.autoRemove }">
                  <svg
                    v-if="form.autoRemove"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="3"
                  >
                    <polyline points="20 6 9 17 4 12" />
                  </svg>
                </div>
                <span class="checkbox-label">{{ t('containers.autoRemove') }}</span>
              </div>
              <div class="form-hint">{{ t('containers.autoRemoveHint') }}</div>
            </div>
            <div class="form-field">
              <div class="checkbox-item" @click="form.interactive = !form.interactive">
                <div class="checkbox" :class="{ checked: form.interactive }">
                  <svg
                    v-if="form.interactive"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="3"
                  >
                    <polyline points="20 6 9 17 4 12" />
                  </svg>
                </div>
                <span class="checkbox-label">{{ t('containers.interactive') }}</span>
              </div>
            </div>
            <div class="form-field">
              <div class="checkbox-item" @click="form.tty = !form.tty">
                <div class="checkbox" :class="{ checked: form.tty }">
                  <svg
                    v-if="form.tty"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="3"
                  >
                    <polyline points="20 6 9 17 4 12" />
                  </svg>
                </div>
                <span class="checkbox-label">{{ t('containers.tty') }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div class="dialog-footer">
        <button class="dialog-btn secondary" @click="$emit('close')">
          {{ t('common.cancel') }}
        </button>
        <button class="dialog-btn primary" @click="create" :disabled="creating || !form.image">
          {{ creating ? t('containers.creating') : t('containers.create') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, inject } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import api from '../services/api'

const { t } = useI18n()
const route = useRoute()
const showToast = inject('showToast')

const emit = defineEmits(['close', 'created'])

const showAdvanced = ref(false)

const form = ref({
  name: '',
  image: '',
  ports: [{ host: '', container: '', protocol: 'tcp' }],
  volumes: [{ host: '', container: '', mode: 'rw' }],
  envs: [{ key: '', value: '' }],
  network: 'bridge',
  restart: 'no',
  cmd: '',
  entrypoint: '',
  workdir: '',
  user: '',
  hostname: '',
  memory: null,
  memoryUnit: 'm',
  cpuShares: null,
  privileged: false,
  autoRemove: false,
  interactive: false,
  tty: false
})

const creating = ref(false)
const showImageDropdown = ref(false)
const images = ref([])
const filteredImages = ref([])

// 加载本地镜像列表
async function loadImages() {
  try {
    const data = await api.get('/api/images')
    images.value = data.images || []
    filteredImages.value = images.value
  } catch (e) {
    console.error('Failed to load images:', e)
  }
}

// 过滤镜像
function filterImages() {
  const query = form.value.image.toLowerCase()
  if (!query) {
    filteredImages.value = images.value
  } else {
    filteredImages.value = images.value.filter(image => {
      const name = getImageName(image).toLowerCase()
      return name.includes(query)
    })
  }
}

// 获取镜像名称
function getImageName(image) {
  if (image.RepoTags && image.RepoTags.length > 0) {
    return image.RepoTags[0]
  }
  if (image.RepoDigests && image.RepoDigests.length > 0) {
    return image.RepoDigests[0].split('@')[0]
  }
  return image.Id?.substring(7, 19) || 'untagged'
}

// 格式化镜像大小
function formatImageSize(bytes) {
  if (!bytes) return '-'
  const gb = bytes / 1024 / 1024 / 1024
  if (gb >= 1) {
    return gb.toFixed(2) + ' GB'
  }
  const mb = bytes / 1024 / 1024
  if (mb >= 1) {
    return mb.toFixed(1) + ' MB'
  }
  return (bytes / 1024).toFixed(0) + ' KB'
}

// 选择镜像
function selectImage(image) {
  form.value.image = getImageName(image)
  showImageDropdown.value = false
}

// 点击外部关闭下拉框
function handleClickOutside(e) {
  const wrapper = e.target.closest('.image-input-wrapper')
  if (!wrapper) {
    showImageDropdown.value = false
  }
}

function addPort() {
  form.value.ports.push({ host: '', container: '', protocol: 'tcp' })
}

function removePort(index) {
  if (form.value.ports.length > 1) {
    form.value.ports.splice(index, 1)
  }
}

function addVolume() {
  form.value.volumes.push({ host: '', container: '', mode: 'rw' })
}

function removeVolume(index) {
  if (form.value.volumes.length > 1) {
    form.value.volumes.splice(index, 1)
  }
}

function addEnv() {
  form.value.envs.push({ key: '', value: '' })
}

function removeEnv(index) {
  form.value.envs.splice(index, 1)
}

onMounted(() => {
  if (route.query.image) {
    form.value.image = route.query.image
  }
  loadImages()
  document.addEventListener('click', handleClickOutside)
})

// 自动设置内存（获取系统总内存的 1/4）
async function autoSetMemory() {
  try {
    const data = await api.get('/api/system/info')
    const totalMemory = data.info?.MemTotal || 0
    if (totalMemory > 0) {
      // 设置为总内存的 1/4，最小 512MB，最大 8GB
      const suggestedMemory = Math.min(
        Math.max(Math.floor(totalMemory / 1024 / 1024 / 4), 512),
        8192
      )
      if (suggestedMemory >= 1024) {
        form.value.memory = Math.floor(suggestedMemory / 1024)
        form.value.memoryUnit = 'g'
      } else {
        form.value.memory = suggestedMemory
        form.value.memoryUnit = 'm'
      }
      showToast(`已自动设置内存为 ${form.value.memory}${form.value.memoryUnit.toUpperCase()}`)
    } else {
      // 默认值
      form.value.memory = 512
      form.value.memoryUnit = 'm'
      showToast('无法获取系统内存，已设置为默认 512MB')
    }
  } catch (e) {
    // 默认值
    form.value.memory = 512
    form.value.memoryUnit = 'm'
    showToast('无法获取系统内存，已设置为默认 512MB')
  }
}

// 自动设置 CPU（获取系统 CPU 核心数 * 1024）
async function autoSetCPU() {
  try {
    const data = await api.get('/api/system/info')
    const cpuCount = data.info?.NCPU || 0
    if (cpuCount > 0) {
      // 设置为 CPU 核心数对应的权重
      form.value.cpuShares = cpuCount * 1024
      showToast(`已自动设置 CPU 权重为 ${form.value.cpuShares}（${cpuCount} 核心）`)
    } else {
      // 默认值
      form.value.cpuShares = 1024
      showToast('无法获取 CPU 信息，已设置为默认 1024')
    }
  } catch (e) {
    // 默认值
    form.value.cpuShares = 1024
    showToast('无法获取 CPU 信息，已设置为默认 1024')
  }
}

async function create() {
  if (!form.value.image) return
  creating.value = true
  try {
    const body = {
      name: form.value.name || undefined,
      image: form.value.image,
      network: form.value.network,
      restart: form.value.restart
    }

    // 处理端口映射
    const ports = form.value.ports
      .filter(p => p.host.trim() && p.container.trim())
      .map(p => {
        const hostParts = p.host.split(':')
        const portData = { containerPort: `${p.container}/${p.protocol || 'tcp'}` }
        if (hostParts.length === 2) {
          portData.hostIp = hostParts[0]
          portData.hostPort = hostParts[1]
        } else {
          portData.hostPort = p.host
        }
        return portData
      })
    if (ports.length > 0) {
      body.ports = ports
    }

    // 处理存储卷
    const volumes = form.value.volumes
      .filter(v => v.host.trim() && v.container.trim())
      .map(v => {
        const mode = v.mode === 'ro' ? 'ro' : 'rw'
        return `${v.host}:${v.container}:${mode}`
      })
    if (volumes.length > 0) {
      body.volumes = volumes
    }

    const envs = form.value.envs.filter(e => e.key.trim()).map(e => `${e.key.trim()}=${e.value}`)
    if (envs.length > 0) {
      body.env = envs
    }

    if (form.value.cmd?.trim()) {
      body.cmd = form.value.cmd.trim().split(/\s+/)
    }
    if (form.value.entrypoint?.trim()) {
      body.entrypoint = form.value.entrypoint.trim()
    }
    if (form.value.workdir?.trim()) {
      body.workdir = form.value.workdir.trim()
    }
    if (form.value.user?.trim()) {
      body.user = form.value.user.trim()
    }
    if (form.value.hostname?.trim()) {
      body.hostname = form.value.hostname.trim()
    }
    if (form.value.memory && form.value.memory > 0) {
      body.memory = form.value.memory * (form.value.memoryUnit === 'g' ? 1024 : 1)
    }
    if (form.value.cpuShares && form.value.cpuShares > 0) {
      body.cpuShares = form.value.cpuShares
    }
    if (form.value.privileged) {
      body.privileged = true
    }
    if (form.value.autoRemove) {
      body.autoRemove = true
    }
    if (form.value.interactive) {
      body.interactive = true
    }
    if (form.value.tty) {
      body.tty = true
    }

    await api.post('/api/container/create', body)
    showToast(t('containers.createSuccess'))
    emit('created')
  } catch (e) {
    showToast(t('containers.createFailed') + ': ' + e.message)
  } finally {
    creating.value = false
  }
}
</script>

<style scoped>
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
  max-width: 420px;
  max-height: calc(100vh - 32px);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

@media (min-width: 768px) {
  .dialog {
    max-width: 520px;
  }
}

@media (max-height: 600px) {
  .dialog {
    max-height: 95vh;
  }
}

.dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px;
  border-bottom: 1px solid var(--border-color);
  flex-shrink: 0;
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
  transition: background var(--transition-fast);
}

.dialog-close:hover {
  background: var(--hover-bg);
}

.dialog-close svg {
  width: 20px;
  height: 20px;
}

.dialog-body {
  padding: 16px 20px;
  overflow-y: auto;
  max-height: calc(100vh - 200px);
}

@media (max-height: 600px) {
  .dialog-body {
    max-height: calc(95vh - 150px);
  }
}

.form-section {
  margin-bottom: 20px;
}

.form-section:last-child {
  margin-bottom: 0;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-color);
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}

.section-header.clickable {
  cursor: pointer;
  padding: 8px 12px;
  margin: -8px -12px;
  border-radius: 10px;
  transition: background var(--transition-fast);
}

.section-header.clickable:hover {
  background: var(--hover-bg);
}

.expand-icon {
  width: 18px;
  height: 18px;
  color: var(--text-secondary);
  transition: transform var(--transition-fast);
}

.form-section.collapsible .expand-icon {
  transform: rotate(-90deg);
}

.form-section.collapsible.expanded .expand-icon {
  transform: rotate(0deg);
}

.advanced-content {
  padding-top: 12px;
}

.form-field {
  margin-bottom: 12px;
}

.form-field.half {
  flex: 1;
  margin-bottom: 0;
}

.form-label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
  margin-bottom: 6px;
}

.form-hint {
  font-size: 11px;
  color: var(--text-tertiary);
  margin-top: 4px;
}

.form-input {
  width: 100%;
  padding: 10px 12px;
  background: var(--input-bg);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  color: var(--text-color);
  font-size: 14px;
  font-family: inherit;
  transition:
    border-color var(--transition-fast),
    box-shadow var(--transition-fast);
}

.form-input:focus {
  outline: none;
  border-color: #007dff;
  box-shadow: 0 0 0 3px rgba(0, 125, 255, 0.12);
}

.form-input::placeholder {
  color: var(--text-tertiary);
}

.form-row {
  display: flex;
  gap: 12px;
}

.input-group {
  display: flex;
  gap: 8px;
}

.input-group .form-input {
  flex: 1;
}

.unit-select {
  width: 80px;
  flex-shrink: 0;
}

.add-btn {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 125, 255, 0.1);
  border: none;
  border-radius: 8px;
  color: #007dff;
  cursor: pointer;
  transition: background var(--transition-fast);
}

.add-btn:hover {
  background: rgba(0, 125, 255, 0.2);
}

.add-btn svg {
  width: 16px;
  height: 16px;
}

.dynamic-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.dynamic-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.dynamic-item .form-input {
  flex: 1;
}

/* 三字段布局（端口映射、存储卷） */
.dynamic-item.triple-item {
  display: flex;
  gap: 8px;
  align-items: center;
}

.dynamic-item.triple-item .form-input {
  flex: 1;
  min-width: 0;
}

.dynamic-item.triple-item select.form-input {
  flex: 0 0 80px;
  cursor: pointer;
}

.remove-btn {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(250, 42, 45, 0.1);
  border: none;
  border-radius: 8px;
  color: #fa2a2d;
  cursor: pointer;
  flex-shrink: 0;
  transition: background var(--transition-fast);
}

.remove-btn:hover {
  background: rgba(250, 42, 45, 0.2);
}

.remove-btn svg {
  width: 16px;
  height: 16px;
}

.env-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.env-item .env-key {
  flex: 1;
  min-width: 0;
}

.env-item .env-value {
  flex: 2;
  min-width: 0;
}

.env-equal {
  color: var(--text-tertiary);
  font-weight: 500;
  flex-shrink: 0;
}

.checkbox-item {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  padding: 8px 0;
}

.checkbox {
  width: 22px;
  height: 22px;
  border: 2px solid var(--border-color);
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--transition-fast);
  flex-shrink: 0;
}

.checkbox.checked {
  background: #007dff;
  border-color: #007dff;
}

.checkbox svg {
  width: 14px;
  height: 14px;
  color: white;
}

.checkbox-label {
  font-size: 14px;
  color: var(--text-color);
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid var(--border-color);
  flex-shrink: 0;
}

.dialog-btn {
  padding: 10px 20px;
  border: none;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.dialog-btn.primary {
  background: #007dff;
  color: white;
}

.dialog-btn.primary:hover {
  background: #0066cc;
}

.dialog-btn.primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.dialog-btn.secondary {
  background: var(--hover-bg);
  color: var(--text-color);
}

.dialog-btn.secondary:hover {
  background: var(--active-bg);
}

.field-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 6px;
}

.auto-set-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  background: rgba(0, 125, 255, 0.1);
  border: none;
  border-radius: 6px;
  color: #007dff;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: background var(--transition-fast);
}

.auto-set-btn:hover {
  background: rgba(0, 125, 255, 0.2);
}

.auto-set-btn svg {
  width: 14px;
  height: 14px;
}

.image-input-wrapper {
  position: relative;
}

.image-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  margin-top: 4px;
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  max-height: 200px;
  overflow-y: auto;
  z-index: 100;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.image-dropdown-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  cursor: pointer;
  transition: background var(--transition-fast);
}

.image-dropdown-item:hover {
  background: var(--hover-bg);
}

.image-dropdown-item:first-child {
  border-radius: 10px 10px 0 0;
}

.image-dropdown-item:last-child {
  border-radius: 0 0 10px 10px;
}

.image-name {
  font-size: 14px;
  color: var(--text-color);
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.image-size {
  font-size: 12px;
  color: var(--text-secondary);
  margin-left: 8px;
  flex-shrink: 0;
}
</style>
