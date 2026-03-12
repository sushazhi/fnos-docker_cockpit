<template>
  <div v-if="isCheckingAuth" class="auth-loading">
    <div class="spinner"></div>
  </div>

  <SetupModal v-else-if="!isInitialized" @setup="handleSetup" />

  <LoginModal v-else-if="!isLoggedIn" @success="handleLogin" />

  <div v-else class="app-container">
    <router-view />
    <BottomNav />
  </div>

  <Toast v-if="toastMessage" :message="toastMessage" />

  <!-- 更新通知弹窗 -->
  <UpdateNotification
    :show="showUpdateModal"
    :current-version="appVersion"
    :latest-version="updateInfo.latestVersion"
    :release-url="updateInfo.releaseUrl"
    :changelog="formatChangelog(updateInfo.changelog)"
    @close="closeUpdateModal"
    @ignore="ignoreUpdate"
  />
</template>

<script setup>
import { ref, onMounted, provide } from 'vue'
import { useRouter } from 'vue-router'
// import { useI18n } from 'vue-i18n'
import api, { setOnSessionExpired } from './services/api'
import BottomNav from './components/BottomNav.vue'
import LoginModal from './components/LoginModal.vue'
import SetupModal from './components/SetupModal.vue'
import Toast from './components/Toast.vue'
import UpdateNotification from './components/UpdateNotification.vue'
import { useUpdateCheck } from './composables/useUpdateCheck'

// const { t } = useI18n()
const router = useRouter()
const isLoggedIn = ref(false)
const isInitialized = ref(false)
const isCheckingAuth = ref(true)
const toastMessage = ref('')
const theme = ref('light')
const appVersion = ref('1.0.0')

// 更新检测
const {
  updateInfo,
  showUpdateModal,
  // checkUpdate,
  ignoreUpdate,
  closeUpdateModal,
  formatChangelog
} = useUpdateCheck(appVersion.value)

function setTheme(newTheme) {
  theme.value = newTheme
  document.documentElement.setAttribute('data-theme', newTheme)
  localStorage.setItem('theme', newTheme)
}

function toggleTheme() {
  setTheme(theme.value === 'light' ? 'dark' : 'light')
}

function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (savedTheme) {
    setTheme(savedTheme)
  } else {
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
    setTheme(prefersDark ? 'dark' : 'light')
  }
}

function showToast(message, duration = 2000) {
  toastMessage.value = message
  setTimeout(() => {
    toastMessage.value = ''
  }, duration)
}

provide('showToast', showToast)
provide('theme', theme)
provide('setTheme', setTheme)
provide('toggleTheme', toggleTheme)

function handleLogin() {
  isLoggedIn.value = true
  // 登录成功后跳转到首页
  router.push('/')
  // 获取应用版本并检查更新
  loadAppInfo()
}

function handleSetup() {
  isInitialized.value = true
  isLoggedIn.value = true
  // 设置成功后跳转到首页
  router.push('/')
}

async function loadAppInfo() {
  try {
    const data = await api.get('/api/system/app-info')
    appVersion.value = data.version || '1.0.0'
    // 更新检查会在 composable 的 onMounted 中自动执行
  } catch (e) {
    console.error('Failed to load app info:', e)
  }
}

async function checkAuth() {
  isCheckingAuth.value = true
  try {
    const data = await api.get('/api/auth/check')
    isInitialized.value = data.initialized === true

    if (isInitialized.value) {
      const meData = await api.get('/api/me').catch(() => null)
      if (meData && meData.authenticated) {
        isLoggedIn.value = true
        if (!api.getCSRFToken()) {
          await api.fetchCSRFToken()
        }
      } else {
        isLoggedIn.value = false
      }
    }
  } catch (e) {
    console.error('Auth check failed:', e)
    isLoggedIn.value = false
  } finally {
    isCheckingAuth.value = false
  }
}

onMounted(() => {
  initTheme()
  checkAuth()

  // 注册会话过期回调
  setOnSessionExpired(() => {
    isLoggedIn.value = false
    showToast('会话已过期，请重新登录')
  })
})
</script>

<style>
.auth-loading {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-color);
  z-index: 9999;
}

.app-container {
  min-height: 100vh;
}
</style>
