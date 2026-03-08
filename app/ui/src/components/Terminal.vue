<template>
  <div class="terminal-wrapper">
    <div class="terminal-toolbar">
      <div class="toolbar-left">
        <button class="toolbar-btn" @click="createTerminal" :disabled="loading || isConnected">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="4 17 10 11 4 5"/>
            <line x1="12" y1="19" x2="20" y2="19"/>
          </svg>
          {{ loading ? '...' : connectText }}
        </button>
        <button v-if="isConnected" class="toolbar-btn danger" @click="disconnect" :disabled="loading">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"/>
            <line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
          断开连接
        </button>
      </div>
      <div class="terminal-status" :class="{ connected: isConnected }">
        <span class="status-dot"></span>
        {{ isConnected ? '已连接' : '未连接' }}
      </div>
    </div>
    <div ref="terminalContainer" class="terminal-container"></div>
  </div>
</template>

<script setup>
import { ref, onUnmounted } from 'vue'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import '@xterm/xterm/css/xterm.css'
import api from '../services/api'

const props = defineProps({
  containerId: {
    type: String,
    required: true
  },
  connectText: {
    type: String,
    default: '连接终端'
  }
})

const emit = defineEmits(['error', 'connected', 'disconnected'])

const terminalContainer = ref(null)
const loading = ref(false)
const isConnected = ref(false)

let terminal = null
let fitAddon = null
let execId = null
let ws = null
let dataDisposable = null

async function createTerminal() {
  if (loading.value || isConnected.value) return

  loading.value = true

  try {
    const res = await api.post(`/api/container/${props.containerId}/terminal`, {
      cmd: ['/bin/sh'],
      width: 80,
      height: 24
    })

    execId = res.execId

    if (!terminal) {
      terminal = new Terminal({
        theme: {
          background: '#1e1e1e',
          foreground: '#d4d4d4',
          cursor: '#ffffff',
          cursorAccent: '#1e1e1e',
          selection: 'rgba(255, 255, 255, 0.3)',
          black: '#000000',
          red: '#cd3131',
          green: '#0dbc79',
          yellow: '#e5e510',
          blue: '#2472c8',
          magenta: '#bc3fbc',
          cyan: '#11a8cd',
          white: '#e5e5e5',
          brightBlack: '#666666',
          brightRed: '#f14c4c',
          brightGreen: '#23d18b',
          brightYellow: '#f5f543',
          brightBlue: '#3b8eea',
          brightMagenta: '#d670d6',
          brightCyan: '#29b8db',
          brightWhite: '#e5e5e5'
        },
        fontFamily: '"HarmonyOS Sans SC", "Cascadia Code", "Fira Code", Consolas, "Courier New", monospace',
        fontSize: 13,
        lineHeight: 1.4,
        cursorBlink: true,
        cursorStyle: 'block',
        scrollback: 1000,
        tabStopWidth: 4,
        bellStyle: 'none'
      })
      
      fitAddon = new FitAddon()
      terminal.loadAddon(fitAddon)
      terminal.loadAddon(new WebLinksAddon())
      
      terminal.open(terminalContainer.value)
      fitAddon.fit()

      window.addEventListener('resize', handleResize)
    }

    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsHost = window.location.host
    ws = new WebSocket(`${wsProtocol}//${wsHost}/api/exec/${execId}/ws`)

    ws.onopen = () => {
      isConnected.value = true
      emit('connected')
      terminal.clear()
      terminal.writeln('\x1b[1;32m✓ 终端已连接\x1b[0m')
      terminal.writeln('')

      setTimeout(() => {
        if (fitAddon) {
          fitAddon.fit()
          resizeTerminal(fitAddon.proposeDimensions().cols, fitAddon.proposeDimensions().rows)
        }
      }, 100)
    }

    ws.onmessage = (event) => {
      if (terminal) {
        terminal.write(event.data)
      }
    }

    ws.onerror = () => {
      emit('error', '终端连接错误')
      disconnect()
    }

    ws.onclose = () => {
      disconnect()
    }

    // 移除之前的监听器（如果有）
    if (dataDisposable) {
      dataDisposable.dispose()
    }

    // 添加新的数据监听器
    dataDisposable = terminal.onData((data) => {
      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(data)
      }
    })
    
  } catch (e) {
    emit('error', e.message || '连接终端失败')
  } finally {
    loading.value = false
  }
}

async function resizeTerminal(cols, rows) {
  if (!execId || !cols || !rows) return

  try {
    await api.post(`/api/exec/${execId}/resize`, {
      width: Math.floor(cols),
      height: Math.floor(rows)
    })

    if (terminal && fitAddon) {
      terminal.resize(Math.floor(cols), Math.floor(rows))
      fitAddon.fit()
    }
  } catch (e) {
    console.error('Failed to resize terminal:', e)
  }
}

function handleResize() {
  if (fitAddon && terminal) {
    fitAddon.fit()
    const dims = fitAddon.proposeDimensions()
    if (dims && dims.cols && dims.rows && execId) {
      resizeTerminal(dims.cols, dims.rows)
    }
  }
}

function disconnect() {
  // 清理 WebSocket 连接
  if (ws) {
    ws.onopen = null
    ws.onmessage = null
    ws.onerror = null
    ws.onclose = null
    ws.close()
    ws = null
  }

  // 清理数据监听器
  if (dataDisposable) {
    dataDisposable.dispose()
    dataDisposable = null
  }

  execId = null
  isConnected.value = false
  emit('disconnected')

  if (terminal) {
    terminal.writeln('')
    terminal.writeln('\x1b[1;33m⚠ 终端已断开\x1b[0m')
  }
}

onUnmounted(() => {
  disconnect()

  if (terminal) {
    terminal.dispose()
    terminal = null
  }

  if (fitAddon) {
    fitAddon = null
  }

  window.removeEventListener('resize', handleResize)
})

defineExpose({
  createTerminal,
  disconnect,
  isConnected
})
</script>

<style scoped>
.terminal-wrapper {
  display: flex;
  flex-direction: column;
  background: var(--card-bg);
  border-radius: 16px;
  overflow: hidden;
  box-shadow: var(--shadow-sm);
}

[data-theme="dark"] .terminal-wrapper {
  box-shadow: none;
}

.terminal-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-color);
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.toolbar-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  background: var(--hover-bg);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  color: var(--text-color);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.toolbar-btn:hover:not(:disabled) {
  background: var(--active-bg);
}

.toolbar-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.toolbar-btn.danger {
  color: #ff4d4f;
  border-color: #ff4d4f;
}

.toolbar-btn.danger:hover:not(:disabled) {
  background: #ff4d4f;
  color: white;
}

.toolbar-btn svg {
  width: 16px;
  height: 16px;
}

.terminal-status {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--text-secondary);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #ff4d4f;
  transition: background var(--transition-fast);
}

.terminal-status.connected .status-dot {
  background: #52c41a;
}

.terminal-container {
  background: #1e1e1e;
  padding: 14px;
  min-height: 300px;
  max-height: 400px;
  overflow: hidden;
}

/* xterm.js 样式覆盖 */
:deep(.xterm) {
  padding: 0;
}

:deep(.xterm-viewport) {
  overflow-y: auto !important;
}

:deep(.xterm-viewport::-webkit-scrollbar) {
  width: 8px;
}

:deep(.xterm-viewport::-webkit-scrollbar-track) {
  background: transparent;
}

:deep(.xterm-viewport::-webkit-scrollbar-thumb) {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 4px;
}

:deep(.xterm-viewport::-webkit-scrollbar-thumb:hover) {
  background: rgba(255, 255, 255, 0.3);
}
</style>
