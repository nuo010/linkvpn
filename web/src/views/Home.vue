<template>
  <div class="dashboard">
    <p v-if="loadError" class="load-error-bar">加载失败：{{ loadError }}，请检查网络或重新登录。</p>

    <div class="hero-card">
      <div class="hero-copy">
        <p class="hero-eyebrow">控制台总览</p>
        <h3 class="hero-title">当前 OpenVPN 服务状态一目了然</h3>
        <p class="hero-desc">
          首页会自动刷新运行状态、在线人数和当前连接流量，方便你先看整体健康度，再进入具体模块排查。
        </p>
      </div>
      <div class="hero-side">
        <span :class="['hero-status', stats.openvpn_running ? 'success' : 'danger']">
          {{ stats.openvpn_running ? '服务运行中' : '服务未运行' }}
        </span>
        <span class="hero-meta">自动刷新：每 10 秒</span>
      </div>
    </div>

    <div class="metrics-row">
      <div class="metric-card card">
        <h4 class="metric-title"><span class="metric-icon">⏱</span>系统运行时长</h4>
        <div class="metric-main">{{ uptimeStr }}</div>
        <div class="metric-sub">
          <span class="metric-sub-text">面板启动以来的运行时间</span>
        </div>
      </div>
      <div class="metric-card card">
        <h4 class="metric-title"><span class="metric-icon">👥</span>用户统计</h4>
        <div class="metric-main">{{ stats.user_count }}</div>
        <div class="metric-sub metric-sub-two">
          <span class="metric-sub-label">总用户</span>
          <span class="metric-sub-value">{{ stats.user_count }}</span>
          <span class="metric-sub-label">当前在线</span>
          <span class="metric-sub-value" :class="stats.online_count > 0 ? 'online' : ''">
            {{ stats.online_count }}
          </span>
        </div>
      </div>
      <div class="metric-card card">
        <h4 class="metric-title"><span class="metric-icon">📶</span>数据流量（当前连接）</h4>
        <div class="metric-main">
          {{ formatBytes(stats.total_bytes_recv + stats.total_bytes_sent) }}
        </div>
        <div class="metric-sub metric-sub-two">
          <span class="metric-sub-label">上传</span>
          <span class="metric-sub-value">{{ formatBytes(stats.total_bytes_recv) }}</span>
          <span class="metric-sub-label">下载</span>
          <span class="metric-sub-value">{{ formatBytes(stats.total_bytes_sent) }}</span>
        </div>
      </div>
      <div class="metric-card card">
        <h4 class="metric-title"><span class="metric-icon">🖥</span>系统信息</h4>
        <div class="metric-main">
          <span :class="['badge', stats.openvpn_running ? 'success' : 'danger']">
            {{ stats.openvpn_running ? 'OpenVPN 运行中' : 'OpenVPN 未运行' }}
          </span>
        </div>
        <div class="metric-sub metric-sub-stack">
          <span class="metric-sub-line">
            PKI 状态：
            <span :class="stats.pki_initialized ? 'ok' : 'warn'">
              {{ stats.pki_initialized ? '已初始化' : '未初始化' }}
            </span>
          </span>
          <span v-if="stats.openvpn_running && stats.openvpn_pid" class="metric-sub-line">
            进程 PID：{{ stats.openvpn_pid }}
          </span>
          <span v-if="stats.openvpn_version" class="metric-sub-line">
            版本：{{ stats.openvpn_version }}
          </span>
          <span class="metric-sub-line path-dd">工作目录：{{ stats.base_path || '-' }}</span>
        </div>
      </div>
    </div>

    <div class="two-col">
      <div class="card table-card">
        <h4 class="section-title">上传流量 Top 10（当前连接）</h4>
        <table class="data-table">
          <thead>
            <tr>
              <th>排名</th>
              <th>用户名</th>
              <th>上传流量</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(item, i) in stats.top10_upload" :key="'up-' + item.username">
              <td>{{ i + 1 }}</td>
              <td>{{ item.username }}</td>
              <td>{{ formatBytes(item.bytes_recv) }}</td>
            </tr>
          </tbody>
        </table>
        <div v-if="stats.top10_upload.length === 0" class="empty-state">
          <p class="empty-state-title">暂无上传流量数据</p>
          <p class="empty-state-desc">当前没有活跃连接，或者还没有形成可统计的上传流量。</p>
        </div>
      </div>
      <div class="card table-card">
        <h4 class="section-title">下载流量 Top 10（当前连接）</h4>
        <table class="data-table">
          <thead>
            <tr>
              <th>排名</th>
              <th>用户名</th>
              <th>下载流量</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(item, i) in stats.top10_download" :key="'down-' + item.username">
              <td>{{ i + 1 }}</td>
              <td>{{ item.username }}</td>
              <td>{{ formatBytes(item.bytes_sent) }}</td>
            </tr>
          </tbody>
        </table>
        <div v-if="stats.top10_download.length === 0" class="empty-state">
          <p class="empty-state-title">暂无下载流量数据</p>
          <p class="empty-state-desc">当前没有活跃连接，或者还没有形成可统计的下载流量。</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import client from '../api/client'

