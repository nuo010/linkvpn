<template>
  <div class="vpn-logs-page">
    <div class="page-header">
      <h2 class="page-title">VPN 服务日志</h2>
      <div class="header-actions">
        <label class="label">日志文件：</label>
        <select v-model="currentFile" class="log-select" @change="load">
          <option value="openvpn.log">openvpn.log（服务主日志）</option>
          <option value="openvpn-status.log">openvpn-status.log（连接状态）</option>
        </select>
        <label class="label">行数：</label>
        <select v-model.number="lines" class="lines-select" @change="load">
          <option :value="200">200</option>
          <option :value="500">500</option>
          <option :value="1000">1000</option>
          <option :value="2000">2000</option>
        </select>
        <button class="primary" :disabled="loading" @click="load">{{ loading ? '加载中…' : '刷新' }}</button>
        <label class="auto-refresh">
          <input v-model="autoRefresh" type="checkbox" />
          自动刷新（每 3 秒）
        </label>
      </div>
    </div>

    <div v-if="path" class="path-tip">路径：{{ path }}</div>
    <div v-if="message" class="msg error">{{ message }}</div>

    <div class="card log-wrap">
      <pre class="log-content">{{ content || (loading ? '加载中…' : '暂无内容或文件不存在') }}</pre>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, onMounted, onUnmounted } from 'vue'
import client from '../api/client'

const currentFile = ref('openvpn.log')
const lines = ref(500)
const content = ref('')
const path = ref('')
const loading = ref(false)
const message = ref(null)
const autoRefresh = ref(false)
let timer = null

async function load() {
  loading.value = true
  message.value = null
  try {
    const { data } = await client.get('/logs/vpn-file', {
      params: { name: currentFile.value, lines: lines.value },
    })
    content.value = data.content || ''
    path.value = data.path || ''
    if (data.message) message.value = data.message
  } catch (e) {
    content.value = ''
    message.value = e.response?.data?.error || '加载失败'
  } finally {
    loading.value = false
  }
}

watch(autoRefresh, (on) => {
  if (timer) clearInterval(timer)
  timer = null
  if (on) timer = setInterval(load, 3000)
})

onMounted(load)
onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<style scoped>
.vpn-logs-page {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.page-header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
}
.label {
  font-size: 0.9rem;
  color: var(--muted);
}
.log-select {
  width: 240px;
}
.lines-select {
  width: 90px;
}
.auto-refresh {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.9rem;
  color: var(--muted);
  cursor: pointer;
}
.path-tip {
  font-size: 0.85rem;
  color: var(--muted);
}
.log-wrap {
  padding: 1rem;
  overflow: hidden;
}
.log-content {
  margin: 0;
  font-family: ui-monospace, monospace;
  font-size: 0.85rem;
  line-height: 1.4;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 70vh;
  overflow: auto;
  color: #333;
}
</style>
