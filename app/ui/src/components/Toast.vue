<template>
  <Transition name="toast">
    <div v-if="visible" class="toast">{{ message }}</div>
  </Transition>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue'

const props = defineProps({
  message: String,
  duration: { type: Number, default: 2500 }
})

const visible = ref(false)

onMounted(() => {
  if (props.message) {
    visible.value = true
    setTimeout(() => {
      visible.value = false
    }, props.duration)
  }
})

watch(() => props.message, (val) => {
  if (val) {
    visible.value = true
    setTimeout(() => {
      visible.value = false
    }, props.duration)
  }
})
</script>

<style scoped>
.toast {
  position: fixed;
  bottom: calc(100px + var(--safe-area-bottom));
  left: 50%;
  transform: translateX(-50%);
  background: var(--text-color);
  color: var(--card-bg);
  padding: 14px 24px;
  border-radius: 12px;
  font-size: 14px;
  font-weight: 500;
  z-index: 2000;
  max-width: 80%;
  text-align: center;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.2);
}

[data-theme="dark"] .toast {
  background: #2a2a2a;
  color: #f0f0f0;
}

.toast-enter-active {
  animation: toastIn 0.2s ease;
}

.toast-leave-active {
  animation: toastOut 0.2s ease;
}

@keyframes toastIn {
  from { 
    opacity: 0;
    transform: translateX(-50%) translateY(10px);
  }
  to { 
    opacity: 1;
    transform: translateX(-50%) translateY(0);
  }
}

@keyframes toastOut {
  from { 
    opacity: 1;
    transform: translateX(-50%) translateY(0);
  }
  to { 
    opacity: 0;
    transform: translateX(-50%) translateY(10px);
  }
}
</style>
