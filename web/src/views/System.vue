<template>
  <div class="system-page">
    <div class="card alert-tip">
      <span class="alert-icon">ℹ️</span>
      系统配置提示：若无法正确获取数据，请刷新页面。
    </div>

    <div class="card section-card">
      <h3 class="section-title">客户端下载配置</h3>
      <p class="muted-tip">下载的 .ovpn 文件中的「服务器地址」和「端口」将使用下方配置，未填写时默认 127.0.0.1:1194。</p>
      <div class="form-row">
        <label>服务器地址</label>
        <el-input v-model="clientRemoteHost" placeholder="如 vpn.example.com 或 1.2.3.4" size="small" />
      </div>
      <div class="form-row">
        <label>端口</label>
        <el-input
          v-model.number="clientRemotePort"
          type="number"
          min="1"
          max="65535"
          placeholder="1194"
          size="small"
        />
      </div>
      <div class="form-actions">
        <el-button type="primary" :loading="saveRemoteLoading" size="small" @click="saveClientRemote">
          保存
        </el-button>
      </div>
    </div>

    <div class="card section-card">
      <h3 class="section-title">创建 CA / 服务端证书</h3>
      <p v-if="status.pki_initialized" class="success-tip">
        您已成功设置 CA 根证书以及 Server 端证书，可在上方「VPN 配置」标签中查看与编辑 server.conf。
      </p>
      <p v-else class="muted-tip">
        首次部署请点击下方「初始化 PKI」，生成 CA、服务端证书与默认 server.conf（需在已安装 easy-rsa 的环境中执行）。
      </p>
      <div class="form-actions">
        <el-button type="primary" :loading="initLoading" size="small" @click="initVPN">
          初始化 PKI
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import client from '../api/client'

const status = ref({ pki_initialized: false })
const initLoading = ref(false)
const clientRemoteHost = ref('')
const clientRemotePort = ref(1194)
const saveRemoteLoading = ref(false)

async function loadStatus() {
  try {
    const { data } = await client.get('/vpn/status')
    status.value = data || {}
  } catch (_) {
    status.value = {}
  }
}

async function loadConfig() {
  try {
    const { data } = await client.get('/config')
    clientRemoteHost.value = data.client_remote_host || ''
    const p = data.client_remote_port
    clientRemotePort.value = p ? parseInt(p, 10) || 1194 : 1194
  } catch (_) {
    clientRemoteHost.value = ''
    clientRemotePort.value = 1194
  }
}

async function saveClientRemote() {
  saveRemoteLoading.value = true
  try {
    const port = clientRemotePort.value
    await client.post('/config', {
      client_remote_host: clientRemoteHost.value.trim(),
      client_remote_port: (port > 0 && port <= 65535 ? port : 1194).toString(),
    })
  } finally {
    saveRemoteLoading.value = false
  }
}

async function initVPN() {
  initLoading.value = true
  try {
    await client.post('/vpn/init')
    await loadStatus()
  } finally {
    initLoading.value = false
  }
}

onMounted(() => {
  loadStatus()
  loadConfig()
})
</script>

<style scoped>
.system-page {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.alert-tip {
  background: #ecf5ff;
  border-color: #b3d8ff;
  color: #409eff;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
.alert-icon {
  font-size: 1.2rem;
}
.section-card {
  width: 100%;
}
.section-title {
  margin: 0 0 0.75rem;
  font-size: 1rem;
}
.form-row {
  margin-bottom: 0.75rem;
}
.form-row label {
  display: block;
  margin-bottom: 0.25rem;
  font-size: 0.9rem;
  color: var(--muted);
}
.form-row input {
  width: 100%;
}
.success-tip {
  margin: 0 0 1rem;
  color: var(--success);
  font-size: 0.9rem;
}
.muted-tip {
  margin: 0 0 1rem;
  color: var(--muted);
  font-size: 0.9rem;
}
.form-actions {
  margin-top: 0.5rem;
}
@media (max-width: 720px) {
  .alert-tip,
  .section-card {
    padding-left: 1rem;
    padding-right: 1rem;
  }
  .section-title {
    font-size: 0.98rem;
  }
  .form-actions {
    display: flex;
  }
  .form-actions :deep(.el-button) {
    width: 100%;
  }
}
</style>
