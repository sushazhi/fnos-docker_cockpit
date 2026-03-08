<template>
  <Teleport to="body">
    <Transition name="slide">
      <div v-if="show" class="update-notification-overlay" @click.self="close">
        <div class="update-notification">
          <button class="update-close" @click="close">&times;</button>
          
          <div class="update-header">
            <div class="update-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                <polyline points="7 10 12 15 17 10"/>
                <line x1="12" y1="15" x2="12" y2="3"/>
              </svg>
            </div>
            <div class="update-title">{{ t('settings.updateAvailable') }}</div>
          </div>
          
          <div class="update-version">
            <span class="version-label">{{ t('settings.currentVersion') }}:</span>
            <span class="version-value">v{{ currentVersion }}</span>
            <span class="version-arrow">→</span>
            <span class="version-label">{{ t('settings.latestVersion') }}:</span>
            <span class="version-value new">v{{ latestVersion }}</span>
          </div>
          
          <div v-if="changelog" class="update-changelog">
            <div class="changelog-title">更新内容</div>
            <div class="changelog-content">{{ changelog }}</div>
          </div>
          
          <div class="update-actions">
            <button class="update-btn secondary" @click="ignore">
              忽略此版本
            </button>
            <a :href="releaseUrl" target="_blank" class="update-btn primary" @click="close">
              {{ t('settings.downloadUpdate') }}
            </a>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps({
  show: {
    type: Boolean,
    default: false
  },
  currentVersion: {
    type: String,
    default: '1.0.0'
  },
  latestVersion: {
    type: String,
    default: ''
  },
  releaseUrl: {
    type: String,
    default: ''
  },
  changelog: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['close', 'ignore'])

const { t } = useI18n()

function close() {
  emit('close')
}

function ignore() {
  emit('ignore')
}
</script>

<style scoped>
.update-notification-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  padding: 16px;
}

.update-notification {
  background: var(--card-bg);
  border-radius: 20px;
  width: 100%;
  max-width: 400px;
  padding: 24px;
  position: relative;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
}

.update-close {
  position: absolute;
  top: 16px;
  right: 16px;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: none;
  border: none;
  font-size: 24px;
  color: var(--text-secondary);
  cursor: pointer;
  border-radius: 8px;
  transition: all 0.2s;
}

.update-close:hover {
  background: var(--hover-bg);
  color: var(--text-color);
}

.update-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.update-icon {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #28a745 0%, #20c997 100%);
  border-radius: 14px;
  color: white;
}

.update-icon svg {
  width: 24px;
  height: 24px;
}

.update-title {
  font-size: 20px;
  font-weight: 700;
  color: var(--text-color);
}

.update-version {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  padding: 12px 16px;
  background: var(--hover-bg);
  border-radius: 12px;
  font-size: 14px;
  margin-bottom: 16px;
}

.version-label {
  color: var(--text-secondary);
}

.version-value {
  font-weight: 600;
  color: var(--text-color);
  font-family: 'SF Mono', 'Consolas', monospace;
}

.version-value.new {
  color: #28a745;
}

.version-arrow {
  color: var(--text-tertiary);
  margin: 0 4px;
}

.update-changelog {
  margin-bottom: 20px;
}

.changelog-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-color);
  margin-bottom: 8px;
}

.changelog-content {
  font-size: 14px;
  color: var(--text-secondary);
  line-height: 1.6;
  max-height: 150px;
  overflow-y: auto;
  padding: 12px;
  background: var(--hover-bg);
  border-radius: 10px;
  white-space: pre-wrap;
}

.update-actions {
  display: flex;
  gap: 12px;
}

.update-btn {
  flex: 1;
  padding: 12px 20px;
  border-radius: 12px;
  font-size: 15px;
  font-weight: 600;
  text-align: center;
  text-decoration: none;
  cursor: pointer;
  border: none;
  transition: all 0.2s;
}

.update-btn.primary {
  background: linear-gradient(135deg, #007DFF 0%, #0066CC 100%);
  color: white;
  box-shadow: 0 4px 12px rgba(0, 125, 255, 0.3);
}

.update-btn.primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(0, 125, 255, 0.4);
}

.update-btn.secondary {
  background: var(--hover-bg);
  color: var(--text-secondary);
  border: 1px solid var(--border-color);
}

.update-btn.secondary:hover {
  background: var(--active-bg);
  color: var(--text-color);
}

/* 动画 */
.slide-enter-active,
.slide-leave-active {
  transition: all 0.3s ease;
}

.slide-enter-from,
.slide-leave-to {
  opacity: 0;
}

.slide-enter-from .update-notification,
.slide-leave-to .update-notification {
  transform: scale(0.95) translateY(20px);
}

/* 移动端适配 */
@media (max-width: 480px) {
  .update-notification {
    padding: 20px;
    margin: 10px;
  }

  .update-title {
    font-size: 18px;
  }

  .update-version {
    font-size: 13px;
    padding: 10px 12px;
  }

  .update-actions {
    flex-direction: column-reverse;
  }

  .update-btn {
    width: 100%;
  }
}
</style>
