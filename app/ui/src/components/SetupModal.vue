<template>
  <div class="dialog-overlay">
    <div class="dialog">
      <div class="dialog-header">
        <h3 class="dialog-title">{{ t('setup.title') }}</h3>
      </div>
      <div class="dialog-body">
        <p class="setup-desc">{{ t('setup.desc') }}</p>
        <div class="form-field">
          <label class="form-label">{{ t('setup.username') }}</label>
          <input type="text" class="form-input" v-model="form.username" />
        </div>
        <div class="form-field">
          <label class="form-label">{{ t('setup.password') }}</label>
          <input type="password" class="form-input" v-model="form.password" />
          <p class="field-hint" :class="{ 'error': passwordError }">{{ passwordHint }}</p>
        </div>
        <div class="form-field">
          <label class="form-label">{{ t('setup.confirmPassword') }}</label>
          <input type="password" class="form-input" v-model="form.confirmPassword" />
          <p class="field-hint" :class="{ 'error': confirmPasswordError }" v-if="confirmPasswordHint">{{ confirmPasswordHint }}</p>
        </div>
      </div>
      <div class="dialog-footer">
        <button class="dialog-btn primary btn-block" @click="setup" :disabled="setting || !isValid">
          {{ setting ? t('setup.setting') : t('setup.submit') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, inject, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '../services/api'

const { t } = useI18n()
const showToast = inject('showToast')

const emit = defineEmits(['setup'])

const form = ref({ username: '', password: '', confirmPassword: '' })
const setting = ref(false)

const passwordError = computed(() => {
  return form.value.password.length > 0 && form.value.password.length < 8
})

const passwordHint = computed(() => {
  if (form.value.password.length === 0) return t('setup.passwordMin')
  if (passwordError.value) return t('setup.passwordMin')
  return '✓ ' + t('setup.passwordMin')
})

const confirmPasswordError = computed(() => {
  return form.value.confirmPassword.length > 0 && form.value.password !== form.value.confirmPassword
})

const confirmPasswordHint = computed(() => {
  if (form.value.confirmPassword.length === 0) return ''
  if (confirmPasswordError.value) return t('setup.passwordMismatch')
  return '✓ ' + t('setup.passwordMatch')
})

const isValid = computed(() => {
  return form.value.username && 
         form.value.password.length >= 8 && 
         form.value.password === form.value.confirmPassword
})

async function setup() {
  if (!isValid.value) {
    if (!form.value.username || !form.value.password) {
      showToast(t('setup.required'))
    } else if (form.value.password.length < 8) {
      showToast(t('setup.passwordMin'))
    } else if (form.value.password !== form.value.confirmPassword) {
      showToast(t('setup.passwordMismatch'))
    }
    return
  }
  setting.value = true
  try {
    await api.post('/api/auth/setup', {
      username: form.value.username,
      password: form.value.password
    })
    emit('setup')
  } catch (e) {
    showToast(t('setup.failed') + ': ' + e.message)
  } finally {
    setting.value = false
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
  background: var(--bg-color);
  display: flex;
  align-items: center;
  justify-content: center;
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
  padding: 24px 20px 0;
}

.dialog-title {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-color);
  text-align: center;
}

.dialog-body {
  padding: 24px 20px;
}

.setup-desc {
  font-size: 14px;
  color: var(--text-secondary);
  text-align: center;
  margin-bottom: 20px;
  line-height: 1.5;
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
  border-color: #007DFF;
  box-shadow: 0 0 0 3px rgba(0, 125, 255, 0.12);
}

.field-hint {
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 6px;
  margin-bottom: 0;
}

.field-hint.error {
  color: #ff4d4f;
}

.dialog-footer {
  padding: 0 20px 24px;
}

.dialog-btn {
  padding: 14px 20px;
  border: none;
  border-radius: 12px;
  font-size: 15px;
  font-weight: 500;
  cursor: pointer;
}

.dialog-btn.primary {
  background: #007DFF;
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
