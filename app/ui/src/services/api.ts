// API 类型定义
export interface ApiResponse<T = unknown> {
  data?: T
  error?: string
  Error?: string
  code?: string
  [key: string]: unknown
}

export interface ApiError extends Error {
  code?: string
  status?: number
}

export type RequestMethod = 'GET' | 'POST' | 'PUT' | 'DELETE'

export interface RequestOptions extends RequestInit {
  method?: RequestMethod
}

// 在开发环境使用相对路径，让 vite 代理处理请求
// 在生产环境使用当前 origin
const API_BASE = import.meta.env.DEV ? '' : window.location.origin
let CSRF_TOKEN = ''
let onSessionExpiredCallback: (() => void) | null = null

export function setOnSessionExpired(callback: () => void) {
  onSessionExpiredCallback = callback
}

export function setCSRFToken(csrfToken: string) {
  CSRF_TOKEN = csrfToken || ''
  if (csrfToken) {
    sessionStorage.setItem('dpanel_csrf_token', csrfToken)
  } else {
    sessionStorage.removeItem('dpanel_csrf_token')
  }
}

export function getCSRFToken(): string {
  return CSRF_TOKEN || sessionStorage.getItem('dpanel_csrf_token') || ''
}

export function clearCSRFToken() {
  CSRF_TOKEN = ''
  sessionStorage.removeItem('dpanel_csrf_token')
}

export async function fetchCSRFToken(): Promise<string | null> {
  try {
    const response = await fetch(`${API_BASE}/api/csrf-token`, { credentials: 'include' })
    if (response.ok) {
      const result: ApiResponse = await response.json()
      const csrfToken = (result.data as Record<string, unknown>)?.csrfToken as string || result.csrfToken as string
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

async function request<T = unknown>(endpoint: string, options: RequestOptions = {}): Promise<T> {
  const url = `${API_BASE}${endpoint}`
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  if (options.headers) {
    Object.assign(headers, options.headers)
  }
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
    const errorData: ApiResponse = await response.json().catch(() => ({}))
    const errorMessage = errorData.error || errorData.Error || '请求失败'
    const error = new Error(errorMessage) as ApiError
    error.code = errorData.code
    error.status = response.status
    
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
          const retryResult: ApiResponse = await retryResponse.json()
          return retryResult.data as T || retryResult as T
        }
        const retryErrorData: ApiResponse = await retryResponse.json().catch(() => ({}))
        const retryError = new Error(retryErrorData.error || retryErrorData.Error || `HTTP ${retryResponse.status}`) as ApiError
        retryError.code = retryErrorData.code
        throw retryError
      }
    }
    throw error
  }
  
  const result: ApiResponse = await response.json()
  return (result.data || result) as T
}

export const api = {
  get<T = unknown>(endpoint: string): Promise<T> {
    return request<T>(endpoint)
  },
  post<T = unknown>(endpoint: string, data?: unknown): Promise<T> {
    return request<T>(endpoint, { method: 'POST', body: data ? JSON.stringify(data) : undefined })
  },
  put<T = unknown>(endpoint: string, data?: unknown): Promise<T> {
    return request<T>(endpoint, { method: 'PUT', body: data ? JSON.stringify(data) : undefined })
  },
  delete<T = unknown>(endpoint: string, data?: unknown): Promise<T> {
    return request<T>(endpoint, { method: 'DELETE', body: data ? JSON.stringify(data) : undefined })
  },
  setCSRFToken,
  getCSRFToken,
  clearCSRFToken,
  fetchCSRFToken
}

export default api
