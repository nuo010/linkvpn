<template>
  <div class="logs-page">
    <div class="page-header">
      <h2 class="page-title">连接记录</h2>
      <div class="header-actions">
        <label class="date-label">日期</label>
        <el-date-picker
          v-model="dateFilter"
          type="date"
          value-format="YYYY-MM-DD"
          class="date-input"
          size="small"
        />
        <el-button type="primary" size="small" @click="currentPage = 1; load()">查看</el-button>
        <el-button type="danger" size="small" @click="clearLogs">清空记录</el-button>
        <span class="refresh-hint">每 8 秒自动刷新</span>
      </div>
    </div>

    <div class="card table-wrap">
      <el-table
        :data="list"
        size="small"
        stripe
        border
        style="width: 100%"
        :loading="loading"
      >
        <el-table-column label="序号" width="70" align="center">
          <template #default="{ $index }">
            {{ (currentPage - 1) * pageSize + $index + 1 }}
          </template>
        </el-table-column>
        <el-table-column prop="username" label="用户名" min-width="120" align="center" />
        <el-table-column label="状态" min-width="120" align="center">
          <template #default="{ row }">
            <span :class="['status-tag', row.status === 'success' ? 'success' : 'failed']">
              {{
                row.status === 'success'
                  ? '成功'
                  : row.status === 'tls_error'
                    ? 'TLS 错误'
                    : '未成功（账号或密码错误等）'
              }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="real_ip" label="来源 IP" min-width="140" align="center">
          <template #default="{ row }">
            {{ row.real_ip || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="连接时间" min-width="180" align="center">
          <template #default="{ row }">
            {{ formatTime(row.connected_at) }}
          </template>
        </el-table-column>
      </el-table>
      <p v-if="list.length === 0 && !loading" class="empty-tip">暂无 VPN 连接记录。记录在访问「用户管理」或刷新 status 时同步生成。</p>
      <div v-if="total > 0" class="pagination-wrap">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next"
          small
          @size-change="load"
          @current-change="load"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import client from '../api/client'

const list = ref([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)
const dateFilter = ref(new Date().toISOString().slice(0, 10))
const loading = ref(false)

function formatTime(s) {
  if (!s) return '-'
  return new Date(s).toLocaleString('zh-CN')
}

function formatBytes(n) {
  if (n == null || n < 0) return '0 B'
  if (n < 1024) return n + ' B'
  if (n < 1024 * 1024) return (n / 1024).toFixed(2) + ' KB'
  return (n / (1024 * 1024)).toFixed(2) + ' MB'
}

async function load() {
  loading.value = true
  try {
    const params = {
      page: currentPage.value,
      page_size: pageSize.value,
    }
    if (dateFilter.value) params.date = dateFilter.value
    const { data } = await client.get('/logs/vpn', { params })
    list.value = data.list || []
    total.value = data.total ?? 0
  } catch (e) {
    ElMessage.error(e.response?.data?.error || '加载失败')
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

async function clearLogs() {
  if (!confirm('确定要清空所有 VPN 连接记录吗？')) return
  try {
    await client.delete('/logs/vpn')
    ElMessage.success('已清空')
    total.value = 0
    currentPage.value = 1
    load()
  } catch (e) {
    ElMessage.error(e.response?.data?.error || '清空失败')
  }
}

let refreshTimer = null
onMounted(() => {
  load()
  refreshTimer = setInterval(load, 8000)
})
onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer)
})
</script>

<style scoped>
.logs-page {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.page-header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.75rem;
  justify-content: space-between;
}
.date-label {
  font-size: 0.9rem;
  color: var(--muted);
}
.date-input {
  width: 150px;
}
.refresh-hint {
  font-size: 0.8rem;
  color: var(--muted);
}
.empty-tip {
  padding: 1.5rem;
  margin: 0;
  color: var(--muted);
  text-align: center;
}
.pagination-wrap {
  padding: 0.75rem 0;
  display: flex;
  justify-content: flex-end;
}

.status-tag {
  display: inline-block;
  padding: 0.1rem 0.4rem;
  border-radius: 999px;
  font-size: 0.8rem;
}
.status-tag.success {
  background: #e1f3d8;
  color: #67c23a;
}
.status-tag.failed {
  background: #fef0f0;
  color: #f56c6c;
}
</style>
