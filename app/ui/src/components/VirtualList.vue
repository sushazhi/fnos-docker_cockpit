<template>
  <div ref="containerRef" class="virtual-list" @scroll="handleScroll">
    <div class="virtual-list-phantom" :style="{ height: totalHeight + 'px' }"></div>
    <div class="virtual-list-content" :style="{ transform: `translateY(${offset}px)` }">
      <slot name="item" v-for="item in visibleItems" :key="getItemKey(item)" :item="item" />
    </div>
  </div>
</template>

<script setup lang="ts" generic="T">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { throttle } from '../utils/debounce'

interface Props {
  items: T[]
  itemHeight: number
  buffer?: number
  keyField?: keyof T
}

const props = withDefaults(defineProps<Props>(), {
  buffer: 5,
  keyField: 'id' as keyof T
})

const containerRef = ref<HTMLElement>()
const scrollTop = ref(0)
const containerHeight = ref(0)

const totalHeight = computed(() => props.items.length * props.itemHeight)

const startIndex = computed(() => {
  const index = Math.floor(scrollTop.value / props.itemHeight) - props.buffer
  return Math.max(0, index)
})

const endIndex = computed(() => {
  const visibleCount = Math.ceil(containerHeight.value / props.itemHeight)
  const index = startIndex.value + visibleCount + props.buffer * 2
  return Math.min(props.items.length, index)
})

const visibleItems = computed(() => {
  return props.items.slice(startIndex.value, endIndex.value)
})

const offset = computed(() => startIndex.value * props.itemHeight)

function getItemKey(item: T): string | number {
  const key = item[props.keyField]
  return typeof key === 'string' || typeof key === 'number' ? key : String(key)
}

const handleScroll = throttle(() => {
  if (containerRef.value) {
    scrollTop.value = containerRef.value.scrollTop
  }
}, 16) // ~60fps

function updateContainerHeight() {
  if (containerRef.value) {
    containerHeight.value = containerRef.value.clientHeight
  }
}

onMounted(() => {
  updateContainerHeight()
  window.addEventListener('resize', updateContainerHeight)
})

onUnmounted(() => {
  window.removeEventListener('resize', updateContainerHeight)
})

watch(() => props.items.length, () => {
  // Reset scroll when items change significantly
  if (containerRef.value && scrollTop.value > totalHeight.value) {
    containerRef.value.scrollTop = 0
    scrollTop.value = 0
  }
})
</script>

<style scoped>
.virtual-list {
  height: 100%;
  overflow-y: auto;
  position: relative;
}

.virtual-list-phantom {
  position: absolute;
  left: 0;
  top: 0;
  right: 0;
  z-index: -1;
}

.virtual-list-content {
  position: absolute;
  left: 0;
  right: 0;
  top: 0;
}
</style>
