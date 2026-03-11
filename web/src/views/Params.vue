<template>
  <div class="params-page">
    <div class="card alert-warning">
      <span class="alert-icon">⚠️</span>
      注意：需点击「应用到 OpenVPN 配置」才会写入 server.conf；且重启 OpenVPN 服务后修改的配置才生效。
    </div>

    <div class="card params-card">
      <h3 class="section-title">配置</h3>
      <form @submit.prevent="save" class="params-form">
        <div class="params-grid">
          <div class="form-group">
            <label>端口</label>
            <el-input v-model="form.port" type="number" min="1" max="65535" size="small" />
          </div>
          <div class="form-group">
            <label>协议</label>
            <el-select v-model="form.protocol" size="small">
              <el-option label="udp" value="udp" />
              <el-option label="tcp" value="tcp" />
            </el-select>
          </div>
          <div class="form-group">
            <label>最大连接数</label>
            <el-input v-model="form.max_connections" type="number" min="1" size="small" />
          </div>
          <div class="form-group">
            <label>子网</label>
            <el-input v-model="form.subnet" placeholder="10.8.8.0/24" size="small" />
          </div>
          <div class="form-group">
            <label>管理接口地址</label>
            <el-input v-model="form.management" placeholder="127.0.0.1:7505" size="small" />
          </div>
          <div class="form-group checkbox-row">
            <el-checkbox v-model="form.vpn_gateway">VPN 网关（启用则推送默认路由）</el-checkbox>
          </div>
          <div class="form-group">
            <label>推送 DNS1</label>
            <el-input v-model="form.push_dns1" placeholder="8.8.8.8" size="small" />
          </div>
          <div class="form-group">
            <label>推送 DNS2</label>
            <el-input v-model="form.push_dns2" placeholder="2001:4860:4860::8888" size="small" />
          </div>
          <div class="form-group checkbox-row">
            <el-checkbox v-model="form.ipv6">IPv6</el-checkbox>
          </div>
          <div class="form-group">
            <label>IPv6 子网</label>
            <el-input v-model="form.ipv6_subnet" placeholder="fd00:8::/64" size="small" />
          </div>
          <div class="form-group checkbox-row">
            <el-checkbox v-model="form.auto_apply_to_config">
              保存时自动应用到 server.conf
            </el-checkbox>
          </div>
        </div>
        <div class="form-actions">
          <el-button type="primary" native-type="submit" size="small">保存参数</el-button>
          <el-button size="small" @click="apply">应用到 OpenVPN 配置</el-button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import client from '../api/client'
const form = ref({
  port: '1194',
  protocol: 'udp',
  max_connections: '200',
  vpn_gateway: false,
  subnet: '10.8.8.0/24',
  management: '127.0.0.1:7505',
  ipv6: false,
  ipv6_subnet: 'fd00:8::/64',
  push_dns1: '8.8.8.8',
  push_dns2: '2001:4860:4860::8888',
  auto_apply_to_config: false,
})

async function load() {
  try {
    const { data } = await client.get('/config/params')
    form.value = {
      port: String(data.port ?? 1194),
      protocol: data.protocol || 'udp',
      max_connections: String(data.max_connections ?? 200),
      vpn_gateway: !!data.vpn_gateway,
      subnet: data.subnet || '10.8.8.0/24',
      management: data.management || '127.0.0.1:7505',
      ipv6: !!data.ipv6,
      ipv6_subnet: data.ipv6_subnet || 'fd00:8::/64',
      push_dns1: data.push_dns1 || '',
      push_dns2: data.push_dns2 || '',
      auto_apply_to_config: !!data.auto_apply_to_config,
    }
  } catch (_) {
    ElMessage.error('加载失败')
  }
}

async function save() {
  const payload = {
    ...form.value,
    port: String(form.value.port),
    max_connections: String(form.value.max_connections),
  }
  try {
    await client.post('/config/params', payload)
    ElMessage.success('已保存')
  } catch (e) {
    ElMessage.error(e.response?.data?.error || '保存失败')
  }
}

async function apply() {
  try {
    const { data } = await client.post('/config/params/apply')
    ElMessage.success(data.message || '已应用到配置文件')
  } catch (e) {
    ElMessage.error(e.response?.data?.error || '应用失败')
  }
}

onMounted(load)
</script>

<style scoped>
.params-page {
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
.section-title {
  margin: 0 0 1rem;
  font-size: 1rem;
}
.params-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 1rem;
}
.checkbox-row {
  grid-column: 1 / -1;
}
.form-actions {
  display: flex;
  gap: 0.5rem;
  margin-top: 1.25rem;
  flex-wrap: wrap;
}
</style>
