<template>
  <div class="user-page">
    <div class="page-header">
      <h2 class="page-title">VPN 用户列表</h2>
      <div class="header-actions">
        <div class="search-wrap">
          <span class="search-label">综合查询</span>
          <input
            v-model="searchKeyword"
            type="text"
            class="search-input"
            placeholder="请输入 VPN 用户名"
            @keyup.enter="applySearch"
          />
          <input
            v-model="searchIP"
            type="text"
            class="search-input ip-input"
            placeholder="IP（获取到的 IP / 静态 IP）"
            @keyup.enter="applySearch"
          />
          <el-select
            v-model="filterKind"
            placeholder="全部类型"
            class="filter-select"
            size="default"
          >
            <el-option label="全部类型" value="all" />
            <el-option label="用户" value="user" />
            <el-option label="客户端" value="client" />
          </el-select>
          <el-select
            v-model="filterEnabled"
            placeholder="全部状态"
            class="filter-select"
            size="default"
          >
            <el-option label="全部状态" value="all" />
            <el-option label="仅启用" value="enabled" />
            <el-option label="仅禁用" value="disabled" />
          </el-select>
          <el-select
            v-model="filterOnline"
            placeholder="全部在线状态"
            class="filter-select filter-select-wide"
            size="default"
          >
            <el-option label="全部在线状态" value="all" />
            <el-option label="仅在线" value="online" />
            <el-option label="仅离线" value="offline" />
          </el-select>
          <button class="primary" @click="applySearch">查询</button>
        </div>
        <div class="toolbar-meta">
          <span class="refresh-hint">每 5 秒自动刷新</span>
          <button class="primary add-btn" @click="openForm()">+ 添加用户</button>
        </div>
      </div>
    </div>

    <div class="card user-table-card">
      <div class="table-wrap">
      <el-table
        :data="paginatedUsers"
        size="small"
        stripe
        border
        table-layout="auto"
        style="width: 100%"
        :header-cell-style="{ background: '#f8fafc', color: '#475569', fontWeight: '700' }"
      >
        <el-table-column label="序号" width="70" align="center">
          <template #default="{ $index }">
            {{ (currentPage - 1) * pageSize + $index + 1 }}
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="昵称" min-width="120" align="center">
          <template #default="{ row }">
            {{ row.remark || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="name" label="VPN 用户名" min-width="140" align="center" />
        <el-table-column label="类型" min-width="100" align="center">
          <template #default="{ row }">
            <el-tag
              :type="row.kind === 'client' ? 'success' : 'info'"
              effect="light"
              size="small"
              class="tag-type"
            >
              {{ row.kind === 'client' ? '客户端' : '用户' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="分配 IP" min-width="160" align="center">
          <template #default="{ row }">
            <div class="ip-with-mode">
              <el-tag effect="plain" size="small" class="tag-ip">
                {{ displayIP(row) }}
              </el-tag>
              <el-tag
                v-if="displayIP(row) !== '-'"
                effect="plain"
                size="small"
                class="tag-ip-mode"
                :type="ipMode(row) === 'DHCP' ? 'info' : 'success'"
              >
                {{ ipMode(row) }}
              </el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="用户状态" min-width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'danger'" effect="light" size="small" class="tag-enabled">
              {{ row.enabled ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="当前在线" min-width="100" align="center">
          <template #default="{ row }">
            <span :class="['online-chip', row.online ? 'on' : 'off']">
              <span class="dot" />
              {{ row.online ? '在线' : '离线' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="上线时间" min-width="160" align="center">
          <template #default="{ row }">
            {{ row.online ? formatTimeToMinute(row.connected_at) : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="在线时长" min-width="120" align="center">
          <template #default="{ row }">
            {{ row.online ? formatDuration(durationSecsForRow(row)) : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="上传流量" min-width="120" align="center">
          <template #default="{ row }">
            {{ formatBytes(row.bytes_recv) }}
          </template>
        </el-table-column>
        <el-table-column label="下载流量" min-width="120" align="center">
          <template #default="{ row }">
            {{ formatBytes(row.bytes_sent) }}
          </template>
        </el-table-column>
        <el-table-column label="创建时间" min-width="160" align="center">
          <template #default="{ row }">
            {{ formatTimeToMinute(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="功能操作" min-width="220" fixed="right" align="center">
          <template #default="{ row }">
            <div class="actions-cell">
              <el-dropdown trigger="click" popper-class="user-actions-dropdown" @command="(cmd) => onActionCommand(cmd, row)">
                <el-button size="small" type="primary" plain>
                  操作
                  <span class="dropdown-caret">▼</span>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="download">下载配置</el-dropdown-item>
                    <el-dropdown-item command="copy">复制用户名和密码</el-dropdown-item>
                    <el-dropdown-item command="edit">编辑</el-dropdown-item>
                    <el-dropdown-item command="delete" divided>
                      <span class="dropdown-danger-text">删除</span>
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </template>
        </el-table-column>
      </el-table>
      </div>
      <div v-if="filteredUsers.length === 0" class="empty-state user-empty-state">
        <p class="empty-state-title">暂无用户</p>
        <p class="empty-state-desc">当前还没有 VPN 用户，点击上方“添加用户”就可以开始创建。</p>
      </div>
    </div>

    <div v-if="filteredUsers.length > 0" class="pagination">
      <span class="pagination-info">共 {{ filteredUsers.length }} 条</span>
      <el-select
        v-model="pageSize"
        class="page-size-select"
        size="default"
        @change="currentPage = 1"
      >
        <el-option :value="10" label="10 条/页" />
        <el-option :value="20" label="20 条/页" />
        <el-option :value="50" label="50 条/页" />
      </el-select>
      <div class="page-btns">
        <button :disabled="currentPage <= 1" @click="currentPage--">上一页</button>
        <span class="page-num">{{ currentPage }} / {{ totalPages }}</span>
        <button :disabled="currentPage >= totalPages" @click="currentPage++">下一页</button>
      </div>
    </div>

    <div v-if="showForm" class="modal" @click.self="showForm = false">
      <div class="card modal-inner user-form-modal modal-inner-wide">
        <h3 class="modal-title">{{ editTarget ? '编辑用户' : '添加用户' }}</h3>
        <form class="user-form" @submit.prevent="editTarget ? updateUser() : createUser()">
          <!-- 基本信息：分块 + 统一网格，避免左右错位 -->
          <div class="form-section form-section-basic">
            <h4 class="form-section-title">基本信息</h4>
            <div class="form-basic-card">
              <div class="form-grid-2">
                <div class="form-group">
                  <label>昵称</label>
                  <el-input v-model="form.remark" placeholder="选填，便于区分用户" size="default" clearable />
                </div>
                <div class="form-group">
                  <label>VPN 用户名 <span class="req">*</span></label>
                  <el-input
                    v-model="form.name"
                    :readonly="!!editTarget"
                    placeholder="仅英文字母、数字、下划线"
                    size="default"
                    @input="validateVPNName"
                  />
                  <span v-if="editTarget" class="form-hint">编辑时不可修改用户名</span>
                  <span v-else-if="form.name && !isValidVPNName(form.name)" class="form-hint error">
                    不能含中文或特殊字符
                  </span>
                </div>
              </div>
              <div class="form-grid-2">
                <div class="form-group">
                  <label>类型</label>
                  <el-radio-group v-model="form.kind" size="default" class="kind-radio">
                    <el-radio-button label="user">用户</el-radio-button>
                    <el-radio-button label="client">客户端</el-radio-button>
                  </el-radio-group>
                </div>
                <div class="form-group">
                  <label>{{ editTarget ? '密码（留空不改）' : '密码' }} <span v-if="!editTarget" class="req">*</span></label>
                  <el-input
                    v-model="form.password"
                    type="password"
                    show-password
                    :placeholder="editTarget ? '不修改请留空' : '设置登录密码'"
                    autocomplete="new-password"
                    size="default"
                  />
                </div>
              </div>
              <div class="form-grid-2">
                <div class="form-group">
                  <label>账号到期</label>
                  <el-date-picker
                    v-model="form.expires_at"
                    type="date"
                    value-format="YYYY-MM-DD"
                    placeholder="不选则永不过期"
                    class="form-date-full"
                    size="default"
                    clearable
                  />
                </div>
                <div class="form-group">
                  <label>联系邮箱</label>
                  <el-input v-model="form.email" type="email" placeholder="选填" size="default" clearable />
                </div>
              </div>
            </div>
          </div>
          <div v-if="editTarget" class="form-section form-section-ccd">
            <div class="ccd-section-title">CCD 配置</div>
            <p class="ccd-warn-inline">以下配置写入 client-config-dir；修改后需客户端重连生效。每用户仅允许一条 ifconfig-push。</p>
            <div class="form-group">
              <label>给客户端指定静态 IP（ifconfig-push）</label>
              <div class="form-row-2">
                <el-input v-model="form.ccd_ifconfig_ip" placeholder="客户端 IP，如 10.8.8.99" size="small" />
                <el-input v-model="form.ccd_ifconfig_mask" placeholder="子网掩码，填 255.255.255.0" size="small" />
              </div>
              <span class="form-hint">第二项必须为子网掩码 255.255.255.0（勿填 10.8.8.1）。保存后需客户端重连生效。</span>
            </div>
            <div class="form-group">
              <label>内网 IP 地址范围（iroute，可多行）</label>
              <textarea v-model="form.ccd_iroutes" class="ccd-textarea-small" rows="3" placeholder="每行一条，如：iroute 10.44.111.0 255.255.255.0"></textarea>
            </div>
            <div class="form-group">
              <label>推送给客户端的路由（push "route"，可多行）</label>
              <textarea v-model="form.ccd_push_routes" class="ccd-textarea-small" rows="3" placeholder='每行一条，如：push "route 172.16.10.0 255.255.255.0"'></textarea>
            </div>
            <div class="form-group form-group-checkbox">
              <el-checkbox v-model="form.ccd_redirect_gateway">
                重定向客户端所有流量（通过 VPN 服务器上网）
              </el-checkbox>
            </div>
            <div class="form-group form-group-checkbox">
              <el-checkbox v-model="form.route_nopull">
                忽略服务端推送的路由（下载的配置中添加 route-nopull）
              </el-checkbox>
              <span class="form-hint">开启后，下载的 .ovpn 中会加入「# 忽略服务端推送的路由」与「route-nopull」</span>
            </div>
          </div>
          <div class="form-section form-footer-bar">
            <el-checkbox v-model="form.enabled" size="default">启用账号</el-checkbox>
            <div class="form-actions">
              <el-button @click="closeForm">取消</el-button>
              <el-button type="primary" native-type="submit">
                {{ editTarget ? '保存' : '创建' }}
              </el-button>
            </div>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import client from '../api/client'

const users = ref([])
const showForm = ref(false)
const editTarget = ref(null) // 编辑时的用户对象
// 仅在当前前端会话中缓存「最近设置的明文密码」，便于下拉菜单一键复制。
// 不会持久化到后端数据库。
const lastPasswords = ref({})
const defaultForm = () => ({
  name: '', kind: 'user', password: '', email: '', remark: '', expires_at: '', enabled: true,
  ccd_ifconfig_ip: '', ccd_ifconfig_mask: '', ccd_iroutes: '', ccd_push_routes: '', ccd_redirect_gateway: false,
  route_nopull: false,
})
const form = ref(defaultForm())
const searchKeyword = ref('')
const searchIP = ref('')
const filterKind = ref('all') // all | user | client
const filterEnabled = ref('all') // all | enabled | disabled
const filterOnline = ref('all') // all | online | offline
const currentPage = ref(1)
const pageSize = ref(20)
function onActionCommand(cmd, row) {
  if (cmd === 'download') downloadConfig(row)
  else if (cmd === 'copy') copyCredentials(row)
  else if (cmd === 'edit') openEdit(row)
  else if (cmd === 'delete') del(row)
}

function formatDate(s) {
  if (!s) return '-'
  return new Date(s).toLocaleString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}
function formatTimeToMinute(s) {
  if (!s) return '-'
  let d = new Date(s)
  if (isNaN(d.getTime())) {
    const ms = parseConnectedAtMs(s)
    if (isNaN(ms)) return '-'
    d = new Date(ms)
  }
  if (isNaN(d.getTime())) return '-'
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  const h = String(d.getHours()).padStart(2, '0')
  const min = String(d.getMinutes()).padStart(2, '0')
  return `${y}/${m}/${day} ${h}:${min}`
}
function isValidVPNName(name) {
  return /^[a-zA-Z0-9_]+$/.test(name || '')
}
function validateVPNName() {
  // 仅用于触发校验提示
}

function formatBytes(n) {
  if (n == null || n < 0) return '0 B'
  if (n < 1024) return n + ' B'
  if (n < 1024 * 1024) return (n / 1024).toFixed(2) + ' KB'
  return (n / (1024 * 1024)).toFixed(2) + ' MB'
}

function displayIP(u) {
  const raw = u.current_ip || u.static_ip || ''
  if (!raw) return '-'
  // 只提取 IPv4 部分，避免出现类似 "192.168.9.199C" 或附带掩码等脏数据
  const m = raw.match(/\d+\.\d+\.\d+\.\d+/)
  return m ? m[0] : raw
}

function ipMode(u) {
  // 有静态 IP 配置则认为是静态分配，否则为 DHCP/动态分配
  if (u.static_ip && u.static_ip.trim() !== '') return '静态'
  if (u.current_ip && u.current_ip.trim() !== '') return 'DHCP'
  return ''
}

// 解析上线时间戳（ms）。JS 对 "2026-03-11 22:07:00" 无 T 常返回 Invalid Date，需手动按本地时间解析
function parseConnectedAtMs(s) {
  if (s == null || s === '') return NaN
  if (typeof s !== 'string') s = String(s)
  s = s.trim()
  let d = new Date(s)
  if (!isNaN(d.getTime())) return d.getTime()
  // NaiveTime / openvpn status 常见：2026-03-11 22:07:00 或 2026-03-11T22:07:00
  const m = s.match(/^(\d{4})-(\d{2})-(\d{2})[ T](\d{2}):(\d{2}):(\d{2})/)
  if (m) {
    const y = +m[1], mo = +m[2] - 1, day = +m[3], h = +m[4], min = +m[5], sec = +m[6]
    d = new Date(y, mo, day, h, min, sec)
    if (!isNaN(d.getTime())) return d.getTime()
  }
  return NaN
}

// 在线时长：优先用后端 duration_secs；缺失或为 0 时用 connected_at 推算
function durationSecsForRow(row) {
  let secs = row.duration_secs
  if (typeof secs === 'string' && secs !== '') secs = parseInt(secs, 10)
  if (secs != null && !isNaN(secs) && secs > 0) return secs
  if (!row.connected_at) return secs != null && !isNaN(secs) ? secs : 0
  const t = parseConnectedAtMs(row.connected_at)
  if (isNaN(t)) return secs != null && !isNaN(secs) ? secs : 0
  const delta = Math.floor((Date.now() - t) / 1000)
  return delta >= 0 ? delta : 0
}

function formatDuration(secs) {
  if (secs == null || isNaN(secs)) return '-'
  secs = Math.floor(Number(secs))
  if (secs < 0) return '-'
  const h = Math.floor(secs / 3600)
  const m = Math.floor((secs % 3600) / 60)
  const s = Math.floor(secs % 60)
  const parts = []
  if (h > 0) parts.push(h + 'h')
  if (m > 0) parts.push(m + 'm')
  parts.push(s + 's')
  return parts.join('')
}

const filteredUsers = computed(() => {
  let list = users.value
  const kw = searchKeyword.value.trim().toLowerCase()
  const ipKw = searchIP.value.trim().toLowerCase()

  if (kw) {
    list = list.filter(u => (u.name || '').toLowerCase().includes(kw) || (u.remark || '').toLowerCase().includes(kw))
  }
  if (ipKw) {
    list = list.filter(u => {
      const ip = (displayIP(u) || '').toLowerCase()
      return ip && ip.includes(ipKw)
    })
  }
  if (filterKind.value !== 'all') {
    list = list.filter(u => (u.kind || 'user') === filterKind.value)
  }
  if (filterEnabled.value !== 'all') {
    const wantEnabled = filterEnabled.value === 'enabled'
    list = list.filter(u => !!u.enabled === wantEnabled)
  }
  if (filterOnline.value !== 'all') {
    const wantOnline = filterOnline.value === 'online'
    list = list.filter(u => !!u.online === wantOnline)
  }
  return list
})

const totalPages = computed(() => Math.max(1, Math.ceil(filteredUsers.value.length / pageSize.value)))
const paginatedUsers = computed(() => {
  const list = filteredUsers.value
  const start = (currentPage.value - 1) * pageSize.value
  return list.slice(start, start + pageSize.value)
})

function applySearch() {
  currentPage.value = 1
}

async function downloadConfig(u) {
  // 不传 server/port，使用「系统配置」中保存的服务器地址与端口
  const url = `/api/users/${u.id}/download`
  const token = localStorage.getItem('token')
  const res = await fetch(url, { headers: { Authorization: `Bearer ${token}` } })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    ElMessage.error(err.error || '下载失败')
    return
  }
  const blob = await res.blob()
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = `${u.name}.ovpn`
  a.click()
  URL.revokeObjectURL(a.href)
}

async function load() {
  try {
    const { data } = await client.get('/users')
    users.value = data || []
  } catch (_) {
    users.value = []
  }
}

function openForm() {
  editTarget.value = null
  form.value = defaultForm()
  showForm.value = true
}

function parseCCDContent(content) {
  const r = { ifconfig_ip: '', ifconfig_mask: '255.255.255.0', iroutes: [], push_routes: [], redirect_gateway: false }
  if (!content || !content.trim()) return r
  const lines = content.split(/\r?\n/)
  for (const line of lines) {
    const raw = line.trim()
    const uncomment = raw.replace(/^\s*#\s*/, '').trim()
    if (/^ifconfig-push\s+\S+\s+\S+/.test(uncomment)) {
      const m = uncomment.match(/ifconfig-push\s+(\S+)\s+(\S+)/)
      if (m) { r.ifconfig_ip = m[1]; r.ifconfig_mask = m[2] }
    } else if (/^iroute\s+/.test(uncomment)) {
      r.iroutes.push(uncomment)
    } else if (/push\s+"route\s+/.test(uncomment) || /^push "route /.test(uncomment)) {
      const match = uncomment.match(/push\s+"route[^"]*"/) || uncomment.match(/push "route [^"]+"/)
      if (match) r.push_routes.push(match[0])
    } else if (/redirect-gateway/.test(uncomment)) {
      r.redirect_gateway = true
    }
  }
  return r
}

function buildCCDContent(form) {
  const lines = []
  lines.push('# 客户端 CCD 配置（client-config-dir）')
  lines.push('# 修改后需客户端重连 VPN 后生效')
  if (form.ccd_ifconfig_ip && form.ccd_ifconfig_ip.trim()) {
    lines.push('# ifconfig-push 格式（topology subnet）：ifconfig-push <客户端IP> 255.255.255.0')
    let mask = (form.ccd_ifconfig_mask && form.ccd_ifconfig_mask.trim()) || '255.255.255.0'
    if (!/255\.255\.255\.0|255\.255\.0\.0|255\.0\.0\.0/.test(mask)) mask = '255.255.255.0'
    lines.push(`ifconfig-push ${form.ccd_ifconfig_ip.trim()} ${mask}`)
  }
  const iroutes = (form.ccd_iroutes || '').split(/\r?\n/).map(s => s.trim()).filter(s => s && /^iroute\s+/.test(s))
  if (iroutes.length) {
    lines.push('')
    iroutes.forEach(l => lines.push(l))
  }
  const pushRoutes = (form.ccd_push_routes || '').split(/\r?\n/).map(s => s.trim()).filter(s => s && s.includes('push') && s.includes('route'))
  if (pushRoutes.length) {
    lines.push('')
    pushRoutes.forEach(l => lines.push(l))
  }
  if (form.ccd_redirect_gateway) {
    lines.push('')
    lines.push('push "redirect-gateway def1 bypass-dhcp"')
  }
  return lines.join('\n')
}

async function openEdit(u) {
  editTarget.value = u
  form.value = {
    ...defaultForm(),
    name: u.name,
    kind: u.kind === 'client' ? 'client' : 'user',
    password: '',
    email: u.email || '',
    remark: u.remark || '',
    expires_at: u.expires_at ? String(u.expires_at).slice(0, 10) : '',
    enabled: u.enabled,
    route_nopull: !!u.route_nopull,
  }
  showForm.value = true
  try {
    const { data } = await client.get(`/users/${u.id}/ccd`)
    const parsed = parseCCDContent(data.content || '')
    form.value.ccd_ifconfig_ip = parsed.ifconfig_ip
    form.value.ccd_ifconfig_mask = parsed.ifconfig_mask || ''
    form.value.ccd_iroutes = parsed.iroutes.join('\n')
    form.value.ccd_push_routes = parsed.push_routes.join('\n')
    form.value.ccd_redirect_gateway = parsed.redirect_gateway
  } catch (_) {}
}

function closeForm() {
  showForm.value = false
  editTarget.value = null
  form.value = defaultForm()
}

function copyToClipboardFallback(text) {
  const el = document.createElement('textarea')
  el.value = text
  el.style.position = 'fixed'
  el.style.left = '-9999px'
  el.style.top = '0'
  document.body.appendChild(el)
  el.select()
  try {
    return document.execCommand('copy')
  } finally {
    document.body.removeChild(el)
  }
}

async function copyCredentials(u) {
  const plain = lastPasswords.value[u.name]
  if (!plain) {
    ElMessage.warning('当前会话中未记录该用户的明文密码，请在编辑中重新设置密码后再复制。')
    return
  }
  const text = `VPN用户名: ${u.name}\n密码: ${plain}`
  let ok = false
  if (navigator.clipboard && window.isSecureContext) {
    try {
      await navigator.clipboard.writeText(text)
      ok = true
    } catch (_) {}
  }
  if (!ok) ok = copyToClipboardFallback(text)
  if (ok) {
    ElMessage.success('已复制到剪贴板')
  } else {
    ElMessage.error('复制失败，请手动复制（HTTP 下部分浏览器限制剪贴板）')
  }
}

async function createUser() {
  if (!form.value.password || form.value.password.length < 6) {
    ElMessage.error('请设置密码且不少于 6 位')
    return
  }
  if (!isValidVPNName(form.value.name)) {
    ElMessage.error('VPN 用户名仅限英文字母、数字和下划线')
    return
  }
  try {
    await client.post('/users', { name: form.value.name, kind: form.value.kind, email: form.value.email, remark: form.value.remark || '', password: form.value.password, route_nopull: form.value.route_nopull, expires_at: form.value.expires_at || '', enabled: form.value.enabled })
    const name = form.value.name
    const pwd = form.value.password
    lastPasswords.value[name] = pwd
    closeForm()
    load()
    const copyText = `VPN用户名: ${name}\n密码: ${pwd}`
    let copied = false
    if (navigator.clipboard && window.isSecureContext) {
      try {
        await navigator.clipboard.writeText(copyText)
        copied = true
      } catch (_) {}
    }
    if (!copied) copied = copyToClipboardFallback(copyText)
    ElMessage.success(copied ? '用户已创建，用户名与密码已复制到剪贴板' : '用户已创建，证书已生成。请妥善保存您设置的密码。')
  } catch (e) {
    ElMessage.error(e.response?.data?.error || '创建失败')
  }
}

async function updateUser() {
  if (editTarget.value && form.value.password && form.value.password.length < 6) {
    ElMessage.error('密码不少于 6 位')
    return
  }
  const id = editTarget.value?.id
  try {
    const ccd_content = buildCCDContent(form.value)
    await client.put(`/users/${id}`, {
      id,
      name: form.value.name,
      kind: form.value.kind,
      email: form.value.email,
      remark: form.value.remark ?? '',
      password: form.value.password || undefined,
      route_nopull: form.value.route_nopull,
      expires_at: form.value.expires_at || '',
      enabled: form.value.enabled,
      ccd_content,
    })
    if (form.value.password) {
      lastPasswords.value[form.value.name] = form.value.password
    }
    closeForm()
    load()
    ElMessage.success('已保存')
  } catch (e) {
    ElMessage.error(e.response?.data?.error || '保存失败')
  }
}

async function del(u) {
  if (!confirm(`确定删除用户「${u.name}」？`)) return
  try {
    await client.delete(`/users/${u.id}`)
    load()
    ElMessage.success('已删除')
  } catch (e) {
    ElMessage.error(e.response?.data?.error || '删除失败')
  }
}

const refreshInterval = 5000
let refreshTimer = null

onMounted(() => {
  load()
  refreshTimer = setInterval(load, refreshInterval)
})
onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer)
})
</script>

<style scoped>
.user-page {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  min-height: 0;
}
.page-header {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: flex-start;
  gap: 0.85rem;
  padding: 1.1rem 1.15rem 1rem;
  border-radius: 18px;
  border: 1px solid #dbe7f3;
  background: linear-gradient(180deg, #ffffff 0%, #f8fbff 100%);
  box-shadow: 0 16px 34px rgba(15, 23, 42, 0.05);
}
.page-title {
  margin: 0;
  width: 100%;
  font-size: 1.1rem;
  font-weight: 700;
  color: #0f172a;
}
.header-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.8rem;
  justify-content: flex-start;
  width: 100%;
}
.search-wrap {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  flex-wrap: wrap;
  flex: 1 1 760px;
  min-width: min(100%, 320px);
  padding: 0;
  background: transparent;
  border: none;
  border-radius: 0;
  box-shadow: none;
}
.search-label {
  font-size: 0.9rem;
  color: #64748b;
  font-weight: 700;
  margin-right: 0.1rem;
}
.search-input {
  width: 200px;
  flex: 1 1 200px;
  padding: 0 0.85rem;
  height: 40px;
  border-radius: 12px;
  border-color: #d8e3f0;
  background: #fff;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.92);
}
.ip-input {
  width: 220px;
}
/* Element Plus 下拉：与输入框同高、统一圆角 */
.filter-select {
  width: 150px;
  flex: 0 0 150px;
}
.filter-select :deep(.el-select__wrapper) {
  min-height: 40px;
  border-radius: 12px;
  border: 1px solid #d8e3f0;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.92);
  background: #fff;
}
.filter-select-wide {
  width: 160px;
}
.toolbar-meta {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.9rem;
  flex: 0 0 auto;
  margin-left: auto;
  padding-left: 0.25rem;
}
.add-btn {
  flex-shrink: 0;
  height: 40px;
  padding: 0 1.2rem;
  border-radius: 12px;
  box-shadow: 0 12px 26px rgba(64, 158, 255, 0.18);
}
.refresh-hint {
  font-size: 0.88rem;
  color: #64748b;
  white-space: nowrap;
}
.card.table-wrap {
  padding: 0;
}
.user-table-card {
  padding: 0;
  overflow: hidden;
  border-radius: 12px;
  box-shadow: 0 8px 24px rgba(15, 23, 42, 0.04);
  border: 1px solid #e8edf5;
  display: flex;
  flex-direction: column;
}
.table-wrap {
  overflow-x: auto;
  overflow-y: hidden;
  background: #fff;
  min-height: 0;
}
.table-wrap :deep(.el-table) {
  min-width: 100%;
  --el-table-border-color: #e8edf5;
  --el-table-header-bg-color: #f8fafc;
  --el-table-row-hover-bg-color: #f8fbff;
}
.table-wrap :deep(.el-table__inner-wrapper::before) {
  display: none;
}
.table-wrap :deep(.el-table__header-wrapper th) {
  background: #f8fafc !important;
}
.table-wrap :deep(.el-table__body tr td) {
  background: #fff;
}
.table-wrap :deep(.el-table td.el-table__cell),
.table-wrap :deep(.el-table th.el-table__cell) {
  padding: 0.68rem 0.7rem;
}
.table-wrap :deep(.el-table__body tr:hover > td) {
  background: #f8fbff !important;
}
.table-wrap :deep(.el-scrollbar__bar.is-horizontal) {
  bottom: 0;
  height: 10px;
}
.table-wrap :deep(.el-scrollbar__thumb) {
  background: rgba(148, 163, 184, 0.65);
}
.card.table-wrap table,
.user-table-card table {
  table-layout: auto;
  min-width: 100%;
}
.card.table-wrap th,
.user-table-card th {
  font-weight: 600;
  font-size: 0.875rem;
  color: var(--text);
  background: #fafafa;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--border);
  white-space: nowrap;
  text-align: center;
}
.card.table-wrap td,
.user-table-card td {
  text-align: center;
}
.card.table-wrap th.col-no,
.card.table-wrap td.col-no,
.user-table-card th.col-no,
.user-table-card td.col-no {
  min-width: 3.5em;
  white-space: nowrap;
}
.card.table-wrap th.col-type,
.card.table-wrap td.col-type,
.user-table-card th.col-type,
.user-table-card td.col-type {
  min-width: 4.5em;
  white-space: nowrap;
}
.card.table-wrap td.col-type .type-label,
.user-table-card td.col-type .type-label {
  display: inline-block;
  font-weight: 600;
  font-size: 0.875rem;
  padding: 0.25em 0.6em;
  border-radius: 4px;
  border: 1px solid;
}
.card.table-wrap td.col-type .type-user,
.user-table-card td.col-type .type-user {
  color: #0369a1;
  border-color: #0ea5e9;
  background: #e0f2fe;
}
.card.table-wrap td.col-type .type-client,
.user-table-card td.col-type .type-client {
  color: #047857;
  border-color: #10b981;
  background: #d1fae5;
}
.card.table-wrap th.col-remark,
.card.table-wrap td.col-remark,
.user-table-card th.col-remark,
.user-table-card td.col-remark {
  min-width: 8em;
  max-width: 14em;
  word-break: break-all;
  text-align: center;
}
.card.table-wrap th.col-ip,
.card.table-wrap td.col-ip,
.user-table-card th.col-ip,
.user-table-card td.col-ip {
  min-width: 9em;
  white-space: nowrap;
}
.tag-type {
  font-weight: 700;
  letter-spacing: 0.02em;
}
.tag-type:deep(.el-tag--info.el-tag--light) {
  color: #7c3aed;
  background: #f5f3ff;
  border-color: #c4b5fd;
}
.tag-type:deep(.el-tag--success.el-tag--light) {
  color: #15803d;
  background: #ecfdf3;
  border-color: #86efac;
}
.tag-ip {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
  font-weight: 600;
}
.ip-with-mode {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
}
.tag-ip-mode {
  font-size: 0.75rem;
}
.tag-enabled {
  font-weight: 700;
}

.online-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.2rem 0.55rem;
  border-radius: 999px;
  font-size: 0.82rem;
  font-weight: 600;
  border: 1px solid var(--border);
  background: #fff;
}
.online-chip .dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #9ca3af;
}
.online-chip.on {
  color: #16a34a;
  border-color: rgba(22, 163, 74, 0.35);
  background: rgba(22, 163, 74, 0.06);
}
.online-chip.on .dot {
  background: #16a34a;
}
.online-chip.off {
  color: #6b7280;
  border-color: rgba(107, 114, 128, 0.35);
  background: rgba(107, 114, 128, 0.06);
}
.online-chip.off .dot {
  background: #9ca3af;
}
.card.table-wrap th.col-time,
.card.table-wrap td.col-time,
.user-table-card th.col-time,
.user-table-card td.col-time {
  min-width: 10em;
  white-space: nowrap;
}
.card.table-wrap th.col-duration,
.card.table-wrap td.col-duration,
.user-table-card th.col-duration,
.user-table-card td.col-duration {
  min-width: 4.5em;
  white-space: nowrap;
}
.card.table-wrap th.col-traffic,
.card.table-wrap td.col-traffic,
.user-table-card th.col-traffic,
.user-table-card td.col-traffic {
  min-width: 5.5em;
  white-space: nowrap;
}
.card.table-wrap th.actions-col,
.card.table-wrap td.actions-cell,
.user-table-card th.actions-col,
.user-table-card td.actions-cell {
  min-width: 200px;
  box-sizing: border-box;
  text-align: center;
}
.card.table-wrap td,
.user-table-card td {
  padding: 0.62rem 0.85rem;
  font-size: 0.88rem;
  vertical-align: middle;
  border-bottom: 1px solid var(--border);
}
.card.table-wrap tbody tr:hover,
.user-table-card tbody tr:hover {
  background: #fafbfc;
}
.actions-cell {
  white-space: nowrap;
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 0.25rem;
}
.actions-cell .btn {
  white-space: nowrap;
}
.dropdown-caret {
  margin-left: 4px;
  font-size: 9px;
  opacity: 0.85;
  display: inline-block;
  vertical-align: middle;
}
button.small {
  padding: 0.4rem 0.75rem;
  font-size: 0.85rem;
  height: 32px;
}
.badge {
  padding: 0.25rem 0.6rem;
  border-radius: 12px;
  font-size: 0.8rem;
  font-weight: 500;
}
.user-empty-state {
  margin: 1rem;
  min-height: 120px;
}
.actions-cell :deep(.el-button) {
  min-width: 96px;
  height: 32px;
  padding: 0 0.75rem;
  border-radius: 11px;
  font-weight: 600;
  font-size: 0.86rem;
}
.actions-cell :deep(.el-button--primary.is-plain) {
  background: #eef6ff;
  border-color: #93c5fd;
  color: #409eff;
}
.actions-cell :deep(.el-button--primary.is-plain:hover) {
  background: #dbeafe;
  border-color: #60a5fa;
}
.pagination {
  display: flex;
  align-items: center;
  gap: 1rem;
  flex-wrap: wrap;
  justify-content: space-between;
  padding: 0.15rem 0.3rem 0;
  margin-top: -0.1rem;
}
.pagination-info {
  color: var(--muted);
  font-size: 0.9rem;
}
.page-size-select {
  width: 120px;
}
.page-size-select :deep(.el-select__wrapper) {
  min-height: 40px;
  border-radius: 12px;
}
.page-btns {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
.page-btns button {
  height: 40px;
  padding: 0 1rem;
  border-radius: 12px;
}
.page-num {
  font-size: 0.9rem;
  color: var(--muted);
  min-width: 56px;
  text-align: center;
}
.modal {
  position: fixed;
  inset: 0;
  padding: 24px;
  background:
    radial-gradient(circle at top, rgba(59, 130, 246, 0.14), transparent 36%),
    rgba(15, 23, 42, 0.48);
  backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
}
.modal-inner {
  min-width: 420px;
  max-width: 90vw;
  padding: 1rem 1.25rem;
}
.modal-inner-wide {
  min-width: 560px;
  max-width: 720px;
  padding: 1.4rem 1.6rem;
  border-radius: 18px;
  border: 1px solid rgba(226, 232, 240, 0.92);
  background: linear-gradient(180deg, #ffffff 0%, #f8fbff 100%);
  box-shadow: 0 28px 70px rgba(15, 23, 42, 0.22);
}
.user-form-modal .form-group input[type="date"] {
  width: 100%;
  min-width: 0;
}
.ccd-section-title {
  font-weight: 700;
  font-size: 0.95rem;
  margin-bottom: 0.45rem;
  color: #334155;
}
.ccd-warn-inline {
  margin: 0 0 0.75rem;
  padding: 0.65rem 0.8rem;
  background: linear-gradient(180deg, #fff9ee 0%, #fff4da 100%);
  border: 1px solid #f6d595;
  color: #a16207;
  font-size: 0.8rem;
  border-radius: 10px;
  line-height: 1.5;
}
.ccd-textarea-small {
  width: 100%;
  min-height: 100px;
  padding: 0.7rem 0.85rem;
  font-family: ui-monospace, monospace;
  font-size: 0.85rem;
  line-height: 1.55;
  border: 1px solid #dbe5f1;
  border-radius: 10px;
  background: #fff;
  resize: vertical;
  box-sizing: border-box;
}
.form-section-ccd .form-row-2 input {
  min-width: 0;
}
.modal-title {
  margin: 0 0 1rem;
  font-size: 1.18rem;
  font-weight: 700;
  color: #0f172a;
  letter-spacing: 0.01em;
}
.user-form-modal {
  max-height: 90vh;
  overflow-y: auto;
}
.user-form {
  display: flex;
  flex-direction: column;
  gap: 0;
}
.form-section {
  margin-bottom: 1.1rem;
  padding-bottom: 1.1rem;
  border-bottom: 1px solid #e8eef6;
}
.form-section-basic {
  border-bottom: none;
  padding-bottom: 0;
  margin-bottom: 0.5rem;
}
.form-section-title {
  margin: 0 0 0.85rem;
  font-size: 0.98rem;
  font-weight: 700;
  color: #334155;
}
/* 基本信息：浅底卡片 + 两列网格，行距统一 */
.form-basic-card {
  background: linear-gradient(180deg, #f8fbff 0%, #f4f8fc 100%);
  border: 1px solid #dbe7f3;
  border-radius: 14px;
  padding: 1.05rem 1.15rem;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.9);
}
.form-grid-2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem 1.3rem;
  align-items: start;
}
.form-grid-2 + .form-grid-2 {
  margin-top: 1.05rem;
  padding-top: 1.05rem;
  border-top: 1px solid #dbe7f3;
}
.form-basic-card .form-group {
  margin-bottom: 0;
}
.form-basic-card .form-group label {
  margin-bottom: 0.4rem;
}
.form-date-full {
  width: 100%;
}
.kind-radio {
  display: flex;
  flex-wrap: wrap;
}
.kind-radio :deep(.el-radio-button__inner) {
  padding-left: 1rem;
  padding-right: 1rem;
}
.req {
  color: var(--el-color-danger);
  font-weight: 600;
}
@media (max-width: 560px) {
  .form-grid-2 {
    grid-template-columns: 1fr;
  }
  .form-grid-2 + .form-grid-2 {
    border-top: none;
    padding-top: 0;
    margin-top: 1rem;
  }
}
.form-section:last-of-type {
  border-bottom: none;
  margin-bottom: 0;
  padding-bottom: 0;
}
.form-section-ccd {
  border-bottom: none;
}
.form-section .form-group {
  margin-bottom: 0.6rem;
}
.form-section .form-group:last-child {
  margin-bottom: 0;
}
.form-row-2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
}
.form-row-2 .form-group {
  margin-bottom: 0;
  min-width: 0;
}
.form-row-2 .form-group .form-hint {
  margin-top: 0.2rem;
}
.form-footer-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 0.75rem;
  margin-top: 1.2rem;
  padding-top: 1rem;
  border-top: 1px solid #e8eef6;
  border-bottom: none;
  margin-bottom: 0;
}
.form-footer-bar .form-actions {
  margin: 0;
  padding: 0;
  border: none;
  display: flex;
  gap: 0.5rem;
}
.form-group {
  margin-bottom: 0.6rem;
}
.form-group:last-of-type {
  margin-bottom: 0;
}
.form-group label {
  display: block;
  margin-bottom: 0.38rem;
  font-size: 0.885rem;
  font-weight: 600;
  color: #334155;
}
.form-group input[type="text"],
.form-group input[type="password"],
.form-group input[type="email"] {
  width: 100%;
  height: 38px;
  padding: 0 0.75rem;
  border: 1px solid #dbe5f1;
  border-radius: 10px;
  background: #fff;
  box-sizing: border-box;
  font-size: 0.9rem;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}
.form-group input[type="text"]:focus,
.form-group input[type="password"]:focus,
.form-group input[type="email"]:focus,
.ccd-textarea-small:focus {
  outline: none;
  border-color: #60a5fa;
  box-shadow: 0 0 0 3px rgba(96, 165, 250, 0.14);
}
.form-group-checkbox {
  margin-top: 0.2rem;
  margin-bottom: 0.5rem;
}
.checkbox-label,
.radio-label {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  cursor: pointer;
  font-weight: 500;
  color: var(--text);
  font-size: 0.9rem;
}
.checkbox-label input,
.radio-label input {
  width: 1rem;
  height: 1rem;
  margin: 0;
  cursor: pointer;
  accent-color: var(--accent, #1890ff);
}
.radio-group {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}
.radio-group .radio-label {
  margin-bottom: 0;
}
.form-hint {
  display: block;
  margin-top: 0.3rem;
  font-size: 0.79rem;
  color: var(--muted);
  line-height: 1.5;
}
.form-hint.error {
  color: var(--danger, #dc2626);
}
.checkbox-label-inline {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  font-weight: normal;
  cursor: pointer;
}
.checkbox-label-inline input {
  margin: 0;
}
.form-actions {
  display: flex;
  gap: 0.6rem;
  margin-top: 0;
  padding-top: 0;
  border-top: none;
}
.form-actions button {
  height: 38px;
  padding: 0 1rem;
  font-size: 0.92rem;
  border-radius: 10px;
  font-weight: 600;
}
@media (max-width: 960px) {
  .header-actions {
    align-items: stretch;
  }
  .search-wrap {
    width: 100%;
    flex: 1 1 100%;
  }
  .toolbar-meta {
    width: 100%;
    margin-left: 0;
    justify-content: space-between;
  }
  .pagination {
    justify-content: flex-start;
  }
}
@media (max-width: 820px) {
  .modal {
    padding: 16px;
    align-items: flex-start;
    overflow-y: auto;
  }
  .modal-inner,
  .modal-inner-wide {
    min-width: 0;
    width: min(100%, 720px);
    max-width: 100%;
    margin: 18px auto;
    padding: 1.1rem 1rem;
  }
  .form-row-2,
  .form-grid-2 {
    grid-template-columns: 1fr;
  }
  .form-footer-bar {
    align-items: stretch;
  }
  .form-footer-bar .form-actions,
  .form-actions {
    width: 100%;
    justify-content: flex-end;
    flex-wrap: wrap;
  }
  .page-header {
    gap: 0.85rem;
  }
  .search-wrap {
    gap: 0.55rem;
  }
  .search-label {
    width: 100%;
  }
  .search-input,
  .ip-input,
  .filter-select,
  .filter-select-wide {
    width: 100%;
    flex: 1 1 100%;
  }
  .search-wrap button.primary {
    width: 100%;
  }
  .toolbar-meta {
    flex-direction: column-reverse;
    align-items: stretch;
    padding-left: 0;
  }
  .add-btn {
    width: 100%;
    justify-content: center;
  }
  .refresh-hint {
    text-align: center;
  }
  .pagination {
    flex-direction: column;
    align-items: stretch;
    gap: 0.75rem;
  }
  .pagination-info {
    text-align: center;
  }
  .page-size-select {
    width: 100%;
  }
  .page-btns {
    justify-content: center;
    flex-wrap: wrap;
  }
}
</style>

<!-- 下拉挂在 body 上，需非 scoped 才能作用到 popper -->
<style>
.user-actions-dropdown {
  min-width: 170px;
  padding: 6px;
  border-radius: 12px;
  border: 1px solid #e6edf7;
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.12);
  background: rgba(255, 255, 255, 0.98);
  backdrop-filter: blur(10px);
}
.user-actions-dropdown .el-popper__arrow::before {
  border-color: #e6edf7 !important;
  background: rgba(255, 255, 255, 0.98) !important;
}
.user-actions-dropdown .el-dropdown-menu {
  padding: 0;
  border: none;
  box-shadow: none;
}
.user-actions-dropdown .el-dropdown-menu__item {
  min-height: 36px;
  border-radius: 9px;
  margin: 1px 0;
  padding: 0 12px;
  font-size: 0.9rem;
  font-weight: 600;
  color: #334155;
  transition: background-color 0.18s ease, color 0.18s ease, transform 0.18s ease;
}
.user-actions-dropdown .el-dropdown-menu__item:hover {
  background: #f3f8ff;
  color: #2563eb;
  transform: translateX(2px);
}
.user-actions-dropdown .el-dropdown-menu__item.is-disabled {
  opacity: 0.5;
}
.user-actions-dropdown .el-dropdown-menu__item--divided {
  margin-top: 6px;
  border-top-color: #eef2f7;
  padding-top: 8px;
}
.dropdown-danger-text {
  color: #ef4444;
  font-weight: 700;
}
</style>
