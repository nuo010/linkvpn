<template>
  <div class="config-page">
    <div class="card alert-warning">
      <span class="alert-icon">⚠️</span>
      注意：编辑 VPN 主配置文件之前请仔细阅读注释说明，修改后需重启 VPN 服务才能生效。
    </div>

    <div class="card config-card">
      <h3 class="section-title">VPN 服务配置</h3>
      <div class="config-actions">
        <el-button type="primary" @click="saveFile">提交更新配置</el-button>
        <el-button @click="loadFile">重新加载</el-button>
        <el-button @click="loadDefault">加载默认配置</el-button>
        <el-button
          type="warning"
          plain
          class="restart-service-btn"
          :loading="restartLoading"
          @click="restartService"
        >
          <span class="restart-icon">↻</span>
          {{ restartLoading ? '正在重启…' : '重启 OpenVPN' }}
        </el-button>
      </div>
      <textarea v-model="fileContent" class="config-editor" rows="24" placeholder="server.conf 内容…"></textarea>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import client from '../api/client'

const fileContent = ref('')
const restartLoading = ref(false)

async function loadFile() {
  try {
    const { data } = await client.get('/config/file')
    fileContent.value = data.content || ''
    if (!fileContent.value.trim()) {
      ElMessage.warning('当前无配置文件，已填充默认内容，保存后生效。')
    }
  } catch (e) {
    ElMessage.error(e.response?.data?.error || '加载失败')
  }
}

async function loadDefault() {
  try {
    const { data } = await client.get('/config/default')
    fileContent.value = data.content || ''
    ElMessage.success('已加载默认配置，点击「提交更新配置」保存到服务器。')
  } catch (e) {
    ElMessage.error(e.response?.data?.error || '加载失败')
  }
}

async function saveFile() {
  try {
    await client.put('/config/file', { content: fileContent.value })
    ElMessage.success('配置已保存')
  } catch (e) {
    ElMessage.error(e.response?.data?.error || '保存失败')
  }
}

async function restartService() {
  restartLoading.value = true
  try {
    const { data } = await client.post('/config/restart')
    ElMessage.success(data.message || '重启已触发')
  } catch (e) {
    ElMessage.error(e.response?.data?.error || '重启失败')
  } finally {
    restartLoading.value = false
  }
}

onMounted(loadFile)
</script>

<style scoped>
.config-page {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.alert-warning {
  background: #fdf6ec;
  border-color: #f5dab1;
  color: #b88230;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
.alert-icon {
  font-size: 1.2rem;
}
.config-card {
  overflow: hidden;
}
.section-title {
  margin: 0 0 0.75rem;
  font-size: 1rem;
}
.config-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}
/* 重启：醒目但不抢眼，与主按钮区分 */
.restart-service-btn {
  margin-left: auto;
  font-weight: 600;
  border-width: 2px;
  box-shadow: 0 2px 8px rgba(230, 162, 60, 0.25);
}
.restart-service-btn:hover:not(:disabled) {
  box-shadow: 0 4px 12px rgba(230, 162, 60, 0.35);
}
.restart-icon {
  display: inline-block;
  margin-right: 6px;
  font-size: 1.1em;
  line-height: 1;
  vertical-align: -0.05em;
}
.config-editor {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 0.85rem;
  resize: vertical;
  min-height: 320px;
}
</style>
