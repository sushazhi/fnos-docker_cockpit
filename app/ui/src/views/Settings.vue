<template>
  <div class="page">
    <div class="header">
      <span class="header-title">{{ t('settings.title') }}</span>
    </div>

    <div class="settings-content">
      <div class="setting-group">
        <div class="group-header">
          <div class="group-icon appearance">
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <circle cx="12" cy="12" r="5" />
              <line x1="12" y1="1" x2="12" y2="3" />
              <line x1="12" y1="21" x2="12" y2="23" />
              <line x1="4.22" y1="4.22" x2="5.64" y2="5.64" />
              <line x1="18.36" y1="18.36" x2="19.78" y2="19.78" />
              <line x1="1" y1="12" x2="3" y2="12" />
              <line x1="21" y1="12" x2="23" y2="12" />
              <line x1="4.22" y1="19.78" x2="5.64" y2="18.36" />
              <line x1="18.36" y1="5.64" x2="19.78" y2="4.22" />
            </svg>
          </div>
          <div class="group-info">
            <div class="group-title">{{ t('appearance.title') }}</div>
            <div class="group-desc">{{ t('appearance.desc') }}</div>
          </div>
        </div>

        <div class="setting-card">
          <div class="setting-item" @click="toggleTheme">
            <div class="item-left">
              <div class="item-icon theme-icon">
                <svg
                  v-if="theme === 'light'"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                >
                  <circle cx="12" cy="12" r="5" />
                  <line x1="12" y1="1" x2="12" y2="3" />
                  <line x1="12" y1="21" x2="12" y2="23" />
                  <line x1="4.22" y1="4.22" x2="5.64" y2="5.64" />
                  <line x1="18.36" y1="18.36" x2="19.78" y2="19.78" />
                  <line x1="1" y1="12" x2="3" y2="12" />
                  <line x1="21" y1="12" x2="23" y2="12" />
                  <line x1="4.22" y1="19.78" x2="5.64" y2="18.36" />
                  <line x1="18.36" y1="5.64" x2="19.78" y2="4.22" />
                </svg>
                <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
                </svg>
              </div>
              <div class="item-text">
                <div class="item-title">{{ t('appearance.darkMode') }}</div>
                <div class="item-subtitle">
                  {{ theme === 'dark' ? t('appearance.enabled') : t('appearance.disabled') }}
                </div>
              </div>
            </div>
            <div class="toggle-switch" :class="{ active: theme === 'dark' }">
              <div class="toggle-thumb"></div>
            </div>
          </div>
        </div>
      </div>

      <div class="setting-group">
        <div class="group-header">
          <div class="group-icon features">
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <rect x="3" y="3" width="7" height="7" />
              <rect x="14" y="3" width="7" height="7" />
              <rect x="14" y="14" width="7" height="7" />
              <rect x="3" y="14" width="7" height="7" />
            </svg>
          </div>
          <div class="group-info">
            <div class="group-title">{{ t('settings.features') }}</div>
            <div class="group-desc">{{ t('settings.featuresDesc') }}</div>
          </div>
        </div>

        <div class="setting-card">
          <div class="setting-item" @click="$router.push('/networks')">
            <div class="item-left">
              <div class="item-icon network-icon">
                <svg
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                >
                  <circle cx="12" cy="12" r="10" />
                  <line x1="2" y1="12" x2="22" y2="12" />
                  <path
                    d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"
                  />
                </svg>
              </div>
              <div class="item-text">
                <div class="item-title">{{ t('settings.networkManage') }}</div>
                <div class="item-subtitle">{{ t('settings.dockerNetworkConfig') }}</div>
              </div>
            </div>
            <div class="item-arrow">
              <svg
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <polyline points="9 18 15 12 9 6" />
              </svg>
            </div>
          </div>

          <div class="setting-item" @click="$router.push('/system')">
            <div class="item-left">
              <div class="item-icon system-icon">
                <svg
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                >
                  <circle cx="12" cy="12" r="10" />
                  <line x1="12" y1="16" x2="12" y2="12" />
                  <line x1="12" y1="8" x2="12.01" y2="8" />
                </svg>
              </div>
              <div class="item-text">
                <div class="item-title">{{ t('settings.systemInfo') }}</div>
                <div class="item-subtitle">{{ t('settings.dockerVersionStatus') }}</div>
              </div>
            </div>
            <div class="item-arrow">
              <svg
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <polyline points="9 18 15 12 9 6" />
              </svg>
            </div>
          </div>

          <div class="setting-item" @click="$router.push('/volumes')">
            <div class="item-left">
              <div class="item-icon volume-icon">
                <svg
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                >
                  <ellipse cx="12" cy="5" rx="9" ry="3" />
                  <path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3" />
                  <path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5" />
                </svg>
              </div>
              <div class="item-text">
                <div class="item-title">{{ t('nav.volumes') }}</div>
                <div class="item-subtitle">{{ t('volumes.list') }}</div>
              </div>
            </div>
            <div class="item-arrow">
              <svg
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <polyline points="9 18 15 12 9 6" />
              </svg>
            </div>
          </div>

          <div class="setting-item" @click="showRegistryModal = true">
            <div class="item-left">
              <div class="item-icon registry-icon">
                <svg
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                >
                  <circle cx="12" cy="12" r="10" />
                  <circle cx="12" cy="12" r="4" />
                  <line x1="4.93" y1="4.93" x2="9.17" y2="9.17" />
                  <line x1="14.83" y1="14.83" x2="19.07" y2="19.07" />
                  <line x1="14.83" y1="9.17" x2="19.07" y2="4.93" />
                  <line x1="14.83" y1="9.17" x2="18.36" y2="5.64" />
                  <line x1="4.93" y1="19.07" x2="9.17" y2="14.83" />
                </svg>
              </div>
              <div class="item-text">
                <div class="item-title">镜像加速源</div>
                <div class="item-subtitle">配置 Docker 镜像仓库加速地址</div>
              </div>
            </div>
            <div class="item-arrow">
              <svg
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <polyline points="9 18 15 12 9 6" />
              </svg>
            </div>
          </div>

          <div v-if="searchAvailable" class="setting-item" @click="showSearchModal = true">
            <div class="item-left">
              <div class="item-icon search-icon">
                <svg
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                >
                  <circle cx="11" cy="11" r="8" />
                  <line x1="21" y1="21" x2="16.65" y2="16.65" />
                </svg>
              </div>
              <div class="item-text">
                <div class="item-title">搜索镜像</div>
                <div class="item-subtitle">从 Docker Hub 搜索并拉取镜像</div>
              </div>
            </div>
            <div class="item-arrow">
              <svg
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <polyline points="9 18 15 12 9 6" />
              </svg>
            </div>
          </div>
        </div>
      </div>

      <!-- 只有在搜索功能可用时才显示 -->
      <div v-if="searchAvailable" class="setting-group"></div>

      <div class="setting-group">
        <div class="group-header">
          <div class="group-icon account">
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
              <path d="M7 11V7a5 5 0 0 1 10 0v4" />
            </svg>
          </div>
          <div class="group-info">
            <div class="group-title">{{ t('settings.account') }}</div>
            <div class="group-desc">{{ t('settings.accountDesc') }}</div>
          </div>
        </div>

        <div class="setting-card">
          <div class="setting-item" @click="showChangePassword = true">
            <div class="item-left">
              <div class="item-icon password-icon">
                <svg
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                >
                  <path
                    d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4"
                  />
                </svg>
              </div>
              <div class="item-text">
                <div class="item-title">{{ t('settings.changePassword') }}</div>
                <div class="item-subtitle">{{ t('settings.changePasswordDesc') }}</div>
              </div>
            </div>
            <div class="item-arrow">
              <svg
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <polyline points="9 18 15 12 9 6" />
              </svg>
            </div>
          </div>

          <div class="setting-item" @click="loadAuditLogs">
            <div class="item-left">
              <div class="item-icon audit-icon">
                <svg
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                >
                  <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
                  <polyline points="14 2 14 8 20 8" />
                  <line x1="16" y1="13" x2="8" y2="13" />
                  <line x1="16" y1="17" x2="8" y2="17" />
                  <polyline points="10 9 9 9 8 9" />
                </svg>
              </div>
              <div class="item-text">
                <div class="item-title">{{ t('settings.auditLogs') }}</div>
                <div class="item-subtitle">{{ t('settings.auditLogsDesc') }}</div>
              </div>
            </div>
            <div class="item-arrow">
              <svg
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <polyline points="9 18 15 12 9 6" />
              </svg>
            </div>
          </div>
        </div>
      </div>

      <div class="setting-group">
        <div class="group-header">
          <div class="group-icon about">
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <circle cx="12" cy="12" r="10" />
              <line x1="12" y1="16" x2="12" y2="12" />
              <line x1="12" y1="8" x2="12.01" y2="8" />
            </svg>
          </div>
          <div class="group-info">
            <div class="group-title">{{ t('settings.about') }}</div>
            <div class="group-desc">{{ t('settings.aboutDesc') }}</div>
          </div>
        </div>

        <div class="setting-card">
          <div class="setting-item-static">
            <div class="static-label">{{ t('settings.version') }}</div>
            <div class="static-value">{{ appInfo.version || '1.0.0' }}</div>
          </div>
          <div class="setting-item-static">
            <div class="static-label">Docker SDK</div>
            <div class="static-value">{{ appInfo.dockerSDKVersion || '-' }}</div>
          </div>
          <div class="setting-item-static">
            <div class="static-label">{{ t('settings.author') }}</div>
            <a href="https://github.com/sushazhi" target="_blank" class="static-link">
              <span>yukihana</span>
              <svg
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
                <polyline points="15 3 21 3 21 9" />
                <line x1="10" y1="14" x2="21" y2="3" />
              </svg>
            </a>
          </div>
          <div class="setting-item-static">
            <div class="static-label">{{ t('settings.repository') }}</div>
            <a
              href="https://github.com/sushazhi/fnos-docker_cockpit"
              target="_blank"
              class="static-link"
            >
              <span>GitHub</span>
              <svg
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
                <polyline points="15 3 21 3 21 9" />
                <line x1="10" y1="14" x2="21" y2="3" />
              </svg>
            </a>
          </div>
        </div>
      </div>

      <div class="logout-section">
        <button class="logout-btn" @click="logout">
          <svg
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
            <polyline points="16 17 21 12 16 7" />
            <line x1="21" y1="12" x2="9" y2="12" />
          </svg>
          <span>{{ t('settings.signOut') }}</span>
        </button>
      </div>
    </div>

    <div v-if="showChangePassword" class="dialog-overlay" @click.self="showChangePassword = false">
      <div class="dialog">
        <div class="dialog-header">
          <h3 class="dialog-title">{{ t('settings.changePassword') }}</h3>
          <button class="dialog-close" @click="showChangePassword = false">
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>
        <div class="dialog-body">
          <div class="form-field">
            <label class="form-label">{{ t('settings.currentPassword') }}</label>
            <input type="password" class="form-input" v-model="passwordForm.currentPassword" />
          </div>
          <div class="form-field">
            <label class="form-label">{{ t('settings.newPassword') }}</label>
            <input
              type="password"
              class="form-input"
              v-model="passwordForm.newPassword"
              :placeholder="t('settings.passwordHint')"
            />
          </div>
          <div class="form-field">
            <label class="form-label">{{ t('settings.confirmPassword') }}</label>
            <input type="password" class="form-input" v-model="passwordForm.confirmPassword" />
          </div>
        </div>
        <div class="dialog-footer">
          <button class="dialog-btn secondary" @click="showChangePassword = false">
            {{ t('settings.cancel') }}
          </button>
          <button class="dialog-btn primary" @click="changePassword" :disabled="changing">
            {{ changing ? t('settings.changing') : t('settings.confirm') }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="showAuditLogs" class="dialog-overlay" @click.self="showAuditLogs = false">
      <div class="dialog dialog-large">
        <div class="dialog-header">
          <h3 class="dialog-title">{{ t('settings.auditLogs') }}</h3>
          <button class="dialog-close" @click="showAuditLogs = false">
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>
        <div class="dialog-body audit-body">
          <div v-for="log in auditLogs" :key="log.id" class="audit-item">
            <div class="audit-header">
              <span class="audit-type" :class="getAuditTypeClass(log.action)">
                {{ getAuditTypeText(log.action) }}
              </span>
              <span class="audit-time">{{ formatTime(log.timestamp) }}</span>
            </div>
            <div class="audit-action">
              {{ getAuditActionText(log.action, log.details, log.clientIp) }}
            </div>
          </div>
          <div v-if="auditLogs.length === 0" class="empty-audit">{{ t('settings.noRecords') }}</div>
        </div>
        <div class="dialog-footer">
          <button class="dialog-btn primary" @click="showAuditLogs = false">
            {{ t('settings.close') }}
          </button>
        </div>
      </div>
    </div>

    <!-- 镜像加速源设置弹窗 -->
    <div v-if="showRegistryModal" class="dialog-overlay" @click.self="showRegistryModal = false">
      <div class="dialog dialog-large">
        <div class="dialog-header">
          <h3 class="dialog-title">镜像加速源设置</h3>
          <button class="dialog-close" @click="showRegistryModal = false">
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>
        <div class="dialog-body">
          <div class="form-field">
            <label class="form-label">镜像加速地址</label>
            <input
              type="text"
              class="form-input"
              v-model="registryUrl"
              placeholder="https://docker.1ms.run"
            />
            <div class="form-hint">
              推荐使用：1ms.run（毫秒）、轩辕镜像、1Panel、DaoCloud
            </div>
          </div>
          <div class="registry-presets">
            <div class="preset-label">快速选择：</div>
            <div class="preset-list">
              <button class="preset-btn" @click="registryUrl = 'https://docker.1ms.run'">
                毫秒 (1ms.run)
              </button>
              <button class="preset-btn" @click="registryUrl = 'https://docker.xuanyuan.me'">
                轩辕镜像
              </button>
              <button class="preset-btn" @click="registryUrl = 'https://docker.1panel.live'">
                1Panel
              </button>
              <button class="preset-btn" @click="registryUrl = 'https://docker.m.daocloud.io'">
                DaoCloud
              </button>
            </div>
          </div>
        </div>
        <div class="dialog-footer">
          <button class="dialog-btn secondary" @click="showRegistryModal = false">取消</button>
          <button class="dialog-btn primary" @click="saveRegistry" :disabled="savingRegistry">
            {{ savingRegistry ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 搜索镜像弹窗 -->
    <div v-if="showSearchModal" class="dialog-overlay" @click.self="closeSearchModal">
      <div class="dialog dialog-large">
        <div class="dialog-header">
          <h3 class="dialog-title">搜索镜像</h3>
          <button class="dialog-close" @click="closeSearchModal">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>
        <div class="dialog-body">
          <div class="form-field">
            <label class="form-label">搜索关键词</label>
            <div class="search-input-wrapper">
              <input
                type="text"
                class="form-input"
                v-model="searchKeyword"
                placeholder="nginx"
                @keyup.enter="doSearch"
                @input="onSearchInput"
              />
              <button class="search-btn" @click="doSearch" :disabled="searching">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <circle cx="11" cy="11" r="8" />
                  <line x1="21" y1="21" x2="16.65" y2="16.65" />
                </svg>
              </button>
            </div>
          </div>

          <!-- 搜索中 -->
          <div v-if="searching" class="search-loading">
            <div class="spinner"></div>
            <span>搜索中...</span>
          </div>

          <!-- 搜索结果 -->
          <div v-else-if="searchResults.length > 0" class="search-results">
            <div class="search-results-title">搜索结果</div>
            <div class="search-results-list">
              <div v-for="result in searchResults" :key="result.name" class="search-result-item">
                <div class="result-icon">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
                    <circle cx="8.5" cy="8.5" r="1.5" />
                    <polyline points="21 15 16 10 5 21" />
                  </svg>
                </div>
                <div class="result-content">
                  <div class="result-name">{{ result.name }}</div>
                  <div class="result-description">{{ result.description || '暂无描述' }}</div>
                  <div class="result-meta">
                    <span class="result-stars">⭐ {{ result.star_count || 0 }}</span>
                    <span v-if="result.is_official" class="result-official">官方</span>
                  </div>
                </div>
                <button class="result-download-btn" @click="openDownloadModal(result.name)">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                    <polyline points="7 10 12 15 17 10" />
                    <line x1="12" y1="15" x2="12" y2="3" />
                  </svg>
                </button>
              </div>
            </div>
          </div>

          <!-- 无结果 -->
          <div
            v-else-if="hasSearched && !searching && searchResults.length === 0"
            class="search-empty"
          >
            <div class="empty-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="11" cy="11" r="8" />
                <line x1="21" y1="21" x2="16.65" y2="16.65" />
              </svg>
            </div>
            <div class="empty-text">未找到相关镜像</div>
            <div class="empty-hint">
              <p>可能的原因：</p>
              <ul>
                <li>网络连接问题，无法访问Docker Hub</li>
                <li>镜像加速源不支持搜索功能</li>
                <li>搜索关键词不准确</li>
              </ul>
              <p class="hint-tip">
                💡
                提示：您可以直接在"镜像管理"页面手动拉取镜像，输入完整的镜像名称（如：nginx:latest）
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- 下载镜像弹窗 -->
  <div v-if="showDownloadModal" class="dialog-overlay" @click.self="showDownloadModal = false">
    <div class="dialog">
      <div class="dialog-header">
        <h3 class="dialog-title">下载镜像</h3>
        <button class="dialog-close" @click="showDownloadModal = false">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18" />
            <line x1="6" y1="6" x2="18" y2="18" />
          </svg>
        </button>
      </div>
      <div class="dialog-body">
        <div class="form-field">
          <label class="form-label">镜像名称</label>
          <input type="text" class="form-input" :value="selectedImage" disabled />
        </div>
        <div class="form-field">
          <label class="form-label">标签</label>
          <div v-if="loadingTags" class="tags-loading">
            <div class="spinner-small"></div>
            <span>加载标签中...</span>
          </div>
          <select v-else-if="availableTags.length > 0" class="form-input" v-model="downloadTag">
            <option v-for="tag in availableTags" :key="tag" :value="tag">{{ tag }}</option>
          </select>
          <input
            v-else
            type="text"
            class="form-input"
            v-model="downloadTag"
            placeholder="latest"
            @keyup.enter="downloadImage"
          />
          <div class="form-hint" v-if="availableTags.length > 0">
            共 {{ availableTags.length }} 个标签可选
          </div>
        </div>
      </div>
      <div class="dialog-footer">
        <button class="dialog-btn secondary" @click="showDownloadModal = false">取消</button>
        <button class="dialog-btn primary" @click="downloadImage" :disabled="downloading">
          {{ downloading ? '下载中...' : '下载' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, inject, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import api from '../services/api'

const { t } = useI18n()
const router = useRouter()
const showToast = inject('showToast')
const theme = inject('theme')
const toggleTheme = inject('toggleTheme')

const showChangePassword = ref(false)
const showAuditLogs = ref(false)
const showRegistryModal = ref(false)
const showSearchModal = ref(false)
const showDownloadModal = ref(false)
const auditLogs = ref([])
const changing = ref(false)
const savingRegistry = ref(false)
const registryUrl = ref('')
const searchKeyword = ref('')
const searching = ref(false)
const searchResults = ref([])
const hasSearched = ref(false)
const searchAvailable = ref(false)
const selectedImage = ref('')
const downloadTag = ref('latest')
const downloading = ref(false)
const availableTags = ref([])
const loadingTags = ref(false)
let searchTimeout = null
const appInfo = ref({
  version: '',
  dockerSDKVersion: '',
  goVersion: '',
  platform: ''
})
const passwordForm = ref({
  currentPassword: '',
  newPassword: '',
  confirmPassword: ''
})

// 检查搜索功能是否可用
async function checkSearchAvailable() {
  // 获取镜像加速源
  let registry = localStorage.getItem('docker_registry_mirror') || ''
  registry = registry.replace(/[`'"]/g, '').trim()

  // 如果配置了加速源，直接认为搜索可用
  if (registry) {
    searchAvailable.value = true
    return
  }

  // 没有配置加速源，检查 Docker Hub 是否可连接
  try {
    const data = await api.get('/api/image/search/check')
    searchAvailable.value = data.available || false
  } catch (e) {
    console.error('检查搜索功能失败:', e)
    searchAvailable.value = false
  }
}

// 加载应用信息
async function loadAppInfo() {
  try {
    const data = await api.get('/api/system/app-info')
    appInfo.value = data
  } catch (e) {
    console.error('Failed to load app info:', e)
  }
}

// 加载保存的镜像加速源
onMounted(() => {
  loadAppInfo()
  checkSearchAvailable()
  const saved = localStorage.getItem('docker_registry_mirror')
  if (saved) {
    registryUrl.value = saved
  }
})

// 保存镜像加速源
async function saveRegistry() {
  savingRegistry.value = true
  try {
    // 验证 URL 格式
    if (registryUrl.value && !registryUrl.value.startsWith('http')) {
      showToast(t('settings.invalidUrl'))
      return
    }

    // 保存到本地存储（清理引号）
    if (registryUrl.value) {
      let cleanUrl = registryUrl.value.replace(/[`'"]/g, '').trim()
      localStorage.setItem('docker_registry_mirror', cleanUrl)
    } else {
      localStorage.removeItem('docker_registry_mirror')
    }

    showToast(t('settings.registrySaved'))
    showRegistryModal.value = false
  } finally {
    savingRegistry.value = false
  }
}

function formatTime(timestamp) {
  return new Date(timestamp).toLocaleString()
}

// 审计日志类型映射
const auditTypeMap = {
  login_success: { type: 'auth', text: '登录成功' },
  login_failed: { type: 'auth', text: '登录失败' },
  login_locked: { type: 'auth', text: '账户锁定' },
  logout: { type: 'auth', text: '登出' },
  password_change: { type: 'auth', text: '密码修改' },
  container_start: { type: 'container', text: '容器启动' },
  container_stop: { type: 'container', text: '容器停止' },
  container_restart: { type: 'container', text: '容器重启' },
  container_remove: { type: 'container', text: '容器删除' },
  container_create: { type: 'container', text: '容器创建' },
  image_pull: { type: 'image', text: '镜像拉取' },
  image_remove: { type: 'image', text: '镜像删除' },
  image_push: { type: 'image', text: '镜像推送' },
  network_create: { type: 'network', text: '网络创建' },
  network_remove: { type: 'network', text: '网络删除' },
  volume_create: { type: 'volume', text: '存储卷创建' },
  volume_remove: { type: 'volume', text: '存储卷删除' },
  compose_up: { type: 'compose', text: 'Compose启动' },
  compose_down: { type: 'compose', text: 'Compose停止' },
  compose_delete: { type: 'compose', text: 'Compose删除' },
  file_write: { type: 'file', text: '文件写入' },
  file_delete: { type: 'file', text: '文件删除' }
}

function getAuditTypeText(action) {
  return auditTypeMap[action]?.text || action
}

function getAuditTypeClass(action) {
  const type = auditTypeMap[action]?.type || 'other'
  return `type-${type}`
}

function getAuditActionText(action, details, clientIp) {
  const info = auditTypeMap[action]
  if (!info) return action

  // 根据不同类型添加详细信息
  switch (action) {
    case 'login_success':
      return clientIp ? `${info.text}，IP: ${clientIp}` : info.text
    case 'login_failed':
      const remaining = details?.remaining ? `（剩余 ${details.remaining} 次）` : ''
      return clientIp ? `${info.text}${remaining}，IP: ${clientIp}` : `${info.text}${remaining}`
    case 'login_locked':
      return clientIp ? `${info.text}，IP: ${clientIp}` : info.text
    case 'logout':
      return info.text
    case 'password_change':
      return info.text
    case 'container_start':
    case 'container_stop':
    case 'container_restart':
    case 'container_remove':
    case 'container_create':
      return details?.name ? `${info.text}: ${details.name}` : info.text
    case 'image_pull':
    case 'image_remove':
    case 'image_push':
      return details?.image ? `${info.text}: ${details.image}` : info.text
    case 'network_create':
    case 'network_remove':
      return details?.name ? `${info.text}: ${details.name}` : info.text
    case 'volume_create':
    case 'volume_remove':
      return details?.name ? `${info.text}: ${details.name}` : info.text
    case 'compose_up':
    case 'compose_down':
    case 'compose_delete':
      return details?.name ? `${info.text}: ${details.name}` : info.text
    case 'file_write':
    case 'file_delete':
      return details?.path ? `${info.text}: ${details.path}` : info.text
    default:
      return info.text
  }
}

async function changePassword() {
  if (!passwordForm.value.currentPassword) {
    showToast(t('settings.enterCurrentPassword'))
    return
  }
  if (passwordForm.value.newPassword.length < 8) {
    showToast(t('settings.passwordMinLength'))
    return
  }
  if (passwordForm.value.newPassword !== passwordForm.value.confirmPassword) {
    showToast(t('settings.passwordMismatch'))
    return
  }

  changing.value = true
  try {
    await api.post('/api/change-password', {
      currentPassword: passwordForm.value.currentPassword,
      newPassword: passwordForm.value.newPassword
    })
    showToast(t('settings.passwordChanged'))
    showChangePassword.value = false
    passwordForm.value = { currentPassword: '', newPassword: '', confirmPassword: '' }
  } catch (e) {
    showToast(t('settings.passwordFailed') + ': ' + e.message)
  } finally {
    changing.value = false
  }
}

async function loadAuditLogs() {
  try {
    const data = await api.get('/api/audit')
    auditLogs.value = data.logs || []
    showAuditLogs.value = true
  } catch (e) {
    showToast(t('settings.loadFailed'))
  }
}

async function logout() {
  try {
    await api.post('/api/logout')
    api.clearCSRFToken()
    router.go(0)
  } catch (e) {
    router.go(0)
  }
}

// 搜索输入处理（防抖）
function onSearchInput() {
  if (searchTimeout) {
    clearTimeout(searchTimeout)
  }
  if (!searchKeyword.value.trim()) {
    searchResults.value = []
    hasSearched.value = false
    return
  }
  searchTimeout = setTimeout(() => {
    doSearch()
  }, 500)
}

// 执行搜索
async function doSearch() {
  if (!searchKeyword.value.trim() || searching.value) return

  hasSearched.value = true
  searching.value = true
  try {
    // 获取镜像加速源
    let registry = localStorage.getItem('docker_registry_mirror') || ''
    registry = registry.replace(/[`'"]/g, '').trim()

    const url = registry
      ? `/api/image/search?term=${encodeURIComponent(searchKeyword.value.trim())}&limit=10&registry=${encodeURIComponent(registry)}`
      : `/api/image/search?term=${encodeURIComponent(searchKeyword.value.trim())}&limit=10`
    const data = await api.get(url)
    searchResults.value = data.results || []
  } catch (e) {
    console.error('搜索失败:', e)
    searchResults.value = []
  } finally {
    searching.value = false
  }
}

// 打开下载弹窗
async function openDownloadModal(name) {
  selectedImage.value = name
  downloadTag.value = 'latest'
  availableTags.value = []
  showDownloadModal.value = true

  // 获取镜像标签
  loadingTags.value = true
  try {
    let registry = localStorage.getItem('docker_registry_mirror') || ''
    registry = registry.replace(/[`'"]/g, '').trim()

    const url = registry
      ? `/api/image/tags?image=${encodeURIComponent(name)}&registry=${encodeURIComponent(registry)}`
      : `/api/image/tags?image=${encodeURIComponent(name)}`
    const data = await api.get(url)
    availableTags.value = data.tags || []
  } catch (e) {
    console.error('获取标签失败:', e)
    availableTags.value = []
  } finally {
    loadingTags.value = false
  }
}

// 下载镜像
async function downloadImage() {
  if (!selectedImage.value || downloading.value) return
  downloading.value = true
  try {
    let registry = localStorage.getItem('docker_registry_mirror') || ''
    registry = registry.replace(/[`'"]/g, '').trim()
    await api.post('/api/image/pull', {
      image: selectedImage.value + ':' + downloadTag.value,
      registry: registry
    })
    showToast(t('settings.downloadStarted', { name: selectedImage.value }))
    showDownloadModal.value = false
    closeSearchModal()
  } catch (e) {
    showToast(t('settings.downloadFailed') + ': ' + e.message)
  } finally {
    downloading.value = false
  }
}

// 关闭搜索模态框
function closeSearchModal() {
  showSearchModal.value = false
  searchKeyword.value = ''
  searchResults.value = []
  hasSearched.value = false
}
</script>

<style scoped>
.page {
  min-height: 100vh;
  padding-bottom: calc(84px + var(--safe-area-bottom));
  background: var(--bg-color);
}

.header {
  position: sticky;
  top: 0;
  z-index: 100;
  background: var(--card-bg);
  border-bottom: 1px solid var(--border-color);
  padding: 0 16px;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.header-title {
  font-size: 17px;
  font-weight: 600;
  color: var(--text-color);
}

.settings-content {
  padding: 12px;
}

.setting-group {
  margin-bottom: 12px;
}

.group-header {
  display: flex;
  align-items: center;
  padding: 12px 8px;
  gap: 12px;
}

.group-icon {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
  flex-shrink: 0;
}

.group-icon svg {
  width: 20px;
  height: 20px;
}

.group-icon.appearance {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.group-icon.features {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  color: white;
}

.group-icon.account {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
  color: white;
}

.group-icon.about {
  background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
  color: white;
}

.group-info {
  flex: 1;
}

.group-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-color);
}

.group-desc {
  font-size: 13px;
  color: var(--text-secondary);
  margin-top: 1px;
}

.setting-card {
  background: var(--card-bg);
  border-radius: 16px;
  overflow: hidden;
  box-shadow: var(--shadow-sm);
}

[data-theme='dark'] .setting-card {
  box-shadow: none;
}

.setting-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
  cursor: pointer;
  transition: background var(--transition-fast);
}

.setting-item:hover {
  background: var(--hover-bg);
}

.setting-item:active {
  background: var(--active-bg);
}

.setting-item + .setting-item {
  border-top: 1px solid var(--border-color);
}

.item-left {
  display: flex;
  align-items: center;
  flex: 1;
  min-width: 0;
}

.item-icon {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 10px;
  margin-right: 12px;
  flex-shrink: 0;
}

.item-icon svg {
  width: 18px;
  height: 18px;
}

.item-icon.theme-icon {
  background: rgba(102, 126, 234, 0.12);
  color: #667eea;
}

.item-icon.network-icon {
  background: rgba(0, 125, 255, 0.1);
  color: #007dff;
}

.item-icon.system-icon {
  background: rgba(0, 200, 83, 0.1);
  color: #00c853;
}

.item-icon.compose-icon {
  background: rgba(255, 152, 0, 0.1);
  color: #ff9800;
}

.item-icon.volume-icon {
  background: rgba(255, 152, 0, 0.1);
  color: #ff9800;
}

.item-icon.password-icon {
  background: rgba(250, 42, 45, 0.1);
  color: #fa2a2d;
}

.item-icon.audit-icon {
  background: rgba(102, 126, 234, 0.1);
  color: #667eea;
}

.item-text {
  flex: 1;
  min-width: 0;
}

.item-title {
  font-size: 15px;
  font-weight: 500;
  color: var(--text-color);
}

.item-subtitle {
  font-size: 13px;
  color: var(--text-secondary);
  margin-top: 2px;
}

.item-arrow {
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-tertiary);
  flex-shrink: 0;
}

.item-arrow svg {
  width: 18px;
  height: 18px;
}

.toggle-switch {
  width: 50px;
  height: 30px;
  background: var(--border-color);
  border-radius: 15px;
  position: relative;
  transition: background var(--transition-normal);
  flex-shrink: 0;
}

.toggle-switch.active {
  background: #007dff;
}

.toggle-thumb {
  width: 26px;
  height: 26px;
  background: white;
  border-radius: 13px;
  position: absolute;
  top: 2px;
  left: 2px;
  transition: transform var(--transition-normal);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.15);
}

.toggle-switch.active .toggle-thumb {
  transform: translateX(20px);
}

.setting-item-static {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
}

.setting-item-static + .setting-item-static {
  border-top: 1px solid var(--border-color);
}

.static-label {
  font-size: 15px;
  color: var(--text-color);
}

.static-value {
  font-size: 14px;
  color: var(--text-secondary);
  font-family: 'HarmonyOS Sans SC', 'SF Mono', 'Consolas', monospace;
}

.static-link {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 14px;
  color: #007dff;
  text-decoration: none;
  padding: 4px 8px;
  margin: -4px -8px;
  border-radius: 8px;
  transition: background var(--transition-fast);
}

.static-link:hover {
  background: rgba(0, 125, 255, 0.08);
}

.static-link svg {
  width: 14px;
  height: 14px;
  opacity: 0.7;
}

.logout-section {
  margin-top: 16px;
  padding: 0 4px;
}

.logout-btn {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 14px;
  background: var(--card-bg);
  border: none;
  border-radius: 16px;
  color: #fa2a2d;
  font-size: 15px;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
  box-shadow: var(--shadow-sm);
}

[data-theme='dark'] .logout-btn {
  box-shadow: none;
}

.logout-btn:hover {
  background: rgba(250, 42, 45, 0.08);
}

.logout-btn:active {
  background: rgba(250, 42, 45, 0.12);
}

.logout-btn svg {
  width: 18px;
  height: 18px;
}

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
  animation: fadeIn var(--transition-fast);
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.dialog {
  background: var(--card-bg);
  border-radius: 20px;
  width: 100%;
  max-width: 420px;
  max-height: calc(100vh - 32px);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  box-shadow: var(--shadow-lg);
  animation: slideUp var(--transition-normal);
}

@media (min-width: 768px) {
  .dialog {
    max-width: 520px;
  }
}

@media (max-height: 600px) {
  .dialog {
    max-height: 95vh;
  }
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
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
  transition: all var(--transition-fast);
}

.dialog-close:hover {
  background: var(--hover-bg);
  color: var(--text-color);
}

.dialog-close svg {
  width: 20px;
  height: 20px;
}

.dialog-body {
  padding: 20px;
  overflow-y: auto;
  max-height: calc(100vh - 200px);
}

@media (max-height: 600px) {
  .dialog-body {
    max-height: calc(95vh - 150px);
  }
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid var(--border-color);
  flex-shrink: 0;
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
  transition:
    border-color var(--transition-fast),
    box-shadow var(--transition-fast);
}

.form-input:focus {
  outline: none;
  border-color: #007dff;
  box-shadow: 0 0 0 3px rgba(0, 125, 255, 0.12);
}

.form-input::placeholder {
  color: var(--text-tertiary);
}

.dialog-btn {
  padding: 10px 20px;
  border: none;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.dialog-btn.primary {
  background: #007dff;
  color: white;
}

.dialog-btn.primary:hover {
  background: #0066cc;
}

.dialog-btn.primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.dialog-btn.secondary {
  background: var(--hover-bg);
  color: var(--text-color);
}

.dialog-btn.secondary:hover {
  background: var(--active-bg);
}

.audit-body {
  max-height: 400px;
}

.audit-item {
  padding: 12px 0;
  border-bottom: 1px solid var(--border-color);
}

.audit-item:last-child {
  border-bottom: none;
}

.audit-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 6px;
}

.audit-type {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.audit-type.type-auth {
  background: rgba(102, 126, 234, 0.12);
  color: #667eea;
}

.audit-type.type-container {
  background: rgba(0, 125, 255, 0.1);
  color: #007dff;
}

.audit-type.type-image {
  background: rgba(255, 152, 0, 0.1);
  color: #ff9800;
}

.audit-type.type-network {
  background: rgba(0, 200, 83, 0.1);
  color: #00c853;
}

.audit-type.type-volume {
  background: rgba(156, 39, 176, 0.1);
  color: #9c27b0;
}

.audit-type.type-compose {
  background: rgba(233, 30, 99, 0.1);
  color: #e91e63;
}

.audit-type.type-file {
  background: rgba(96, 125, 139, 0.1);
  color: #607d8b;
}

.audit-type.type-other {
  background: var(--hover-bg);
  color: var(--text-secondary);
}

.audit-time {
  font-size: 12px;
  color: var(--text-tertiary);
}

.audit-action {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-color);
}

.empty-audit {
  text-align: center;
  padding: 40px 20px;
  color: var(--text-secondary);
  font-size: 14px;
}

.item-icon.registry-icon {
  background: rgba(0, 200, 83, 0.1);
  color: #00c853;
}

.item-icon.search-icon {
  background: rgba(102, 126, 234, 0.1);
  color: #667eea;
}

.registry-presets {
  margin-top: 16px;
}

.preset-label {
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 10px;
}

.preset-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.preset-btn {
  padding: 8px 14px;
  background: var(--hover-bg);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  color: var(--text-color);
  font-size: 13px;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.preset-btn:hover {
  background: #007dff;
  border-color: #007dff;
  color: white;
}

/* 搜索相关样式 */
.search-input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.search-input-wrapper .form-input {
  padding-right: 44px;
}

.search-btn {
  position: absolute;
  right: 4px;
  top: 50%;
  transform: translateY(-50%);
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #007dff;
  border: none;
  border-radius: 8px;
  color: white;
  cursor: pointer;
  z-index: 1;
}

.search-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.search-btn svg {
  width: 18px;
  height: 18px;
}

.search-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 20px;
  color: var(--text-secondary);
}

.search-results {
  margin-top: 16px;
  max-height: 300px;
  overflow-y: auto;
  position: relative;
  z-index: 10;
}

.search-results-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
  margin-bottom: 8px;
}

.search-results-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.search-result-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px;
  background: var(--hover-bg);
  border-radius: 12px;
  transition: background 0.2s;
}

.search-result-item:hover {
  background: var(--active-bg);
}

.result-icon {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 125, 255, 0.1);
  border-radius: 8px;
  color: #007dff;
  flex-shrink: 0;
}

.result-icon svg {
  width: 20px;
  height: 20px;
}

.result-content {
  flex: 1;
  min-width: 0;
}

.result-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-color);
  margin-bottom: 4px;
}

.result-description {
  font-size: 12px;
  color: var(--text-secondary);
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  margin-bottom: 6px;
}

.result-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 12px;
}

.result-stars {
  color: var(--text-secondary);
}

.result-official {
  padding: 2px 8px;
  background: #007dff;
  color: white;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
}

.result-download-btn {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #007dff;
  border: none;
  border-radius: 8px;
  color: white;
  cursor: pointer;
  flex-shrink: 0;
  transition: all 0.2s;
}

.result-download-btn:hover {
  background: #0066cc;
  transform: scale(1.05);
}

.result-download-btn svg {
  width: 18px;
  height: 18px;
}

.search-empty {
  text-align: center;
  padding: 40px 20px;
  color: var(--text-secondary);
  font-size: 14px;
}

.search-empty .empty-icon {
  margin-bottom: 16px;
  opacity: 0.5;
}

.search-empty .empty-icon svg {
  width: 48px;
  height: 48px;
}

.search-empty .empty-text {
  font-size: 16px;
  color: var(--text-color);
  margin-bottom: 16px;
}

.search-empty .empty-hint {
  text-align: left;
  background: var(--card-bg);
  padding: 16px;
  border-radius: 8px;
  margin-top: 16px;
  font-size: 13px;
  line-height: 1.6;
}

.search-empty .empty-hint p {
  margin: 0 0 8px 0;
  font-weight: 500;
  color: var(--text-color);
}

.search-empty .empty-hint ul {
  margin: 0 0 12px 0;
  padding-left: 20px;
  color: var(--text-secondary);
}

.search-empty .empty-hint li {
  margin-bottom: 4px;
}

.search-empty .empty-hint .hint-tip {
  color: #007dff;
  font-weight: 500;
  padding: 8px;
  background: rgba(0, 125, 255, 0.1);
  border-radius: 6px;
  margin: 0;
}

.spinner {
  width: 20px;
  height: 20px;
  border: 2px solid var(--border-color);
  border-top-color: #007dff;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.spinner-small {
  width: 16px;
  height: 16px;
  border: 2px solid var(--border-color);
  border-top-color: #007dff;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.tags-loading {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 14px;
  background: var(--input-bg);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  color: var(--text-secondary);
  font-size: 14px;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
