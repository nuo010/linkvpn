<template>
  <div class="log-view-page">
    <div class="page-header">
      <h2 class="page-title">{{ pageTitle }}</h2>
      <div class="header-actions">
        <label class="label">行数</label>
        <el-select v-model="lines" class="lines-select" size="small" @change="load">
          <el-option :value="200" label="200" />
          <el-option :value="500" label="500" />
          <el-option :value="1000" label="1000" />
          <el-option :value="2000" label="2000" />
        </el-select>
        <el-button type="primary" size="small" :loading="loading" @click="load">刷新</el-button>
        <el-checkbox v-model="autoRefresh" class="auto-refresh" size="small">
          自动刷新（每 3 秒）
        </el-checkbox>
        <el-checkbox v-model="autoScroll" class="auto-refresh" size="small">
          自动滚动到底部
        </el-checkbox>
      </div>
    </div>
    <div v-if="path" class="path-tip">路径：{{ path }}</div>
    <div v-if="message" class="msg error">{{ message }}</div>
    <div class="card log-wrap">
      <pre ref="logContentRef" class="log-content">{{ content || (loading ? '加载中…' : '暂无内容或文件不存在') }}</pre>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, onMounted, onUnmounted, computed, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import client from '../api/client'

const route = useRoute()
const logFile = computed(() => route.meta.logFile || 'openvpn.log')
const pageTitle = computed(() => route.meta.title || 'VPN 日志')

const lines = ref(500)
const content = ref('')
const path = ref('')
const loading = ref(false)
const message = ref(null)
const autoRefresh = ref(false)
const autoScroll = ref(true)
const logContentRef = ref(null)
let timer = null

async function scrollToBottom() {
  await nextTick()
  const el = logContentRef.value
  if (!el) return
  el.scrollTop = el.scrollHeight
}

async function load() {
  loading.value = true
  message.value = null
  try {
    const { data } = await client.get('/logs/vpn-file', {
      params: { name: logFile.value, lines: lines.value },
    })
    content.value = data.content || ''
    path.value = data.path || ''
    if (data.message) message.value = data.message
    if (autoScroll.value) {
      await scrollToBottom()
    }
  } catch (e) {
    content.value = ''
    message.value = e.response?.data?.error || '加载失败'
    ElMessage.error(message.value)
  } finally {
    loading.value = false
  }
}

watch(autoRefresh, (on) => {
  if (timer) clearInterval(timer)
  timer = null
  if (on) timer = setInterval(load, 3000)
})
watch(autoScroll, (on) => {
  if (on) scrollToBottom()
})
watch(logFile, () => { load() })

onMounted(load)
onUnmounted(() => { if (timer) clearInterval(timer) })
</script>

<style scoped>
.log-view-page {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}
.page-header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
}
.header-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
}
.page-title {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 600;
}
.label {
  font-size: 0.9rem;
  color: var(--muted);
}
.lines-select {
  width: 90px;
  padding: 0.4rem 0.5rem;
  height: 34px;
}
.path-tip {
  font-size: 0.85rem;
  color: var(--muted);
}
.log-wrap {
  padding: 1rem;
  overflow: hidden;
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.log-content {
  margin: 0;
  font-family: ui-monospace, monospace;
  font-size: 0.8rem;
  line-height: 1.4;
  white-space: pre-wrap;
  word-break: break-all;
  flex: 1;
  min-height: 200px;
  max-height: calc(100vh - 220px);
  overflow: auto;
  color: #333;
}
.auto-refresh {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.85rem;
  color: var(--muted);
  cursor: pointer;
}
@media (max-width: 820px) {
  .page-header {
    align-items: stretch;
  }
  .header-actions {
    width: 100%;
    align-items: stretch;
  }
  .lines-select {
    width: 100%;
    padding: 0;
  }
  .header-actions :deep(.el-button) {
    width: 100%;
  }
  .auto-refresh {
    width: 100%;
  }
  .log-wrap {
    padding: 0.85rem;
  }
  .log-content {
    max-height: calc(100vh - 260px);
    min-height: 160px;
    font-size: 0.76rem;
  }
}
</style>
