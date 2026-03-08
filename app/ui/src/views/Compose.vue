<template>
  <div class="page">
    <div class="header">
      <button class="header-back" @click="$router.back()">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
          <polyline points="15 18 9 12 15 6"/>
        </svg>
      </button>
      <span class="header-title">{{ t('compose.title') }}</span>
      <button class="header-action" @click="refresh">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="23 4 23 10 17 10"/>
          <polyline points="1 20 1 14 7 14"/>
          <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/>
        </svg>
      </button>
    </div>
    
    <div v-if="loading" class="loading">
      <div class="spinner"></div>
    </div>
    
    <div v-else-if="projects.length === 0" class="empty-state">
      <div class="empty-icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="8" y1="6" x2="21" y2="6"/>
          <line x1="8" y1="12" x2="21" y2="12"/>
          <line x1="8" y1="18" x2="21" y2="18"/>
          <line x1="3" y1="6" x2="3.01" y2="6"/>
          <line x1="3" y1="12" x2="3.01" y2="12"/>
          <line x1="3" y1="18" x2="3.01" y2="18"/>
        </svg>
      </div>
      <div class="empty-text">{{ t('common.noData') }}</div>
    </div>
    
    <div v-else class="list-card">
      <div 
        v-for="project in projects" 
        :key="project.name" 
        class="list-item"
        @click="$router.push(`/compose/${project.name}`)"
      >
        <div class="item-icon compose-icon" :class="{ running: project.status === 'running' }">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="8" y1="6" x2="21" y2="6"/>
            <line x1="8" y1="12" x2="21" y2="12"/>
            <line x1="8" y1="18" x2="21" y2="18"/>
            <line x1="3" y1="6" x2="3.01" y2="6"/>
            <line x1="3" y1="12" x2="3.01" y2="12"/>
            <line x1="3" y1="18" x2="3.01" y2="18"/>
          </svg>
        </div>
        <div class="item-content">
          <div class="item-title">{{ project.name }}</div>
          <div class="item-subtitle">
            {{ project.services ?? 0 }} {{ t('compose.services') }} · 
            <span class="status-text" :class="project.status">{{ getStatusText(project.status) }}</span>
          </div>
        </div>
        <div class="item-arrow">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="9 18 15 12 9 6"/>
          </svg>
        </div>
      </div>
    </div>
    
    <button class="fab" @click="showCreateModal = true">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <line x1="12" y1="5" x2="12" y2="19"/>
        <line x1="5" y1="12" x2="19" y2="12"/>
      </svg>
    </button>
    
    <div v-if="showCreateModal" class="dialog-overlay" @click.self="showCreateModal = false">
      <div class="dialog dialog-large">
        <div class="dialog-header">
          <h3 class="dialog-title">{{ t('compose.create') }}</h3>
          <button class="dialog-close" @click="showCreateModal = false">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"/>
              <line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>
        <div class="dialog-body">
          <div class="form-field">
            <label class="form-label">{{ t('compose.projectName') }}</label>
            <input type="text" class="form-input" v-model="newProject.name" placeholder="my-project" />
          </div>
          <div class="form-field">
            <label class="form-label">{{ t('compose.composeYaml') }}</label>
            <textarea class="form-input textarea" v-model="newProject.yaml" placeholder="version: '3'
services:
  web:
    image: nginx
    ports:
      - '80:80'"></textarea>
          </div>
        </div>
        <div class="dialog-footer">
          <button class="dialog-btn secondary" @click="showCreateModal = false">{{ t('common.cancel') }}</button>
          <button class="dialog-btn primary" @click="createProject" :disabled="creating || !newProject.name">{{ creating ? t('compose.creating') : t('compose.create') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { inject } from 'vue'
import api from '../services/api'

const { t } = useI18n()
const showToast = inject('showToast')

const loading = ref(true)
const projects = ref([])
const showCreateModal = ref(false)
const creating = ref(false)
const newProject = ref({ name: '', yaml: '' })

async function refresh() {
  loading.value = true
  try {
    const data = await api.get('/api/compose')
    console.log('Compose API response:', data)
    console.log('Projects:', data.projects)
    console.log('Projects length:', data.projects?.length)
    projects.value = data.projects || []
    console.log('projects.value set to:', projects.value)
  } catch (e) {
    console.error('Failed to load compose projects:', e)
  } finally {
    loading.value = false
  }
}

async function createProject() {
  if (!newProject.value.name) return
  creating.value = true
  try {
    await api.post('/api/compose/save', {
      name: newProject.value.name,
      content: newProject.value.yaml
    })
    showToast(t('compose.createSuccess'))
    showCreateModal.value = false
    newProject.value = { name: '', yaml: '' }
    refresh()
  } catch (e) {
    showToast(t('compose.createFailed') + ': ' + e.message)
  } finally {
    creating.value = false
  }
}

onMounted(() => {
  refresh()
})

function getStatusText(status) {
  if (status === 'running') {
    return t('containers.state.running')
  }
  return t('containers.state.exited')
}
</script>

<style scoped>
.list-card {
  background: var(--card-bg);
  margin: 12px;
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

.item-icon.compose-icon {
  background: rgba(255, 152, 0, 0.1);
  color: #FF9800;
}

.item-icon.compose-icon.running {
  background: rgba(0, 200, 83, 0.1);
  color: #00C853;
}

.status-text {
  font-weight: 500;
}

.status-text.running {
  color: #00C853;
}

.status-text.stopped {
  color: var(--text-tertiary);
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
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
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
}

.dialog-close svg {
  width: 20px;
  height: 20px;
}

.dialog-body {
  padding: 20px;
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
  border-color: #007DFF;
  box-shadow: 0 0 0 3px rgba(0, 125, 255, 0.12);
}

.form-input.textarea {
  min-height: 200px;
  resize: vertical;
  font-family: 'HarmonyOS Sans SC', 'Consolas', 'Monaco', monospace;
  font-size: 13px;
  line-height: 1.5;
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
}

.dialog-btn.primary {
  background: #007DFF;
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
</style>
