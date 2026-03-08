// 在开发环境使用相对路径，让 vite 代理处理请求
// 在生产环境使用当前 origin
const API_BASE = import.meta.env.DEV ? '' : window.location.origin
let CSRF_TOKEN = ''
let onSessionExpiredCallback = null

export function setOnSessionExpired(callback) {
  onSessionExpiredCallback = callback
}

export function setCSRFToken(csrfToken) {
  CSRF_TOKEN = csrfToken || ''
  if (csrfToken) {
    sessionStorage.setItem('dpanel_csrf_token', csrfToken)
  } else {
    sessionStorage.removeItem('dpanel_csrf_token')
  }
}

export function getCSRFToken() {
  return CSRF_TOKEN || sessionStorage.getItem('dpanel_csrf_token') || ''
}

export function clearCSRFToken() {
  CSRF_TOKEN = ''
  sessionStorage.removeItem('dpanel_csrf_token')
}

export async function fetchCSRFToken() {
  try {
    const response = await fetch(`${API_BASE}/api/csrf-token`, { credentials: 'include' })
    if (response.ok) {
      const result = await response.json()
      const csrfToken = result.data?.csrfToken || result.csrfToken
      if (csrfToken) {
        setCSRFToken(csrfToken)
        return csrfToken
      }
    }
  } catch (e) {
    console.error('Failed to fetch CSRF token:', e)
  }
  return null
}

// 不需要 CSRF Token 的接口
const NO_CSRF_ENDPOINTS = ['/api/auth/setup', '/api/auth/login']

async function request(endpoint, options = {}) {
  const url = `${API_BASE}${endpoint}`
  const headers = { 'Content-Type': 'application/json', ...options.headers }
  const method = options.method || 'GET'
  const needCSRF = method === 'POST' || method === 'PUT' || method === 'DELETE'
  const skipCSRF = NO_CSRF_ENDPOINTS.includes(endpoint)
  
  if (needCSRF && !skipCSRF && !CSRF_TOKEN) {
    await fetchCSRFToken()
  }
  
  if (needCSRF && !skipCSRF && CSRF_TOKEN) {
    headers['X-CSRF-Token'] = CSRF_TOKEN
  }
  
  const response = await fetch(url, { ...options, headers, credentials: 'include' })
  
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}))
    const errorMessage = errorData.error || errorData.Error || '请求失败'
    const error = new Error(errorMessage)
    Object.assign(error, errorData)
    
    if (response.status === 401) {
      clearCSRFToken()
      // 会话过期，通知应用显示登录界面
      if (onSessionExpiredCallback) {
        onSessionExpiredCallback()
      }
    }
    if (response.status === 403 && (errorData.error === 'CSRF验证失败' || errorData.code === 'CSRF_ERROR')) {
      clearCSRFToken()
      const newCSRFToken = await fetchCSRFToken()
      if (newCSRFToken) {
        headers['X-CSRF-Token'] = newCSRFToken
        const retryResponse = await fetch(url, { ...options, headers, credentials: 'include' })
        if (retryResponse.ok) {
          const retryResult = await retryResponse.json()
          return retryResult.data || retryResult
        }
        const retryErrorData = await retryResponse.json().catch(() => ({}))
        const retryError = new Error(retryErrorData.error || retryErrorData.Error || `HTTP ${retryResponse.status}`)
        Object.assign(retryError, retryErrorData)
        throw retryError
      }
    }
    throw error
  }
  
  const result = await response.json()
  return result.data || result
}

export const api = {
  get(endpoint) { return request(endpoint) },
  post(endpoint, data) { return request(endpoint, { method: 'POST', body: JSON.stringify(data) }) },
  put(endpoint, data) { return request(endpoint, { method: 'PUT', body: JSON.stringify(data) }) },
  delete(endpoint, data) { return request(endpoint, { method: 'DELETE', body: JSON.stringify(data) }) },
  setCSRFToken,
  getCSRFToken,
  clearCSRFToken,
  fetchCSRFToken
}

export default api
