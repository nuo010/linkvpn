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
  background:
    radial-gradient(circle at top left, rgba(96, 165, 250, 0.08), transparent 24%),
    linear-gradient(180deg, #f8fbff 0%, #f4f7fb 100%);
}
.sidebar {
  width: var(--sidebar-width);
  background: rgba(255, 255, 255, 0.88);
  border-right: 1px solid rgba(226, 232, 240, 0.95);
  backdrop-filter: blur(14px);
  box-shadow: 10px 0 30px rgba(15, 23, 42, 0.04);
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
  padding: 0.75rem 1rem;
  border-bottom: 1px solid rgba(226, 232, 240, 0.9);
  gap: 0.65rem;
  flex-shrink: 0;
}
.logo {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  object-fit: contain;
  background: #fff;
  border: 1px solid rgba(148, 163, 184, 0.25);
  box-shadow: 0 8px 20px rgba(37, 99, 235, 0.08);
  padding: 3px;
  flex-shrink: 0;
}
.logo-title {
  font-weight: 700;
  font-size: 0.98rem;
  line-height: 1.35;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  color: #0f172a;
}
.nav {
  flex: 1;
  padding: 0.8rem 0.55rem;
}
.nav-item {
  display: flex;
  align-items: center;
  gap: 0.7rem;
  min-height: 44px;
  padding: 0.68rem 0.9rem;
  margin: 0 0 0.22rem;
  border-radius: 14px;
  color: #334155;
  text-decoration: none;
  border: none;
  background: none;
  width: 100%;
  cursor: pointer;
  font-size: 0.93rem;
  font-weight: 600;
  text-align: left;
  transition: background 0.2s ease, color 0.2s ease, transform 0.2s ease, box-shadow 0.2s ease;
}
.nav-item:hover {
  background: rgba(239, 246, 255, 0.95);
  color: #2563eb;
  transform: translateX(1px);
}
.nav-item.router-link-active {
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
  color: #2563eb;
  box-shadow: inset 0 0 0 1px rgba(147, 197, 253, 0.65);
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
  transition: opacity 0.2s ease, transform 0.2s ease;
}
.nav-item.router-link-active .nav-icon-img {
  opacity: 1;
  filter: none;
  transform: scale(1.03);
}
.nav-item:not(.router-link-active) .nav-icon-img {
  opacity: 0.82;
}
.nav-text {
  white-space: nowrap;
  overflow: hidden;
  flex: 1;
}
.logout {
  margin-top: auto;
  color: #475569;
}
.collapse-btn {
  margin: 0.65rem;
  padding: 0.48rem;
  font-size: 1rem;
  border-radius: 12px;
  border: 1px solid rgba(203, 213, 225, 0.9);
  background: rgba(255, 255, 255, 0.94);
  color: #64748b;
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.05);
}
.main-wrap {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}
.header {
  height: var(--header-height);
  background: rgba(255, 255, 255, 0.78);
  border-bottom: 1px solid rgba(226, 232, 240, 0.92);
  backdrop-filter: blur(12px);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 1.4rem 0 1.5rem;
  box-shadow: 0 6px 24px rgba(15, 23, 42, 0.03);
}
.header-title {
  font-size: 1.14rem;
  font-weight: 700;
  color: #0f172a;
}
.greeting {
  display: inline-flex;
  align-items: center;
  min-height: 34px;
  padding: 0 0.85rem;
  border-radius: 999px;
  color: #64748b;
  font-size: 0.86rem;
  font-weight: 600;
  background: rgba(248, 250, 252, 0.95);
  border: 1px solid rgba(226, 232, 240, 0.9);
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
  background:
    radial-gradient(circle at top, rgba(59, 130, 246, 0.16), transparent 34%),
    rgba(15, 23, 42, 0.52);
  backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
}
.initial-setup-modal {
  background: linear-gradient(180deg, #ffffff 0%, #f8fbff 100%);
  border-radius: 18px;
  border: 1px solid rgba(226, 232, 240, 0.92);
  padding: 1.5rem;
  max-width: 420px;
  width: 90%;
  box-shadow: 0 28px 64px rgba(15, 23, 42, 0.24);
}
.initial-setup-title {
  margin: 0 0 0.5rem;
  font-size: 1.18rem;
  color: #0f172a;
}
.initial-setup-desc {
  margin: 0 0 1rem;
  color: #64748b;
  font-size: 0.92rem;
  line-height: 1.6;
}
.initial-setup-modal .form-row {
  margin-bottom: 0.85rem;
}
.initial-setup-modal .form-row label {
  display: block;
  margin-bottom: 0.38rem;
  font-size: 0.9rem;
  font-weight: 600;
  color: #334155;
}
.initial-setup-modal .form-row input {
  width: 100%;
  min-height: 40px;
  padding: 0 0.8rem;
  border: 1px solid #dbe5f1;
  border-radius: 10px;
  background: #fff;
  box-sizing: border-box;
}
.initial-setup-modal .form-row input:focus {
  outline: none;
  border-color: #60a5fa;
  box-shadow: 0 0 0 3px rgba(96, 165, 250, 0.14);
}
.initial-setup-error {
  margin: 0 0 0.75rem;
  color: #dc2626;
  font-size: 0.9rem;
}
.initial-setup-modal .form-actions {
  margin-top: 1rem;
}
@media (max-width: 820px) {
  .header {
    padding: 0 1rem;
  }
  .main {
    padding: 1rem;
  }
}
</style>