const defaultStats = () => ({
  uptime_seconds: 0,
  user_count: 0,
  online_count: 0,
  openvpn_running: false,
  openvpn_pid: 0,
  openvpn_version: '',
  total_bytes_recv: 0,
  total_bytes_sent: 0,
  top10_upload: [],
  top10_download: [],
  pki_initialized: false,
  base_path: '',
})
const stats = ref(defaultStats())
const loadError = ref('')

const uptimeStr = computed(() => {
  const s = stats.value.uptime_seconds || 0
  if (s < 60) return s + ' 秒'
  const m = Math.floor(s / 60)
  if (m < 60) return m + ' 分钟'
  const h = Math.floor(m / 60)
  if (h < 24) return h + ' 小时 ' + (m % 60) + ' 分钟'
  const d = Math.floor(h / 24)
  return d + ' 天 ' + (h % 24) + ' 小时'
})

function formatBytes(n) {
  if (n == null || n < 0) return '0 B'
  if (n < 1024) return n + ' B'
  if (n < 1024 * 1024) return (n / 1024).toFixed(2) + ' KB'
  return (n / (1024 * 1024)).toFixed(2) + ' MB'
}

async function load() {
  loadError.value = ''
  try {
    const { data } = await client.get('/home')
    stats.value = { ...defaultStats(), ...data }
    if (!Array.isArray(stats.value.top10_upload)) stats.value.top10_upload = []
    if (!Array.isArray(stats.value.top10_download)) stats.value.top10_download = []
  } catch (e) {
    const msg = e.response?.status === 401 ? '未登录或已过期' : (e.response?.data?.message || e.message || '网络错误')
    loadError.value = msg
  }
}

