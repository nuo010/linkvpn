<template>
  <div class="users-page">
    <div class="page-header">
      <h2 class="page-title">用户管理</h2>
      <button class="primary" @click="showForm = true">添加用户</button>
    </div>

    <div v-if="message" :class="['msg', message.type]">{{ message.text }}</div>

    <div class="card table-wrap">
      <table>
        <thead>
          <tr>
            <th>ID</th>
            <th>用户名</th>
            <th>邮箱</th>
            <th>启用</th>
            <th>创建时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="u in users" :key="u.id">
            <td>{{ u.id }}</td>
            <td>{{ u.name }}</td>
            <td>{{ u.email || '-' }}</td>
            <td>{{ u.enabled ? '是' : '否' }}</td>
            <td>{{ formatDate(u.created_at) }}</td>
            <td class="actions-cell">
              <button class="primary small" @click="downloadConfig(u)">下载 .ovpn</button>
              <button class="danger small" @click="del(u)">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
      <p v-if="users.length === 0" class="empty-tip">暂无用户，点击「添加用户」创建。</p>
    </div>

    <div v-if="showForm" class="modal" @click.self="showForm = false">
      <div class="card modal-inner">
        <h3 class="modal-title">添加用户</h3>
        <form @submit.prevent="createUser">
          <div class="form-group">
            <label>用户名 *</label>
            <input v-model="form.name" required />
          </div>
          <div class="form-group">
            <label>邮箱</label>
            <input v-model="form.email" type="email" />
          </div>
          <div class="form-group">
            <label>
              <input type="checkbox" v-model="form.enabled" /> 启用
            </label>
          </div>
          <div class="form-actions">
            <button type="submit" class="primary">创建</button>
            <button type="button" @click="showForm = false">取消</button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import client from '../api/client'

const users = ref([])
const showForm = ref(false)
const message = ref(null)
const form = ref({ name: '', email: '', enabled: true })

function formatDate(s) {
  if (!s) return '-'
  return new Date(s).toLocaleString('zh-CN')
}

async function downloadConfig(u) {
  // 不传 server/port，使用「系统配置」中保存的服务器地址与端口
  const url = `/api/users/${u.id}/download`
  const token = localStorage.getItem('token')
  const res = await fetch(url, { headers: { Authorization: `Bearer ${token}` } })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    message.value = { type: 'error', text: err.error || '下载失败' }
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
  const { data } = await client.get('/users')
  users.value = data
}

async function createUser() {
  message.value = null
  try {
    await client.post('/users', form.value)
    showForm.value = false
    form.value = { name: '', email: '', enabled: true }
    load()
    message.value = { type: 'success', text: '用户已创建，证书已生成。' }
  } catch (e) {
    message.value = { type: 'error', text: e.response?.data?.error || '创建失败' }
  }
}

async function del(u) {
  if (!confirm(`确定删除用户「${u.name}」？`)) return
  try {
    await client.delete(`/users/${u.id}`)
    load()
    message.value = { type: 'success', text: '已删除' }
  } catch (e) {
    message.value = { type: 'error', text: e.response?.data?.error || '删除失败' }
  }
}

onMounted(load)
</script>

<style scoped>
.users-page {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.actions-cell {
  white-space: nowrap;
}
.actions-cell button:first-child {
  margin-right: 0.5rem;
}
.modal-title {
  margin: 0 0 1rem;
  font-size: 1rem;
  font-weight: 600;
}
.form-actions {
  display: flex;
  gap: 0.5rem;
  margin-top: 1rem;
}
.modal {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
}
.modal-inner {
  min-width: 320px;
  max-width: 90vw;
  padding: 1.5rem 1.75rem;
}
</style>
