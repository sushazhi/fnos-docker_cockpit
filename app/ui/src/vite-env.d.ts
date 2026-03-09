/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

declare module '*/locales/zh-CN' {
  const zhCN: Record<string, any>
  export default zhCN
}

declare module '*/locales/en-US' {
  const enUS: Record<string, any>
  export default enUS
}

interface ImportMetaEnv {
  readonly VITE_APP_TITLE: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
