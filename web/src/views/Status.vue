<template>
  <div class="status-page">
    <h2 class="page-title">VPN 状态</h2>
    <div class="card status-card">
      <dl class="status-dl">
        <div class="status-item">
          <dt>PKI 已初始化</dt>
          <dd>{{ status.pki_initialized ? '是' : '否' }}</dd>
        </div>
        <div class="status-item">
          <dt>工作目录</dt>
          <dd>{{ status.base_path || '-' }}</dd>
        </div>
      </dl>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import client from '../api/client'

const status = ref({})

async function load() {
  const { data } = await client.get('/vpn/status')
  status.value = data
}

onMounted(load)
</script>

<style scoped>
.status-page {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.status-card {
  max-width: 480px;
}
.status-dl {
  margin: 0;
  display: grid;
  gap: 0.75rem;
}
.status-item dt {
  margin: 0;
  font-size: 0.85rem;
  color: var(--muted);
}
.status-item dd {
  margin: 0.25rem 0 0;
}
</style>
