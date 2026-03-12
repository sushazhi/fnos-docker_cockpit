<template>
  <div class="page">
    <div class="header">
      <button class="header-back" @click="$router.back()">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
          <polyline points="15 18 9 12 15 6" />
        </svg>
      </button>
      <span class="header-title">{{ t('containers.title') }}</span>
      <button class="header-action" @click="refresh(true)">
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

    <div class="tabs-container">
      <div class="tabs">
        <button class="tab" :class="{ active: filter === 'all' }" @click="filter = 'all'">
          {{ t('containers.all') }}
        </button>
        <button class="tab" :class="{ active: filter === 'running' }" @click="filter = 'running'">
          {{ t('containers.running') }}
        </button>
        <button class="tab" :class="{ active: filter === 'stopped' }" @click="filter = 'stopped'">
          {{ t('containers.stopped') }}
        </button>
      </div>
    </div>

    <div v-if="loading" class="loading">
      <div class="spinner"></div>
    </div>

    <div v-else-if="filteredContainers.length === 0" class="empty-state">
      <div class="empty-icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <rect x="2" y="7" width="20" height="14" rx="2" ry="2" />
          <path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16" />
        </svg>
      </div>
      <div class="empty-text">{{ t('common.noData') }}</div>
    </div>

    <div v-else class="list-card">
      <div
        v-for="c in filteredContainers"
        :key="c.Id"
        class="list-item"
        @click="$router.push(`/container/${c.Id}`)"
      >
        <div class="item-icon" :class="getIconClass(c.State)">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="2" y="7" width="20" height="14" rx="2" ry="2" />
            <path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16" />
          </svg>
        </div>
        <div class="item-content">
          <div class="item-title">{{ getContainerName(c) }}</div>
          <div class="item-subtitle">{{ c.Image }}</div>
        </div>
        <span class="badge" :class="getStatusClass(c.State)">{{ getStatusText(c.State) }}</span>
      </div>
    </div>

    <button class="fab" @click="showCreateModal = true">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <line x1="12" y1="5" x2="12" y2="19" />
        <line x1="5" y1="12" x2="19" y2="12" />
      </svg>
    </button>

    <CreateContainerModal
      v-if="showCreateModal"
      @close="showCreateModal = false"
      @created="handleCreated"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { inject } from 'vue'
import api from '../services/api'
import CreateContainerModal from '../components/CreateContainerModal.vue'

const { t } = useI18n()
const route = useRoute()
const _showToast = inject('showToast')

const loading = ref(true)
const containers = ref([])
const filter = ref('all')
const showCreateModal = ref(false)

function getContainerName(c) {
  if (c.Names && c.Names.length > 0) {
    return c.Names[0].replace(/^\//, '')
  }
  return c.Id?.substring(0, 12) || 'unknown'
}

const filteredContainers = computed(() => {
  if (filter.value === 'running') {
    return containers.value.filter(c => c.State === 'running')
  } else if (filter.value === 'stopped') {
    return containers.value.filter(c => c.State !== 'running')
  }
  return containers.value
})

function getStatusClass(state) {
  switch (state) {
    case 'running':
      return 'badge-success'
    case 'paused':
      return 'badge-warning'
    case 'exited':
      return 'badge-danger'
    default:
      return 'badge-info'
  }
}

function getIconClass(state) {
  switch (state) {
    case 'running':
      return 'icon-running'
    case 'paused':
      return 'icon-paused'
    case 'exited':
      return 'icon-stopped'
    default:
      return 'icon-default'
  }
}

function getStatusText(state) {
  return t(`containers.state.${state}`) || state
}

async function refresh(showLoading = true) {
  if (showLoading) {
    loading.value = true
  }
  try {
    const data = await api.get('/api/containers')
    containers.value = data.containers || []
  } catch (e) {
    console.error('Failed to load containers:', e)
  } finally {
    if (showLoading) {
      loading.value = false
    }
  }
}

function handleCreated() {
  showCreateModal.value = false
  refresh()
}

onMounted(() => {
  if (route.query.image) {
    showCreateModal.value = true
  }
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

.item-icon.icon-running {
  background: rgba(0, 200, 83, 0.1);
  color: #00c853;
}

.item-icon.icon-paused {
  background: rgba(255, 152, 0, 0.1);
  color: #ff9800;
}

.item-icon.icon-stopped {
  background: rgba(250, 42, 45, 0.1);
  color: #fa2a2d;
}

.item-icon.icon-default {
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
</style>
