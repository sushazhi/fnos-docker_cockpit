import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Container, BatchOperationRequest, BatchOperationResult } from '../types'
import api from '../services/api'

export const useContainerStore = defineStore('container', () => {
  const containers = ref<Container[]>([])
  const loading = ref(false)
  const selectedIds = ref<Set<string>>(new Set())

  const runningContainers = computed(() =>
    containers.value.filter((c) => c.State === 'running')
  )

  const stoppedContainers = computed(() =>
    containers.value.filter((c) => c.State !== 'running')
  )

  async function fetchContainers() {
    loading.value = true
    try {
      const data = await api.get<{ containers: Container[] }>('/api/containers')
      containers.value = data.containers || []
    } catch (error) {
      console.error('Failed to fetch containers:', error)
      throw error
    } finally {
      loading.value = false
    }
  }

  async function batchOperation(
    operation: BatchOperationRequest['operation'],
    ids?: string[],
    options?: { force?: boolean; timeout?: number }
  ): Promise<BatchOperationResult> {
    const targetIds = ids || Array.from(selectedIds.value)
    if (targetIds.length === 0) {
      throw new Error('No containers selected')
    }

    const request: BatchOperationRequest = {
      ids: targetIds,
      operation,
      ...options
    }

    const result = await api.post<BatchOperationResult>('/api/containers/batch', request)
    
    // Refresh container list after batch operation
    await fetchContainers()
    
    // Clear selection
    selectedIds.value.clear()
    
    return result
  }

  function toggleSelection(id: string) {
    if (selectedIds.value.has(id)) {
      selectedIds.value.delete(id)
    } else {
      selectedIds.value.add(id)
    }
  }

  function selectAll() {
    containers.value.forEach((c) => selectedIds.value.add(c.Id))
  }

  function clearSelection() {
    selectedIds.value.clear()
  }

  return {
    containers,
    loading,
    selectedIds,
    runningContainers,
    stoppedContainers,
    fetchContainers,
    batchOperation,
    toggleSelection,
    selectAll,
    clearSelection
  }
})
