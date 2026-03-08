<template>
  <div class="page">
    <div class="header">
      <button class="header-back" @click="$router.back()">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
          <polyline points="15 18 9 12 15 6"/>
        </svg>
      </button>
      <span class="header-title">{{ getContainerName(container) || t('containers.container') }}</span>

    </div>
    
    <div v-if="loading" class="loading">
      <div class="spinner"></div>
    </div>
    
    <template v-else-if="container">
      <div class="section-header">
        <div class="section-icon status">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <polyline points="12 6 12 12 16 14"/>
          </svg>
        </div>
        <div class="section-info">
          <div class="section-title">{{ t('containers.status') }}</div>
        </div>
      </div>
      <div class="info-card">
        <div class="status-row">
          <span class="badge" :class="getStatusClass(container.State)">{{ getStatusText(container.State) }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">{{ t('containers.id') }}</span>
          <span class="info-value monospace">{{ container.Id?.substring(0, 12) }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">{{ t('containers.image') }}</span>
          <span class="info-value">{{ container.Image }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">{{ t('containers.status') }}</span>
          <span class="info-value">{{ container.Status }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">{{ t('containers.created') }}</span>
          <span class="info-value">{{ formatTime(container.Created) }}</span>
        </div>
      </div>

      <!-- 操作按钮 -->
      <div class="quick-actions">
        <button v-if="container.State !== 'running'" class="quick-action-btn start" @click="startContainer">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polygon points="5 3 19 12 5 21 5 3"/>
          </svg>
          <span>{{ t('containers.actions.start') }}</span>
        </button>
        <button v-if="container.State === 'running'" class="quick-action-btn stop" @click="stopContainer">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="6" y="4" width="4" height="16"/>
            <rect x="14" y="4" width="4" height="16"/>
          </svg>
          <span>{{ t('containers.actions.stop') }}</span>
        </button>
        <button v-if="container.State === 'running'" class="quick-action-btn restart" @click="restartContainer">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="23 4 23 10 17 10"/>
            <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
          </svg>
          <span>{{ t('containers.actions.restart') }}</span>
        </button>
        <button v-if="container.State === 'running'" class="quick-action-btn pause" @click="pauseContainer">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="6" y="4" width="4" height="16"/>
            <rect x="14" y="4" width="4" height="16"/>
          </svg>
          <span>{{ t('containers.actions.pause') }}</span>
        </button>
        <button v-if="container.State === 'paused'" class="quick-action-btn resume" @click="unpauseContainer">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polygon points="5 3 19 12 5 21 5 3"/>
          </svg>
          <span>{{ t('containers.actions.resume') }}</span>
        </button>
        <button class="quick-action-btn remove" @click="confirmRemove">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="3 6 5 6 21 6"/>
            <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
          </svg>
          <span>{{ t('containers.actions.remove') }}</span>
        </button>
        <button class="quick-action-btn edit" @click="openEditModal">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
            <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
          </svg>
          <span>{{ t('containers.actions.edit') }}</span>
        </button>
        <button class="quick-action-btn more" @click="showMoreActions = true">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="1"/>
            <circle cx="19" cy="12" r="1"/>
            <circle cx="5" cy="12" r="1"/>
          </svg>
          <span>{{ t('common.more') }}</span>
        </button>
      </div>
      
      <div class="tabs-container">
        <div class="tabs">
          <button class="tab" :class="{ active: tab === 'overview' }" @click="tab = 'overview'">{{ t('containers.overview') }}</button>
          <button class="tab" :class="{ active: tab === 'logs' }" @click="tab = 'logs'">{{ t('containers.logs') }}</button>
          <button class="tab" :class="{ active: tab === 'terminal' }" @click="tab = 'terminal'" v-if="container?.State === 'running'">{{ t('containers.terminal') }}</button>
        </div>
      </div>
      
      <div class="tab-content">
        <!-- 概览标签 -->
        <div v-if="tab === 'overview'" class="info-section">
          <!-- 端口映射 -->
          <div v-if="containerInfo?.NetworkSettings?.Ports && Object.keys(containerInfo.NetworkSettings.Ports).length > 0" class="info-block">
            <div class="info-block-title">{{ t('containers.ports') }}</div>
            <div class="info-block-content">
              <div v-for="(bindings, port) in containerInfo.NetworkSettings.Ports" :key="port" class="info-line">
                <span class="info-line-label">{{ port }}</span>
                <span class="info-line-value">
                  <template v-if="bindings && bindings.length > 0">
                    <span v-for="(binding, idx) in bindings" :key="idx">
                      {{ binding.HostIp === '0.0.0.0' ? '' : binding.HostIp + ':' }}{{ binding.HostPort }}
                    </span>
                  </template>
                  <template v-else>-</template>
                </span>
              </div>
            </div>
          </div>

          <!-- 环境变量 -->
          <div v-if="containerInfo?.Config?.Env && containerInfo.Config.Env.length > 0" class="info-block">
            <div class="info-block-title">{{ t('containers.env') }}</div>
            <div class="info-block-content">
              <div v-for="env in containerInfo.Config.Env" :key="env" class="info-line">
                <span class="info-line-value env-value">{{ env }}</span>
              </div>
            </div>
          </div>

          <!-- 卷挂载 -->
          <div v-if="containerInfo?.Mounts && containerInfo.Mounts.length > 0" class="info-block">
            <div class="info-block-title">{{ t('containers.volumes') }}</div>
            <div class="info-block-content">
              <div v-for="(mount, idx) in containerInfo.Mounts" :key="idx" class="info-line">
                <span class="info-line-value">
                  <span class="mount-type">{{ mount.Type }}</span>
                  {{ mount.Source }} → {{ mount.Destination }}
                </span>
              </div>
            </div>
          </div>

          <!-- 网络 -->
          <div v-if="containerInfo?.NetworkSettings?.Networks" class="info-block">
            <div class="info-block-title">{{ t('containers.networks') }}</div>
            <div class="info-block-content">
              <div v-for="(config, name) in containerInfo.NetworkSettings.Networks" :key="name" class="info-line">
                <span class="info-line-label">{{ name }}</span>
                <span class="info-line-value">{{ config.IPAddress || '-' }}</span>
              </div>
            </div>
          </div>

          <!-- 重启策略 -->
          <div v-if="containerInfo?.HostConfig?.RestartPolicy?.Name" class="info-block">
            <div class="info-block-title">{{ t('containers.restartPolicy') }}</div>
            <div class="info-block-content">
              <div class="info-line">
                <span class="info-line-value">{{ containerInfo.HostConfig.RestartPolicy.Name }}</span>
              </div>
            </div>
          </div>

          <!-- 网络模式 -->
          <div v-if="containerInfo?.HostConfig?.NetworkMode" class="info-block">
            <div class="info-block-title">{{ t('containers.networkMode') }}</div>
            <div class="info-block-content">
              <div class="info-line">
                <span class="info-line-value">{{ containerInfo.HostConfig.NetworkMode }}</span>
              </div>
            </div>
          </div>

          <!-- 原始 JSON -->
          <div class="info-block">
            <div class="info-block-title">{{ t('containers.rawJson') }}</div>
            <pre class="log-viewer">{{ JSON.stringify(containerInfo, null, 2) }}</pre>
          </div>
        </div>

        <div v-else-if="tab === 'logs'" class="logs-section">
          <div class="section-toolbar">
            <button class="toolbar-btn" @click="loadLogs">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="23 4 23 10 17 10"/>
                <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
              </svg>
              {{ t('containers.refresh') }}
            </button>
          </div>
          <div v-if="logsLoading" class="loading"><div class="spinner"></div></div>
          <div v-else class="log-viewer" v-html="formattedLogs || t('containers.noLogs')"></div>
        </div>

        <div v-else-if="tab === 'terminal'" class="terminal-tab">
          <Terminal 
            ref="terminalRef"
            :container-id="containerId"
            :connect-text="t('containers.newTerminal')"
            @error="showToast($event)"
            @connected="showToast(t('containers.terminalConnected'))"
            @disconnected="showToast(t('containers.terminalDisconnected'))"
          />
        </div>
      </div>
    </template>
    
    <div v-else class="empty-state">
      <div class="empty-icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"/>
          <line x1="15" y1="9" x2="9" y2="15"/>
          <line x1="9" y1="9" x2="15" y2="15"/>
        </svg>
      </div>
      <div class="empty-text">{{ t('containers.notFound') }}</div>
    </div>
    
    <!-- 更多操作菜单 -->
    <div v-if="showMoreActions" class="action-sheet-overlay" @click.self="showMoreActions = false">
      <div class="action-sheet">
        <div class="action-sheet-handle"></div>
        <div class="action-sheet-content">
          <button class="sheet-btn" @click="generateDockerRun">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="4 17 10 11 4 5"/>
              <line x1="12" y1="19" x2="20" y2="19"/>
            </svg>
            {{ t('containers.generateDockerRun') }}
          </button>
          <button class="sheet-btn" @click="generateCompose">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
              <polyline points="14 2 14 8 20 8"/>
              <line x1="16" y1="13" x2="8" y2="13"/>
              <line x1="16" y1="17" x2="8" y2="17"/>
              <polyline points="10 9 9 9 8 9"/>
            </svg>
            {{ t('containers.generateCompose') }}
          </button>
        </div>
      </div>
    </div>

    <!-- 编辑容器模态框 -->
    <div v-if="showEditModal" class="modal-overlay" @click.self="showEditModal = false">
      <div class="modal-container modal-large">
        <div class="modal-header">
          <h3 class="modal-title">{{ t('containers.edit') }}</h3>
          <button class="modal-close" @click="showEditModal = false">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"/>
              <line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>
        <div class="modal-body edit-modal-body">
          <!-- 基本信息 -->
          <div class="form-section">
            <div class="section-title">{{ t('containers.basicInfo') }}</div>
            <div class="form-field">
              <label class="form-label">{{ t('containers.name') }}</label>
              <input type="text" class="form-input" v-model="editForm.name" :placeholder="t('containers.namePlaceholder')" />
            </div>
            <div class="form-field">
              <label class="form-label">{{ t('containers.hostname') }}</label>
              <input type="text" class="form-input" v-model="editForm.hostname" placeholder="my-container" />
            </div>
          </div>

          <!-- 端口映射 -->
          <div class="form-section">
            <div class="section-header">
              <div class="section-title">{{ t('containers.ports') }}</div>
              <button class="add-btn" @click="addEditPort">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <line x1="12" y1="5" x2="12" y2="19"/>
                  <line x1="5" y1="12" x2="19" y2="12"/>
                </svg>
              </button>
            </div>
            <div class="dynamic-list">
              <div v-for="(port, index) in editForm.ports" :key="'port-' + index" class="dynamic-item triple-item">
                <input type="text" class="form-input" v-model="port.host" placeholder="主机端口" />
                <input type="text" class="form-input" v-model="port.container" placeholder="容器端口" />
                <select class="form-input" v-model="port.protocol">
                  <option value="tcp">TCP</option>
                  <option value="udp">UDP</option>
                </select>
                <button class="remove-btn" @click="removeEditPort(index)">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <line x1="18" y1="6" x2="6" y2="18"/>
                    <line x1="6" y1="6" x2="18" y2="18"/>
                  </svg>
                </button>
              </div>
            </div>
          </div>

          <!-- 存储卷 -->
          <div class="form-section">
            <div class="section-header">
              <div class="section-title">{{ t('containers.volumes') }}</div>
              <button class="add-btn" @click="addEditVolume">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <line x1="12" y1="5" x2="12" y2="19"/>
                  <line x1="5" y1="12" x2="19" y2="12"/>
                </svg>
              </button>
            </div>
            <div class="dynamic-list">
              <div v-for="(volume, index) in editForm.volumes" :key="'volume-' + index" class="dynamic-item triple-item">
                <input type="text" class="form-input" v-model="volume.host" placeholder="主机路径" />
                <input type="text" class="form-input" v-model="volume.container" placeholder="容器路径" />
                <select class="form-input" v-model="volume.mode">
                  <option value="rw">读写</option>
                  <option value="ro">只读</option>
                </select>
                <button class="remove-btn" @click="removeEditVolume(index)">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <line x1="18" y1="6" x2="6" y2="18"/>
                    <line x1="6" y1="6" x2="18" y2="18"/>
                  </svg>
                </button>
              </div>
            </div>
          </div>

          <!-- 环境变量 -->
          <div class="form-section">
            <div class="section-header">
              <div class="section-title">{{ t('containers.env') }}</div>
              <button class="add-btn" @click="addEditEnv">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <line x1="12" y1="5" x2="12" y2="19"/>
                  <line x1="5" y1="12" x2="19" y2="12"/>
                </svg>
              </button>
            </div>
            <div class="dynamic-list">
              <div v-for="(env, index) in editForm.envs" :key="'env-' + index" class="dynamic-item env-item">
                <input type="text" class="form-input env-key" v-model="env.key" placeholder="KEY" />
                <span class="env-equal">=</span>
                <input type="text" class="form-input env-value" v-model="env.value" placeholder="value" />
                <button class="remove-btn" @click="removeEditEnv(index)">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <line x1="18" y1="6" x2="6" y2="18"/>
                    <line x1="6" y1="6" x2="18" y2="18"/>
                  </svg>
                </button>
              </div>
            </div>
          </div>

          <!-- 高级选项 -->
          <div class="form-section collapsible" :class="{ expanded: editShowAdvanced }">
            <div class="section-header clickable" @click="editShowAdvanced = !editShowAdvanced">
              <div class="section-title">{{ t('containers.advancedOptions') }}</div>
              <svg class="expand-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="6 9 12 15 18 9"/>
              </svg>
            </div>
            <div class="advanced-content" v-if="editShowAdvanced">
              <div class="form-field">
                <label class="form-label">{{ t('containers.restartPolicy') }}</label>
                <select class="form-input" v-model="editForm.restartPolicy">
                  <option value="no">no</option>
                  <option value="always">always</option>
                  <option value="unless-stopped">unless-stopped</option>
                  <option value="on-failure">on-failure</option>
                </select>
              </div>
              <div class="form-row">
                <div class="form-field half">
                  <label class="form-label">{{ t('containers.memoryLimit') }} (MB)</label>
                  <input type="number" class="form-input" v-model.number="editForm.memory" placeholder="0" min="0" />
                </div>
                <div class="form-field half">
                  <label class="form-label">{{ t('containers.cpuLimit') }}</label>
                  <input type="number" class="form-input" v-model.number="editForm.cpuShares" placeholder="1024" min="0" />
                </div>
              </div>
              <div class="form-field checkbox-field">
                <label class="checkbox-item" @click="editForm.privileged = !editForm.privileged">
                  <div class="checkbox" :class="{ checked: editForm.privileged }">
                    <svg v-if="editForm.privileged" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
                      <polyline points="20 6 9 17 4 12"/>
                    </svg>
                  </div>
                  <span class="checkbox-label">{{ t('containers.privileged') }}</span>
                </label>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showEditModal = false">{{ t('common.cancel') }}</button>
          <button class="btn btn-primary" @click="saveEdit" :disabled="editLoading">
            {{ editLoading ? t('common.saving') : t('common.confirm') }}
          </button>
        </div>
      </div>
    </div>
    
    <!-- 生成 Docker Run 命令模态框 -->
    <div v-if="showDockerRunModal" class="modal-overlay" @click.self="showDockerRunModal = false">
      <div class="modal-container modal-large">
        <div class="modal-header">
          <h3 class="modal-title">{{ t('containers.dockerRunCommand') }}</h3>
          <button class="modal-close" @click="showDockerRunModal = false">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"/>
              <line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>
        <div class="modal-body">
          <pre class="compose-viewer">{{ dockerRunCommand }}</pre>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showDockerRunModal = false">{{ t('common.close') }}</button>
          <button class="btn btn-primary" @click="copyDockerRun">
            {{ dockerRunCopied ? t('common.copied') : t('common.copy') }}
          </button>
        </div>
      </div>
    </div>

    <!-- 生成 Compose 文件模态框 -->
    <!-- Docker Run 命令模态框 -->
    <div v-if="showDockerRunModal" class="modal-overlay" @click.self="showDockerRunModal = false">
      <div class="modal-container modal-large">
        <div class="modal-header">
          <h3 class="modal-title">{{ t('containers.dockerRunCommand') }}</h3>
          <button class="modal-close" @click="showDockerRunModal = false">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"/>
              <line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>
        <div class="modal-body">
          <pre class="compose-viewer">{{ dockerRunCommand }}</pre>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showDockerRunModal = false">{{ t('common.close') }}</button>
          <button class="btn btn-primary" @click="copyDockerRun">
            {{ dockerRunCopied ? t('common.copied') : t('common.copy') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Compose 文件模态框 -->
    <div v-if="showComposeModal" class="modal-overlay" @click.self="showComposeModal = false">
      <div class="modal-container modal-large">
        <div class="modal-header">
          <h3 class="modal-title">{{ t('containers.composeFile') }}</h3>
          <button class="modal-close" @click="showComposeModal = false">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"/>
              <line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>
        <div class="modal-body">
          <pre class="compose-viewer">{{ composeContent }}</pre>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showComposeModal = false">{{ t('common.close') }}</button>
          <button class="btn btn-primary" @click="copyCompose">
            {{ copied ? t('common.copied') : t('common.copy') }}
          </button>
        </div>
      </div>
    </div>

    <ConfirmModal 
      v-if="showConfirm"
      :title="t('containers.deleteContainer')"
      :message="t('containers.confirmDelete') + ' ' + getContainerName(container) + '?'"
      :confirm-text="t('containers.delete')"
      danger
      @close="showConfirm = false"
      @confirm="removeContainer"
    />
  </div>
</template>

<script setup>
import { ref, onMounted, watch, computed, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { inject } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '../services/api'
import ConfirmModal from '../components/ConfirmModal.vue'
import Terminal from '../components/Terminal.vue'
import AnsiToHtml from 'ansi-to-html'
import DOMPurify from 'dompurify'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const showToast = inject('showToast')

const loading = ref(true)
const container = ref(null)
const containerInfo = ref(null)
const logs = ref('')
const logsLoading = ref(false)
const tab = ref('overview')
const showMoreActions = ref(false)
const showConfirm = ref(false)
const terminalRef = ref(null)
const showEditModal = ref(false)
const editLoading = ref(false)
const editShowAdvanced = ref(false)
const editForm = ref({
  name: '',
  hostname: '',
  ports: [],
  volumes: [],
  envs: [],
  restartPolicy: 'no',
  memory: null,
  cpuShares: null,
  privileged: false
})
const showComposeModal = ref(false)
const composeContent = ref('')
const copied = ref(false)
const showDockerRunModal = ref(false)
const dockerRunCommand = ref('')
const dockerRunCopied = ref(false)
let prevNetworkRx = ref(0)
let prevNetworkTx = ref(0)
let prevNetworkTime = ref(0)
let logsInterval = null

const containerId = route.params.id

const ansiConverter = new AnsiToHtml({
  fg: '#d4d4d4',
  bg: '#1e1e1e',
  newline: true,
  escapeXML: true,
  stream: false
})

// 清理日志中的特殊控制字符
function cleanLogOutput(logText) {
  if (!logText) return ''

  // 移除或替换常见的控制字符
  return logText
    // 移除SOH (Start of Heading) \u0001
    .replace(/\x01/g, '')
    // 移除STX (Start of Text) \u0002
    .replace(/\x02/g, '')
    // 移除其他常见的控制字符（保留换行\x0A、回车\x0D、制表符\x09等）
    .replace(/[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]/g, '')
    // 修复缺少ESC字符的ANSI颜色代码
    // 将 [数字m 格式转换为标准的 \x1B[数字m 格式
    .replace(/\[(\d+(?:;\d+)*)m/g, '\x1B[$1m')
    // 移除ANSI光标控制序列(保留颜色代码)
    .replace(/\x1B\[[0-9;]*[A-HJKSTfmin]/g, '')
    // 移除其他ANSI转义序列
    .replace(/\x1B\][^\x07]*\x07/g, '')
    .replace(/\x1B[PX^_].*?\x1B\\/g, '')
}

const formattedLogs = computed(() => {
  if (!logs.value) return ''
  try {
    // 先清理控制字符
    const cleanedLogs = cleanLogOutput(logs.value)
    // 转换ANSI颜色代码（ansi-to-html 已经将 \n 转换为 <br>）
    const html = ansiConverter.toHtml(cleanedLogs)
    // 使用DOMPurify进行二次过滤，防止XSS攻击
    return DOMPurify.sanitize(html, {
      ALLOWED_TAGS: ['span', 'b', 'i', 'u', 's', 'strong', 'em', 'br'],
      ALLOWED_ATTR: ['style', 'class']
    })
  } catch (e) {
    console.error('Failed to convert ANSI:', e)
    // 出错时，将换行符转换为 <br> 标签
    return logs.value.replace(/\n/g, '<br>')
  }
})

function getContainerName(c) {
  if (!c) return ''
  if (c.Names && c.Names.length > 0) {
    return c.Names[0].replace(/^\//, '')
  }
  return c.Id?.substring(0, 12) || 'unknown'
}

function formatTime(timestamp) {
  if (!timestamp) return '-'
  const date = new Date(timestamp * 1000)
  return date.toLocaleString()
}

function getStatusClass(state) {
  switch (state) {
    case 'running': return 'badge-success'
    case 'paused': return 'badge-warning'
    case 'exited': return 'badge-danger'
    default: return 'badge-info'
  }
}

function getStatusText(state) {
  switch (state) {
    case 'running': return t('containers.state.running')
    case 'paused': return t('containers.state.paused')
    case 'exited': return t('containers.state.exited')
    default: return state
  }
}

async function loadContainer() {
  loading.value = true
  try {
    const [containersRes, infoRes] = await Promise.all([
      api.get('/api/containers'),
      api.get(`/api/container/${containerId}`).catch(() => null)
    ])
    const c = (containersRes.containers || []).find(c => c.Id === containerId || c.Id?.startsWith(containerId))
    container.value = c
    containerInfo.value = infoRes?.info || null
    
  } catch (e) {
    console.error('Failed to load container:', e)
  } finally {
    loading.value = false
  }
}

async function loadLogs(showLoading = true) {
  if (showLoading) {
    logsLoading.value = true
  }
  try {
    const data = await api.get(`/api/container/${containerId}/logs?tail=500`)
    logs.value = data.logs || ''
  } catch (e) {
    showToast(t('containers.loadLogsFailed'))
  } finally {
    if (showLoading) {
      logsLoading.value = false
    }
  }
}

async function startContainer() {
  try {
    await api.post(`/api/container/${containerId}/start`)
    showToast(t('containers.started'))
    loadContainer()
  } catch (e) {
    showToast(t('containers.startFailed') + ': ' + e.message)
  }
}

async function stopContainer() {
  try {
    await api.post(`/api/container/${containerId}/stop`, { timeout: 10 })
    showToast(t('containers.stopped'))
    loadContainer()
  } catch (e) {
    showToast(t('containers.stopFailed') + ': ' + e.message)
  }
}

async function restartContainer() {
  try {
    await api.post(`/api/container/${containerId}/restart`, { timeout: 10 })
    showToast(t('containers.restarted'))
    loadContainer()
  } catch (e) {
    showToast(t('containers.restartFailed') + ': ' + e.message)
  }
}

async function pauseContainer() {
  try {
    await api.post(`/api/container/${containerId}/pause`)
    showToast(t('containers.paused'))
    loadContainer()
  } catch (e) {
    showToast(t('containers.pauseFailed') + ': ' + e.message)
  }
}

async function unpauseContainer() {
  try {
    await api.post(`/api/container/${containerId}/unpause`)
    showToast(t('containers.resumed'))
    loadContainer()
  } catch (e) {
    showToast(t('containers.resumeFailed') + ': ' + e.message)
  }
}

function confirmRemove() {
  showMoreActions.value = false
  showConfirm.value = true
}

async function removeContainer() {
  try {
    await api.post(`/api/container/${containerId}/remove`, { force: true })
    showToast(t('containers.removed'))
    router.push('/containers')
  } catch (e) {
    showToast(t('containers.removeFailed') + ': ' + e.message)
  }
}

function renameContainer() {
  showMoreActions.value = false
  const newName = prompt(t('containers.actions.rename') + ':', getContainerName(container.value))
  if (newName && newName !== getContainerName(container.value)) {
    api.post(`/api/container/${containerId}/rename`, { name: newName })
      .then(() => { showToast(t('containers.renamed')); loadContainer() })
      .catch(e => showToast(t('containers.renameFailed') + ': ' + e.message))
  }
}

// 生成 Docker Run 命令
function generateDockerRun() {
  showMoreActions.value = false
  if (!containerInfo.value) {
    showToast(t('containers.noContainerInfo'))
    return
  }

  const info = containerInfo.value
  const config = info.Config || {}
  const hostConfig = info.HostConfig || {}
  const name = getContainerName(container.value) || 'container'

  const args = []

  // 容器名称
  args.push(`--name ${name}`)

  // 主机名
  if (config.Hostname) {
    args.push(`--hostname ${config.Hostname}`)
  }

  // 工作目录
  if (config.WorkingDir) {
    args.push(`--workdir ${escapeShellArg(config.WorkingDir)}`)
  }

  // 用户
  if (config.User) {
    args.push(`--user ${config.User}`)
  }

  // 端口映射
  if (info.NetworkSettings?.Ports) {
    for (const [containerPort, bindings] of Object.entries(info.NetworkSettings.Ports)) {
      if (bindings && bindings.length > 0) {
        bindings.forEach(binding => {
          const hostPort = binding.HostPort
          const hostIp = binding.HostIp
          if (hostIp && hostIp !== '0.0.0.0') {
            args.push(`-p ${hostIp}:${hostPort}:${containerPort}`)
          } else {
            args.push(`-p ${hostPort}:${containerPort}`)
          }
        })
      }
    }
  }

  // 暴露端口（不映射到主机）
  if (config.ExposedPorts) {
    for (const port of Object.keys(config.ExposedPorts)) {
      args.push(`--expose ${port}`)
    }
  }

  // 卷挂载
  if (info.Mounts && info.Mounts.length > 0) {
    info.Mounts.forEach(mount => {
      let volumeArg
      if (mount.Type === 'volume') {
        volumeArg = `-v ${mount.Name}:${mount.Destination}`
      } else {
        volumeArg = `-v ${mount.Source}:${mount.Destination}`
      }
      if (mount.Mode === 'ro' || mount.RW === false) {
        volumeArg += ':ro'
      }
      args.push(volumeArg)
    })
  }

  // 环境变量
  if (config.Env && config.Env.length > 0) {
    config.Env.forEach(env => {
      args.push(`-e ${escapeShellArg(env)}`)
    })
  }

  // 标签
  if (config.Labels) {
    for (const [key, value] of Object.entries(config.Labels)) {
      if (!key.startsWith('com.docker.')) {
        args.push(`--label ${escapeShellArg(key + '=' + value)}`)
      }
    }
  }

  // 重启策略
  if (hostConfig.RestartPolicy?.Name && hostConfig.RestartPolicy.Name !== 'no') {
    let restartArg = `--restart ${hostConfig.RestartPolicy.Name}`
    if (hostConfig.RestartPolicy.MaximumRetryCount > 0) {
      restartArg += `:${hostConfig.RestartPolicy.MaximumRetryCount}`
    }
    args.push(restartArg)
  }

  // 网络模式
  if (hostConfig.NetworkMode && hostConfig.NetworkMode !== 'default') {
    args.push(`--network ${hostConfig.NetworkMode}`)
  }

  // DNS
  if (hostConfig.Dns && hostConfig.Dns.length > 0) {
    hostConfig.Dns.forEach(dns => args.push(`--dns ${dns}`))
  }
  if (hostConfig.DnsSearch && hostConfig.DnsSearch.length > 0) {
    hostConfig.DnsSearch.forEach(ds => args.push(`--dns-search ${ds}`))
  }

  // Extra Hosts
  if (hostConfig.ExtraHosts && hostConfig.ExtraHosts.length > 0) {
    hostConfig.ExtraHosts.forEach(eh => args.push(`--add-host ${escapeShellArg(eh)}`))
  }

  // 特权模式
  if (hostConfig.Privileged) {
    args.push('--privileged')
  }

  // Capabilities
  if (hostConfig.CapAdd && hostConfig.CapAdd.length > 0) {
    hostConfig.CapAdd.forEach(cap => args.push(`--cap-add ${cap}`))
  }
  if (hostConfig.CapDrop && hostConfig.CapDrop.length > 0) {
    hostConfig.CapDrop.forEach(cap => args.push(`--cap-drop ${cap}`))
  }

  // 设备映射
  if (hostConfig.Devices && hostConfig.Devices.length > 0) {
    hostConfig.Devices.forEach(device => {
      let deviceArg = `--device ${device.PathOnHost}:${device.PathInContainer}`
      if (device.CgroupPermissions) {
        deviceArg += `:${device.CgroupPermissions}`
      }
      args.push(deviceArg)
    })
  }

  // 内存限制
  if (hostConfig.Memory) {
    args.push(`--memory ${hostConfig.Memory}B`)
  }
  if (hostConfig.MemoryReservation) {
    args.push(`--memory-reservation ${hostConfig.MemoryReservation}B`)
  }

  // CPU 限制
  if (hostConfig.CpuQuota && hostConfig.CpuPeriod) {
    const cpus = hostConfig.CpuQuota / hostConfig.CpuPeriod
    args.push(`--cpus ${cpus.toFixed(2)}`)
  }
  if (hostConfig.CpuShares && hostConfig.CpuShares !== 1024) {
    args.push(`--cpu-shares ${hostConfig.CpuShares}`)
  }

  // PID 模式
  if (hostConfig.PidMode) {
    args.push(`--pid ${hostConfig.PidMode}`)
  }

  // IPC 模式
  if (hostConfig.IpcMode) {
    args.push(`--ipc ${hostConfig.IpcMode}`)
  }

  // 只读根文件系统
  if (hostConfig.ReadonlyRootfs) {
    args.push('--read-only')
  }

  // 日志驱动
  if (hostConfig.LogConfig?.Type && hostConfig.LogConfig.Type !== 'json-file') {
    args.push(`--log-driver ${hostConfig.LogConfig.Type}`)
  }

  // Entrypoint
  if (config.Entrypoint && config.Entrypoint.length > 0) {
    args.push(`--entrypoint ${escapeShellArg(config.Entrypoint.join(' '))}`)
  }

  // 镜像
  args.push(config.Image)

  // 命令
  if (config.Cmd && config.Cmd.length > 0) {
    config.Cmd.forEach(cmd => args.push(escapeShellArg(cmd)))
  }

  // 构建命令
  dockerRunCommand.value = 'docker run -d \\\n  ' + args.join(' \\\n  ')
  dockerRunCopied.value = false
  showDockerRunModal.value = true
}

// 转义 shell 参数
function escapeShellArg(arg) {
  if (/^[a-zA-Z0-9_.:@/-]+$/.test(arg)) {
    return arg
  }
  return `"${arg.replace(/"/g, '\\"')}"`
}

// 复制 Docker Run 命令到剪贴板
async function copyDockerRun() {
  try {
    await navigator.clipboard.writeText(dockerRunCommand.value)
    dockerRunCopied.value = true
    showToast(t('common.copied'))
    setTimeout(() => {
      dockerRunCopied.value = false
    }, 2000)
  } catch (e) {
    showToast(t('common.copyFailed'))
  }
}

// 生成 Compose 文件
function generateCompose() {
  showMoreActions.value = false
  if (!containerInfo.value) {
    showToast(t('containers.noContainerInfo'))
    return
  }

  const info = containerInfo.value
  const config = info.Config || {}
  const hostConfig = info.HostConfig || {}
  const name = getContainerName(container.value) || 'container'

  // 构建 Compose 服务配置
  const service = {
    image: config.Image,
    container_name: name
  }

  // 添加主机名
  if (config.Hostname) {
    service.hostname = config.Hostname
  }

  // 添加域名
  if (config.Domainname) {
    service.domainname = config.Domainname
  }

  // 添加工作目录
  if (config.WorkingDir) {
    service.working_dir = config.WorkingDir
  }

  // 添加用户
  if (config.User) {
    service.user = config.User
  }

  // 添加启动命令（Entrypoint）
  if (config.Entrypoint && config.Entrypoint.length > 0) {
    service.entrypoint = config.Entrypoint
  }

  // 添加命令
  if (config.Cmd && config.Cmd.length > 0) {
    service.command = config.Cmd
  }

  // 添加环境变量
  if (config.Env && config.Env.length > 0) {
    service.environment = config.Env.reduce((env, item) => {
      const [key, ...valueParts] = item.split('=')
      env[key] = valueParts.join('=') || ''
      return env
    }, {})
  }

  // 添加端口映射
  if (info.NetworkSettings?.Ports) {
    const ports = []
    for (const [containerPort, bindings] of Object.entries(info.NetworkSettings.Ports)) {
      if (bindings && bindings.length > 0) {
        bindings.forEach(binding => {
          const hostPort = binding.HostPort
          const hostIp = binding.HostIp
          if (hostIp && hostIp !== '0.0.0.0') {
            ports.push(`${hostIp}:${hostPort}:${containerPort}`)
          } else {
            ports.push(`${hostPort}:${containerPort}`)
          }
        })
      }
    }
    if (ports.length > 0) {
      service.ports = ports
    }
  }

  // 添加暴露端口（不映射到主机的端口）
  if (config.ExposedPorts && Object.keys(config.ExposedPorts).length > 0) {
    const exposedPorts = Object.keys(config.ExposedPorts)
    if (!service.ports) {
      service.expose = exposedPorts.map(p => p.split('/')[0])
    }
  }

  // 添加卷挂载
  if (info.Mounts && info.Mounts.length > 0) {
    service.volumes = info.Mounts.map(mount => {
      let volumeStr = ''
      if (mount.Type === 'volume') {
        volumeStr = `${mount.Name}:${mount.Destination}`
      } else {
        volumeStr = `${mount.Source}:${mount.Destination}`
      }
      // 添加读写模式
      if (mount.Mode === 'ro' || mount.RW === false) {
        volumeStr += ':ro'
      }
      return volumeStr
    })
  }

  // 添加网络
  if (info.NetworkSettings?.Networks) {
    const networks = Object.keys(info.NetworkSettings.Networks)
    if (networks.length > 0 && !networks.includes('bridge')) {
      service.networks = networks
    }
  }

  // 添加网络模式（如果不是默认网络）
  if (hostConfig.NetworkMode && hostConfig.NetworkMode !== 'default') {
    service.network_mode = hostConfig.NetworkMode
  }

  // 添加标签
  if (config.Labels && Object.keys(config.Labels).length > 0) {
    // 过滤掉 Docker 自动添加的标签
    const filteredLabels = {}
    for (const [key, value] of Object.entries(config.Labels)) {
      if (!key.startsWith('com.docker.')) {
        filteredLabels[key] = value
      }
    }
    if (Object.keys(filteredLabels).length > 0) {
      service.labels = filteredLabels
    }
  }

  // 添加重启策略
  if (hostConfig.RestartPolicy?.Name && hostConfig.RestartPolicy.Name !== 'no') {
    service.restart = hostConfig.RestartPolicy.Name
  }

  // 添加资源限制
  if (hostConfig.Memory || hostConfig.CpuQuota || hostConfig.CpuShares) {
    service.deploy = {
      resources: {
        limits: {},
        reservations: {}
      }
    }
    if (hostConfig.Memory) {
      service.deploy.resources.limits.memory = `${Math.round(hostConfig.Memory / 1024 / 1024)}M`
    }
    if (hostConfig.MemoryReservation) {
      service.deploy.resources.reservations.memory = `${Math.round(hostConfig.MemoryReservation / 1024 / 1024)}M`
    }
    // CPU 限制
    if (hostConfig.CpuQuota && hostConfig.CpuPeriod) {
      const cpus = hostConfig.CpuQuota / hostConfig.CpuPeriod
      service.deploy.resources.limits.cpus = cpus.toFixed(2)
    }
    // 如果只有 CpuShares，作为相对权重
    if (hostConfig.CpuShares && hostConfig.CpuShares !== 1024) {
      service.cpu_shares = hostConfig.CpuShares
    }
  }

  // 添加特权模式
  if (hostConfig.Privileged) {
    service.privileged = true
  }

  // 添加 capabilities
  if (hostConfig.CapAdd && hostConfig.CapAdd.length > 0) {
    service.cap_add = hostConfig.CapAdd
  }
  if (hostConfig.CapDrop && hostConfig.CapDrop.length > 0) {
    service.cap_drop = hostConfig.CapDrop
  }

  // 添加设备映射
  if (hostConfig.Devices && hostConfig.Devices.length > 0) {
    service.devices = hostConfig.Devices.map(device => {
      let deviceStr = `${device.PathOnHost}:${device.PathInContainer}`
      if (device.CgroupPermissions) {
        deviceStr += `:${device.CgroupPermissions}`
      }
      return deviceStr
    })
  }

  // 添加 DNS 配置
  if (hostConfig.Dns && hostConfig.Dns.length > 0) {
    service.dns = hostConfig.Dns
  }
  if (hostConfig.DnsSearch && hostConfig.DnsSearch.length > 0) {
    service.dns_search = hostConfig.DnsSearch
  }
  if (hostConfig.DnsOptions && hostConfig.DnsOptions.length > 0) {
    service.dns_opt = hostConfig.DnsOptions
  }

  // 添加 extra_hosts
  if (hostConfig.ExtraHosts && hostConfig.ExtraHosts.length > 0) {
    service.extra_hosts = hostConfig.ExtraHosts
  }

  // 添加日志配置
  if (hostConfig.LogConfig?.Type && hostConfig.LogConfig.Type !== 'json-file') {
    service.logging = {
      driver: hostConfig.LogConfig.Type,
      options: hostConfig.LogConfig.Config || {}
    }
  }

  // 添加 PID 模式
  if (hostConfig.PidMode) {
    service.pid = hostConfig.PidMode
  }

  // 添加 IPC 模式
  if (hostConfig.IpcMode) {
    service.ipc = hostConfig.IpcMode
  }

  // 添加 UTS 模式
  if (hostConfig.UTSMode) {
    service.uts = hostConfig.UTSMode
  }

  // 添加只读根文件系统
  if (hostConfig.ReadonlyRootfs) {
    service.read_only = true
  }

  // 添加停止信号
  if (config.StopSignal) {
    service.stop_signal = config.StopSignal
  }

  // 添加健康检查
  if (config.Healthcheck) {
    service.healthcheck = {
      test: config.Healthcheck.Test,
      interval: config.Healthcheck.Interval ? `${Math.round(config.Healthcheck.Interval / 1000000000)}s` : undefined,
      timeout: config.Healthcheck.Timeout ? `${Math.round(config.Healthcheck.Timeout / 1000000000)}s` : undefined,
      retries: config.Healthcheck.Retries,
      start_period: config.Healthcheck.StartPeriod ? `${Math.round(config.Healthcheck.StartPeriod / 1000000000)}s` : undefined
    }
    // 移除 undefined 值
    for (const key of Object.keys(service.healthcheck)) {
      if (service.healthcheck[key] === undefined) {
        delete service.healthcheck[key]
      }
    }
    if (Object.keys(service.healthcheck).length === 0) {
      delete service.healthcheck
    }
  }

  // 构建完整的 Compose 文件
  const compose = {
    version: '3.8',
    services: {
      [name]: service
    }
  }

  // 如果有自定义网络，添加网络配置
  if (info.NetworkSettings?.Networks) {
    const customNetworks = Object.entries(info.NetworkSettings.Networks)
      .filter(([name]) => name !== 'bridge' && name !== 'host' && name !== 'none')
    if (customNetworks.length > 0) {
      compose.networks = {}
      customNetworks.forEach(([name]) => {
        compose.networks[name] = { external: true }
      })
    }
  }

  composeContent.value = yamlStringify(compose)
  copied.value = false
  showComposeModal.value = true
}

// 将对象转换为 YAML 字符串（简化版）
function yamlStringify(obj, indent = 0) {
  const spaces = '  '.repeat(indent)
  let result = ''

  for (const [key, value] of Object.entries(obj)) {
    if (value === null || value === undefined) {
      continue
    }

    if (Array.isArray(value)) {
      result += `${spaces}${key}:\n`
      value.forEach(item => {
        if (typeof item === 'object' && item !== null) {
          result += `${spaces}-\n${yamlStringify(item, indent + 1).replace(/^/gm, '  ')}
`
        } else {
          result += `${spaces}- ${item}\n`
        }
      })
    } else if (typeof value === 'object' && value !== null) {
      result += `${spaces}${key}:\n`
      result += yamlStringify(value, indent + 1)
    } else {
      // 处理包含特殊字符的字符串
      const strValue = String(value)
      if (strValue.includes(':') || strValue.includes('#') || strValue.includes('{') || strValue.includes('"')) {
        result += `${spaces}${key}: "${strValue.replace(/"/g, '\\"')}"\n`
      } else {
        result += `${spaces}${key}: ${strValue}\n`
      }
    }
  }

  return result
}

// 复制 Compose 内容到剪贴板
async function copyCompose() {
  try {
    await navigator.clipboard.writeText(composeContent.value)
    copied.value = true
    showToast(t('common.copied'))
    setTimeout(() => {
      copied.value = false
    }, 2000)
  } catch (e) {
    showToast(t('common.copyFailed'))
  }
}

watch(tab, (newTab, oldTab) => {
  if (newTab === 'logs') {
    loadLogs(true) // 切换标签时显示 loading
    startLogsAutoRefresh()
  } else {
    stopLogsAutoRefresh()
  }
})

function startLogsAutoRefresh() {
  stopLogsAutoRefresh()
  logsInterval = setInterval(() => {
    if (tab.value === 'logs' && !logsLoading.value) {
      loadLogs(false) // 自动刷新时不显示 loading，避免界面跳动
    }
  }, 3000)
}

function stopLogsAutoRefresh() {
  if (logsInterval) {
    clearInterval(logsInterval)
    logsInterval = null
  }
}

onMounted(() => {
  loadContainer()
  if (tab.value === 'logs') {
    loadLogs(true) // 初始加载时显示 loading
    startLogsAutoRefresh()
  }
})

onUnmounted(() => {
  stopLogsAutoRefresh()
})

// 打开编辑模态框
function openEditModal() {
  if (!containerInfo.value) {
    showToast(t('containers.noContainerInfo'))
    return
  }

  const info = containerInfo.value
  const config = info.Config || {}
  const hostConfig = info.HostConfig || {}

  // 填充基本信息
  editForm.value.name = getContainerName(container.value) || ''
  editForm.value.hostname = config.Hostname || ''

  // 填充端口映射
  editForm.value.ports = []
  if (info.NetworkSettings?.Ports) {
    for (const [containerPort, bindings] of Object.entries(info.NetworkSettings.Ports)) {
      if (bindings && bindings.length > 0) {
        bindings.forEach(binding => {
          const hostPort = binding.HostPort
          const hostIp = binding.HostIp
          const [portNum, protocol] = containerPort.split('/')
          const hostValue = (hostIp && hostIp !== '0.0.0.0') ? `${hostIp}:${hostPort}` : hostPort
          editForm.value.ports.push({
            host: hostValue,
            container: portNum,
            protocol: protocol || 'tcp'
          })
        })
      }
    }
  }

  // 填充存储卷
  editForm.value.volumes = []
  if (info.Mounts && info.Mounts.length > 0) {
    info.Mounts.forEach(mount => {
      if (mount.Type === 'bind' || mount.Type === 'volume') {
        const mode = mount.RW === false ? 'ro' : 'rw'
        editForm.value.volumes.push({
          host: mount.Source,
          container: mount.Destination,
          mode: mode
        })
      }
    })
  }

  // 填充环境变量
  editForm.value.envs = []
  if (config.Env && config.Env.length > 0) {
    config.Env.forEach(env => {
      const [key, ...valueParts] = env.split('=')
      editForm.value.envs.push({ key, value: valueParts.join('=') || '' })
    })
  }

  // 填充高级选项
  editForm.value.restartPolicy = hostConfig.RestartPolicy?.Name || 'no'
  editForm.value.memory = hostConfig.Memory ? Math.round(hostConfig.Memory / 1024 / 1024) : null
  editForm.value.cpuShares = hostConfig.CpuShares || null
  editForm.value.privileged = hostConfig.Privileged || false

  editShowAdvanced.value = false
  showEditModal.value = true
}

// 添加端口
function addEditPort() {
  editForm.value.ports.push({ host: '', container: '', protocol: 'tcp' })
}

// 移除端口
function removeEditPort(index) {
  editForm.value.ports.splice(index, 1)
}

// 添加存储卷
function addEditVolume() {
  editForm.value.volumes.push({ host: '', container: '', mode: 'rw' })
}

// 移除存储卷
function removeEditVolume(index) {
  editForm.value.volumes.splice(index, 1)
}

// 添加环境变量
function addEditEnv() {
  editForm.value.envs.push({ key: '', value: '' })
}

// 移除环境变量
function removeEditEnv(index) {
  editForm.value.envs.splice(index, 1)
}

// 保存编辑
async function saveEdit() {
  editLoading.value = true
  try {
    const updateData = {
      restartPolicy: editForm.value.restartPolicy,
      hostname: editForm.value.hostname || undefined,
      memory: editForm.value.memory ? editForm.value.memory * 1024 * 1024 : undefined,
      cpuShares: editForm.value.cpuShares || undefined,
      privileged: editForm.value.privileged
    }

    // 处理端口映射
    const ports = editForm.value.ports
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
      updateData.ports = ports
    }

    // 处理存储卷
    const volumes = editForm.value.volumes
      .filter(v => v.host.trim() && v.container.trim())
      .map(v => {
        const mode = v.mode === 'ro' ? 'ro' : 'rw'
        return `${v.host}:${v.container}:${mode}`
      })

    if (volumes.length > 0) {
      updateData.volumes = volumes
    }

    // 处理环境变量
    const envs = editForm.value.envs
      .filter(e => e.key.trim())
      .map(e => `${e.key}=${e.value}`)

    if (envs.length > 0) {
      updateData.env = envs
    }

    await api.post(`/api/container/${containerId}/update`, updateData)

    // 如果名称改变了，调用重命名接口
    const currentName = getContainerName(container.value)
    if (editForm.value.name && editForm.value.name !== currentName) {
      await api.post(`/api/container/${containerId}/rename`, { name: editForm.value.name })
    }

    showToast(t('containers.updateSuccess'))
    showEditModal.value = false
    // 刷新容器信息
    await loadContainer()
  } catch (e) {
    showToast(t('containers.updateFailed') + ': ' + e.message)
  } finally {
    editLoading.value = false
  }
}
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

.section-icon.status {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
  color: white;
}

.section-icon.actions {
  background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
  color: white;
}

.section-icon.overview {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
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

[data-theme="dark"] .info-card {
  box-shadow: none;
}

.status-row {
  padding: 16px;
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

.info-value.monospace {
  font-family: 'HarmonyOS Sans SC', 'SF Mono', 'Consolas', monospace;
}

/* 快速操作按钮 */
.quick-actions {
  display: flex;
  gap: 8px;
  padding: 0 12px;
  margin-bottom: 12px;
  overflow-x: auto;
  scrollbar-width: none;
  justify-content: space-between;
}

.quick-actions::-webkit-scrollbar {
  display: none;
}

.quick-action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 10px 16px;
  flex: 1;
  min-width: 0;
  border: none;
  border-radius: 10px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
  flex-shrink: 0;
}

.quick-action-btn svg {
  width: 16px;
  height: 16px;
}

.quick-action-btn.start {
  background: linear-gradient(135deg, #00C853 0%, #00E676 100%);
  color: white;
}

.quick-action-btn.stop {
  background: linear-gradient(135deg, #FF5252 0%, #FF1744 100%);
  color: white;
}

.quick-action-btn.restart {
  background: linear-gradient(135deg, #448AFF 0%, #2979FF 100%);
  color: white;
}

.quick-action-btn.pause,
.quick-action-btn.resume {
  background: linear-gradient(135deg, #FFB300 0%, #FFA000 100%);
  color: white;
}

.quick-action-btn.edit {
  background: linear-gradient(135deg, #9C27B0 0%, #7B1FA2 100%);
  color: white;
}

.quick-action-btn.remove {
  background: linear-gradient(135deg, #FF5252 0%, #FF1744 100%);
  color: white;
}

.quick-action-btn.more {
  background: linear-gradient(135deg, #607D8B 0%, #455A64 100%);
  color: white;
}

.quick-action-btn:active {
  transform: scale(0.95);
}

.overview-card {
  background: var(--card-bg);
  margin: 0 12px 12px;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: var(--shadow-sm);
}

[data-theme="dark"] .overview-card {
  box-shadow: none;
}

.overview-item {
  display: flex;
  align-items: flex-start;
  padding: 14px 16px;
  gap: 12px;
}

.overview-item + .overview-item {
  border-top: 1px solid var(--border-color);
}

.overview-label {
  font-size: 14px;
  color: var(--text-secondary);
  flex-shrink: 0;
  width: 80px;
}

.overview-value {
  flex: 1;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  justify-content: flex-end;
}

.overview-tag {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  background: var(--hover-bg);
  border-radius: 6px;
  font-size: 12px;
  color: var(--text-color);
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.overview-tag.env-tag {
  background: rgba(0, 125, 255, 0.1);
  color: #007DFF;
}

.overview-tag.volume-tag {
  background: rgba(102, 126, 234, 0.1);
  color: #667eea;
  font-size: 11px;
}

.overview-tag.more-tag {
  background: var(--border-color);
  color: var(--text-secondary);
}

.actions-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tabs-container {
  padding: 0 12px 12px;
}

.tabs {
  display: flex;
  background: var(--card-bg);
  border-radius: 12px;
  padding: 4px;
  box-shadow: var(--shadow-sm);
}

[data-theme="dark"] .tabs {
  box-shadow: none;
}

.tab {
  flex: 1;
  padding: 10px 12px;
  text-align: center;
  font-size: 14px;
  font-weight: 500;
  color: var(--text-secondary);
  background: none;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.tab.active {
  background: #007DFF;
  color: white;
}

.tab-content {
  padding: 0 12px 12px;
}

.logs-section, .stats-section, .info-section, .terminal-section {
  background: var(--card-bg);
  border-radius: 16px;
  overflow: hidden;
  box-shadow: var(--shadow-sm);
}

[data-theme="dark"] .logs-section,
[data-theme="dark"] .stats-section,
[data-theme="dark"] .info-section,
[data-theme="dark"] .terminal-section {
  box-shadow: none;
}

.section-toolbar {
  padding: 12px 16px;
  display: flex;
  justify-content: flex-end;
}

.toolbar-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  background: var(--hover-bg);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  color: var(--text-color);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
}

.toolbar-btn svg {
  width: 16px;
  height: 16px;
}

.log-viewer {
  background: #1e1e1e;
  color: #d4d4d4;
  padding: 14px;
  font-family: 'HarmonyOS Sans SC', 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  line-height: 1.6;
  overflow-x: hidden;
  white-space: normal;
  word-wrap: break-word;
  overflow-wrap: break-word;
  max-height: 400px;
  overflow-y: auto;
}

.log-viewer :deep(br) {
  display: block;
  content: "";
  margin-bottom: 0.2em;
}

/* Info 标签页使用 pre 标签，需要保留 pre-wrap */
.info-section .log-viewer {
  white-space: pre-wrap;
  word-break: break-all;
}

.info-block {
  padding: 16px;
  border-bottom: 1px solid var(--border-color);
}

.info-block:last-child {
  border-bottom: none;
}

.info-block-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-color);
  margin-bottom: 12px;
}

.info-block-content {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.info-line {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}

.info-line-label {
  font-size: 13px;
  color: var(--text-secondary);
  flex-shrink: 0;
}

.info-line-value {
  font-size: 13px;
  color: var(--text-color);
  text-align: right;
  word-break: break-all;
}

.info-line-value.env-value {
  text-align: left;
  font-family: 'Consolas', 'Monaco', monospace;
  background: var(--hover-bg);
  padding: 4px 8px;
  border-radius: 4px;
  width: 100%;
}

.mount-type {
  display: inline-block;
  padding: 2px 6px;
  background: #007DFF;
  color: white;
  font-size: 11px;
  border-radius: 4px;
  margin-right: 8px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  padding: 16px;
}

.stat-card {
  background: var(--hover-bg);
  border-radius: 12px;
  padding: 16px;
  text-align: center;
}

.stat-value {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-color);
}

.stat-label {
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 4px;
}

.terminal-tab {
  padding: 0 12px 12px;
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
  animation: fadeIn var(--transition-fast);
}

.action-sheet {
  background: var(--card-bg);
  border-radius: 20px 20px 0 0;
  width: 100%;
  max-height: 80vh;
  animation: slideUp var(--transition-normal);
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
  color: #FA2A2D;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes slideUp {
  from { transform: translateY(100%); }
  to { transform: translateY(0); }
}

/* Compose 文件查看器 */
.modal-large {
  max-width: 600px;
  width: 100%;
}

@media (min-width: 768px) {
  .modal-large {
    max-width: 700px;
  }
}

@media (max-height: 600px) {
  .modal-container {
    max-height: 95vh;
  }
}

.compose-viewer {
  background: #1e1e1e;
  color: #d4d4d4;
  padding: 16px;
  border-radius: 8px;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
  overflow-x: auto;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 400px;
  overflow-y: auto;
  margin: 0;
}

/* 模态框样式 */
.modal-overlay {
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

.modal-container {
  background: var(--card-bg);
  border-radius: 16px;
  width: 100%;
  max-width: 420px;
  max-height: calc(100vh - 32px);
  overflow: hidden;
  animation: modalIn var(--transition-fast);
}

@media (min-width: 768px) {
  .modal-container {
    max-width: 520px;
  }
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
}

.modal-title {
  font-size: 17px;
  font-weight: 600;
  color: var(--text-color);
  margin: 0;
}

.modal-close {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--hover-bg);
  border: none;
  border-radius: 8px;
  color: var(--text-secondary);
  cursor: pointer;
}

.modal-close svg {
  width: 18px;
  height: 18px;
}

.modal-body {
  padding: 20px;
  max-height: calc(100vh - 200px);
  overflow-y: auto;
}

@media (max-height: 600px) {
  .modal-body {
    max-height: calc(95vh - 150px);
  }
}

.modal-footer {
  display: flex;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid var(--border-color);
}

.modal-footer .btn {
  flex: 1;
}

.btn {
  padding: 12px 20px;
  border: none;
  border-radius: 10px;
  font-size: 15px;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.btn-primary {
  background: #007DFF;
  color: white;
}

.btn-primary:hover {
  background: #0056CC;
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-secondary {
  background: var(--hover-bg);
  color: var(--text-color);
}

.btn-secondary:hover {
  background: var(--border-color);
}

@keyframes modalIn {
  from {
    opacity: 0;
    transform: scale(0.95);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

/* 编辑模态框样式 */
.edit-modal-body {
  padding: 0;
}

.form-section {
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
}

.form-section:last-child {
  border-bottom: none;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-color);
  margin-bottom: 12px;
}

.form-field {
  margin-bottom: 12px;
}

.form-field:last-child {
  margin-bottom: 0;
}

.form-label {
  display: block;
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 6px;
}

.form-input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  font-size: 14px;
  background: var(--bg-color);
  color: var(--text-color);
}

.form-input:focus {
  outline: none;
  border-color: #007DFF;
}

.form-row {
  display: flex;
  gap: 12px;
}

.form-field.half {
  flex: 1;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.section-header.clickable {
  cursor: pointer;
}

.add-btn {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #007DFF;
  border: none;
  border-radius: 6px;
  color: white;
  cursor: pointer;
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

.dynamic-item.env-item .env-key {
  flex: 0 0 120px;
}

.dynamic-item.env-item .env-equal {
  color: var(--text-secondary);
}

.dynamic-item.env-item .env-value {
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
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--hover-bg);
  border: none;
  border-radius: 6px;
  color: var(--text-secondary);
  cursor: pointer;
}

.remove-btn svg {
  width: 16px;
  height: 16px;
}

.expand-icon {
  width: 20px;
  height: 20px;
  transition: transform 0.2s;
}

.collapsible.expanded .expand-icon {
  transform: rotate(180deg);
}

.advanced-content {
  padding-top: 12px;
}

.checkbox-field {
  padding: 8px 0;
}

.checkbox-item {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
}

.checkbox {
  width: 20px;
  height: 20px;
  border: 2px solid var(--border-color);
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.checkbox.checked {
  background: #007DFF;
  border-color: #007DFF;
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
</style>
