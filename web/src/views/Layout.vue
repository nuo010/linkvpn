<template>
  <div class="app-layout">
    <aside class="sidebar" :class="{ collapsed: sidebarCollapsed }">
      <div class="logo-wrap">
        <img class="logo" :src="logoUrl" alt="LinkVPN" />
        <span v-show="!sidebarCollapsed" class="logo-title">LinkVPN</span>
      </div>
      <nav class="nav">
        <router-link to="/home" class="nav-item">
          <img class="nav-icon-img" :src="iconHome" alt="" />
          <span v-show="!sidebarCollapsed" class="nav-text">首页</span>
        </router-link>
        <router-link to="/user" class="nav-item">
          <img class="nav-icon-img" :src="iconUserList" alt="" />
          <span v-show="!sidebarCollapsed" class="nav-text">用户管理</span>
        </router-link>
        <router-link to="/logs/login" class="nav-item">
           <img class="nav-icon-img" :src="iconLoginLog" alt="" />
           <span v-show="!sidebarCollapsed" class="nav-text">连接记录</span>
        </router-link>
        <router-link to="/logs/main" class="nav-item">
          <img class="nav-icon-img" :src="iconSysLog" alt="" />
          <span v-show="!sidebarCollapsed" class="nav-text">服务日志</span>
        </router-link>
        <router-link to="/logs/status" class="nav-item">
          <img class="nav-icon-img" :src="iconStatusLog" alt="" />
          <span v-show="!sidebarCollapsed" class="nav-text">状态日志</span>
        </router-link>
        <router-link to="/settings" class="nav-item" :class="{ 'router-link-active': route.path.startsWith('/settings') }">
           <img class="nav-icon-img" :src="iconSettings" alt="" />
           <span v-show="!sidebarCollapsed" class="nav-text">系统配置</span>
        </router-link>
        <button class="nav-item logout" @click="logout">
          <img class="nav-icon-img" :src="iconLogout" alt="" />
          <span v-show="!sidebarCollapsed" class="nav-text">注销登录</span>
        </button>
      </nav>
      <button class="collapse-btn" @click="sidebarCollapsed = !sidebarCollapsed" :title="sidebarCollapsed ? '展开' : '收起'">
        {{ sidebarCollapsed ? '»' : '«' }}
      </button>
    </aside>
    <div class="main-wrap">
      <header class="header">
        <span class="header-title">{{ currentTitle }}</span>
        <div class="header-right">
          <span class="greeting">您好，管理员</span>
        </div>
      </header>
      <main class="main">
        <router-view />
      </main>
    </div>

    <!-- 首次运行：必须设置客户端下载配置（服务器地址与端口），不设置无法使用 -->
    <Teleport to="body">
      <div v-if="showInitialSetupModal" class="initial-setup-overlay">
        <div class="initial-setup-modal" @click.stop>
          <h2 class="initial-setup-title">首次使用：请设置客户端连接信息</h2>
          <p class="initial-setup-desc">下载的 .ovpn 文件将使用下方「服务器地址」和「端口」，请填写后保存，否则无法继续使用系统。</p>
          <div class="form-row">
            <label>服务器地址</label>
            <input v-model="initialHost" type="text" placeholder="如 vpn.example.com 或公网 IP" />
          </div>
          <div class="form-row">
            <label>端口</label>
            <input v-model.number="initialPort" type="number" min="1" max="65535" placeholder="1194" />
          </div>
          <p v-if="initialSetupError" class="initial-setup-error">{{ initialSetupError }}</p>
          <div class="form-actions">
            <button class="primary" :disabled="initialSetupSaving" @click="saveInitialSetup">
              {{ initialSetupSaving ? '保存中…' : '保存并继续' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import client from '../api/client'
import logoUrl from '../logo.png'
// 侧栏菜单图标（accets 目录，按文件名对应菜单）
import iconHome from '../accets/首页.svg?url'
import iconUserList from '../accets/用户列表.svg?url'
import iconSettings from '../accets/系统配置.svg?url'
import iconLoginLog from '../accets/登录记录.svg?url'
import iconSysLog from '../accets/系统日志.svg?url'
import iconStatusLog from '../accets/状态日志.svg?url'
import iconLogout from '../accets/退出.svg?url'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const sidebarCollapsed = ref(false)

const showInitialSetupModal = ref(false)
const initialHost = ref('')
const initialPort = ref(1194)
const initialSetupSaving = ref(false)
const initialSetupError = ref('')

async function checkNeedInitialSetup() {
  try {
    const { data } = await client.get('/config/need-initial-setup')
    if (data && data.need) {
      showInitialSetupModal.value = true
      // 默认用当前访问地址栏中的主机名作为服务器地址，方便用户首填
      if (typeof window !== 'undefined' && window.location && window.location.hostname) {
        initialHost.value = window.location.hostname
      } else {
        initialHost.value = ''
      }
      initialPort.value = 1194
      initialSetupError.value = ''
    }
  } catch (_) {
    // 接口失败时不阻塞使用，仅不弹窗
  }
}

async function saveInitialSetup() {
  const host = (initialHost.value || '').trim()
  const port = Number(initialPort.value)
  if (!host) {
    initialSetupError.value = '请填写服务器地址'
    return
  }
  if (!Number.isInteger(port) || port < 1 || port > 65535) {
    initialSetupError.value = '请填写有效端口（1–65535）'
    return
  }
  initialSetupError.value = ''
  initialSetupSaving.value = true
  try {
    await client.post('/config', {
      client_remote_host: host,
      client_remote_port: String(port),
    })
    showInitialSetupModal.value = false
  } catch (e) {
    initialSetupError.value = e.response?.data?.error || '保存失败，请重试'
  } finally {
    initialSetupSaving.value = false
  }
}

const titleMap = {
  home: '首页',
  user: '用户管理',
  settingsConfig: '系统配置',
  settingsSystem: '系统配置',
  stats: '在线统计',
  logsLogin: '连接记录',
  logMain: '服务日志',
  logStatus: '连接日志',
}
const currentTitle = computed(() => titleMap[route.name] || 'LinkVPN')

function logout() {
  auth.logout()
  router.push('/login')
}

onMounted(() => {
  checkNeedInitialSetup()
})
</script>

<style scoped>
.app-layout {
  display: flex;
  min-height: 100vh;
}
.sidebar {
  width: var(--sidebar-width);
  background: var(--card);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  transition: width 0.2s;
  flex-shrink: 0;
}
.sidebar.collapsed {
  width: 64px;
}
.logo-wrap {
  min-height: var(--header-height);
  display: flex;
  align-items: center;
  padding: 0.5rem 1rem;
  border-bottom: 1px solid var(--border);
  gap: 0.5rem;
  flex-shrink: 0;
}
.logo {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  object-fit: contain;
  background: #fff;
  border: 1px solid rgba(148, 163, 184, 0.35);
  padding: 2px;
  flex-shrink: 0;
}
.logo-title {
  font-weight: 600;
  font-size: 0.95rem;
  line-height: 1.35;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.nav {
  flex: 1;
  padding: 0.5rem 0;
}
.nav-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.6rem 1rem;
  margin: 0 0.25rem;
  border-radius: var(--radius);
  color: var(--text);
  text-decoration: none;
  border: none;
  background: none;
  width: calc(100% - 0.5rem);
  cursor: pointer;
  font-size: 0.9rem;
  text-align: left;
}
.nav-item:hover {
  background: var(--bg);
  color: var(--accent);
}
.nav-item.router-link-active {
  background: #ecf5ff;
  color: var(--accent);
}
.nav-icon {
  font-size: 1.1rem;
  flex-shrink: 0;
  width: 1.5rem;
  text-align: center;
}
.nav-icon-img {
  width: 22px;
  height: 22px;
  flex-shrink: 0;
  object-fit: contain;
  display: block;
}
.nav-item.router-link-active .nav-icon-img {
  opacity: 1;
  filter: none;
}
.nav-item:not(.router-link-active) .nav-icon-img {
  opacity: 0.85;
}
.nav-text {
  white-space: nowrap;
  overflow: hidden;
  flex: 1;
}
.logout {
  margin-top: auto;
}
.collapse-btn {
  margin: 0.5rem;
  padding: 0.35rem;
  font-size: 1rem;
  border-radius: var(--radius);
}
.main-wrap {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}
.header {
  height: var(--header-height);
  background: var(--card);
  border-bottom: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 1.5rem;
}
.header-title {
  font-size: 1.1rem;
  font-weight: 600;
}
.greeting {
  color: var(--muted);
  font-size: 0.9rem;
}
.main {
  flex: 1;
  padding: 1.25rem;
  overflow: auto;
}

/* 首次设置弹窗：不可点击遮罩关闭，必须保存 */
.initial-setup-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
}
.initial-setup-modal {
  background: var(--card);
  border-radius: var(--radius);
  padding: 1.5rem;
  max-width: 420px;
  width: 90%;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
}
.initial-setup-title {
  margin: 0 0 0.5rem;
  font-size: 1.15rem;
}
.initial-setup-desc {
  margin: 0 0 1rem;
  color: var(--muted);
  font-size: 0.9rem;
  line-height: 1.4;
}
.initial-setup-modal .form-row {
  margin-bottom: 0.75rem;
}
.initial-setup-modal .form-row label {
  display: block;
  margin-bottom: 0.25rem;
  font-size: 0.9rem;
  color: var(--muted);
}
.initial-setup-modal .form-row input {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid var(--border);
  border-radius: var(--radius);
}
.initial-setup-error {
  margin: 0 0 0.75rem;
  color: var(--danger, #e74c3c);
  font-size: 0.9rem;
}
.initial-setup-modal .form-actions {
  margin-top: 1rem;
}
</style>
