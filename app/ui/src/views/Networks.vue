<template>
  <div class="page">
    <div class="header">
      <button class="header-back" @click="$router.back()">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
          <polyline points="15 18 9 12 15 6" />
        </svg>
      </button>
      <span class="header-title">{{ t('networks.title') }}</span>
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

    <div v-else-if="networks.length === 0" class="empty-state">
      <div class="empty-icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10" />
          <line x1="2" y1="12" x2="22" y2="12" />
          <path
            d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"
          />
        </svg>
      </div>
      <div class="empty-text">{{ t('common.noData') }}</div>
    </div>

    <div v-else class="list-card">
      <div v-for="net in networks" :key="net.Id" class="list-item" @click="showNetworkDetail(net)">
        <div class="item-icon network-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10" />
            <line x1="2" y1="12" x2="22" y2="12" />
            <path
              d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"
            />
          </svg>
        </div>
        <div class="item-content">
          <div class="item-title">{{ net.Name }}</div>
          <div class="item-subtitle">
            {{ net.Driver }} · {{ getContainerCount(net) }} {{ t('networks.containers') }}
          </div>
        </div>
        <span class="badge" :class="getScopeClass(net.Scope)">{{ net.Scope }}</span>
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
          <h3 class="dialog-title">{{ t('networks.create') }}</h3>
          <button class="dialog-close" @click="showCreateModal = false">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>
        <div class="dialog-body">
          <div class="form-field">
            <label class="form-label">{{ t('networks.name') }}</label>
            <input
              type="text"
              class="form-input"
              v-model="newNetworkName"
              placeholder="my-network"
            />
          </div>
          <div class="form-field">
            <label class="form-label">{{ t('networks.driver') }}</label>
            <select class="form-input" v-model="newNetworkDriver">
              <option value="bridge">bridge</option>
              <option value="overlay">overlay</option>
              <option value="macvlan">macvlan</option>
            </select>
          </div>
        </div>
        <div class="dialog-footer">
          <button class="dialog-btn secondary" @click="showCreateModal = false">
            {{ t('common.cancel') }}
          </button>
          <button class="dialog-btn primary" @click="createNetwork" :disabled="!newNetworkName">
            {{ t('networks.create') }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="showDetailModal" class="dialog-overlay" @click.self="showDetailModal = false">
      <div class="dialog dialog-large">
        <div class="dialog-header">
          <h3 class="dialog-title">{{ selectedNetwork?.Name }}</h3>
          <button class="dialog-close" @click="showDetailModal = false">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>
        <div class="dialog-body">
          <div class="detail-section">
            <div class="detail-row">
              <span class="detail-label">{{ t('networks.driver') }}</span>
              <span class="detail-value">{{ networkDetail?.Driver }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">{{ t('networks.scope') }}</span>
              <span class="detail-value">{{ networkDetail?.Scope }}</span>
            </div>
            <div class="detail-row" v-if="networkDetail?.IPAM?.Config?.length">
              <span class="detail-label">{{ t('networks.subnet') }}</span>
              <span class="detail-value">{{ networkDetail.IPAM.Config[0].Subnet }}</span>
            </div>
          </div>

          <div class="containers-section">
            <div class="section-title">{{ t('networks.connectedContainers') }}</div>
            <div v-if="connectedContainers.length === 0" class="empty-containers">
              {{ t('networks.noContainers') }}
            </div>
            <div v-else class="container-list">
              <div
                v-for="container in connectedContainers"
                :key="container.EndpointID"
                class="container-item"
                @click="goToContainer(container.ContainerId)"
              >
                <div class="container-icon">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <rect x="2" y="7" width="20" height="14" rx="2" ry="2" />
                    <path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16" />
                  </svg>
                </div>
                <div class="container-info">
                  <div class="container-name">
                    {{ container.Name || container.ContainerId?.substring(0, 12) }}
                  </div>
                  <div class="container-ip" v-if="container.IPv4Address">
                    IP: {{ container.IPv4Address }}
                  </div>
                </div>
                <div class="container-arrow">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="9 18 15 12 9 6" />
                  </svg>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div class="dialog-footer">
          <button class="dialog-btn danger" @click="confirmRemove" v-if="canDelete">
            {{ t('networks.remove') }}
          </button>
          <button class="dialog-btn primary" @click="showDetailModal = false">
            {{ t('common.close') }}
          </button>
        </div>
      </div>
    </div>

    <ConfirmModal
      v-if="showConfirm"
      :title="t('networks.remove')"
      :message="t('common.confirmDelete') + ' ' + selectedNetwork?.Name + '?'"
      :confirm-text="t('networks.remove')"
      danger
      @close="showConfirm = false"
      @confirm="removeNetwork"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { inject } from 'vue'
import api from '../services/api'
import ConfirmModal from '../components/ConfirmModal.vue'

const { t } = useI18n()
const router = useRouter()
const showToast = inject('showToast')

const loading = ref(true)
const networks = ref([])
const showCreateModal = ref(false)
const newNetworkName = ref('')
const newNetworkDriver = ref('bridge')
const showDetailModal = ref(false)
const showConfirm = ref(false)
const selectedNetwork = ref(null)
const networkDetail = ref(null)

const connectedContainers = computed(() => {
  if (!networkDetail.value?.Containers) return []
  return Object.entries(networkDetail.value.Containers).map(([id, info]) => ({
    ContainerId: id,
    Name: info.Name,
    EndpointID: info.EndpointID,
    IPv4Address: info.IPv4Address,
    MacAddress: info.MacAddress
  }))
})

const canDelete = computed(() => {
  if (!selectedNetwork.value) return false
  return !['bridge', 'host', 'none'].includes(selectedNetwork.value.Name)
})

function getContainerCount(net) {
  return net.ContainerCount || 0
}

function getScopeClass(scope) {
  switch (scope) {
    case 'local':
      return 'badge-info'
    case 'swarm':
      return 'badge-success'
    default:
      return 'badge-info'
  }
}

async function refresh() {
  loading.value = true
  try {
    const data = await api.get('/api/networks')
    networks.value = data.networks || []
  } catch (e) {
    console.error('Failed to load networks:', e)
  } finally {
    loading.value = false
  }
}

async function showNetworkDetail(net) {
  selectedNetwork.value = net
  showDetailModal.value = true

  try {
    const data = await api.get(`/api/network/${net.Id}`)
    networkDetail.value = data.info
  } catch (e) {
    showToast(t('networks.loadFailed'))
    networkDetail.value = null
  }
}

function goToContainer(containerId) {
  showDetailModal.value = false
  router.push(`/container/${containerId}`)
}

async function createNetwork() {
  if (!newNetworkName.value) return
  try {
    await api.post('/api/network/create', {
      name: newNetworkName.value,
      driver: newNetworkDriver.value
    })
    showToast(t('networks.createSuccess'))
    showCreateModal.value = false
    newNetworkName.value = ''
    refresh()
  } catch (e) {
    showToast(t('networks.createFailed') + ': ' + e.message)
  }
}

function confirmRemove() {
  showDetailModal.value = false
  showConfirm.value = true
}

async function removeNetwork() {
  try {
    await api.post(`/api/network/${selectedNetwork.value.Name}/remove`)
    showToast(t('networks.removeSuccess'))
    showConfirm.value = false
    refresh()
  } catch (e) {
    showToast(t('networks.removeFailed') + ': ' + e.message)
  }
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

.item-icon.network-icon {
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
}

.badge {
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
  flex-shrink: 0;
}

.badge-info {
  background: rgba(0, 125, 255, 0.1);
  color: #007dff;
}

.badge-success {
  background: rgba(0, 200, 83, 0.1);
  color: #00c853;
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

.dialog-large {
  max-width: 420px;
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
  max-height: 60vh;
  overflow-y: auto;
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

.dialog-btn.danger {
  background: rgba(250, 42, 45, 0.1);
  color: #fa2a2d;
}

.detail-section {
  margin-bottom: 20px;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  padding: 10px 0;
  border-bottom: 1px solid var(--border-color);
}

.detail-row:last-child {
  border-bottom: none;
}

.detail-label {
  font-size: 14px;
  color: var(--text-secondary);
}

.detail-value {
  font-size: 14px;
  color: var(--text-color);
  font-family: 'HarmonyOS Sans SC', 'SF Mono', 'Consolas', monospace;
}

.section-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-color);
  margin-bottom: 12px;
}

.empty-containers {
  text-align: center;
  padding: 24px;
  color: var(--text-secondary);
  font-size: 14px;
  background: var(--hover-bg);
  border-radius: 12px;
}

.container-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.container-item {
  display: flex;
  align-items: center;
  padding: 12px;
  background: var(--hover-bg);
  border-radius: 12px;
  cursor: pointer;
  transition: background var(--transition-fast);
}

.container-item:hover {
  background: var(--active-bg);
}

.container-icon {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 200, 83, 0.1);
  color: #00c853;
  border-radius: 10px;
  margin-right: 12px;
  flex-shrink: 0;
}

.container-icon svg {
  width: 18px;
  height: 18px;
}

.container-info {
  flex: 1;
  min-width: 0;
}

.container-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-color);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.container-ip {
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 2px;
  font-family: 'HarmonyOS Sans SC', 'SF Mono', 'Consolas', monospace;
}

.container-arrow {
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-tertiary);
  flex-shrink: 0;
}

.container-arrow svg {
  width: 16px;
  height: 16px;
}
</style>
