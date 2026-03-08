<template>
  <div class="page">
    <div class="header">
      <button class="header-back" @click="$router.back()">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
          <polyline points="15 18 9 12 15 6"/>
        </svg>
      </button>
      <span class="header-title">{{ projectName }}</span>

    </div>
    
    <div v-if="loading" class="loading">
      <div class="spinner"></div>
    </div>
    
    <template v-else>
      <div class="section-header">
        <div class="section-icon services">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="2" y="7" width="20" height="14" rx="2" ry="2"/>
            <path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16"/>
          </svg>
        </div>
        <div class="section-info">
          <div class="section-title">{{ t('compose.services') }}</div>
        </div>
      </div>
      <div class="list-card">
        <div 
          v-for="service in services" 
          :key="service.name" 
          class="list-item"
          @click="service.containerId && $router.push(`/container/${service.containerId}`)"
        >
          <div class="item-icon" :class="getServiceIconClass(service.state)">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="2" y="7" width="20" height="14" rx="2" ry="2"/>
              <path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16"/>
            </svg>
          </div>
          <div class="item-content">
            <div class="item-title">{{ service.name }}</div>
            <div class="item-subtitle">{{ service.image }}</div>
          </div>
          <span class="badge" :class="getStatusClass(service.state)">{{ getStatusText(service.state) }}</span>
        </div>
      </div>
      
      <!-- 快捷操作按钮 -->
      <div class="quick-actions">
        <button v-if="!isRunning" class="quick-action-btn start" @click="startProject" :disabled="operating">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polygon points="5 3 19 12 5 21 5 3"/>
          </svg>
          <span>{{ t('compose.up') }}</span>
        </button>
        <button v-if="isRunning" class="quick-action-btn stop" @click="stopProject" :disabled="operating">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="6" y="4" width="4" height="16"/>
            <rect x="14" y="4" width="4" height="16"/>
          </svg>
          <span>{{ t('compose.down') }}</span>
        </button>
        <button v-if="isRunning" class="quick-action-btn restart" @click="restartProject" :disabled="operating">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="23 4 23 10 17 10"/>
            <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
          </svg>
          <span>{{ t('compose.restart') }}</span>
        </button>
        <button class="quick-action-btn build" @click="buildProject" :disabled="operating">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 2L2 7l10 5 10-5-10-5z"/>
            <path d="M2 17l10 5 10-5"/>
            <path d="M2 12l10 5 10-5"/>
          </svg>
          <span>{{ t('compose.build') }}</span>
        </button>
        <button class="quick-action-btn edit" @click="editYaml">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
            <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
          </svg>
          <span>{{ t('compose.edit') }}</span>
        </button>
        <button class="quick-action-btn remove" @click="confirmRemove">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="3 6 5 6 21 6"/>
            <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
          </svg>
          <span>{{ t('compose.remove') }}</span>
        </button>
      </div>
      
      <div class="tabs-container">
        <div class="tabs">
          <button class="tab" :class="{ active: tab === 'logs' }" @click="tab = 'logs'">{{ t('compose.logs') }}</button>
          <button class="tab" :class="{ active: tab === 'yaml' }" @click="tab = 'yaml'">YAML</button>
        </div>
      </div>
      
      <div class="tab-content">
        <div v-if="tab === 'logs'" class="logs-section">
          <div v-if="logsLoading" class="loading"><div class="spinner"></div></div>
          <div v-else class="log-viewer" v-html="formattedLogs || t('compose.noLogs')"></div>
        </div>
        <div v-else-if="tab === 'yaml'" class="yaml-section">
          <pre class="log-viewer">{{ yaml }}</pre>
        </div>
      </div>
    </template>
    
    <ConfirmModal 
      v-if="showConfirm"
      :title="t('compose.remove')"
      :message="t('common.confirmDelete') + ' ' + projectName + '?'"
      :confirm-text="t('compose.remove')"
      danger
      @close="showConfirm = false"
      @confirm="removeProject"
    />
    
    <!-- 编辑 YAML 模态框 -->
    <div v-if="showEditModal" class="modal-overlay" @click.self="closeEditModal">
      <div class="modal-container modal-large">
        <div class="modal-header">
          <h3 class="modal-title">{{ t('compose.edit') }}</h3>
          <button class="modal-close" @click="closeEditModal">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"/>
              <line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>
        <div class="modal-body">
          <textarea 
            class="yaml-editor" 
            v-model="editedYaml" 
            placeholder="version: '3.8'
