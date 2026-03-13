<template>
  <div class="skeleton-loader">
    <div v-for="i in count" :key="i" class="skeleton-item">
      <div class="skeleton-icon"></div>
      <div class="skeleton-content">
        <div class="skeleton-title"></div>
        <div class="skeleton-subtitle"></div>
      </div>
      <div class="skeleton-badge"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  count?: number
}

withDefaults(defineProps<Props>(), {
  count: 3
})
</script>

<style scoped>
.skeleton-loader {
  background: var(--card-bg);
  margin: 0 12px 12px;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: var(--shadow-sm);
}

.skeleton-item {
  display: flex;
  align-items: center;
  padding: 14px 16px;
}

.skeleton-item + .skeleton-item {
  border-top: 1px solid var(--border-color);
}

.skeleton-icon {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  margin-right: 12px;
  background: linear-gradient(90deg, var(--hover-bg) 25%, var(--active-bg) 50%, var(--hover-bg) 75%);
  background-size: 200% 100%;
  animation: skeleton-loading 1.5s ease-in-out infinite;
}

.skeleton-content {
  flex: 1;
  min-width: 0;
}

.skeleton-title {
  height: 16px;
  width: 60%;
  border-radius: 4px;
  background: linear-gradient(90deg, var(--hover-bg) 25%, var(--active-bg) 50%, var(--hover-bg) 75%);
  background-size: 200% 100%;
  animation: skeleton-loading 1.5s ease-in-out infinite;
}

.skeleton-subtitle {
  height: 14px;
  width: 80%;
  margin-top: 8px;
  border-radius: 4px;
  background: linear-gradient(90deg, var(--hover-bg) 25%, var(--active-bg) 50%, var(--hover-bg) 75%);
  background-size: 200% 100%;
  animation: skeleton-loading 1.5s ease-in-out infinite;
  animation-delay: 0.1s;
}

.skeleton-badge {
  width: 60px;
  height: 24px;
  border-radius: 12px;
  background: linear-gradient(90deg, var(--hover-bg) 25%, var(--active-bg) 50%, var(--hover-bg) 75%);
  background-size: 200% 100%;
  animation: skeleton-loading 1.5s ease-in-out infinite;
  animation-delay: 0.2s;
}

@keyframes skeleton-loading {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}
</style>
