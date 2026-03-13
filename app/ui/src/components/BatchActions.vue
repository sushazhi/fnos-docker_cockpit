<template>
  <div v-if="selectedCount > 0" class="batch-actions">
    <div class="batch-info">
      <span>{{ t('containers.selected', { count: selectedCount }) }}</span>
      <button class="btn-text" @click="clearSelection">{{ t('common.clear') }}</button>
    </div>
    <div class="batch-buttons">
      <button class="btn-batch btn-success" @click="handleBatch('start')">
        {{ t('containers.start') }}
      </button>
      <button class="btn-batch btn-warning" @click="handleBatch('stop')">
        {{ t('containers.stop') }}
      </button>
      <button class="btn-batch btn-info" @click="handleBatch('restart')">
        {{ t('containers.restart') }}
      </button>
      <button class="btn-batch btn-danger" @click="handleBatch('remove')">
        {{ t('containers.remove') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, inject } from 'vue'
import { useI18n } from 'vue-i18n'
import { useContainerStore } from '../stores/container'
import type { BatchOperationRequest } from '../types'

const { t } = useI18n()
const containerStore = useContainerStore()
const showToast = inject<(message: string) => void>('showToast')

const selectedCount = computed(() => containerStore.selectedIds.size)

const emit = defineEmits<{
  success: []
}>()

function clearSelection() {
  containerStore.clearSelection()
}

async function handleBatch(operation: BatchOperationRequest['operation']) {
  if (selectedCount.value === 0) return

  const confirmMsg = t(`containers.batch.confirm.${operation}`, { count: selectedCount.value })
  if (!confirm(confirmMsg)) return

  try {
    const result = await containerStore.batchOperation(operation, undefined, {
      force: operation === 'remove'
    })

    const successMsg = t('containers.batch.success', {
      success: result.success.length,
      failed: result.failed.length
    })
    showToast?.(successMsg)

    if (result.failed.length > 0) {
      console.error('Batch operation errors:', result.failed)
    }

    emit('success')
  } catch (error) {
    showToast?.(t('common.error') + ': ' + (error as Error).message)
  }
}
</script>

<style scoped>
.batch-actions {
  position: fixed;
  bottom: 70px;
  left: 12px;
  right: 12px;
  background: var(--card-bg);
  border-radius: 16px;
  padding: 12px;
  box-shadow: var(--shadow-lg);
  z-index: 100;
}

.batch-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-size: 14px;
  color: var(--text-color);
}

.btn-text {
  background: none;
  border: none;
  color: #007dff;
  font-size: 14px;
  cursor: pointer;
  padding: 4px 8px;
}

.batch-buttons {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
}

.btn-batch {
  padding: 10px;
  border: none;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
  color: white;
}

.btn-success {
  background: #00c853;
}

.btn-warning {
  background: #ff9800;
}

.btn-info {
  background: #007dff;
}

.btn-danger {
  background: #fa2a2d;
}

.btn-batch:active {
  transform: scale(0.95);
}
</style>
