<template>
  <div class="dialog-overlay" @click.self="$emit('close')">
    <div class="dialog">
      <div class="dialog-header">
        <h3 class="dialog-title">{{ t('auth.login') }}</h3>
        <button class="dialog-close" @click="$emit('close')">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18" />
            <line x1="6" y1="6" x2="18" y2="18" />
          </svg>
        </button>
      </div>
      <div class="dialog-body">
        <div class="form-field">
          <label class="form-label">{{ t('auth.password') }}</label>
          <div class="password-input-wrapper">
            <input
              :type="showPassword ? 'text' : 'password'"
              class="form-input password-input"
              v-model="form.password"
              @keyup.enter="login"
            />
            <button
              type="button"
              class="password-toggle-btn"
              @click="showPassword = !showPassword"
              :aria-label="showPassword ? t('auth.hidePassword') : t('auth.showPassword')"
            >
              <svg
                v-if="showPassword"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
              >
                <path
                  d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"
                />
                <line x1="1" y1="1" x2="23" y2="23" />
              </svg>
              <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
                <circle cx="12" cy="12" r="3" />
              </svg>
            </button>
          </div>
        </div>
        <div class="form-field remember-field">
          <label class="checkbox-item" @click="rememberPassword = !rememberPassword">
            <div class="checkbox" :class="{ checked: rememberPassword }">
              <svg
                v-if="rememberPassword"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
              >
                <polyline points="20 6 9 17 4 12" />
              </svg>
            </div>
            <span class="checkbox-label">{{ t('auth.rememberPassword') }}</span>
          </label>
        </div>
      </div>
      <div class="dialog-footer">
        <button class="dialog-btn primary btn-block" @click="login" :disabled="logging">
          {{ logging ? t('auth.logging') : t('auth.login') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, inject, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '../services/api'

const { t } = useI18n()
const showToast = inject('showToast')

const emit = defineEmits(['close', 'success'])

const form = ref({ password: '' })
const logging = ref(false)
const showPassword = ref(false)
const rememberPassword = ref(false)

// 简单的编码/解码函数（非加密，仅混淆）
// 注意：这不是真正的加密，只是增加一层混淆，防止明文存储
function encodePassword(password) {
  try {
    // 使用 base64 编码 + 简单的字符替换
    const encoded = btoa(encodeURIComponent(password))
    // 字符替换增加混淆
    return encoded
      .split('')
      .map(c => String.fromCharCode(c.charCodeAt(0) + 1))
      .join('')
  } catch {
    return ''
  }
}

function decodePassword(encoded) {
  try {
    // 反向字符替换
    const decoded = encoded
      .split('')
      .map(c => String.fromCharCode(c.charCodeAt(0) - 1))
      .join('')
    return decodeURIComponent(atob(decoded))
  } catch {
    return ''
  }
}

// 加载保存的密码（使用 sessionStorage，会话级别存储）
onMounted(() => {
  // 优先从 sessionStorage 读取（当前会话）
  const sessionPassword = sessionStorage.getItem('dockpit_session_password')
  if (sessionPassword) {
    try {
      form.value.password = decodePassword(sessionPassword)
      rememberPassword.value = true
      return
    } catch (e) {
      // 解析失败，忽略
    }
  }

  // 如果 sessionStorage 没有，尝试从 localStorage 读取（持久存储）
  const savedPassword = localStorage.getItem('dockpit_remembered_password')
  if (savedPassword) {
    try {
      form.value.password = decodePassword(savedPassword)
      rememberPassword.value = true
    } catch (e) {
      // 解析失败，忽略
    }
  }
})

async function login() {
  if (!form.value.password) return
  logging.value = true
  try {
    await api.post('/api/auth/login', { password: form.value.password })

    // 保存或清除密码
    if (rememberPassword.value) {
      const encoded = encodePassword(form.value.password)
      // 同时保存到 sessionStorage（当前会话）和 localStorage（持久）
      sessionStorage.setItem('dockpit_session_password', encoded)
      localStorage.setItem('dockpit_remembered_password', encoded)
    } else {
      sessionStorage.removeItem('dockpit_session_password')
      localStorage.removeItem('dockpit_remembered_password')
    }

    emit('success')
  } catch (e) {
    // 显示具体的错误信息
    const errorMsg = e.message || t('auth.loginFailed')
    showToast(t('auth.loginFailed') + ': ' + errorMsg)
  } finally {
    logging.value = false
  }
}
</script>

<style scoped>
.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 16px;
}

.dialog {
  background: var(--card-bg);
  border-radius: 20px;
  width: 100%;
  max-width: 360px;
  overflow: hidden;
  box-shadow: var(--shadow-lg);
}

.dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px;
  border-bottom: 1px solid var(--border-color);
}

.dialog-title {
  font-size: 17px;
  font-weight: 600;
  color: var(--text-color);
}

.dialog-close {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: none;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  border-radius: 8px;
}

.dialog-close svg {
  width: 20px;
  height: 20px;
}

.dialog-body {
  padding: 20px;
}

.form-field {
  margin-bottom: 16px;
}

.form-field:last-child {
  margin-bottom: 0;
}

.form-label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
  margin-bottom: 8px;
}

.form-input {
  width: 100%;
  padding: 12px 14px;
  background: var(--input-bg);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  color: var(--text-color);
  font-size: 15px;
  font-family: inherit;
}

.form-input:focus {
  outline: none;
  border-color: #007dff;
  box-shadow: 0 0 0 3px rgba(0, 125, 255, 0.12);
}

/* 密码输入框容器 */
.password-input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.password-input {
  padding-right: 48px;
}

/* 密码显示/隐藏按钮 - 鸿蒙6风格 */
.password-toggle-btn {
  position: absolute;
  right: 0;
  width: 48px;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  flex-shrink: 0;
  transition: color var(--transition-fast);
}

.password-toggle-btn:hover {
  color: var(--text-color);
}

.password-toggle-btn:active {
  color: #007dff;
}

.password-toggle-btn svg {
  width: 20px;
  height: 20px;
}

/* 记住密码 */
.remember-field {
  margin-bottom: 0;
  margin-top: 8px;
}

.checkbox-item {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  user-select: none;
}

.checkbox {
  width: 20px;
  height: 20px;
  border: 2px solid var(--border-color);
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--transition-fast);
  flex-shrink: 0;
}

.checkbox.checked {
  background: #007dff;
  border-color: #007dff;
}

.checkbox svg {
  width: 14px;
  height: 14px;
  color: white;
}

.checkbox-label {
  font-size: 14px;
  color: var(--text-secondary);
}

.dialog-footer {
  padding: 16px 20px;
  border-top: 1px solid var(--border-color);
}

.dialog-btn {
  padding: 12px 20px;
  border: none;
  border-radius: 12px;
  font-size: 15px;
  font-weight: 500;
  cursor: pointer;
}

.dialog-btn.primary {
  background: #007dff;
  color: white;
}

.dialog-btn.primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-block {
  width: 100%;
}
</style>
