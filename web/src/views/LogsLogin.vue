<template>
  <div class="logs-page">
    <div class="hero-card">
      <div class="hero-copy">
        <p class="hero-eyebrow">访问审计</p>
        <h2 class="page-title">连接记录</h2>
        <p class="hero-desc">这里展示账号登录成功、TLS 错误和认证失败记录，方便按日期快速回看最近的接入情况。</p>
      </div>
      <span class="refresh-hint">每 8 秒自动刷新</span>
    </div>

    <div class="toolbar-card">
      <div class="page-header">
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
      <div v-if="list.length === 0 && !loading" class="empty-state">
        <p class="empty-state-title">暂无 VPN 连接记录</p>
        <p class="empty-state-desc">记录会在访问“用户管理”或刷新 status 日志时同步生成，你也可以切换日期后重新查看。</p>
      </div>
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
.hero-card {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
  padding: 1.15rem 1.2rem;
  border-radius: 18px;
  border: 1px solid #dbe7f3;
  background:
    radial-gradient(circle at top right, rgba(96, 165, 250, 0.12), transparent 32%),
    linear-gradient(135deg, #ffffff 0%, #f7fbff 100%);
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
.page-title {
  margin: 0;
  font-size: 1.3rem;
  color: #0f172a;
}
.hero-desc {
  margin: 0.45rem 0 0;
  color: #64748b;
  line-height: 1.65;
  font-size: 0.92rem;
}
.toolbar-card {
  padding: 0.95rem 1rem;
  border-radius: 16px;
  border: 1px solid #dbe7f3;
  background: linear-gradient(180deg, #ffffff 0%, #fbfdff 100%);
  box-shadow: 0 12px 28px rgba(15, 23, 42, 0.04);
}
.page-header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.75rem;
  justify-content: flex-start;
}
.date-label {
  font-size: 0.9rem;
  color: #64748b;
  font-weight: 600;
}
.date-input {
  width: 150px;
}
.refresh-hint {
  display: inline-flex;
  align-items: center;
  min-height: 34px;
  padding: 0 0.85rem;
  border-radius: 999px;
  font-size: 0.82rem;
  color: #2563eb;
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  font-weight: 600;
}
.pagination-wrap {
  padding: 0.95rem 0 0.15rem;
  display: flex;
  justify-content: flex-end;
}

.table-wrap {
  padding: 0.7rem 0.8rem 0.9rem;
  border-radius: 18px;
  border: 1px solid #dbe7f3;
  background: linear-gradient(180deg, #ffffff 0%, #fbfdff 100%);
  box-shadow: 0 16px 34px rgba(15, 23, 42, 0.05);
}
.table-wrap :deep(.el-table) {
  border-radius: 14px;
  overflow: hidden;
  --el-table-border-color: #e2e8f0;
  --el-table-header-bg-color: #f8fafc;
  --el-table-row-hover-bg-color: #f8fbff;
}
.table-wrap :deep(.el-table th.el-table__cell) {
  color: #64748b;
  font-weight: 700;
}
.table-wrap :deep(.el-table td.el-table__cell) {
  color: #0f172a;
}

.status-tag {
  display: inline-block;
  padding: 0.22rem 0.7rem;
  border-radius: 999px;
  font-size: 0.8rem;
  font-weight: 700;
  border: 1px solid transparent;
}
.status-tag.success {
  background: #ecfdf3;
  color: #15803d;
  border-color: #bbf7d0;
}
.status-tag.failed {
  background: #fff1f2;
  color: #dc2626;
  border-color: #fecdd3;
}
@media (max-width: 900px) {
  .hero-card {
    flex-direction: column;
  }
}
@media (max-width: 640px) {
  .toolbar-card {
    padding: 0.85rem;
  }
  .date-input {
    width: 100%;
  }
  .page-header {
    align-items: stretch;
  }
  .page-header > * {
    width: 100%;
  }
}
</style>
