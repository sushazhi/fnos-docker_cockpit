/**
 * Dockpit 更新检测 Composable
 * 
 * 功能：
 * 1. 后台自动检测 GitHub 最新版本（24小时一次）
 * 2. 比较版本号判断是否有更新
 * 3. 有更新时显示通知弹窗
 * 4. 支持忽略更新
 * 5. 显示更新日志
 */

import { ref, onMounted } from 'vue'

const IGNORE_KEY = 'dockpit_ignore_version'
const CHECK_TIME_KEY = 'dockpit_update_check_time'
const CHECK_INTERVAL = 24 * 60 * 60 * 1000 // 24小时检查一次

export function useUpdateCheck(currentVersion = '1.0.0') {
  const updateInfo = ref({
    hasUpdate: false,
    latestVersion: '',
    releaseUrl: '',
    changelog: ''
  })

  const showUpdateModal = ref(false)

  function getIgnoredVersion() {
    try {
      return localStorage.getItem(IGNORE_KEY) || ''
    } catch (e) {
      return ''
    }
  }

  function setIgnoredVersion(version) {
    try {
      localStorage.setItem(IGNORE_KEY, version)
    } catch (e) {}
  }

  function getLastCheckTime() {
    try {
      const time = localStorage.getItem(CHECK_TIME_KEY)
      return time ? parseInt(time, 10) : 0
    } catch (e) {
      return 0
    }
  }

  function setCheckTime() {
    try {
      localStorage.setItem(CHECK_TIME_KEY, Date.now().toString())
    } catch (e) {}
  }

  function shouldCheckUpdate() {
    const lastCheck = getLastCheckTime()
    return Date.now() - lastCheck >= CHECK_INTERVAL
  }

  function compareVersions(current, latest) {
    const cur = (current || '').split('.').map(n => parseInt(n, 10) || 0)
    const lat = (latest || '').split('.').map(n => parseInt(n, 10) || 0)
    
    const minLen = Math.min(cur.length, lat.length)
    for (let i = 0; i < minLen; i++) {
      if (lat[i] > cur[i]) return 1
      if (lat[i] < cur[i]) return -1
    }
    
    if (lat.length > cur.length) return 1
    if (cur.length > lat.length) return -1
    return 0
  }

  async function checkUpdate() {
    // 检查是否需要更新（24小时一次）
    if (!shouldCheckUpdate()) {
      console.log('[Dockpit] 24小时内已检查过更新，跳过')
      return
    }

    try {
      const response = await fetch('https://api.github.com/repos/sushazhi/fnos-docker_cockpit/releases/latest', {
        headers: { 'Accept': 'application/vnd.github.v3+json' },
        cache: 'no-store'
      })

      if (!response.ok) throw new Error('HTTP ' + response.status)

      const data = await response.json()
      const latestVersion = (data.tag_name || '').replace(/^v/, '')

      // 更新检查时间
      setCheckTime()

      const hasUpdate = compareVersions(currentVersion, latestVersion) < 0

      if (hasUpdate) {
        // 检查是否已忽略此版本
        if (getIgnoredVersion() === latestVersion) {
          console.log('[Dockpit] 已忽略版本', latestVersion)
          return
        }

        // 有更新，显示弹窗
        updateInfo.value = {
          hasUpdate: true,
          latestVersion,
          releaseUrl: data.html_url || '',
          changelog: data.body || ''
        }
        showUpdateModal.value = true
        console.log('[Dockpit] 发现新版本:', latestVersion)
      } else {
        console.log('[Dockpit] 当前已是最新版本')
      }
    } catch (error) {
      console.error('[Dockpit] 检查更新失败:', error.message)
    }
  }

  function ignoreUpdate() {
    setIgnoredVersion(updateInfo.value.latestVersion)
    showUpdateModal.value = false
  }

  function closeUpdateModal() {
    showUpdateModal.value = false
  }

  function openReleasePage() {
    if (updateInfo.value.releaseUrl) {
      window.open(updateInfo.value.releaseUrl, '_blank')
    } else {
      window.open('https://github.com/sushazhi/fnos-docker_cockpit/releases', '_blank')
    }
  }

  // 格式化更新日志
  function formatChangelog(markdown) {
    if (!markdown) return ''
    let text = markdown.substring(0, 500)
    text = text.split('\n').filter(line => line.trim().length > 0)
    text = text.map(line => line.replace(/^-\s*/, '• '))
    return text.join('\n')
  }

  // 自动检查更新（延迟3秒，避免影响页面加载）
  onMounted(() => {
    setTimeout(checkUpdate, 3000)
  })

  return {
    updateInfo,
    showUpdateModal,
    checkUpdate,
    ignoreUpdate,
    closeUpdateModal,
    openReleasePage,
    formatChangelog
  }
}