const REFRESH_INTERVAL = 10000
let timer = null
onMounted(() => {
  load()
  timer = setInterval(load, REFRESH_INTERVAL)
})
onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<style scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.load-error-bar {
  margin: 0;
  padding: 0.75rem 0.9rem;
  font-size: 0.86rem;
  color: #b91c1c;
  background: linear-gradient(180deg, #fff5f5 0%, #fef2f2 100%);
  border-radius: 12px;
  border: 1px solid #fecaca;
}
.hero-card {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
  padding: 1.15rem 1.2rem;
  border-radius: 18px;
  border: 1px solid #dbe7f3;
  background:
    radial-gradient(circle at top right, rgba(96, 165, 250, 0.14), transparent 32%),
    linear-gradient(135deg, #ffffff 0%, #f6fbff 100%);
  box-shadow: 0 16px 36px rgba(15, 23, 42, 0.06);
}
.hero-copy {
  max-width: 760px;
}
.hero-eyebrow {
  margin: 0 0 0.3rem;
  font-size: 0.8rem;
  letter-spacing: 0.08em;
  color: #3b82f6;
  font-weight: 700;
}
.hero-title {
  margin: 0;
  font-size: 1.28rem;
  color: #0f172a;
}
.hero-desc {
  margin: 0.45rem 0 0;
  line-height: 1.65;
  color: #64748b;
  font-size: 0.92rem;
}
.hero-side {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.55rem;
  min-width: 150px;
}
.hero-status {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 34px;
  padding: 0 0.9rem;
  border-radius: 999px;
  font-size: 0.86rem;
  font-weight: 700;
  border: 1px solid transparent;
}
.hero-status.success {
  color: #15803d;
  background: #ecfdf3;
  border-color: #bbf7d0;
}
.hero-status.danger {
  color: #b91c1c;
  background: #fff1f2;
  border-color: #fecdd3;
}
.hero-meta {
  color: #64748b;
  font-size: 0.82rem;
}
.metrics-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 0.9rem;
}
.metric-card {
  padding: 0.95rem 1rem;
  border-radius: 16px;
  box-shadow: 0 14px 30px rgba(15, 23, 42, 0.05);
  border: 1px solid rgba(203, 213, 225, 0.8);
  background: linear-gradient(180deg, #ffffff 0%, #f8fbff 100%);
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
}
.metric-title {
  margin: 0 0 0.45rem;
  font-size: 0.92rem;
  color: #1e293b;
  font-weight: 700;
  display: flex;
  align-items: center;
  gap: 0.4rem;
}
.metric-icon {
  font-size: 1rem;
}
.metric-main {
  font-size: 1.55rem;
  font-weight: 700;
  color: #2563eb;
  text-align: left;
  margin: 0.15rem 0 0.35rem;
}
.metric-sub {
  font-size: 0.83rem;
  color: #64748b;
}
.metric-sub-text {
  display: inline-block;
}
.metric-sub-two {
  display: grid;
  grid-template-columns: auto auto;
  column-gap: 0.4rem;
  row-gap: 0.15rem;
  align-items: center;
}
.metric-sub-label {
  color: var(--muted);
}
.metric-sub-value {
  font-weight: 600;
  color: #111827;
}
.metric-sub-value.online {
  color: #16a34a;
}
.metric-sub-stack {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}
.metric-sub-line {
  font-size: 0.82rem;
}
.metric-sub-line .ok {
  color: #16a34a;
}
.metric-sub-line .warn {
  color: #f97316;
}
.path-dd {
  word-break: break-all;
  font-size: 0.8rem;
}
.pid-text {
  margin-left: 0.25rem;
  font-size: 0.8rem;
  color: var(--muted);
}
.metrics-row .metric-card:nth-child(1) {
  border-top: 3px solid #409eff;
}
.metrics-row .metric-card:nth-child(2) {
  border-top: 3px solid #67c23a;
}
.metrics-row .metric-card:nth-child(3) {
  border-top: 3px solid #e6a23c;
}
.metrics-row .metric-card:nth-child(4) {
  border-top: 3px solid #909399;
}
.two-col {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.9rem;
}
.section-title {
  margin: 0 0 0.85rem;
  font-size: 0.96rem;
  font-weight: 700;
  color: #1e293b;
}
.data-table {
  width: 100%;
  font-size: 0.86rem;
  border-collapse: collapse;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  overflow: hidden;
}
.data-table th {
  text-align: center;
  padding: 0.7rem 0.75rem;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  color: #64748b;
  font-weight: 700;
}
.data-table td {
  text-align: center;
  padding: 0.7rem 0.75rem;
  border: 1px solid #e2e8f0;
  color: #0f172a;
  background: #fff;
}
.table-card {
  overflow-x: auto;
  padding: 1rem 1rem 0.9rem;
  border-radius: 18px;
  border: 1px solid #dbe7f3;
  background: linear-gradient(180deg, #ffffff 0%, #fbfdff 100%);
  box-shadow: 0 16px 34px rgba(15, 23, 42, 0.05);
}
@media (max-width: 900px) {
  .hero-card {
    flex-direction: column;
  }
  .hero-side {
    align-items: flex-start;
  }
  .metrics-row {
    grid-template-columns: repeat(2, 1fr);
  }
  .two-col {
    grid-template-columns: 1fr;
  }
}
@media (max-width: 520px) {
  .metrics-row {
    grid-template-columns: 1fr;
  }
}
</style>
