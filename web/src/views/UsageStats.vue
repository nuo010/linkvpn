<template>
  <div class="stats-page">
    <div class="page-header">
      <h2 class="page-title">在线统计</h2>
      <div class="header-actions">
        <span class="search-label">搜索：</span>
        <input v-model="searchKeyword" type="text" class="search-input" placeholder="用户名 / VPN IP / 登录 IP" />
        <button class="primary" @click="load">刷新</button>
        <button @click="exportCSV">导出</button>
      </div>
    </div>

    <div v-if="info" class="msg info">{{ info }}</div>

    <div class="card table-wrap">
      <el-table
        :data="paginatedList"
        size="small"
        stripe
        border
        style="width: 100%"
      >
        <el-table-column type="index" label="序号" width="70" align="center" />
        <el-table-column prop="common_name" label="用户" min-width="120" align="center" />
        <el-table-column prop="real_ip" label="用户登录 IP" min-width="140" align="center" />
        <el-table-column prop="virtual_ip" label="用户 DHCP IP (VPN IP)" min-width="150" align="center">
          <template #default="{ row }">
            {{ row.virtual_ip || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="上传流量" min-width="110" align="center">
          <template #default="{ row }">
            {{ formatBytes(row.bytes_recv) }}
          </template>
        </el-table-column>
        <el-table-column label="下载流量" min-width="110" align="center">
          <template #default="{ row }">
            {{ formatBytes(row.bytes_sent) }}
          </template>
        </el-table-column>
        <el-table-column prop="connected_at" label="上线时间" min-width="170" align="center" />
        <el-table-column label="在线时长" min-width="110" align="center">
          <template #default="{ row }">
            {{ formatDuration(row.duration_secs) }}
          </template>
        </el-table-column>
      </el-table>
      <p v-if="filteredList.length === 0 && !loading" class="empty-tip">
        {{ info || '暂无连接数据，请确认 OpenVPN 已运行且配置了 status 文件。' }}
      </p>
    </div>

    <div v-if="filteredList.length > 0" class="pagination">
      <span class="pagination-info">显示 {{ (currentPage - 1) * pageSize + 1 }} 至 {{ Math.min(currentPage * pageSize, filteredList.length) }} 项，共 {{ filteredList.length }} 项</span>
      <select v-model.number="pageSize" class="page-size">
        <option :value="10">10 条/页</option>
        <option :value="20">20 条/页</option>
        <option :value="50">50 条/页</option>
      </select>
      <div class="page-btns">
        <button :disabled="currentPage <= 1" @click="currentPage--">上一页</button>
        <span class="page-num">{{ currentPage }} / {{ totalPages }}</span>
        <button :disabled="currentPage >= totalPages" @click="currentPage++">下一页</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import client from '../api/client'

const loading = ref(false)
const info = ref('')
const list = ref([])
const searchKeyword = ref('')
const currentPage = ref(1)
const pageSize = ref(10)

function formatBytes(n) {
  if (n == null || n < 0) return '0 B'
  if (n < 1024) return n + ' B'
  if (n < 1024 * 1024) return (n / 1024).toFixed(2) + ' KB'
  return (n / (1024 * 1024)).toFixed(2) + ' MB'
}

function formatDuration(secs) {
  if (secs == null || secs < 0) return '-'
  const h = Math.floor(secs / 3600)
  const m = Math.floor((secs % 3600) / 60)
  const s = Math.floor(secs % 60)
  const parts = []
  if (h > 0) parts.push(h + 'h')
  if (m > 0) parts.push(m + 'm')
  parts.push(s + 's')
  return parts.join('')
}

const filteredList = computed(() => {
  let arr = list.value
  const kw = searchKeyword.value.trim().toLowerCase()
  if (kw) {
    arr = arr.filter(
      (r) =>
        (r.common_name && r.common_name.toLowerCase().includes(kw)) ||
        (r.real_ip && r.real_ip.toLowerCase().includes(kw)) ||
        (r.virtual_ip && r.virtual_ip.toLowerCase().includes(kw))
    )
  }
  return arr
})

const totalPages = computed(() => Math.max(1, Math.ceil(filteredList.value.length / pageSize.value)))
const paginatedList = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  return filteredList.value.slice(start, start + pageSize.value)
})

async function load() {
  loading.value = true
  info.value = ''
  try {
    const { data } = await client.get('/stats/usage')
    list.value = data.list || []
    if (data.message) info.value = data.message
  } catch (e) {
    ElMessage.error(e.response?.data?.error || '加载失败')
    list.value = []
  } finally {
    loading.value = false
  }
}

function exportCSV() {
  const rows = filteredList.value
  const headers = ['用户', '用户登录IP', '用户DHCP的IP', '上传流量', '下载流量', '上线时间', '在线时长']
  const lines = [headers.join(',')]
  for (const r of rows) {
    lines.push(
      [
        r.common_name || '',
        r.real_ip || '',
        r.virtual_ip || '',
        formatBytes(r.bytes_recv),
        formatBytes(r.bytes_sent),
        r.connected_at || '',
        formatDuration(r.duration_secs),
      ].join(',')
    )
  }
  const blob = new Blob(['\ufeff' + lines.join('\n')], { type: 'text/csv;charset=utf-8' })
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = 'openvpn-usage-' + new Date().toISOString().slice(0, 10) + '.csv'
  a.click()
  URL.revokeObjectURL(a.href)
}

onMounted(load)
</script>

<style scoped>
.stats-page {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.page-header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.75rem;
}
.page-title {
  margin: 0;
  font-size: 1.1rem;
}
.header-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}
.search-label {
  font-size: 0.9rem;
  color: var(--muted);
}
.search-input {
  width: 220px;
}
.empty-tip {
  padding: 1.5rem;
  margin: 0;
  color: var(--muted);
  text-align: center;
}
.pagination {
  display: flex;
  align-items: center;
  gap: 1rem;
  flex-wrap: wrap;
  justify-content: flex-end;
}
.pagination-info {
  color: var(--muted);
  font-size: 0.9rem;
}
.page-size {
  width: auto;
  padding: 0.35rem 0.5rem;
}
.page-btns {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
.page-num {
  font-size: 0.9rem;
  color: var(--muted);
}
</style>
