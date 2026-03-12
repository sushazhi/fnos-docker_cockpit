import { createI18n, type I18n } from 'vue-i18n'
import zhCN from './locales/zh-CN'
import enUS from './locales/en-US'

const savedLocale = localStorage.getItem('dockpit_locale') || 'zh-CN'

const i18n: I18n = createI18n({
  legacy: false,
  locale: savedLocale,
  fallbackLocale: 'en-US',
  messages: {
    'zh-CN': zhCN,
    'en-US': enUS
  }
})

export function setLocale(locale: string) {
  ;(i18n.global.locale as unknown as { value: string }).value = locale
  localStorage.setItem('dockpit_locale', locale)
  document.documentElement.setAttribute('lang', locale)
}

export function getLocale(): string {
  return (i18n.global.locale as unknown as { value: string }).value
}

export function t(key: string, params?: Record<string, unknown>): string {
  return (i18n.global.t as (key: string, params?: Record<string, unknown>) => string)(key, params)
}

export default i18n
