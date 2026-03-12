<template>
  <div class="page">
    <div class="header">
      <button class="header-back" @click="$router.back()">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
          <polyline points="15 18 9 12 15 6" />
        </svg>
      </button>
      <span class="header-title">{{ t('volumes.title') }}</span>
      <button class="header-action" @click="refresh">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="23 4 23 10 17 10" />
          <polyline points="1 20 1 14 7 14" />
          <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15" />
        </svg>
      </button>
    </div>

    <div class="tabs-container">
      <div class="tabs">
        <button class="tab" :class="{ active: filter === 'all' }" @click="filter = 'all'">
          {{ t('volumes.all') }}
        </button>
        <button class="tab" :class="{ active: filter === 'used' }" @click="filter = 'used'">
          {{ t('volumes.used') }}
        </button>
        <button class="tab" :class="{ active: filter === 'unused' }" @click="filter = 'unused'">
          {{ t('volumes.unused') }}
        </button>
      </div>
    </div>

    <div v-if="loading" class="loading">
      <div class="spinner"></div>
    </div>

    <div v-else-if="filteredVolumes.length === 0" class="empty-state">
      <div class="empty-icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <ellipse cx="12" cy="5" rx="9" ry="3" />
          <path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3" />
          <path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5" />
        </svg>
      </div>
      <div class="empty-text">{{ t('common.noData') }}</div>
    </div>

    <div v-else class="list-card">
      <div
        v-for="vol in filteredVolumes"
        :key="vol.Name"
        class="list-item"
        @click="showVolumeActions(vol)"
      >
        <div class="item-icon volume-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <ellipse cx="12" cy="5" rx="9" ry="3" />
            <path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3" />
            <path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5" />
          </svg>
        </div>
        <div class="item-content">
          <div class="item-title">{{ vol.Name }}</div>
          <div class="item-subtitle">
            {{ vol.Driver }} · {{ vol.Mountpoint }}
            <span v-if="vol.containers && vol.containers.length > 0" class="volume-containers">
              · {{ t('volumes.usedBy') }}: {{ vol.containers.join(', ') }}
            </span>
          </div>
        </div>
        <span class="badge" :class="isVolumeUsed(vol) ? 'badge-success' : 'badge-warning'">
          {{ isVolumeUsed(vol) ? t('volumes.used') : t('volumes.unused') }}
        </span>
      </div>
    </div>

    <button class="fab" @click="showCreateModal = true">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <line x1="12" y1="5" x2="12" y2="19" />
        <line x1="5" y1="12" x2="19" y2="12" />
      </svg>
    </button>

    <div v-if="showCreateModal" class="dialog-overlay" @click.self="showCreateModal = false">
      <div class="dialog">
        <div class="dialog-header">
          <h3 class="dialog-title">{{ t('volumes.create') }}</h3>
          <button class="dialog-close" @click="showCreateModal = false">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>
        <div class="dialog-body">
          <div class="form-field">
            <label class="form-label">{{ t('volumes.name') }}</label>
            <input type="text" class="form-input" v-model="newVolumeName" placeholder="my-volume" />
          </div>
        </div>
        <div class="dialog-footer">
          <button class="dialog-btn secondary" @click="showCreateModal = false">
            {{ t('common.cancel') }}
          </button>
          <button class="dialog-btn primary" @click="createVolume" :disabled="!newVolumeName">
            {{ t('volumes.create') }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="showActions" class="action-sheet-overlay" @click.self="showActions = false">
      <div class="action-sheet">
        <div class="action-sheet-handle"></div>
        <div class="action-sheet-content">
          <button class="sheet-btn danger" @click="confirmRemove">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="3 6 5 6 21 6" />
              <path
                d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"
              />
            </svg>
            {{ t('volumes.remove') }}
          </button>
        </div>
      </div>
    </div>

    <ConfirmModal
      v-if="showConfirm"
      :title="t('volumes.remove')"
      :message="t('common.confirmDelete') + ' ' + selectedVolume?.Name + '?'"
      :confirm-text="t('volumes.remove')"
      danger
      @close="showConfirm = false"
      @confirm="removeVolume"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { inject } from 'vue'
import api from '../services/api'
import ConfirmModal from '../components/ConfirmModal.vue'

const { t } = useI18n()
const showToast = inject('showToast')

const loading = ref(true)
const volumes = ref([])
const _containers = ref([])
const filter = ref('all')
const showCreateModal = ref(false)
const newVolumeName = ref('')
const showActions = ref(false)
const showConfirm = ref(false)
const selectedVolume = ref(null)

function isVolumeUsed(vol) {
  // 优先使用后端计算的used字段（包含停止的容器）
  if (vol.used !== undefined) {
    return vol.used
  }
  // 降级使用UsageData.RefCount（仅统计运行中的容器）
  if (!vol.UsageData || vol.UsageData.RefCount === 0) {
    return false
  }
  return vol.UsageData.RefCount > 0
}

const filteredVolumes = computed(() => {
  if (filter.value === 'used') {
    return volumes.value.filter(v => isVolumeUsed(v))
  } else if (filter.value === 'unused') {
    return volumes.value.filter(v => !isVolumeUsed(v))
  }
  return volumes.value
})

async function refresh() {
  loading.value = true
  try {
    const data = await api.get('/api/volumes')
    volumes.value = data.volumes || []
  } catch (e) {
    console.error('Failed to load volumes:', e)
  } finally {
    loading.value = false
  }
}

async function createVolume() {
  if (!newVolumeName.value) return
  try {
    await api.post('/api/volume/create', { name: newVolumeName.value })
    showToast(t('volumes.createSuccess'))
    showCreateModal.value = false
    newVolumeName.value = ''
    refresh()
  } catch (e) {
    showToast(t('volumes.createFailed') + ': ' + e.message)
  }
}

function showVolumeActions(vol) {
  selectedVolume.value = vol
  showActions.value = true
}

function confirmRemove() {
  showActions.value = false
  showConfirm.value = true
}

async function removeVolume() {
  try {
    await api.post(`/api/volume/${selectedVolume.value.Name}/remove`)
    showToast(t('volumes.removeSuccess'))
    showConfirm.value = false
    refresh()
  } catch (e) {
    showToast(t('volumes.removeFailed') + ': ' + e.message)
  }
}

onMounted(() => {
  refresh()
})
</script>

<style scoped>
.tabs-container {
  padding: 12px;
}

.tabs {
  display: flex;
  background: var(--card-bg);
  border-radius: 12px;
  padding: 4px;
  box-shadow: var(--shadow-sm);
}

[data-theme='dark'] .tabs {
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
  background: #007dff;
  color: white;
}

.badge {
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
  flex-shrink: 0;
}

.badge-success {
  background: rgba(0, 200, 83, 0.1);
  color: #00c853;
}

.badge-warning {
  background: rgba(255, 152, 0, 0.1);
  color: #ff9800;
}

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

.item-icon.volume-icon {
  background: rgba(255, 152, 0, 0.1);
  color: #ff9800;
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
  z-index: 999;
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
</style>