services:
  web:
    image: nginx"
          ></textarea>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeEditModal">{{ t('common.cancel') }}</button>
          <button class="btn btn-primary" @click="saveYaml" :disabled="saving">
            {{ saving ? t('common.saving') : t('common.confirm') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { inject } from 'vue'
import api from '../services/api'
import ConfirmModal from '../components/ConfirmModal.vue'
import AnsiToHtml from 'ansi-to-html'
import DOMPurify from 'dompurify'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const showToast = inject('showToast')

const projectName = route.params.name
const loading = ref(true)
const services = ref([])
const yaml = ref('')
const logs = ref('')
const logsLoading = ref(false)
const tab = ref('logs')
const showConfirm = ref(false)
const operating = ref(false)
const showEditModal = ref(false)
const editedYaml = ref('')
const saving = ref(false)

// 计算是否有运行中的服务
const isRunning = computed(() => {
  return services.value.some(s => s.state === 'running')
})

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
    // 移除其他常见的控制字符（保留换行、制表符等）
    .replace(/[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]/g, '')
    // 移除ANSI光标控制序列
    .replace(/\x1B\[[0-9;]*[A-Za-z]/g, '')
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

function getServiceIconClass(state) {
  switch (state) {
    case 'running': return 'icon-running'
    case 'exited': return 'icon-stopped'
    default: return 'icon-default'
  }
}

function getStatusClass(state) {
  switch (state) {
    case 'running': return 'badge-success'
    case 'exited': return 'badge-danger'
    default: return 'badge-info'
  }
}

function getStatusText(state) {
  switch (state) {
    case 'running': return t('compose.state.running')
    case 'exited': return t('compose.state.exited')
    default: return state
  }
}

async function load() {
  loading.value = true
  try {
    const data = await api.get(`/api/compose/${projectName}`)
    yaml.value = data.content || ''
    services.value = data.services || []
  } catch (e) {
    console.error('Failed to load compose project:', e)
  } finally {
    loading.value = false
  }
}

async function loadLogs() {
  logsLoading.value = true
  try {
    const data = await api.get(`/api/compose/${projectName}/logs`)
    logs.value = data.logs || ''
  } catch (e) {
    showToast(t('compose.logsFailed'))
  } finally {
    logsLoading.value = false
  }
}

async function startProject() {
  operating.value = true
  try {
    await api.post(`/api/compose/${projectName}/up`)
    showToast(t('compose.upSuccess'))
    load()
  } catch (e) {
    showToast(t('compose.upFailed') + ': ' + e.message)
  } finally {
    operating.value = false
  }
}

async function stopProject() {
  operating.value = true
  try {
    await api.post(`/api/compose/${projectName}/down`)
    showToast(t('compose.downSuccess'))
    load()
  } catch (e) {
    showToast(t('compose.downFailed') + ': ' + e.message)
  } finally {
    operating.value = false
  }
}

async function restartProject() {
  operating.value = true
  try {
    await api.post(`/api/compose/${projectName}/restart`)
    showToast(t('compose.restartSuccess'))
    load()
  } catch (e) {
    showToast(t('compose.restartFailed') + ': ' + e.message)
  } finally {
    operating.value = false
  }
}

async function buildProject() {
  operating.value = true
  try {
    await api.post(`/api/compose/${projectName}/build`)
    showToast(t('compose.buildSuccess'))
    load()
  } catch (e) {
    showToast(t('compose.buildFailed') + ': ' + e.message)
  } finally {
    operating.value = false
  }
}

function editYaml() {
  editedYaml.value = yaml.value
  showEditModal.value = true
}

function closeEditModal() {
  showEditModal.value = false
  editedYaml.value = ''
}

async function saveYaml() {
  if (!editedYaml.value.trim()) {
    showToast(t('compose.yamlEmpty'))
    return
  }

  saving.value = true
  try {
    await api.post(`/api/compose/${projectName}/save`, {
      content: editedYaml.value
    })
    showToast(t('compose.saveSuccess'))
    yaml.value = editedYaml.value
    showEditModal.value = false
    // 重新加载服务列表
    await loadProject()
  } catch (e) {
    showToast(t('compose.saveFailed') + ': ' + e.message)
  } finally {
    saving.value = false
  }
}

function confirmRemove() {
  showConfirm.value = true
}

async function removeProject() {
  try {
    await api.post(`/api/compose/${projectName}/delete`)
    showToast(t('compose.removeSuccess'))
    router.push('/compose')
  } catch (e) {
    showToast(t('compose.removeFailed') + ': ' + e.message)
  }
}

watch(tab, (newTab) => {
  if (newTab === 'logs') loadLogs()
})

onMounted(() => {
  load()
  loadLogs()
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

.section-icon.services {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.section-icon.actions {
  background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
  color: white;
}

.section-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-color);
}

.list-card {
  background: var(--card-bg);
  margin: 0 12px 12px;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: var(--shadow-sm);
}

[data-theme="dark"] .list-card {
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

.item-icon.icon-running {
  background: rgba(0, 200, 83, 0.1);
  color: #00C853;
}

.item-icon.icon-stopped {
  background: rgba(250, 42, 45, 0.1);
  color: #FA2A2D;
}

.item-icon.icon-default {
  background: rgba(0, 125, 255, 0.1);
  color: #007DFF;
}

.item-content {
  flex: 1;
  min-width: 0;
}

.item-title {
  font-size: 15px;
  font-weight: 500;
  color: var(--text-color);
}

.item-subtitle {
  font-size: 13px;
  color: var(--text-secondary);
  margin-top: 2px;
}

.actions-card {
  background: var(--card-bg);
  margin: 0 12px 12px;
  border-radius: 16px;
  padding: 12px;
  box-shadow: var(--shadow-sm);
}

[data-theme="dark"] .actions-card {
  box-shadow: none;
}

/* 快捷操作按钮 */
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
  border: none;
  border-radius: 10px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
  flex: 1;
  min-width: 0;
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

.quick-action-btn.build {
  background: linear-gradient(135deg, #FF9800 0%, #F57C00 100%);
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

.quick-action-btn:active {
  transform: scale(0.95);
}

.quick-action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.action-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 12px;
  background: var(--hover-bg);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  color: var(--text-color);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.action-btn svg {
  width: 18px;
  height: 18px;
}

.action-btn.primary {
  background: #007DFF;
  border-color: #007DFF;
  color: white;
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
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

.logs-section, .yaml-section {
  background: var(--card-bg);
  border-radius: 16px;
  overflow: hidden;
  box-shadow: var(--shadow-sm);
}

[data-theme="dark"] .logs-section,
[data-theme="dark"] .yaml-section {
  box-shadow: none;
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

.yaml-section .log-viewer {
  white-space: pre-wrap;
  word-break: break-all;
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
  color: #FA2A2D;
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
  padding: 20px;
}

.modal-container {
  background: var(--card-bg);
  border-radius: 16px;
  width: 100%;
  max-width: 400px;
  max-height: 90vh;
  overflow: hidden;
  animation: modalIn 0.2s ease;
}

.modal-container.modal-large {
  max-width: 600px;
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
  max-height: 60vh;
  overflow-y: auto;
}

.modal-footer {
  display: flex;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid var(--border-color);
}

.btn {
  flex: 1;
  padding: 12px 20px;
  border: none;
  border-radius: 10px;
  font-size: 15px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
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

/* YAML 编辑器 */
.yaml-editor {
  width: 100%;
  min-height: 300px;
  padding: 16px;
  background: #1e1e1e;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  color: #d4d4d4;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.6;
  resize: vertical;
}

.yaml-editor:focus {
  outline: none;
  border-color: #007DFF;
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
</style>
